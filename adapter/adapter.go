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

	log "github.com/sirupsen/logrus"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/getgort/gort/bundles"
	"github.com/getgort/gort/config"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/dataaccess"
	"github.com/getgort/gort/dataaccess/errs"
	gerrs "github.com/getgort/gort/errors"
	"github.com/getgort/gort/telemetry"
	"github.com/getgort/gort/version"
)

var (
	// All existant adapters keyed by name
	adapterLookup = map[string]Adapter{}
)

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

const ServiceName = "gort-controller"

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
	Listen() <-chan *ProviderEvent

	// SendErrorMessage sends an error message to a specified channel.
	// TODO Create a MessageBuilder at some point to replace this.
	SendErrorMessage(channelID string, title string, text string) error

	// SendMessage sends a standard output message to a specified channel.
	// TODO Create a MessageBuilder at some point to replace this.
	SendMessage(channel string, message string) error
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
func GetCommandEntry(ctx context.Context, tokens []string) (data.CommandEntry, error) {
	bundleName, commandName, err := bundles.SplitCommand(tokens[0])
	if err != nil {
		return data.CommandEntry{}, err
	}

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
		log.
			WithField("requested", tokens[0]).
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
	tr := otel.GetTracerProvider().Tracer(ServiceName)
	ctx, sp := tr.Start(ctx, "OnConnected")
	defer sp.End()

	ctx, err := setContextIdentityData(ctx, event.Adapter, "", event.Info.User.ID)
	if err != nil {
		log.WithError(err).Error("failed to set context identity data")
		return
	}

	le := adapterLogEntry(ctx, nil, event, sp)

	le.Info("Connection established to provider")

	channels, err := event.Adapter.GetPresentChannels(event.Info.User.ID)
	if err != nil {
		telemetry.Errors().WithError(err).Commit(ctx)
		le.WithError(err).Error("Failed to get channels list")
		return
	}

	for _, c := range channels {
		message := fmt.Sprintf("Gort version %s is online. Hello, %s!", version.Version, c.Name)
		err := event.Adapter.SendMessage(c.ID, message)
		if err != nil {
			telemetry.Errors().WithError(err).Commit(ctx)
			le.WithError(err).Error("Failed to send greeting")
		}
	}
}

// OnChannelMessage handles ChannelMessageEvent events.
// If a command is found in the text, it will emit a data.CommandRequest
// instance to the commands channel.
// TODO Support direct in-channel mentions.
func OnChannelMessage(ctx context.Context, event *ProviderEvent, data *ChannelMessageEvent) (*data.CommandRequest, error) {
	tr := otel.GetTracerProvider().Tracer(ServiceName)
	ctx, sp := tr.Start(ctx, "OnChannelMessage")
	defer sp.End()

	ctx, err := setContextIdentityData(ctx, event.Adapter, data.ChannelID, data.UserID)
	if err != nil {
		return nil, err
	}

	rawCommandText := data.Text

	// If this isn't prepended by a trigger character, ignore.
	if len(rawCommandText) <= 1 || rawCommandText[0] != '!' {
		return nil, nil
	}

	// If this starts with a trigger character but enable_spoken_commands is false, ignore.
	if rawCommandText[0] == '!' && !config.GetGortServerConfigs().EnableSpokenCommands {
		return nil, nil
	}

	// Remove the "trigger character" (!)
	rawCommandText = rawCommandText[1:]

	adapterLogEntry(ctx, nil, event, sp).
		WithField("event", event.EventType).
		WithField("command", rawCommandText).
		Debug("Got message")

	return TriggerCommand(ctx, rawCommandText, event.Adapter, data.ChannelID, data.UserID)
}

// OnDirectMessage handles DirectMessageEvent events.
func OnDirectMessage(ctx context.Context, event *ProviderEvent, data *DirectMessageEvent) (*data.CommandRequest, error) {
	tr := otel.GetTracerProvider().Tracer(ServiceName)
	ctx, sp := tr.Start(ctx, "OnDirectMessage")
	defer sp.End()

	ctx, err := setContextIdentityData(ctx, event.Adapter, data.ChannelID, data.UserID)
	if err != nil {
		return nil, err
	}

	rawCommandText := data.Text

	if rawCommandText[0] == '!' {
		rawCommandText = rawCommandText[1:]
	}

	adapterLogEntry(ctx, nil, event, sp).
		WithField("event", event.EventType).
		WithField("command.raw", rawCommandText).
		Debug("Got direct message")

	return TriggerCommand(ctx, rawCommandText, event.Adapter, data.ChannelID, data.UserID)
}

// StartListening instructs all relays to establish connections, receives all
// events from all relays, and forwards them to the various On* handler functions.
func StartListening() (<-chan data.CommandRequest, chan<- data.CommandResponse, <-chan error) {
	log.Debug("Instructing relays to establish connections")

	commandRequests := make(chan data.CommandRequest)
	commandResponses := make(chan data.CommandResponse)

	allEvents, adapterErrors := startAdapters()

	// Start listening for events coming from the chat provider
	go startProviderEventListening(commandRequests, allEvents, adapterErrors)

	// Start listening for responses coming back from the relay
	go startRelayResponseListening(commandResponses, allEvents, adapterErrors)

	return commandRequests, commandResponses, adapterErrors
}

