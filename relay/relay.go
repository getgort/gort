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
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/worker"
)

func handleRequest(commandRequest data.CommandRequest) data.CommandResponse {
	response := data.CommandResponse{
		Command: commandRequest,
		Output:  []string{},
	}

	worker, err := SpawnWorker(commandRequest)
	if err != nil {
		response.Status = 126
		response.Error = err
		response.Title = "Failed to spawn worker"
		response.Output = []string{err.Error()}
		return response
	}

	stdoutChan, err := worker.Start()
	if err != nil {
		response.Status = 126
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

	if response.Status != 0 {
		response.Title = "Command Error"

		if len(response.Output) == 0 {
			response.Output = []string{"Unknown command error"}
		}
	}

	return response
}

// StartListening instructs the relays to begin listening for incoming command requests.
func StartListening() (chan<- data.CommandRequest, <-chan data.CommandResponse) {
	commandRequests := make(chan data.CommandRequest)
	commandResponses := make(chan data.CommandResponse)

	go func() {
		for commandRequest := range commandRequests {
			go func(request data.CommandRequest) {
				commandResponses <- handleRequest(request)
			}(commandRequest)
		}
	}()

	return commandRequests, commandResponses
}

// SpawnWorker receives a CommandEntry and a slice of command parameters
// strings, and constructs a new worker.Worker.
func SpawnWorker(command data.CommandRequest) (*worker.Worker, error) {
	image := command.Bundle.Docker.Image
	tag := command.Bundle.Docker.Tag
	entrypoint := command.Command.Executable

	return worker.NewWorker(image, tag, entrypoint, command.Parameters...)
}
