package adapter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/clockworksoul/cog2/config"
	"github.com/clockworksoul/cog2/data"
	"github.com/clockworksoul/cog2/data/rest"
	"github.com/clockworksoul/cog2/dataaccess"
	cogerr "github.com/clockworksoul/cog2/errors"
	"github.com/clockworksoul/cog2/meta"
	log "github.com/sirupsen/logrus"
)

var (
	// All existant adapters keyed by name
	adapterLookup map[string]Adapter
)

var (
	// ErrAdapterNameCollision is emitted by startAdapters() if two adapters
	// have the same name.
	ErrAdapterNameCollision = errors.New("adapter name collision")

	// ErrAuthenticationFailure is emitted when an AuthenticationErrorEvent
	// is received.
	ErrAuthenticationFailure = errors.New("authentication failure")

	// ErrChannelNotFound is returned when OnChannelMessage can't find
	// information is the originating channel.
	ErrChannelNotFound = errors.New("channel not found")

	// ErrCogNotBootstrapped is returned by findOrMakeCogUser() if a user
	// attempts to trigger a command but Cog hasn't yet been bopotstrapped.
	ErrCogNotBootstrapped = errors.New("Cog hasn't been bootstrapped yet")

	// ErrSelfRegistrationOff is returned by findOrMakeCogUser() if an unknown
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
	// GetBotUser returns the info for the user associated with this bot in
	// its respective provider.
	GetBotUser() *UserInfo

	// GetChannelInfo provides info on a specific provider channel accessible
	// to the adapter.
	GetChannelInfo(channelID string) (*ChannelInfo, error)

	// GetName provides the name of this adapter as per the configuration.
	GetName() string

	// GetPresentChannels returns a slice of channels that a user is present in.
	GetPresentChannels(userID string) ([]*ChannelInfo, error)

	// GetUserInfo provides info on a specific provider channel accessible
	// to the adapter.
	GetUserInfo(userID string) (*UserInfo, error)

	// Listen causes the Adapter to ititiate a connection to its provider and
	// begin relaying back events via the returned channel.
	Listen() <-chan *ProviderEvent

	// SendErrorMessage sends an error message to a specified channel.
	// TODO Create a MessageBuilder at some point to replace this.
	SendErrorMessage(channelID string, title string, text string) error

	// SendMessage sends a standard output message to a specified channel.
	// TODO Create a MessageBuilder at some point to replace this.
	SendMessage(channel string, message string) error
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
func GetCommandEntry(tokens []string) (data.CommandEntry, error) {
	entries, err := config.FindCommandEntry(tokens[0])
	if err != nil {
		return data.CommandEntry{}, err
	}

	if len(entries) == 0 {
		return data.CommandEntry{}, ErrNoSuchCommand
	}

	if len(entries) > 1 {
		log.Warnf("Multiple commands found: %s:%s vs %s:%s",
			entries[0].Bundle.Name, entries[0].Command.Name,
			entries[1].Bundle.Name, entries[1].Command.Name,
		)

		return data.CommandEntry{}, ErrMultipleCommands
	}

	return entries[0], nil
}

// OnConnected handles ConnectedEvent events.
func OnConnected(event *ProviderEvent, data *ConnectedEvent) {
	log.Infof(
		"[OnConnected] Connection established to %s provider %s. I am @%s!",
		event.Info.Provider.Type,
		event.Info.Provider.Name,
		event.Info.User.Name,
	)

	channels, err := event.Adapter.GetPresentChannels(event.Info.User.ID)
	if err != nil {
		log.Errorf("[OnConnected] Failed to get channels list for %s: %s", event.Info.Provider.Name, err.Error())
		return
	}

	for _, c := range channels {
		message := fmt.Sprintf("Cog2 version %s is online. Hello, %s!", meta.CogVersion, c.Name)
		event.Adapter.SendMessage(c.ID, message)
	}
}

// OnChannelMessage handles ChannelMessageEvent events.
// If a command is found in the text, it will emit a data.CommandRequest
// instance to the commands channel.
// TODO Support direct in-channel mentions.
func OnChannelMessage(event *ProviderEvent, data *ChannelMessageEvent) (*data.CommandRequest, error) {
	channelinfo, err := event.Adapter.GetChannelInfo(data.ChannelID)
	if err != nil {
		return nil, cogerr.Wrap(ErrChannelNotFound, err)
	}

	userinfo, err := event.Adapter.GetUserInfo(data.UserID)
	if err != nil {
		return nil, cogerr.Wrap(ErrUserNotFound, err)
	}

	rawCommandText := data.Text

	log.Debugf("[OnChannelMessage] Message from @%s in %s: %s",
		userinfo.DisplayNameNormalized,
		channelinfo.Name,
		rawCommandText,
	)

	// If this isn't prepended by a trigger character, ignore.
	if len(rawCommandText) <= 1 || rawCommandText[0] != '!' {
		return nil, nil
	}

	// If this starts with a trigger character but enable_spoken_commands is false, ignore.
	if rawCommandText[0] == '!' && !config.GetCogServerConfigs().EnableSpokenCommands {
		return nil, nil
	}

	// Remove the "trigger character" (!)
	rawCommandText = rawCommandText[1:]

	return TriggerCommand(rawCommandText, event.Adapter, data.ChannelID, data.UserID)
}

// OnDirectMessage handles DirectMessageEvent events.
func OnDirectMessage(event *ProviderEvent, data *DirectMessageEvent) (*data.CommandRequest, error) {
	userinfo, err := event.Adapter.GetUserInfo(data.UserID)
	if err != nil {
		return nil, cogerr.Wrap(ErrUserNotFound, err)
	}

	rawCommandText := data.Text

	log.Debugf("[OnDirectMessage] Direct message from @%s: %s",
		userinfo.DisplayNameNormalized,
		data.Text,
	)

	if rawCommandText[0] == '!' {
		rawCommandText = rawCommandText[1:]
	}

	return TriggerCommand(rawCommandText, event.Adapter, data.ChannelID, data.UserID)
}

// StartListening instructs all relays to establish connections, receives all
// events from all relays, and forwards them to the various On* handler functions.
func StartListening() (<-chan data.CommandRequest, chan<- data.CommandResponse, <-chan error) {
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
func TriggerCommand(rawCommand string, adapter Adapter, channelID string, userID string) (*data.CommandRequest, error) {
	params := TokenizeParameters(rawCommand)

	info, err := adapter.GetUserInfo(userID)
	if err != nil {
		return nil, err
	}

	command, err := GetCommandEntry(params)
	if err != nil {
		switch {
		case cogerr.ErrEquals(err, ErrNoSuchCommand):
			msg := fmt.Sprintf("No such bundle is currently installed: %s.\n"+
				"If this is not expected, you should contact a Cog administrator.",
				params[0])
			adapter.SendErrorMessage(channelID, "No Such Command", msg)
		default:
			msg := formatCommandErrorMessage(command, params, err.Error())
			adapter.SendErrorMessage(channelID, "Error", msg)
		}

		return nil, err
	}

	log.Debugf("[TriggerCommand] Found matching command: %s:%s",
		command.Bundle.Name, command.Command.Name)

	user, autocreated, err := findOrMakeCogUser(info)
	if err != nil {
		switch {
		case cogerr.ErrEquals(err, ErrSelfRegistrationOff):
			msg := "I'm terribly sorry, but either I don't " +
				"have a Cog account for you, or your Slack chat handle has " +
				"not been registered. Currently, only registered users can " +
				"interact with me.\n\n\nYou'll need to ask a Cog " +
				"administrator to fix this situation and to register your " +
				"Slack handle."
			adapter.SendErrorMessage(channelID, "No Such Account", msg)
		case cogerr.ErrEquals(err, ErrCogNotBootstrapped):
			msg := "Cog doesn't appear to have been bootstrapped yet! Please " +
				"use `cogctl` to properly bootstrap Cog environment before " +
				"proceeding."
			adapter.SendErrorMessage(channelID, "Not Bootstrapped?", msg)
		default:
			msg := formatCommandErrorMessage(command, params, err.Error())
			adapter.SendErrorMessage(channelID, "Error", msg)
		}

		return nil, err
	} else if autocreated {
		message := fmt.Sprintf("Hello! It's great to meet you! You're the proud "+
			"owner of a shiny new Cog account named `%s`!",
			user.Username)
		adapter.SendMessage(info.ID, message)
	}

	request := data.CommandRequest{
		CommandEntry: command,
		ChannelID:    channelID,
		Adapter:      adapter.GetName(),
		Parameters:   params[1:],
		UserID:       userID,
	}

	return &request, nil
}

// findOrMakeCogUser ...
func findOrMakeCogUser(info *UserInfo) (rest.User, bool, error) {
	// Get the data access interface.
	da, err := dataaccess.Get()
	if err != nil {
		return rest.User{}, false, err
	}

	// Try to figure out what user we're working with here.
	exists := true
	user, err := da.UserGetByEmail(info.Email)
	if err != nil {
		exists = false
	}

	// It already exists. Exist.
	if exists {
		return user, false, nil
	}

	// Now we know it doesn't exist. If self-registration is off, exit with
	// an error.
	if !config.GetCogServerConfigs().AllowSelfRegistration {
		return user, false, ErrSelfRegistrationOff
	}

	// We can create the user... unless the instance hasn't been bootstrapped.
	bootstrapped, err := da.UserExists("admin")
	if err != nil {
		return rest.User{}, false, err
	}
	if !bootstrapped {
		return rest.User{}, false, ErrCogNotBootstrapped
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

	log.Infof("[findOrMakeCogUser] User auto-created: %s (%s)", user.Username, user.Email)

	return user, true, da.UserCreate(user)
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

func startAdapters() (<-chan *ProviderEvent, chan error) {
	adapterLookup = make(map[string]Adapter)

	allEvents := make(chan *ProviderEvent)
	adapterErrors := make(chan error, len(config.GetSlackProviders()))

	for _, sp := range config.GetSlackProviders() {
		if _, ok := adapterLookup[sp.Name]; ok {
			adapterErrors <- ErrAdapterNameCollision
			continue
		}

		adapter := NewSlackAdapter(sp)
		adapterLookup[sp.Name] = adapter

		go func() {
			for event := range adapter.Listen() {
				allEvents <- event
			}
		}()
	}

	return allEvents, adapterErrors
}

func startProviderEventListening(commandRequests chan<- data.CommandRequest,
	allEvents <-chan *ProviderEvent, adapterErrors chan<- error) {

	for event := range allEvents {
		switch ev := event.Data.(type) {
		case *ConnectedEvent:
			OnConnected(event, ev)

		case *AuthenticationErrorEvent:
			adapterErrors <- cogerr.Wrap(ErrAuthenticationFailure, errors.New(ev.Msg))

		case *ChannelMessageEvent:
			request, err := OnChannelMessage(event, ev)
			if request != nil {
				commandRequests <- *request
			}
			if err != nil {
				adapterErrors <- err
			}

		case *DirectMessageEvent:
			request, err := OnDirectMessage(event, ev)
			if request != nil {
				commandRequests <- *request
			}
			if err != nil {
				adapterErrors <- err
			}

		case *ErrorEvent:
			adapterErrors <- ev

		default:
			log.Fatalf("[startProviderEventListening] Unknown data type: %T", ev)
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
