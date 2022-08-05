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

package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/getgort/emoji/v2/emoji"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/getgort/gort/auth"
	"github.com/getgort/gort/command"
	"github.com/getgort/gort/config"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/data/io"
	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/dataaccess"
	"github.com/getgort/gort/dataaccess/errs"
	gerrs "github.com/getgort/gort/errors"
	"github.com/getgort/gort/retrieval"
	"github.com/getgort/gort/telemetry"
	"github.com/getgort/gort/templates"
	"github.com/getgort/gort/version"
)

var (
	// All existant adapters keyed by name
	adapterLookup = map[string]Adapter{}
)

const unexpectedError = "An unexpected error has occurred. Please check the logs for more information."

var (
	// ErrAdapterNameCollision is emitted by AddAdapter() if two adapters
	// have the same name.
	ErrAdapterNameCollision = errors.New("adapter name collision")

	// ErrAuthenticationFailure is emitted when an AuthenticationErrorEvent
	// is received.
	ErrAuthenticationFailure = errors.New("authentication failure")

	// ErrChannelNotFound is returned when OnChannelMessage can't find
	// information is the originating channel.
	ErrChannelNotFound = errors.New("channel not found")

	// ErrGortNotBootstrapped is returned by findOrMakeGortUser() if a user
	// attempts to trigger a command but Gort hasn't yet been bootstrapped.
	ErrGortNotBootstrapped = errors.New("gort hasn't been bootstrapped yet")

	// ErrSelfRegistrationOff is returned by findOrMakeGortUser() if an unknown
	// user attempts to trigger a command but self-registration is configured
	// to false.
	ErrSelfRegistrationOff = errors.New("user doesn't exist and self-registration is off")

	// ErrNoSuchAdapter is returned by GetAdapter if a requested adapter name
	// can't be found.
	ErrNoSuchAdapter = errors.New("no such adapter")

	// ErrUserNotFound is throws by several methods if a provider fails to
	// return requested user information.
	ErrUserNotFound = errors.New("user not found")
)

// Adapter represents a connection to a chat provider.
type Adapter interface {
	// GetChannelInfo provides info on a specific provider channel accessible
	// to the adapter.
	GetChannelInfo(channelID string) (*io.ChannelInfo, error)

	// GetName provides the name of this adapter as per the configuration.
	GetName() string

	// GetPresentChannels returns a slice of channels that the adapter is present in.
	GetPresentChannels() ([]*io.ChannelInfo, error)

	// GetUserInfo provides info on a specific provider user accessible
	// to the adapter.
	GetUserInfo(userID string) (*io.UserInfo, error)

	// Listen causes the Adapter to initiate a connection to its provider and
	// begin relaying back events (including errors) via the returned channel.
	Listen(ctx context.Context) <-chan *ProviderEvent

	// React adds the given emoji to the given message as a reaction.
	React(ctx context.Context, message MessageRef, emoji emoji.Emoji) error

	// Reply creates a message denoted as a reply in the adapter.
	// The exact way these replies vary per adapter implementation.
	Reply(ctx context.Context, message MessageRef, content string) error

	// Send sends the contents of a response envelope to a
	// specified channel. If channelID is empty the value of
	// envelope.Request.ChannelID will be used.
	Send(ctx context.Context, channelID string, elements templates.OutputElements) error

	// SendText sends a simple text message to the specified channel.
	SendText(ctx context.Context, channelID string, message string) error

	// SendError is a break-glass error message function that's used when the
	// templating function fails somehow. Obviously, it does not utilize the
	// templating engine.
	SendError(ctx context.Context, channelID string, title string, err error) error
}

type RequestorIdentity struct {
	Adapter     Adapter
	ChatUser    *io.UserInfo
	ChatChannel *io.ChannelInfo
	Provider    *io.ProviderInfo
	GortUser    *rest.User
}

// AddAdapter adds an adapter.
func AddAdapter(a Adapter) {
	name := a.GetName()

	// No name? Generate a temporary name from the type
	if name == "" {
		name = fmt.Sprintf("%T", a)
	}

	log.WithField("adapter", name).Debug("Adapter added")
	adapterLookup[name] = a
}

// GetAdapter returns the requested adapter instance, if one exists.
// If not, an error is returned.
func GetAdapter(name string) (Adapter, error) {
	if adapter, ok := adapterLookup[name]; ok {
		return adapter, nil
	}

	return nil, ErrNoSuchAdapter
}

