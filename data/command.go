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
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// CommandEntry conveniently wraps a bundle and one command within that bundle.
type CommandEntry struct {
	Bundle  Bundle
	Command BundleCommand
}

type CommandParameters []string

func (c CommandParameters) String() string {
	return strings.Join(c, " ")
}

// CommandRequest represents a user command request as triggered in (probably)
// a chat provider.
type CommandRequest struct {
	CommandEntry

	// Adapter is the name of the adapter this request originated from
	Adapter string

	// ChannelID is the (provider-specific) ID of the channel that the
	// request originated in
	ChannelID string

	// Context is the original request context
	// TODO How can we refactor this out?
	Context context.Context

	// Parameters is the tokenized command parameters
	Parameters CommandParameters

	// RequestID is a unique (within this Gort instance) request identifier
	RequestID int64

	// Timestamp is the time this request was triggered
	Timestamp time.Time

	// UserID is the (provider-specific) ID of user making this request
	UserID string

	// UserEmail is the email address associated with the user making the
	// request (if known)
	UserEmail string

	// UserName is the Gort username of the user making the request
	UserName string
}

// String is a convenience method that outputs the normalized command
// string more or less as the user typed it.
func (r CommandRequest) String() string {
	return fmt.Sprintf("%s:%s %s", r.Bundle.Name, r.Command.Name, r.Parameters)
}

// CommandResponse wraps the response text emitted by an executed command.
type CommandResponse struct {
	// Lines contains the command output (from both stdout and stderr) as
	// a string slice, delimitted along newlines.
	Lines []string

	// Out The command output as a single block of text, with lines joined
	// with newlines.
	Out string

	// Structured is true if the command output is valid JSON. If so, then it
	// also be unmarshalled as Payload; else Payload will be a string (equal
	// to Out).
	Structured bool

	// Title includes a title. Usually only set by the relay for certain
	// internally-detected errors. It can be used to build a user output
	// message, and generally contains a short description of the result.
	Title string
}

// CommandResponseData contains about a command execution, including its
// duration and exit code. If the relay set an an explicit error, it will
// be here as well.
type CommandResponseData struct {
	// Duration is how long the command required to execute.
	// TODO(mtitmus) What are the start and endpoints? Do we want to track
	// multiple durations for "framework time" and "command time" and whatever
	// else?
	Duration time.Duration

	// ExitCode is the exit code reported by the command.
	ExitCode int16

	// Error is set by the relay under certain internal error conditions.
	Error error
}

// CommandResponseEnvelope encapsulates the data and metadata around a command
// execution and response. It's returned by a relay when a command has been
// executed. It is passed directly into the templating engine where it can be
// accessed by the Go templates that describe the response formats.
type CommandResponseEnvelope struct {
	// Request is the original request used to execute the command. It contains
	// the original CommandEntry value as well as the user and adapter data.
	Request CommandRequest

	// Response contains the
	Response CommandResponse

	// Data contains about the command execution, including its duration and exit code.
	// If the relay set an an explicit error, it will be here as well.
	Data CommandResponseData

	// Payload includes the command output. If the output is structured JSON,
	// it will be unmarshalled and placed here where it can be accessible to
	// Go templates. If it's not, this will be a string equal to Out.
	Payload interface{}
}

// NewCommandResponseEnvelope can be used to generate a new
// CommandResponseEnvelope value with the provided options.
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

// CommandResponseEnvelopeOption is returned by the various WithX functions
// and accepted by NewCommandResponseEnvelope.
type CommandResponseEnvelopeOption func(e *CommandResponseEnvelope)

// WithExitCode sets Data.ExitCode. It does NOT set
func WithExitCode(code int16) CommandResponseEnvelopeOption {
	return func(e *CommandResponseEnvelope) {
		e.Data.ExitCode = code
	}
}

// WithError sets Data.Error, Data.ExitCode, Response.Lines, Response.Out,
// Response.Structured, Response.Title, and Payload (as err.Error).
func WithError(title string, err error, code int16) CommandResponseEnvelopeOption {
	return func(e *CommandResponseEnvelope) {
		e.Data.Error = err
		e.Data.ExitCode = code
		e.Response.Lines = []string{err.Error()}
		e.Response.Out = err.Error()
		e.Response.Title = title
		e.Payload, e.Response.Structured = unmarshalResponsePayload(e.Response.Out)
	}
}

// WithResponseLines sets Response.Lines, Response.Out, Response.Structured,
// and Payload.
func WithResponseLines(r []string) CommandResponseEnvelopeOption {
	return func(e *CommandResponseEnvelope) {
		e.Response.Lines = r
		e.Response.Out = strings.Join(r, "\n")
		e.Payload, e.Response.Structured = unmarshalResponsePayload(e.Response.Out)
	}
}

// unmarshalResponsePayload examines the string parameter to determine
// whether it contains valid JSON. If it does it will unmarshal the contents
// and return the result and true; else it will return the original string
// and false.
func unmarshalResponsePayload(s string) (interface{}, bool) {
	b := []byte(s)

	if !json.Valid(b) {
		return s, false
	}

	var i interface{}

	if err := json.Unmarshal(b, &i); err != nil {
		return s, false
	}

	return i, true
}
