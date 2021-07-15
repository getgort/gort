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
	"errors"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/getgort/gort/auth"
	"github.com/getgort/gort/bundles"
	"github.com/getgort/gort/command"
	"github.com/getgort/gort/config"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/dataaccess"
	"github.com/getgort/gort/dataaccess/errs"
	gerrs "github.com/getgort/gort/errors"
	"github.com/getgort/gort/rules"
	"github.com/getgort/gort/telemetry"
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

	// ErrMultipleCommands is returned by GetCommandEntry when the same command
	// shortcut matches commands in two or more bundles.
	ErrMultipleCommands = errors.New("multiple commands match that pattern")

	// ErrNoSuchAdapter is returned by GetAdapter if a requested adapter name
	// can't be found.
	ErrNoSuchAdapter = errors.New("no such adapter")

	// ErrNoSuchCommand is returned by GetCommandEntry if a request command
	// isn't found.
	ErrNoSuchCommand = errors.New("no such bundle")

	// ErrUserNotFound is throws by several methods if a provider fails to
	// return requested user information.
	ErrUserNotFound = errors.New("user not found")
)

// Adapter represents a connection to a chat provider.
type Adapter interface {
	// GetChannelInfo provides info on a specific provider channel accessible
	// to the adapter.
	GetChannelInfo(channelID string) (*ChannelInfo, error)

	// GetName provides the name of this adapter as per the configuration.
	GetName() string

	// GetPresentChannels returns a slice of channels that a user is present in.
	GetPresentChannels(userID string) ([]*ChannelInfo, error)

	// GetUserInfo provides info on a specific provider user accessible
	// to the adapter.
	GetUserInfo(userID string) (*UserInfo, error)

	// Listen causes the Adapter to initiate a connection to its provider and
	// begin relaying back events (including errors) via the returned channel.
	Listen(ctx context.Context) <-chan *ProviderEvent

	// SendErrorMessage sends an error message to a specified channel.
	// TODO Create a MessageBuilder at some point to replace this.
	SendErrorMessage(channelID string, title string, text string) error

	// SendMessage sends a standard output message to a specified channel.
	// TODO Create a MessageBuilder at some point to replace this.
	SendMessage(channel string, message string) error
}

type RequestorIdentity struct {
	Adapter     Adapter
	ChatUser    *UserInfo
	ChatChannel *ChannelInfo
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

// GetCommandEntry accepts a tokenized parameter slice and returns any
// associated data.CommandEntry instances. If the number of matching
// commands is > 1, an error is returned.
func GetCommandEntry(ctx context.Context, bundleName, commandName string) (data.CommandEntry, error) {
	finders, err := allCommandEntryFinders()
	if err != nil {
		return data.CommandEntry{}, err
	}

	entries, err := findAllEntries(ctx, bundleName, commandName, finders...)
	if err != nil {
		return data.CommandEntry{}, err
	}

	if len(entries) == 0 {
		return data.CommandEntry{}, ErrNoSuchCommand
	}

	if len(entries) > 1 {
		cmd := commandName
		if bundleName != "" {
			cmd = bundleName + ":" + commandName
		}

		log.
			WithField("requested", cmd).
			WithField("bundle0", entries[0].Bundle.Name).
			WithField("command0", entries[0].Command.Name).
			WithField("bundle1", entries[1].Bundle.Name).
			WithField("command1", entries[1].Command.Name).
			Warn("Multiple commands found")

		return data.CommandEntry{}, ErrMultipleCommands
	}

	return entries[0], nil
}

// OnConnected handles ConnectedEvent events.
func OnConnected(ctx context.Context, event *ProviderEvent, data *ConnectedEvent) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "adapter.OnConnected")
	defer sp.End()

	le := adapterLogEntry(ctx, nil, event)
	addSpanAttributes(ctx, sp, event)

	le.Info("Connection established to provider")

	channels, err := event.Adapter.GetPresentChannels(event.Info.User.ID)
	if err != nil {
		telemetry.Errors().WithError(err).Commit(ctx)
		addSpanAttributes(ctx, sp, err)
		le.WithError(err).Error("Failed to get channels list")
		return
	}

	for _, c := range channels {
		message := fmt.Sprintf("Gort version %s is online. Hello, %s!", version.Version, c.Name)
		err := event.Adapter.SendMessage(c.ID, message)
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

	// If this isn't prepended by a trigger character, ignore.
	if len(rawCommandText) <= 1 || rawCommandText[0] != '!' {
		return nil, nil
	}

	// If this starts with a trigger character but enable_spoken_commands is false, ignore.
	if rawCommandText[0] == '!' && !config.GetGortServerConfigs().EnableSpokenCommands {
		return nil, nil
	}

	id, err := buildRequestorIdentity(ctx, event.Adapter, data.ChannelID, data.UserID)
	if err != nil {
		telemetry.Errors().WithError(err).Commit(ctx)
		id.Adapter.SendErrorMessage(id.ChatChannel.ID, "Error", unexpectedError)
		return nil, err
	}

	// Remove the "trigger character" (!)
	rawCommandText = rawCommandText[1:]

	adapterLogEntry(ctx, nil, event, id).
		WithField("command.raw", rawCommandText).
		Debug("Got message")
	addSpanAttributes(ctx, sp, event, attribute.String("command.raw", rawCommandText))

	return TriggerCommand(ctx, rawCommandText, id)
}