// TriggerCommand is called by OnChannelMessage or OnDirectMessage when a
// valid command trigger is identified.
func TriggerCommand(ctx context.Context, rawCommand string, adapter Adapter, channelID string, userID string) (*data.CommandRequest, error) {
	// Start trace span
	tr := otel.GetTracerProvider().Tracer(ServiceName)
	ctx, sp := tr.Start(ctx, "TriggerCommand")
	defer sp.End()

	// If identity data isn't set, ensure that it's set
	ctx, err := setContextIdentityData(ctx, adapter, channelID, userID)
	if err != nil {
		return nil, err
	}

	// Define parent log entry
	le := adapterLogEntry(ctx, nil, adapter, sp)

	// Tokenize the raw command and look up the command entry
	params := TokenizeParameters(rawCommand)
	le = le.WithField("command.name", params[0]).
		WithField("command.params", strings.Join(params[1:], " "))

	command, err := GetCommandEntry(ctx, params)
	if err != nil {
		switch {
		case gerrs.Is(err, ErrNoSuchCommand):
			msg := fmt.Sprintf("No such bundle is currently installed: %s.\n"+
				"If this is not expected, you should contact a Gort administrator.",
				params[0])
			adapter.SendErrorMessage(channelID, "No Such Command", msg)
		default:
			msg := formatCommandErrorMessage(command, params, err.Error())
			adapter.SendErrorMessage(channelID, "Error", msg)
		}

		telemetry.Errors().WithError(err).Commit(context.TODO())
		le.WithError(err).Error("Command lookup failure")

		return nil, err
	}

	// Update log entry with command info
	le = adapterLogEntry(ctx, le, command)
	le.Debug("Found matching command+bundle")

	info, ok := GetChatUser(ctx)
	if !ok {
		info, err = adapter.GetUserInfo(userID)
		if err != nil {
			return nil, err
		}
		ctx = WithChatUser(ctx, info)
	}

	gortUser, autocreated, err := findOrMakeGortUser(ctx, info)
	if err != nil {
		switch {
		case gerrs.Is(err, ErrSelfRegistrationOff):
			msg := "I'm terribly sorry, but either I don't " +
				"have a Gort account for you, or your Slack chat handle has " +
				"not been registered. Currently, only registered users can " +
				"interact with me.\n\n\nYou'll need to ask a Gort " +
				"administrator to fix this situation and to register your " +
				"Slack handle."
			adapter.SendErrorMessage(channelID, "No Such Account", msg)
		case gerrs.Is(err, ErrGortNotBootstrapped):
			msg := "Gort doesn't appear to have been bootstrapped yet! Please " +
				"use `gortctl` to properly bootstrap the Gort environment " +
				"before proceeding."
			adapter.SendErrorMessage(channelID, "Not Bootstrapped?", msg)
		default:
			msg := formatCommandErrorMessage(command, params, err.Error())
			adapter.SendErrorMessage(channelID, "Error", msg)
		}

		telemetry.Errors().WithError(err).Commit(context.TODO())
		le.WithError(err).Error("Can't find or create user")

		return nil, err
	}

	ctx = WithGortUser(ctx, gortUser)

	// Update log entry with Gort user info
	le = adapterLogEntry(ctx, le)

	if autocreated {
		message := fmt.Sprintf("Hello! It's great to meet you! You're the proud "+
			"owner of a shiny new Gort account named `%s`!",
			gortUser.Username)
		adapter.SendMessage(info.ID, message)

		le.Info("Autocreating Gort user")
	}

	request := data.CommandRequest{
		CommandEntry: command,
		ChannelID:    channelID,
		Adapter:      adapter.GetName(),
		Parameters:   params[1:],
		UserID:       userID,
	}

	// Update log entry with command info
	le = adapterLogEntry(ctx, le, command)

	le.Info("Triggering command")

	return &request, nil
}