// OnConnected handles ConnectedEvent events.
func OnConnected(ctx context.Context, event *ProviderEvent, data *ConnectedEvent) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "adapter.OnConnected")
	defer sp.End()

	le := adapterLogEntry(ctx, nil, event)
	addSpanAttributes(ctx, sp, event)

	le.Info("Connection established to provider")

	channels, err := event.Adapter.GetPresentChannels()
	if err != nil {
		telemetry.Errors().WithError(err).Commit(ctx)
		addSpanAttributes(ctx, sp, err)
		le.WithError(err).Error("Failed to get channels list")
		return
	}

	for _, c := range channels {
		message := fmt.Sprintf("Gort version %s is online. Hello, %s!", version.Version, c.Name)
		err := SendMessage(ctx, event.Adapter, c.ID, message)
		if err != nil {
			telemetry.Errors().WithError(err).Commit(ctx)
			addSpanAttributes(ctx, sp, err)
			le.WithError(err).Error("Failed to send greeting")
		}
	}
}

// OnChannelMessage handles ChannelMessageEvent events.
// If a command is found in the text, it will emit a data.CommandRequest
// instance to the commands channel.
// TODO Support direct in-channel mentions.
func OnChannelMessage(ctx context.Context, event *ProviderEvent, data *ChannelMessageEvent) (*data.CommandRequest, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "adapter.OnChannelMessage")
	defer sp.End()

	rawCommandText := data.Text

	// Ignore empty messages
	if len(rawCommandText) <= 1 {
		return nil, nil
	}

	// If this starts with a trigger character but enable_spoken_commands is false, ignore.
	if rawCommandText[0] == '!' && !config.GetGortServerConfigs().EnableSpokenCommands {
		return nil, nil
	}

	id, err := buildRequestorIdentity(ctx, event.Adapter, event.Info.Provider, data.ChannelID, data.UserID)
	if err != nil {
		telemetry.Errors().WithError(err).Commit(ctx)
		SendErrorMessage(ctx, id.Adapter, id.ChatChannel.ID, "Error", unexpectedError)
		return nil, err
	}

	adapterLogEntry(ctx, nil, event, id).
		WithField("command.raw", rawCommandText).
		Debug("Got message")
	addSpanAttributes(ctx, sp, event, attribute.String("command.raw", rawCommandText))

	// Find command by Name if the message starts with '!'
	if rawCommandText[0] == '!' {
		rawCommandText = rawCommandText[1:]
		return BuildCommandRequest(ctx, rawCommandText, id, data.MessageRef, retrieval.CommandFromTokensByName)
	}

	// Otherwise attempt to find command by trigger
	return BuildCommandRequest(ctx, rawCommandText, id, data.MessageRef, retrieval.CommandFromTokensByTrigger)
}

// OnDirectMessage handles DirectMessageEvent events.
func OnDirectMessage(ctx context.Context, event *ProviderEvent, data *DirectMessageEvent) (*data.CommandRequest, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "adapter.OnDirectMessage")
	defer sp.End()

	rawCommandText := data.Text

	id, err := buildRequestorIdentity(ctx, event.Adapter, event.Info.Provider, data.ChannelID, data.UserID)
	if err != nil {
		telemetry.Errors().WithError(err).Commit(ctx)
		SendErrorMessage(ctx, id.Adapter, id.ChatChannel.ID, "Error", unexpectedError)
		return nil, err
	}

	adapterLogEntry(ctx, nil, event, id).
		WithField("command.raw", rawCommandText).
		Debug("Got direct message")
	addSpanAttributes(ctx, sp, event, attribute.String("command.raw", rawCommandText))

	if rawCommandText[0] == '!' {
		rawCommandText = rawCommandText[1:]
		return BuildCommandRequest(ctx, rawCommandText, id, data.MessageRef, retrieval.CommandFromTokensByName)
	}
	return BuildCommandRequest(ctx, rawCommandText, id, data.MessageRef, retrieval.CommandFromTokensByNameOrTrigger)
}

// SendErrorMessage sends an error message to a specified channel.
func SendErrorMessage(ctx context.Context, a Adapter, channelID string, title, text string) error {
	e := data.NewCommandResponseEnvelope(data.CommandRequest{}, data.WithError(title, fmt.Errorf(text), 1))
	return SendEnvelope(ctx, a, channelID, e, data.MessageError)
}