// OnDirectMessage handles DirectMessageEvent events.
func OnDirectMessage(ctx context.Context, event *ProviderEvent, data *DirectMessageEvent) (*data.CommandRequest, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "adapter.OnDirectMessage")
	defer sp.End()

	rawCommandText := data.Text

	if rawCommandText[0] == '!' {
		rawCommandText = rawCommandText[1:]
	}

	id, err := buildRequestorIdentity(ctx, event.Adapter, data.ChannelID, data.UserID)
	if err != nil {
		telemetry.Errors().WithError(err).Commit(ctx)
		id.Adapter.SendErrorMessage(id.ChatChannel.ID, "Error", unexpectedError)
		return nil, err
	}

	adapterLogEntry(ctx, nil, event, id).
		WithField("command.raw", rawCommandText).
		Debug("Got direct message")
	addSpanAttributes(ctx, sp, event, attribute.String("command.raw", rawCommandText))

	return TriggerCommand(ctx, rawCommandText, id)
}

// StartListening instructs all relays to establish connections, receives all
// events from all relays, and forwards them to the various On* handler functions.
func StartListening(ctx context.Context) (<-chan data.CommandRequest, chan<- data.CommandResponse, <-chan error) {
	log.Debug("Instructing relays to establish connections")

	commandRequests := make(chan data.CommandRequest)
	commandResponses := make(chan data.CommandResponse)

	allEvents, adapterErrors := startAdapters(ctx)

	// Start listening for events coming from the chat provider
	go startProviderEventListening(commandRequests, allEvents, adapterErrors)

	// Start listening for responses coming back from the relay
	go startRelayResponseListening(commandResponses, allEvents, adapterErrors)

	return commandRequests, commandResponses, adapterErrors
}

