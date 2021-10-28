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

package data

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// CommandEntry conveniently wraps a bundle and one command within that bundle.
type CommandEntry struct {
	Bundle  Bundle
	Command BundleCommand
}

// CommandRequest represents a user command request as triggered in (probably)
// a chat provider.
type CommandRequest struct {
	CommandEntry
	Adapter    string          // The name of the adapter this request originated from.
	ChannelID  string          // The provider ID of the channel that the request originated in.
	Context    context.Context // The request context
	Parameters []string        // Tokenized command parameters
	RequestID  int64           // A unique requestID
	Timestamp  time.Time       // The time this request was triggered
	UserID     string          // The provider ID of user making this request.
	UserEmail  string          // The email address associated with the user making the request
	UserName   string          // The gort username of the user making the request
}

// CommandString is a convenience method that outputs the normalized command
// string, more or less as the user typed it.
func (r CommandRequest) CommandString() string {
	return fmt.Sprintf(
		"%s:%s %s",
		r.Bundle.Name,
		r.Command.Name,
		strings.Join(r.Parameters, " "))
}

type CommandResponse struct {
	// IsStructured is true if the command output is structured as JSON? If
	// so, then it will be unmarshalled as Payload; else Payload will be a
	// string (equal to Out).
	IsStructured bool

	// Title includes a title. Usually only set by the relay for certain
	// internally-detected errors. It can be used to build a user output
	// message, and generally contains a short description of the result.
	Title string

	// Lines contains the command output (from both stdout and stderr) as
	// a string slice, delimitted along newlines.
	Lines []string

	// Out The command output as a single block of text, with lines joined
	// with newlines.
	Out string

	// Payload includes the command output. If the output is structured JSON,
	// it will be unmarshalled and placed here where it can be accessible to
	// Go templates. If it's not, this will be a string equal to Out.
	Payload interface{}
}

// CommandResponseData contains about the command execution, including its
// duration and exit code. If the relay set an an explicit error, it will
// be here as well.
type CommandResponseData struct {
	// Duration is how long the command required to execute.
	// TODO(mtitmus) What are the start and endpoints? Do we want to track
	// multiple durations for "famework time" and "command time" and whatever
	// else?
	Duration time.Duration

	// ExitCode is the exit code reported by the command.
	ExitCode int16

	// IsError is a convenience flag that's set to true if ExitCode isn't 0.
	IsError bool

	// Error can be set by the relay in certain internal error conditions.
	// TODO(mtitmus) Do we even need this? Will it be confusing?
	Error error
}

// CommandResponseEnvelope encapsulates the data and metadata around a command
// execution and response. It's returned by a relay when a command has been
// executed. It is passed directly into the response formatter where it can be
// accessed by the Go templates that describe the response formats.
// https://play.golang.org/p/tYe4zc0E1cB
type CommandResponseEnvelope struct {
	// Request is the original request used to execute the command. It contains
	// the original CommandEntry value as well as the user and adapter data.
	Request CommandRequest

	// Response contains the
	Response CommandResponse

	// Data contains about the command execution, including its duration and exit code.
	// If the relay set an an explicit error, it will be here as well.
	Data CommandResponseData
}

func NewCommandResponseEnvelope(request CommandRequest, opts ...CommandResponseEnvelopeOption) CommandResponseEnvelope {
	envelope := CommandResponseEnvelope{
		Request:  request,
		Response: CommandResponse{Lines: []string{}},
	}

	for _, o := range opts {
		o(&envelope)
	}

	return envelope
}

type CommandResponseEnvelopeOption func(e *CommandResponseEnvelope)

// WithExitCode sets Data.ExitCode and Data.IsError
func WithExitCode(code int16) CommandResponseEnvelopeOption {
	return func(e *CommandResponseEnvelope) {
		e.Data.ExitCode = code
		e.Data.IsError = code != 0
	}
}

// WithError sets e.Data.Error, Data.ExitCode, Data.IsError, Response.Lines,
// Response.Out, and Response.Title.
func WithError(title string, err error, code int16) CommandResponseEnvelopeOption {
	return func(e *CommandResponseEnvelope) {
		e.Data.Error = err
		e.Data.ExitCode = code
		e.Data.IsError = code != 0
		e.Response.Lines = []string{err.Error()}
		e.Response.Out = err.Error()
		e.Response.Title = title
	}
}

// WithResponseLines sets Response.Lines and Response.Out.
func WithResponseLines(r []string) CommandResponseEnvelopeOption {
	return func(e *CommandResponseEnvelope) {
		e.Response.Lines = r
		e.Response.Out = strings.Join(r, "\n")
	}
}