// SendMessage sends a standard output message to a specified channel.
func SendMessage(ctx context.Context, a Adapter, channelID string, message string) error {
	e := data.NewCommandResponseEnvelope(data.CommandRequest{}, data.WithResponseLines([]string{message}))
	return SendEnvelope(ctx, a, channelID, e, data.Message)
}

// SendEnvelope sends the contents of a response envelope to a specified
// channel. If channelID is empty the value of envelope.Request.ChannelID will
// be used.
func SendEnvelope(ctx context.Context, a Adapter, channelID string, envelope data.CommandResponseEnvelope, tt data.TemplateType) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "adapter.SendEnvelope")
	defer sp.End()

	handleAdvancedOutput(ctx, envelope.Response.Advanced) // TODO: handle error

	e := adapterLogEntry(ctx, log.WithContext(ctx), a).WithField("message.type", tt)

	template, err := templates.Get(envelope.Request.Command, envelope.Request.Bundle, tt)
	if err != nil {
		e.WithError(err).Error("failed to get template")
		if err := a.SendError(ctx, channelID, "Failed to Get Template", err); err != nil {
			e.WithError(err).Error("break-glass send error failure!")
		}
		return err
	}

	tf, err := templates.Transform(template, envelope)
	if err != nil {
		e.WithError(err).Error("template engine failed to transform template")
		if err := a.SendError(ctx, channelID, "Failed to Transform Template", err); err != nil {
			e.WithError(err).Error("break-glass send error failure!")
		}
		return err
	}

	elements, err := templates.EncodeElements(tf)
	if err != nil {
		e.WithError(err).Error("template engine failed to encode elements")
		if err := a.SendError(ctx, channelID, "Failed to Transform Template", err); err != nil {
			e.WithError(err).Error("break-glass send error failure!")
		}
		return err
	}

	err = a.Send(ctx, channelID, elements)
	if err == nil {
		return nil
	}

	e.WithError(err).Warn("failed to send rich message to adapter, falling back to alt text")
	err = a.SendText(ctx, channelID, elements.Alt())
	if err != nil {
		e.WithError(err).Error("failed to send message to adapter")
		if err := a.SendError(ctx, channelID, "Failed to Send Message", err); err != nil {
			e.WithError(err).Error("break-glass send error failure!")
		}
		return err
	}

	return nil
}

// handleAdvancedOutput executes a slice of adapter actions specified as
// io.AdvancedInput.
func handleAdvancedOutput(ctx context.Context, output []io.AdvancedOutput) error {
	e := adapterLogEntry(ctx, log.WithContext(ctx))
	for _, o := range output {
		e1 := e.WithField("output.action", o.Action).WithField("output.messageref", o.MessageRef)
		var msgRef MessageRef
		err := json.NewDecoder(strings.NewReader(o.MessageRef)).Decode(&msgRef)
		if err != nil {
			e1.WithError(err).Errorf("Badly formatted MessageRef")
		}
		// An alternate way to get the adapter will be necessary for actions
		// that don't involve existing messages.
		a, err := GetAdapter(msgRef.Adapter)
		if err != nil {
			e.WithError(err).Error("Couldn't find specified adapter")
			return nil
		} else {
			e1 = e1.WithField("adapter.name", a.GetName())
		}
		switch o.Action {
		case io.ActionReply:
			err = a.Reply(ctx, msgRef, o.Content)
			if err != nil {
				e1.WithError(err).Errorf("Failed to create reply")
			} else {
				e1.Debug("Replied!")
			}
		case io.ActionReact:
			err = a.React(ctx, msgRef, emoji.From(o.Content))
			if err != nil {
				e1.WithError(err).Error("Failed to react")
			} else {
				e1.Debug("Reacted!")
			}
		default:
			e1.Error("Unknown action")
		}
	}

	return nil
}

// StartListening instructs all relays to establish connections, receives all
// events from all relays, and forwards them to the various On* handler functions.
func StartListening(ctx context.Context) (<-chan data.CommandRequest, chan<- data.CommandResponseEnvelope, <-chan error) {
	log.Debug("Instructing relays to establish connections")

	commandRequests := make(chan data.CommandRequest)
	commandResponses := make(chan data.CommandResponseEnvelope)

	allEvents, adapterErrors := startAdapters(ctx)

	// Start listening for events coming from the chat provider
	go startProviderEventListening(commandRequests, allEvents, adapterErrors)

	// Start listening for responses coming back from the relay
	go startRelayResponseListening(commandResponses, allEvents, adapterErrors)

	return commandRequests, commandResponses, adapterErrors
}