// TriggerCommand is called by OnChannelMessage or OnDirectMessage when a
// valid command trigger is identified. This function is at the core Gort's
// command response capabilities, and it's the most complex in the project.
// This looks like a long, complicated function, but it's like 75% logging
// and tracing.
func TriggerCommand(ctx context.Context, rawCommand string, id RequestorIdentity) (*data.CommandRequest, error) {
	// Start trace span
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "adapter.TriggerCommand")
	defer sp.End()

	da, err := dataaccess.Get()
	if err != nil {
		return nil, err
	}

	request := buildRequest(ctx, id)
	da.RequestBegin(ctx, &request)

	addSpanAttributes(ctx, sp, id, request)
	le := adapterLogEntry(ctx, nil, id, request)

	if id.GortUser == nil {
		var autocreated bool

		if id.GortUser, autocreated, err = findOrMakeGortUser(ctx, id.ChatUser); err != nil {
			switch {
			case gerrs.Is(err, ErrSelfRegistrationOff):
				msg := "I'm terribly sorry, but either I don't " +
					"have a Gort account for you, or your Slack chat handle has " +
					"not been registered. Currently, only registered users can " +
					"interact with me.\n\n\nYou'll need to ask a Gort " +
					"administrator to fix this situation and to register your " +
					"Slack handle."
				id.Adapter.SendErrorMessage(id.ChatChannel.ID, "No Such Account", msg)
			case gerrs.Is(err, ErrGortNotBootstrapped):
				msg := "Gort doesn't appear to have been bootstrapped yet! Please " +
					"use `gort bootstrap` to properly bootstrap the Gort " +
					"environment before proceeding."
				id.Adapter.SendErrorMessage(id.ChatChannel.ID, "Not Bootstrapped?", msg)
			default:
				msg := "An unexpected error has occurred"
				id.Adapter.SendErrorMessage(id.ChatChannel.ID, "Error", msg)
			}

			da.RequestError(ctx, request, err)
			telemetry.Errors().WithError(err).Commit(ctx)
			le.WithError(err).Error("Can't find or create user")

			return nil, fmt.Errorf("can't find or create user: %w", err)
		}

		request.UserEmail = id.GortUser.Email
		request.UserName = id.GortUser.Username
		da.RequestUpdate(ctx, request)

		addSpanAttributes(ctx, sp, id.GortUser)
		le := adapterLogEntry(ctx, nil, id.GortUser)

		if autocreated {
			message := fmt.Sprintf("Hello! It's great to meet you! You're the proud "+
				"owner of a shiny new Gort account named `%s`!",
				id.GortUser.Username)
			id.Adapter.SendMessage(id.ChatUser.ID, message)

			le.Info("Autocreating Gort user")
		}
	}

	// Tokenize the raw command.
	tokens, err := command.Tokenize(rawCommand)
	if err != nil {
		da.RequestError(ctx, request, err)
		telemetry.Errors().WithError(err).Commit(ctx)
		le.WithError(err).Error("Command tokenization error")
		id.Adapter.SendErrorMessage(id.ChatChannel.ID, "Error", unexpectedError)

		return nil, fmt.Errorf("command tokenization error: %w", err)
	}

	le = le.WithField("command.name", tokens[0]).
		WithField("command.params", strings.Join(tokens[1:], " "))

	request.Parameters = tokens[1:]
	da.RequestUpdate(ctx, request)

	// Build a temporary Command value using default tokenization rules. We'll
	// use this to load the CommandEntry for the relevant command (as defined
	// in a command bundle), which contains the command's parsing rules that
	// we'll use for a final, formal Parse to get the final Command version.
	cmdInput, err := command.Parse(tokens)
	if err != nil {
		da.RequestError(ctx, request, err)
		telemetry.Errors().WithError(err).Commit(ctx)
		le.WithError(err).Error("Command parse error")
		id.Adapter.SendErrorMessage(id.ChatChannel.ID, "Error", unexpectedError)

		return nil, fmt.Errorf("command parse error: %w", err)
	}

	// Try to find the relevant command entry.
	cmdEntry, err := GetCommandEntry(ctx, cmdInput.Bundle, cmdInput.Command)
	if err != nil {
		switch {
		case gerrs.Is(err, ErrNoSuchCommand):
			msg := fmt.Sprintf("No such bundle is currently installed: %s.\n"+
				"If this is not expected, you should contact a Gort administrator.",
				tokens[0])
			id.Adapter.SendErrorMessage(id.ChatChannel.ID, "No Such Command", msg)
		case gerrs.Is(err, ErrMultipleCommands):
			msg := fmt.Sprintf("The command %s matches multiple bundles.\n"+
				"Please namespace your command using the bundle name: `bundle:command`.",
				tokens[0])
			id.Adapter.SendErrorMessage(id.ChatChannel.ID, "No Such Command", msg)
		default:
			msg := formatCommandInputErrorMessage(cmdInput, tokens, err.Error())
			id.Adapter.SendErrorMessage(id.ChatChannel.ID, "Error", msg)
		}

		da.RequestError(ctx, request, err)
		telemetry.Errors().WithError(err).Commit(ctx)
		le.WithError(err).Error("Command lookup error")

		return nil, fmt.Errorf("command lookup error: %w", err)
	}

	// Now that we have a command entry, we can re-create the complete Command value.
	tokens[0] = cmdEntry.Bundle.Name + ":" + cmdEntry.Command.Name
	// TODO Set parse options based on the CommandEntry settings.
	cmdInput, err = command.Parse(tokens)
	if err != nil {
		da.RequestError(ctx, request, err)
		telemetry.Errors().WithError(err).Commit(ctx)
		le.WithError(err).Error("Command parse error")
		id.Adapter.SendErrorMessage(id.ChatChannel.ID, "Error", unexpectedError)

		return nil, fmt.Errorf("command parse error: %w", err)
	}

	request.CommandEntry = cmdEntry
	da.RequestUpdate(ctx, request)

	// Update log entry with cmd info
	le = adapterLogEntry(ctx, le, cmdEntry)
	le.Debug("Found matching command+bundle")
	addSpanAttributes(ctx, sp, cmdEntry)

	env := rules.EvaluationEnvironment{
		"option": cmdInput.OptionsValues(),
		"arg":    cmdInput.Parameters,
	}

	perms, err := da.UserPermissionList(ctx, id.GortUser.Username)
	if err != nil {
		da.RequestError(ctx, request, err)
		telemetry.Errors().WithError(err).Commit(ctx)
		le.WithError(err).Error("User permission load failure")
		id.Adapter.SendErrorMessage(id.ChatChannel.ID, "Error", unexpectedError)

		return nil, fmt.Errorf("user permission load error: %w", err)
	}

	allowed, err := auth.EvaluateCommandEntry(perms.Strings(), cmdEntry, env)
	if err != nil {
		da.RequestError(ctx, request, err)
		telemetry.Errors().WithError(err).Commit(ctx)

		switch {
		case gerrs.Is(err, auth.ErrRuleLoadError):
			le.WithError(err).Error("Rule load error")
			id.Adapter.SendErrorMessage(id.ChatChannel.ID, "Error", unexpectedError)
			return nil, fmt.Errorf("rule load error: %w", err)

		case gerrs.Is(err, auth.ErrNoRulesDefined):
			le.WithError(err).Error("No rules defined")
			msg := fmt.Sprintf("The command %s:%s doesn't have any associated rules.\n"+
				"For a command to be executable, it must have at least one rule.",
				cmdEntry.Bundle.Name, cmdEntry.Command.Name)
			id.Adapter.SendErrorMessage(id.ChatChannel.ID, "No Rules Defined", msg)
			return nil, err
		}
	}

	if !allowed {
		msg := fmt.Sprintf("You do not have the permissions to execute %s:%s.", cmdEntry.Bundle.Name, cmdEntry.Command.Name)
		id.Adapter.SendErrorMessage(id.ChatChannel.ID, "Permission Denied", msg)

		err = fmt.Errorf("permission denied")
		da.RequestError(ctx, request, err)
		telemetry.Errors().WithError(err).Commit(ctx)
		le.WithError(err).Error("Permission Denied")

		return nil, err
	}

	// Update log entry with command info
	le.Info("Triggering command")

	return &request, nil
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

		case ChannelInfo:
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

		case UserInfo:
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

		case ChannelInfo:
			attr = append(attr,
				attribute.String("provider.channel.name", o.Name),
				attribute.String("provider.channel.id", o.ID),
			)

		case *ProviderEvent:
			attr = append(attr,
				attribute.String("event", o.EventType),
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

		case UserInfo:
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

func allCommandEntryFinders() ([]bundles.CommandEntryFinder, error) {
	finders := make([]bundles.CommandEntryFinder, 0)

	// Get the DAL CommandEntryFinder
	dal, err := dataaccess.Get()
	if err != nil {
		return nil, err
	}

	finders = append(finders, dal)

	return finders, nil
}

func buildRequestorIdentity(ctx context.Context, adapter Adapter, channelId, userId string) (RequestorIdentity, error) {
	var err error
	id := RequestorIdentity{Adapter: adapter}

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

		user, err := dal.UserGetByEmail(ctx, id.ChatUser.Email)
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

	le.Info("requestor identity built")

	return id, nil
}

func buildRequest(ctx context.Context, id RequestorIdentity) data.CommandRequest {
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

	return request
}

func findAllEntries(ctx context.Context, bundleName, commandName string, finder ...bundles.CommandEntryFinder) ([]data.CommandEntry, error) {
	entries := make([]data.CommandEntry, 0)

	for _, f := range finder {
		e, err := f.FindCommandEntry(ctx, bundleName, commandName)
		if err != nil {
			return nil, err
		}

		entries = append(entries, e...)
	}

	return entries, nil
}

// findOrMakeGortUser ...
func findOrMakeGortUser(ctx context.Context, info *UserInfo) (*rest.User, bool, error) {
	// Get the data access interface.
	da, err := dataaccess.Get()
	if err != nil {
		return nil, false, err
	}

	// Try to figure out what user we're working with here.
	exists := true
	user, err := da.UserGetByEmail(ctx, info.Email)
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
	}

	log.WithField("user.username", user.Username).
		WithField("user.email", user.Email).
		Info("User auto-created")

	return &user, true, da.UserCreate(ctx, user)
}

