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
	"time"

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

const (
	// Most exit codes borrowed from sysexits.h

	ExitOK              = 0   // successful termination
	ExitGeneral         = 1   // catchall for errors
	ExitNoUser          = 67  // user unknown
	ExitNoRelay         = 68  // relay name unknown
	ExitUnavailable     = 69  // relay unavailable
	ExitInternalError   = 70  // internal software error
	ExitSystemErr       = 71  // system error (e.g., can't spawn worker)
	ExitTimeout         = 72  // timeout exceeded
	ExitIoErr           = 74  // input/output error
	ExitTempFail        = 75  // temp failure; user can retry
	ExitProtocol        = 76  // remote error in protocol
	ExitNoPerm          = 77  // permission denied
	ExitCannotInvoke    = 126 //  Command invoked cannot execute
	ExitCommandNotFound = 127 // "command not found"
)

// AuthorizeUser is not yet implemented, and will not be until remote relays
// become a thing. For now, all authentication is done by the command execution
// framework (which, in turn, invokes a/the relay).
func AuthorizeUser(commandRequest data.CommandRequest, user rest.User) (bool, error) {
	// TODO
	return true, nil
}

// StartListening instructs the relay to begin listening for incoming command requests.
func StartListening() (chan<- data.CommandRequest, <-chan data.CommandResponse) {
	commandRequests := make(chan data.CommandRequest)
	commandResponses := make(chan data.CommandResponse)

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
func SpawnWorker(ctx context.Context, command data.CommandRequest) (*worker.Worker, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "relay.SpawnWorker")
	defer sp.End()

	image := command.Bundle.Docker.Image
	tag := command.Bundle.Docker.Tag
	entrypoint := command.Command.Executable

	return worker.NewWorker(image, tag, entrypoint, command.Parameters...)
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
func handleRequest(ctx context.Context, commandRequest data.CommandRequest) data.CommandResponse {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "relay.handleRequest")
	defer sp.End()

	response := data.CommandResponse{
		Command: commandRequest,
		Output:  []string{},
	}

	user, err := getUser(ctx, commandRequest.UserName)
	switch {
	case err == nil:
		break
	case gerrs.Is(err, errs.ErrNoSuchUser):
		response.Status = ExitNoUser
		response.Error = err
		response.Title = "No such Gort user: " + commandRequest.UserName
		response.Output = []string{err.Error()}
		return response
	case gerrs.Is(err, errs.ErrDataAccess):
		response.Status = ExitIoErr
		response.Error = err
		response.Title = "Data access failure"
		response.Output = []string{err.Error()}
		return response
	default:
		response.Status = ExitGeneral
		response.Error = err
		response.Title = "Failed to get Gort user: " + commandRequest.UserName
		response.Output = []string{err.Error()}
		return response
	}

	if authorized, err := AuthorizeUser(commandRequest, user); err != nil {
		response.Status = ExitGeneral
		response.Error = err
		response.Title = "Authorization system failure"
		response.Output = []string{err.Error()}
		return response
	} else if !authorized {
		response.Status = ExitNoPerm
		response.Error = err
		response.Title = "Permission denied"
		response.Output = []string{err.Error()}
		return response
	}

	worker, err := SpawnWorker(ctx, commandRequest)
	if err != nil {
		response.Status = ExitSystemErr
		response.Error = err
		response.Title = "Failed to spawn worker"
		response.Output = []string{err.Error()}
		return response
	}

	response = runWorker(ctx, worker, response)
	response.Duration = time.Since(response.Command.Timestamp)

	da, err := dataaccess.Get()
	if err != nil {
		response.Status = ExitIoErr
		response.Error = err
		response.Title = "Failed to access data access layer"
		response.Output = []string{err.Error()}
		return response
	}

	da.RequestClose(ctx, response)

	return response
}

// runWorker is called by handleRequest to do the work of starting an
// individual worker, capturing its output, and cleaning up after it.
func runWorker(ctx context.Context, worker *worker.Worker, response data.CommandResponse) data.CommandResponse {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "relay.runWorker")
	defer sp.End()

	// Get configured timeout. Zero (or less) is no timeout.
	timeout := config.GetGlobalConfigs().CommandTimeout()
	if timeout <= 0 {
		timeout = time.Hour * 24 * 365 // No timeout? Just use one year.
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	stdoutChan, err := worker.Start(ctx)
	if err != nil {
		response.Status = ExitSystemErr
		response.Error = err
		response.Title = "Failed to start worker"
		response.Output = []string{err.Error()}
		return response
	}

	// Read input from the worker until the stream closes
	for line := range stdoutChan {
		response.Output = append(response.Output, line)
	}

	select {
	case response.Status = <-worker.Stopped():
		if response.Status != ExitOK {
			response.Title = "Command Error"

			if len(response.Output) == 0 {
				response.Output = []string{"Unknown command error"}
			}
		}

		log.
			WithField("request.id", response.Command.RequestID).
			WithField("status", response.Status).
			Info("Command exited")

	case <-ctx.Done():
		err := ctx.Err()
		response.Status = ExitTimeout
		response.Error = err
		response.Title = err.Error()

		log.
			WithError(err).
			WithField("request.id", response.Command.RequestID).
			WithField("status", response.Status).
			Info("Command exited with error")
	}

	forceTerm := time.Second * 10
	worker.Stop(ctx, &forceTerm)

	return response
}