// requestLog encapsulates objects required to log operations when building a request.
type requestLog struct {
	da      dataaccess.DataAccess
	request *data.CommandRequest
	id      *RequestorIdentity
	le      *log.Entry
}

// logAction defines a function that allows additional logging actions to be
// passed to requestLog.Error.
type logAction func(ctx context.Context, r *requestLog)

// logUserMessage allows an error to be sent to the user via a chat message.
func logUserMessage(title, msg string) logAction {
	return func(ctx context.Context, r *requestLog) {
		SendErrorMessage(ctx, r.id.Adapter, r.id.ChatChannel.ID, title, msg)
	}
}

// Error performs actions required to log an error in telemetry, logs and optionally
// other actions provided by logAction functions.
func (r *requestLog) Error(
	ctx context.Context,
	err error,
	logMessage string,
	actions ...logAction,
) error {
	r.da.RequestError(ctx, *r.request, err)
	telemetry.Errors().WithError(err).Commit(ctx)
	r.le.WithField("command.bundle", r.request).WithError(err).Error(logMessage)
	for _, action := range actions {
		action(ctx, r)
	}
	return fmt.Errorf("%v: %w", logMessage, err)
}

// BuildCommandRequest builds a CommandRequest object based on the provided
// message content and user id. Both user existence and authorization are
// verified. A lookup function for identifying a command based on tokens must
// be provided as a parameter. Telemetry for the lookup operation are handled
// inside this function.
func BuildCommandRequest(
	ctx context.Context,
	rawCommand string,
	id RequestorIdentity,
	m MessageRef,
	fCommandFromTokens retrieval.CommandFromTokens,
) (*data.CommandRequest, error) {
	// Start trace span
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "adapter.BuildCommandRequest")
	defer sp.End()

	da, err := dataaccess.Get()
	if err != nil {
		return nil, err
	}

	// Tokenize the raw command.
	tokens, err := command.Tokenize(rawCommand)
	if err != nil {
		return nil, fmt.Errorf("command tokenization error: %w", err)
	}

	cmdEntry, cmdInput, commandLookupErr := fCommandFromTokens(ctx, tokens)
	if commandLookupErr == nil && cmdEntry == nil {
		return nil, nil
	}

	request, id, rl, err := buildAndBeginRequest(ctx, id)
	if err != nil {
		return nil, err
	}

	if len(tokens) == 0 {
		msg := "Empty command received. Did you forget something?"
		return nil, rl.Error(ctx, err, "command had no tokens", logUserMessage("Empty Command", msg))
	}

	if commandLookupErr != nil {
		err := commandLookupErr
		switch {
		case gerrs.Is(err, retrieval.ErrNoSuchCommand):
			msg := fmt.Sprintf("No such bundle is currently installed: %s.\n"+
				"If this is not expected, you should contact a Gort administrator.",
				tokens[0])
			return nil, rl.Error(ctx, err, "command lookup error", logUserMessage("No Such Command", msg))
		case gerrs.Is(err, retrieval.ErrMultipleCommands):
			msg := fmt.Sprintf("The command %s matches multiple bundles.\n"+
				"Please namespace your command using the bundle name: `bundle:command`.",
				tokens[0])
			return nil, rl.Error(ctx, err, "command lookup error", logUserMessage("No Such Command", msg))
		default:
			return nil, rl.Error(ctx, err, "command lookup error", logUserMessage("Error", err.Error()))
		}
	}

	rl.le = rl.le.WithField("command.name", cmdEntry.Command.Name).
		WithField("command.params", cmdInput.Parameters.String())
	request.Parameters = retrieval.ParametersFromCommand(cmdInput)
	da.RequestUpdate(ctx, request)

	cmdFoundMessage := fmt.Sprintf("Executing command: %s", cmdEntry.Command.Name)
	err = SendMessage(ctx, id.Adapter, id.ChatChannel.ID, cmdFoundMessage)
	if err != nil {
		rl.Error(ctx, err, "failed to send command acknowledgement")
	}

	request.CommandEntry = *cmdEntry
	da.RequestUpdate(ctx, request)

	// Update log entry with cmd info
	rl.le = adapterLogEntry(ctx, rl.le, *cmdEntry)
	rl.le.Debug("Found matching command+bundle")
	addSpanAttributes(ctx, sp, *cmdEntry)

	err = auth.CheckPermissions(ctx, id.GortUser.Username, cmdInput, *cmdEntry)
	if err != nil {
		switch {
		case gerrs.Is(err, auth.ErrRuleLoadError):
			return nil, rl.Error(ctx, err, "rule load error", logUserMessage("Error", unexpectedError))
		case gerrs.Is(err, auth.ErrNoRulesDefined):
			msg := fmt.Sprintf("The command %s:%s doesn't have any associated rules.\n"+
				"For a command to be executable, it must have at least one rule.",
				cmdEntry.Bundle.Name, cmdEntry.Command.Name)
			return nil, rl.Error(ctx, err, "no rules defined", logUserMessage("No Rules Defined", msg))
		case gerrs.Is(err, auth.ErrNotAllowed):
			msg := fmt.Sprintf("You do not have the permissions to execute %s:%s.", cmdEntry.Bundle.Name, cmdEntry.Command.Name)
			return nil, rl.Error(ctx, err, "permission denied", logUserMessage("Permission Denied", msg))
		default:
			return nil, rl.Error(ctx, err, "permission check failure", logUserMessage("Error", unexpectedError))
		}
	}

	if cmdEntry.Command.Input.Advanced {
		request = advancedInput(request, id, cmdInput, m)
	}

	// Update log entry with command info
	rl.le.Info("Triggering command")

	return &request, nil
}