// TODO Replace this with something resembling a template. Eventually.
func formatCommandEntryErrorMessage(command data.CommandEntry, params []string, output string) string {
	rawCommand := fmt.Sprintf(
		"%s:%s %s",
		command.Bundle.Name, command.Command.Name, strings.Join(params, " "))

	return fmt.Sprintf(
		"%s\n```%s```\n%s\n```%s```",
		"The pipeline failed planning the invocation:",
		rawCommand,
		"The specific error was:",
		output,
	)
}

// TODO Replace this with something resembling a template. Eventually.
func formatCommandInputErrorMessage(cmd command.Command, params []string, output string) string {
	rawCommand := fmt.Sprintf(
		"%s:%s %v",
		cmd.Bundle, cmd.Command, cmd.Parameters)

	return fmt.Sprintf(
		"%s\n```%s```\n%s\n```%s```",
		"The pipeline failed planning the invocation:",
		rawCommand,
		"The specific error was:",
		output,
	)
}

// TODO Replace this with something resembling a template. Eventually.
func formatCommandOutput(command data.CommandEntry, params []string, output string) string {
	return fmt.Sprintf("```%s```", output)
}

func handleIncomingEvent(event *ProviderEvent, commandRequests chan<- data.CommandRequest, adapterErrors chan<- error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(context.Background(), "adapter.handleIncomingEvent")
	defer sp.End()

	sp.SetAttributes(attribute.String("event.type", event.EventType))

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

func startProviderEventListening(commandRequests chan<- data.CommandRequest,
	allEvents <-chan *ProviderEvent, adapterErrors chan<- error) {

	for event := range allEvents {
		handleIncomingEvent(event, commandRequests, adapterErrors)
	}
}

func startRelayResponseListening(commandResponses <-chan data.CommandResponse,
	allEvents <-chan *ProviderEvent, adapterErrors chan<- error) {

	for response := range commandResponses {
		adapter, err := GetAdapter(response.Command.Adapter)
		if err != nil {
			adapterErrors <- err
			continue
		}

		channelID := response.Command.ChannelID
		output := strings.Join(response.Output, "\n")
		title := response.Title

		if response.Status != 0 || response.Error != nil {
			formatted := formatCommandEntryErrorMessage(
				response.Command.CommandEntry,
				response.Command.Parameters,
				output,
			)

			err = adapter.SendErrorMessage(channelID, title, formatted)
		} else {
			formatted := formatCommandOutput(
				response.Command.CommandEntry,
				response.Command.Parameters,
				output,
			)

			err = adapter.SendMessage(channelID, formatted)
		}

		if err != nil {
			adapterErrors <- err
		}
	}
}
