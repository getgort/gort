/*
 * Copyright 2021 The Gort Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/getgort/gort/data/io"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"

	"github.com/getgort/gort/config"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/dataaccess"
	"github.com/getgort/gort/dataaccess/errs"
	gerrs "github.com/getgort/gort/errors"
	"github.com/getgort/gort/telemetry"
	"github.com/getgort/gort/worker"
)

// AuthorizeUser is not yet implemented, and will not be until remote relays
// become a thing. For now, all authentication is done by the command execution
// framework (which, in turn, invokes a/the relay).
func AuthorizeUser(commandRequest data.CommandRequest, user rest.User) (bool, error) {
	// TODO
	return true, nil
}

// StartListening instructs the relay to begin listening for incoming command requests.
func StartListening() (chan<- data.CommandRequest, <-chan data.CommandResponseEnvelope) {
	commandRequests := make(chan data.CommandRequest)
	commandResponses := make(chan data.CommandResponseEnvelope)

	go func() {
		for commandRequest := range commandRequests {
			go func(request data.CommandRequest) {
				commandResponses <- handleRequest(request.Context, request)
			}(commandRequest)
		}
	}()

	return commandRequests, commandResponses
}

// SpawnWorker receives a CommandEntry and a slice of command parameters
// strings, and constructs a new worker.Worker.
func SpawnWorker(ctx context.Context, command data.CommandRequest) (worker.Worker, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "relay.SpawnWorker")
	defer sp.End()

	// Generate a token good for 10 seconds.
	dal, err := dataaccess.Get()
	if err != nil {
		return nil, err
	}

	token, err := dal.TokenGenerate(ctx, command.UserName, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return worker.New(command, token)
}

// getUser is just a convenience function for interacting with the DAL.
func getUser(ctx context.Context, username string) (rest.User, error) {
	da, err := dataaccess.Get()
	if err != nil {
		return rest.User{}, err
	}

	if exists, err := da.UserExists(ctx, username); err != nil {
		return rest.User{}, err
	} else if !exists {
		return rest.User{}, errs.ErrNoSuchUser
	}

	return da.UserGet(ctx, username)
}

// handleRequest does the work of spawning, starting, stopping, and cleaning
// up after worker processes. It receives incoming command requests from the
// StartListening() CommandRequest channel, returning a CommandResponse which
// in turn gets forwarded to that function's CommandRequest channel.
func handleRequest(ctx context.Context, request data.CommandRequest) data.CommandResponseEnvelope {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "relay.handleRequest")
	defer sp.End()

	var envelope data.CommandResponseEnvelope

	da, err := dataaccess.Get()
	if err != nil {
		envelope = data.NewCommandResponseEnvelope(
			request,
			data.WithError("Failed to access data access layer", err, ExitIoErr),
		)
		envelope.Data.Duration = time.Since(envelope.Request.Timestamp)
		return envelope
	}

	defer func() {
		envelope.Data.Duration = time.Since(envelope.Request.Timestamp)
		da.RequestClose(ctx, envelope)
	}()

	user, err := getUser(ctx, request.UserName)

	switch {
	case err == nil:
		break
	case gerrs.Is(err, errs.ErrNoSuchUser):
		envelope = data.NewCommandResponseEnvelope(
			request,
			data.WithError("No such Gort user: "+request.UserName, err, ExitNoUser),
		)
		return envelope
	case gerrs.Is(err, errs.ErrDataAccess):
		envelope = data.NewCommandResponseEnvelope(
			request,
			data.WithError("Data access failure", err, ExitIoErr),
		)
		return envelope
	default:
		envelope = data.NewCommandResponseEnvelope(
			request,
			data.WithError("Failed to get Gort user: "+request.UserName, err, ExitGeneral),
		)
		return envelope
	}

	if authorized, err := AuthorizeUser(request, user); err != nil {
		envelope = data.NewCommandResponseEnvelope(
			request,
			data.WithError("Authorization system failure", err, ExitSystemErr),
		)
		return envelope
	} else if !authorized {
		envelope = data.NewCommandResponseEnvelope(
			request,
			data.WithError("Permission denied", err, ExitNoPerm),
		)
		return envelope
	}

	worker, err := SpawnWorker(ctx, request)
	if err != nil {
		envelope = data.NewCommandResponseEnvelope(
			request,
			data.WithError("Failed to spawn worker", err, ExitSystemErr),
		)
		return envelope
	}

	dc, err := loadDynamicConfigurations(ctx, request)

	if err != nil {
		envelope = data.NewCommandResponseEnvelope(
			request,
			data.WithError("Failed to load dynamic configurations", err, ExitSystemErr),
		)
		return envelope
	}

	worker.Initialize(dc)

	envelope = runWorker(ctx, worker, request)

	return envelope
}

func loadDynamicConfigurations(ctx context.Context, command data.CommandRequest) ([]data.DynamicConfiguration, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "relay.loadDynamicConfigurations")
	defer sp.End()

	da, err := dataaccess.Get()
	if err != nil {
		return nil, err
	}

	groups, err := da.UserGroupList(ctx, command.UserName)
	if err != nil {
		return nil, err
	}

	var results = make(chan []data.DynamicConfiguration, 3+len(groups))
	var errs = make(chan error, 3+len(groups))

	wg := sync.WaitGroup{}
	wg.Add(3 + len(groups))

	go func() {
		defer wg.Done()
		dc, err := da.DynamicConfigurationList(ctx, data.LayerBundle, command.Bundle.Name, "", "")
		if err != nil {
			errs <- err
			cancel()
			return
		}
		results <- dc
	}()

	go func() {
		defer wg.Done()
		dc, err := da.DynamicConfigurationList(ctx, data.LayerRoom, command.Bundle.Name, command.ChannelID, "")
		if err != nil {
			errs <- err
			cancel()
			return
		}
		results <- dc
	}()

	go func() {
		defer wg.Done()
		dc, err := da.DynamicConfigurationList(ctx, data.LayerUser, command.Bundle.Name, command.UserName, "")
		if err != nil {
			errs <- err
			cancel()
			return
		}
		results <- dc
	}()

	for _, g := range groups {
		go func(g rest.Group) {
			defer wg.Done()
			dc, err := da.DynamicConfigurationList(ctx, data.LayerGroup, g.Name, "", "")
			if err != nil {
				errs <- err
				cancel()
				return
			}
			results <- dc
		}(g)
	}

	wg.Wait()

	close(results)
	close(errs)

	if err = <-errs; err != nil {
		return nil, err
	}

	var configs []data.DynamicConfiguration

	for dc := range results {
		configs = append(configs, dc...)
	}

	return configs, nil
}

// runWorker is called by handleRequest to do the work of starting an
// individual worker, capturing its output, and cleaning up after it.
func runWorker(ctx context.Context, worker worker.Worker, request data.CommandRequest) data.CommandResponseEnvelope {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "relay.runWorker")
	defer sp.End()

	var envelope data.CommandResponseEnvelope
	defer func() {
		envelope.Data.Duration = time.Since(envelope.Request.Timestamp)
	}()

	// Get configured timeout. Zero (or less) is no timeout.
	var cancel context.CancelFunc
	if timeout := config.GetGlobalConfigs().CommandTimeout; timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	stdoutChan, err := worker.Start(ctx)
	if err != nil {
		envelope = data.NewCommandResponseEnvelope(
			request,
			data.WithError("Failed to start worker", err, ExitSystemErr),
		)
		return envelope
	}

	// Read input from the worker until the stream closes
	var lines []string
	var advanced []io.AdvancedOutput
	for line := range stdoutChan {
		if strings.HasPrefix(line, "#!#") {
			var a io.AdvancedOutput
			err = json.NewDecoder(strings.NewReader(strings.TrimPrefix(line, "#!#"))).Decode(&a)
			if err != nil {
				log.
					WithField("request.id", request.RequestID).
					WithField("output.raw", line).
					WithError(err).
					Warnf("Badly formatted advanced output")
				continue
			}
			log.
				WithField("request.id", request.RequestID).
				WithField("output.action", a.Action).
				Info("Found output action")
			advanced = append(advanced, a)
		} else {
			lines = append(lines, line)
		}
	}

	var exitCode int64

	select {
	case exitCode = <-worker.Stopped():
		var opts []data.CommandResponseEnvelopeOption

		if exitCode != ExitOK {
			if len(lines) == 0 {
				lines = []string{"Unknown error executing command"}
			}

			opts = append(opts, data.WithError(
				"Command Error",
				fmt.Errorf(strings.Join(lines, " ")),
				int16(exitCode),
			))
		}

		opts = append(opts, data.WithResponseLines(lines), data.WithAdvancedOutput(advanced))
		envelope = data.NewCommandResponseEnvelope(request, opts...)

		log.
			WithField("request.id", request.RequestID).
			WithField("status", exitCode).
			Info("Command exited")

	case <-ctx.Done():
		err := ctx.Err()

		envelope = data.NewCommandResponseEnvelope(
			request,
			data.WithError(err.Error(), err, ExitTimeout),
		)

		log.
			WithError(err).
			WithField("request.id", request.RequestID).
			WithField("status", ExitTimeout).
			Info("Command exited with error")
	}

	forceTerm := time.Second * 10
	worker.Stop(ctx, &forceTerm)

	return envelope
}