func advancedInput(req data.CommandRequest, id RequestorIdentity, c command.Command, m MessageRef) data.CommandRequest {
	gu := *id.GortUser
	gu.Mappings = nil
	gu.Password = ""

	mb, err := json.Marshal(m)
	if err != nil {
		// oops
	}

	ai := io.AdvancedInput{
		Channel:      *id.ChatChannel,
		Command:      io.NewCommandInfo(c),
		Provider:     *id.Provider,
		ProviderUser: *id.ChatUser,
		GortUser:     *id.GortUser,
		MessageRef:   string(mb),
	}

	req.Parameters = data.CommandParameters([]string{ai.String()})

	return req
}

// adapterLogEntry is a helper that pre-populates a log event with attributes.
func adapterLogEntry(ctx context.Context, e *log.Entry, obs ...interface{}) *log.Entry {
	if e == nil {
		e = log.WithContext(ctx)
	}

	sp := trace.SpanFromContext(ctx)
	if sp.SpanContext().HasTraceID() {
		e = e.WithField("trace.id", sp.SpanContext().TraceID())
	}

	for _, i := range obs {
		if i == nil {
			continue
		}

		switch o := i.(type) {
		case Adapter:
			e = e.WithField("adapter.name", o.GetName())

		case io.ChannelInfo:
			e = e.WithField("provider.channel.name", o.Name).
				WithField("provider.channel.id", o.ID)

		case *ProviderEvent:
			e = e.WithField("event", o.EventType).
				WithField("adapter.name", o.Info.Provider.Name).
				WithField("adapter.type", o.Info.Provider.Type)

		case RequestorIdentity:
			args := []interface{}{}

			if o.Adapter != nil {
				args = append(args, o.Adapter)
			}
			if o.ChatChannel != nil {
				args = append(args, *o.ChatChannel)
			}
			if o.ChatUser != nil {
				args = append(args, *o.ChatUser)
			}
			if o.GortUser != nil {
				args = append(args, *o.GortUser)
			}

			return adapterLogEntry(ctx, e, args...)

		case io.UserInfo:
			e = e.WithField("provider.user.email", o.Email).
				WithField("provider.user.id", o.ID)

		case *rest.User:
			e = adapterLogEntry(ctx, e, *o)

		case rest.User:
			e = e.WithField("gort.user.name", o.Username)

		case data.Bundle:
			e = e.WithField("bundle.name", o.Name).
				WithField("bundle.version", o.Version).
				WithField("bundle.default", o.Default)

		case data.BundleCommand:
			e = e.WithField("command.executable", o.Executable).
				WithField("command.name", o.Name)

		case data.CommandEntry:
			e = adapterLogEntry(ctx, e, o.Bundle, o.Command)

		case data.CommandRequest:
			e = adapterLogEntry(ctx, e, o.CommandEntry, o.Bundle).
				WithField("command.params", strings.Join(o.Parameters, " ")).
				WithField("request.id", o.RequestID)

		case trace.Span:
			if o.SpanContext().HasTraceID() {
				e = e.WithField("trace.id", o.SpanContext().TraceID())
			}

		case error:
			e = e.WithError(o)

		default:
			panic(fmt.Sprintf("I don't know how to log a %T", i))
		}
	}

	return e
}

