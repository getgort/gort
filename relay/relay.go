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

	"github.com/getgort/gort/data"
	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/dataaccess"
	"github.com/getgort/gort/dataaccess/errs"
	gerrs "github.com/getgort/gort/errors"
	"github.com/getgort/gort/telemetry"
	"github.com/getgort/gort/worker"
	"go.opentelemetry.io/otel"
)

const (
	// Exit codes borrowed from sysexits.h

	ExitOK              = 0   // successful termination
	ExitGeneral         = 1   // catchall for errors
	ExitNoUser          = 67  // user unknown
	ExitNoRelay         = 68  // relay name unknown
	ExitUnavailable     = 69  // relay unavailable
	ExitInternalError   = 70  // internal software error
	ExitSystemErr       = 71  // system error (e.g., can't spawn worker)
	ExitIoErr           = 74  // input/output error
	ExitTempFail        = 75  // temp failure; user can retry
	ExitProtocol        = 76  // remote error in protocol
	ExitNoPerm          = 77  // permission denied
	ExitCannotInvoke    = 126 //  Command invoked cannot execute
	ExitCommandNotFound = 127 // "command not found"
)

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

	response.Duration = time.Since(response.Command.Timestamp)
	response = start(ctx, worker, response)

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

func start(ctx context.Context, worker *worker.Worker, response data.CommandResponse) data.CommandResponse {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "relay.start")
	defer sp.End()

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

	// Wait for the exit status, if necessary
	response.Status = <-worker.Stopped()

	if response.Status != ExitOK {
		response.Title = "Command Error"

		if len(response.Output) == 0 {
			response.Output = []string{"Unknown command error"}
		}
	}

	return response
}

func AuthorizeUser(commandRequest data.CommandRequest, user rest.User) (bool, error) {
	// TODO
	return true, nil
}

// StartListening instructs the relays to begin listening for incoming command requests.
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