// adapterLogEntry is a helper that pre-populates a log event
func adapterLogEntry(ctx context.Context, e *log.Entry, obs ...interface{}) *log.Entry {
	if e == nil {
		e = log.WithContext(ctx)
	}

	if ui, ok := GetChatUser(ctx); ok {
		e = e.WithField("provider.user.email", ui.Email).
			WithField("provider.user.id", ui.ID)
	}

	if ci, ok := GetChatChannel(ctx); ok {
		e = e.WithField("provider.channel.name", ci.Name).
			WithField("provider.channel.id", ci.ID)
	}

	if user, ok := GetGortUser(ctx); ok {
		e = e.WithField("gort.user.name", user.Username)
	}

	for _, i := range obs {
		switch o := i.(type) {
		case Adapter:
			e = e.WithField("adapter.name", o.GetName())

		case *ProviderEvent:
			e = e.WithField("event", o.EventType).
				WithField("adapter.name", o.Info.Provider.Name).
				WithField("adapter.type", o.Info.Provider.Type)

		case data.Bundle:
			e = e.WithField("bundle.name", o.Name).
				WithField("bundle.version", o.Version).
				WithField("bundle.default", o.Default)

		case data.BundleCommand:
			e = e.WithField("command.executable", o.Executable).
				WithField("command.name", o.Name)

		case data.CommandEntry:
			e = adapterLogEntry(ctx, e, o.Bundle, o.Command)

		case trace.Span:
			if o.SpanContext().HasTraceID() {
				e = log.WithField("id", o.SpanContext().TraceID())
			}

		case error:
			e = e.WithError(o)

		default:
			panic(fmt.Sprintf("I don't know how to log a %T", i))
		}
	}

	return e
}

func allCommandEntryFinders() ([]bundles.CommandEntryFinder, error) {
	finders := make([]bundles.CommandEntryFinder, 0)

	// Get the configuration CommandEntryFinder
	finders = append(finders, config.CommandEntryFinder())

	// Get the DAL CommandEntryFinder
	dal, err := dataaccess.Get()
	if err != nil {
		return nil, err
	}

	finders = append(finders, dal)

	return finders, nil
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
func findOrMakeGortUser(ctx context.Context, info *UserInfo) (rest.User, bool, error) {
	// Get the data access interface.
	da, err := dataaccess.Get()
	if err != nil {
		return rest.User{}, false, err
	}

	// Try to figure out what user we're working with here.
	exists := true
	user, err := da.UserGetByEmail(ctx, info.Email)
	if gerrs.Is(err, errs.ErrNoSuchUser) {
		exists = false
	} else if err != nil {
		return user, false, err
	}

	// It already exists. Exist.
	if exists {
		return user, false, nil
	}

	// We can create the user... unless the instance hasn't been bootstrapped.
	bootstrapped, err := da.UserExists(ctx, "admin")
	if err != nil {
		return rest.User{}, false, err
	}
	if !bootstrapped {
		return rest.User{}, false, ErrGortNotBootstrapped
	}

	// Now we know it doesn't exist. If self-registration is off, exit with
	// an error.
	if !config.GetGortServerConfigs().AllowSelfRegistration {
		return user, false, ErrSelfRegistrationOff
	}

	// Generate a random password for the auto-created user.
	randomPassword, err := data.GenerateRandomToken(32)
	if err != nil {
		return rest.User{}, false, err
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

	return user, true, da.UserCreate(ctx, user)
}

// TODO Replace this with something resembling a template. Eventually.
func formatCommandOutput(command data.CommandEntry, params []string, output string) string {
	return fmt.Sprintf("```%s```", output)
}

// TODO Replace this with something resembling a template. Eventually.
func formatCommandErrorMessage(command data.CommandEntry, params []string, output string) string {
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

func setContextIdentityData(ctx context.Context, adapter Adapter, channelId, userId string) (context.Context, error) {
	var err error
	var ok bool

	if _, ok = GetChatChannel(ctx); !ok {
		if channelId != "" {
			channelinfo, err := adapter.GetChannelInfo(channelId)
			if err != nil {
				return nil, gerrs.Wrap(ErrChannelNotFound, err)
			}
			ctx = WithChatChannel(ctx, channelinfo)
		}
	}

	var userinfo *UserInfo
	if userinfo, ok = GetChatUser(ctx); !ok {
		if userId != "" {
			userinfo, err = adapter.GetUserInfo(userId)
			if err != nil {
				return nil, gerrs.Wrap(ErrUserNotFound, err)
			}
			ctx = WithChatUser(ctx, userinfo)
		}
	}

	if guser, ok := GetGortUser(ctx); !ok {
		dal, err := dataaccess.Get()
		if err != nil {
			return ctx, err
		}

		guser, err = dal.UserGetByEmail(ctx, userinfo.Email)
		if err == nil {
			ctx = WithGortUser(ctx, guser)
		}
	}

	return ctx, nil
}

func startAdapters() (<-chan *ProviderEvent, chan error) {
	allEvents := make(chan *ProviderEvent)

	// TODO This isn't currently used. Use this, or remove it.
	adapterErrors := make(chan error, len(config.GetSlackProviders()))

	for k, a := range adapterLookup {
		log.WithField("adapter.name", k).Debug("Starting adapter")

		go func(adapter Adapter) {
			for event := range adapter.Listen() {
				allEvents <- event
			}
		}(a)
	}

	return allEvents, adapterErrors
}

func startProviderEventListening(commandRequests chan<- data.CommandRequest,
	allEvents <-chan *ProviderEvent, adapterErrors chan<- error) {

	tr := otel.GetTracerProvider().Tracer(ServiceName)
	ctx, sp := tr.Start(context.Background(), "Incoming Adapter Event")
	defer sp.End()

	for event := range allEvents {
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
			formatted := formatCommandErrorMessage(
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