// addSpanAttributes is a helper that populates a tracing span with attributes.
func addSpanAttributes(ctx context.Context, sp trace.Span, obs ...interface{}) {
	attr := []attribute.KeyValue{}

	for _, i := range obs {
		if i == nil {
			continue
		}

		switch o := i.(type) {
		case Adapter:
			attr = append(attr,
				attribute.String("adapter.name", o.GetName()),
			)

		case io.ChannelInfo:
			attr = append(attr,
				attribute.String("provider.channel.name", o.Name),
				attribute.String("provider.channel.id", o.ID),
			)

		case *ProviderEvent:
			attr = append(attr,
				attribute.String("event", string(o.EventType)),
				attribute.String("adapter.name", o.Info.Provider.Name),
				attribute.String("adapter.type", o.Info.Provider.Type),
			)

		case RequestorIdentity:
			args := []interface{}{}

			if o.Adapter != nil {
				args = append(args, o.Adapter)
			}
			if o.ChatChannel != nil {
				args = append(args, *o.ChatChannel)
			}
			if o.ChatUser != nil {
				args = append(args, *o.ChatUser)
			}
			if o.GortUser != nil {
				args = append(args, *o.GortUser)
			}

			addSpanAttributes(ctx, sp, args...)

		case io.UserInfo:
			attr = append(attr,
				attribute.String("provider.user.email", o.Email),
				attribute.String("provider.user.id", o.ID),
			)

		case attribute.KeyValue:
			attr = append(attr, o)

		case data.Bundle:
			attr = append(attr,
				attribute.String("bundle.name", o.Name),
				attribute.String("bundle.version", o.Version),
				attribute.Bool("bundle.default", o.Default),
			)

		case data.BundleCommand:
			attr = append(attr,
				attribute.String("command.executable", strings.Join(o.Executable, " ")),
				attribute.String("command.name", o.Name),
			)

		case data.CommandEntry:
			addSpanAttributes(ctx, sp, o.Bundle, o.Command)

		case data.CommandRequest:
			addSpanAttributes(ctx, sp, o.CommandEntry, o.Bundle)
			attr = append(attr,
				attribute.String("command.params", strings.Join(o.Parameters, " ")),
				attribute.Int64("request.id", o.RequestID),
			)

		case *rest.User:
			addSpanAttributes(ctx, sp, *o)

		case rest.User:
			attr = append(attr,
				attribute.String("gort.user.name", o.Username),
			)

		case error:
			attr = append(attr,
				attribute.String("error", o.Error()),
			)

		default:
			panic(fmt.Sprintf("I don't know how to get attributes from a %T", i))
		}
	}

	sp.SetAttributes(attr...)
}

func buildRequestorIdentity(ctx context.Context, adapter Adapter, provider *io.ProviderInfo, channelId, userId string) (RequestorIdentity, error) {
	var err error
	id := RequestorIdentity{
		Adapter:  adapter,
		Provider: provider,
	}

	le := adapterLogEntry(ctx, nil, adapter).
		WithField("channelId", channelId).
		WithField("userId", userId)

	if channelId != "" {
		id.ChatChannel, err = adapter.GetChannelInfo(channelId)
		switch {
		case err == nil:
			le = adapterLogEntry(ctx, le, *id.ChatChannel)
		case gerrs.Is(err, ErrChannelNotFound):
			le.WithError(err).WithField("userId", userId).Debug("can't find user")
		default:
			le.WithError(err).Debug("Unexpected error getting requestor channel")
			return id, err
		}
	}

	if userId != "" {
		id.ChatUser, err = adapter.GetUserInfo(userId)
		switch {
		case err == nil:
			le = adapterLogEntry(ctx, le, *id.ChatUser)
		case gerrs.Is(err, errs.ErrNoSuchUser):
			le.WithError(err).WithField("userId", userId).Debug("can't find user")
		default:
			le.WithError(err).Debug("Unexpected error getting requestor user")
			return id, err
		}
	}

	if id.ChatChannel != nil {
		dal, err := dataaccess.Get()
		if err != nil {
			return id, err
		}

		user, err := dal.UserGetByID(ctx, id.Adapter.GetName(), id.ChatUser.ID)
		switch {
		case err == nil:
			id.GortUser = &user
			le = adapterLogEntry(ctx, le, *id.GortUser)
		case gerrs.Is(err, errs.ErrNoSuchUser):
			le.WithError(err).WithField("userId", userId).Debug("can't find user")
		default:
			le.WithError(err).Debug("Unexpected error getting requestor user by email")
			return id, err
		}
	}

	le.Info("Requestor identity built")

	return id, nil
}

// buildAndBeginRequest sets up a data.CommandRequest.
// User information is verified and populated as needed.
// The user is only required to exist, permission checks take place later.
func buildAndBeginRequest(ctx context.Context, id RequestorIdentity) (data.CommandRequest, RequestorIdentity, requestLog, error) {
	request := data.CommandRequest{
		Adapter:   id.Adapter.GetName(),
		ChannelID: id.ChatChannel.ID,
		Context:   ctx,
		Timestamp: time.Now(),
		UserEmail: id.ChatUser.Email,
		UserID:    id.ChatUser.ID,
	}

	if id.GortUser != nil {
		request.UserEmail = id.GortUser.Email
		request.UserName = id.GortUser.Username
	}

	r := requestLog{
		request: &request,
		id:      &id,
	}
	var err error
	r.le = adapterLogEntry(ctx, nil, id, request)
	r.da, err = dataaccess.Get()
	if err != nil {
		return data.CommandRequest{}, RequestorIdentity{}, requestLog{}, err
	}

	r.da.RequestBegin(ctx, &request)
	addSpanAttributes(ctx, trace.SpanFromContext(ctx), id, request)

	if id.GortUser != nil {
		return request, id, r, nil
	}

	var autocreated bool
	if id.GortUser, autocreated, err = findOrMakeGortUser(ctx, id.Adapter, id.ChatUser); err != nil {
		switch {
		case gerrs.Is(err, ErrSelfRegistrationOff):
			msg := "I'm terribly sorry, but either I don't have a Gort " +
				"account for you, or your chat handle has not been " +
				"registered. Currently, only registered users can " +
				"interact with me.\n\nYou'll need a Gort administrator " +
				"to map your Gort user to the adapter (%s) and chat " +
				"user ID (%s)."
			msg = fmt.Sprintf(msg, id.Adapter.GetName(), id.ChatUser.ID)
			SendErrorMessage(ctx, id.Adapter, id.ChatChannel.ID, "No Such Account", msg)

		case gerrs.Is(err, ErrGortNotBootstrapped):
			msg := "Gort doesn't appear to have been bootstrapped yet! Please " +
				"use `gort bootstrap` to properly bootstrap the Gort " +
				"environment before proceeding."
			SendErrorMessage(ctx, id.Adapter, id.ChatChannel.ID, "Not Bootstrapped?", msg)

		default:
			msg := "An unexpected error has occurred"
			SendErrorMessage(ctx, id.Adapter, id.ChatChannel.ID, "Error", msg)
		}

		r.da.RequestError(ctx, request, err)
		telemetry.Errors().WithError(err).Commit(ctx)
		r.le.WithError(err).Error("Can't find or create user")

		return data.CommandRequest{}, RequestorIdentity{}, requestLog{}, fmt.Errorf("can't find or create user: %w", err)
	}

	request.UserEmail = id.GortUser.Email
	request.UserName = id.GortUser.Username
	r.da.RequestUpdate(ctx, request)

	addSpanAttributes(ctx, trace.SpanFromContext(ctx), id.GortUser)
	r.le = adapterLogEntry(ctx, nil, id.GortUser)

	if autocreated {
		message := fmt.Sprintf("Hello! It's great to meet you! You're the proud "+
			"owner of a shiny new Gort account named `%s`!",
			id.GortUser.Username)
		SendMessage(ctx, id.Adapter, id.ChatUser.ID, message)

		r.le.Info("Autocreating Gort user")
	}

	return request, id, r, nil
}

// findOrMakeGortUser ...
func findOrMakeGortUser(ctx context.Context, adapter Adapter, info *io.UserInfo) (*rest.User, bool, error) {
	// Get the data access interface.
	da, err := dataaccess.Get()
	if err != nil {
		return nil, false, err
	}

	// Try to figure out what user we're working with here.
	exists := true
	user, err := da.UserGetByID(ctx, adapter.GetName(), info.ID)
	if gerrs.Is(err, errs.ErrNoSuchUser) {
		exists = false
	} else if err != nil {
		return nil, false, err
	}

	// It already exists. Exist.
	if exists {
		return &user, false, nil
	}

	// We can create the user... unless the instance hasn't been bootstrapped.
	bootstrapped, err := da.UserExists(ctx, "admin")
	if err != nil {
		return nil, false, err
	}
	if !bootstrapped {
		return nil, false, ErrGortNotBootstrapped
	}

	// Now we know it doesn't exist. If self-registration is off, exit with
	// an error.
	if !config.GetGortServerConfigs().AllowSelfRegistration {
		return nil, false, ErrSelfRegistrationOff
	}

	// Generate a random password for the auto-created user.
	randomPassword, err := data.GenerateRandomToken(32)
	if err != nil {
		return nil, false, err
	}

	// Let's create the user!
	user = rest.User{
		Email:    info.Email,
		FullName: info.RealNameNormalized,
		Password: randomPassword,
		Username: info.Name,
		Mappings: map[string]string{adapter.GetName(): info.ID},
	}

	log.WithField("user.username", user.Username).
		WithField("user.email", user.Email).
		Info("User auto-created")

	return &user, true, da.UserCreate(ctx, user)
}

func handleIncomingEvent(event *ProviderEvent, commandRequests chan<- data.CommandRequest, adapterErrors chan<- error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(context.Background(), "adapter.handleIncomingEvent")
	defer sp.End()

	sp.SetAttributes(attribute.String("event.type", string(event.EventType)))

	switch ev := event.Data.(type) {
	case *ConnectedEvent:
		OnConnected(ctx, event, ev)

	case *DisconnectedEvent:
		// Do nothing.

	case *AuthenticationErrorEvent:
		adapterErrors <- gerrs.Wrap(ErrAuthenticationFailure, errors.New(ev.Msg))

	case *ChannelMessageEvent:
		request, err := OnChannelMessage(ctx, event, ev)
		if request != nil {
			commandRequests <- *request
		}
		if err != nil {
			adapterErrors <- err
		}

	case *DirectMessageEvent:
		request, err := OnDirectMessage(ctx, event, ev)
		if request != nil {
			commandRequests <- *request
		}
		if err != nil {
			adapterErrors <- err
		}

	case *ErrorEvent:
		adapterErrors <- ev

	default:
		log.WithField("type", fmt.Sprintf("%T", ev)).Warn("Unknown event type")
	}
}

func startAdapters(ctx context.Context) (<-chan *ProviderEvent, chan error) {
	allEvents := make(chan *ProviderEvent)

	adapterErrors := make(chan error, len(config.GetSlackProviders()))

	for k, a := range adapterLookup {
		log.WithField("adapter.name", k).Debug("Starting adapter")

		go func(adapter Adapter) {
			for event := range adapter.Listen(ctx) {
				allEvents <- event
			}
		}(a)
	}

	return allEvents, adapterErrors
}

func startProviderEventListening(requests chan<- data.CommandRequest,
	allEvents <-chan *ProviderEvent, adapterErrors chan<- error) {

	for event := range allEvents {
		handleIncomingEvent(event, requests, adapterErrors)
	}
}

func startRelayResponseListening(responses <-chan data.CommandResponseEnvelope,
	allEvents <-chan *ProviderEvent, adapterErrors chan<- error) {

	for envelope := range responses {
		adapter, err := GetAdapter(envelope.Request.Adapter)
		if err != nil {
			adapterErrors <- err
			continue
		}

		tt := data.Command
		if envelope.Data.ExitCode != 0 {
			tt = data.CommandError
		}

		ctx := context.Background()
		channelID := envelope.Request.ChannelID
		if err := SendEnvelope(ctx, adapter, channelID, envelope, tt); err != nil {
			adapterErrors <- err
		}
	}
}
