package relay

import (
	"fmt"
	"log"
	"strings"

	"github.com/clockworksoul/cog2/config"
	"github.com/clockworksoul/cog2/context"
	"github.com/clockworksoul/cog2/worker"
)

type TriggerCommandError struct {
	command          config.CommandEntry
	params           []string
	title            string // Command Error
	shortDescription string // The pipeline failed planning the invocation:
	longDescription  string // <command output>
}

func (e TriggerCommandError) Error() string {
	return e.shortDescription
}

func (e TriggerCommandError) RawCommand() string {
	return fmt.Sprintf(
		"%s:%s %s",
		e.command.Bundle.Name,
		e.command.Command.Name,
		strings.Join(e.params, " "))
}

// Relay represents a single relay. Right now it's pretty different from what
// Cog calls a relay, but that'll probably change.
type Relay interface {
	GetBotUser() *UserInfo
	GetChannelInfo(channelID string) (*ChannelInfo, error)
	GetPresentChannels(userID string) ([]*ChannelInfo, error)
	GetUserInfo(userID string) (*UserInfo, error)
	Listen() <-chan *ProviderEvent
	SendErrorMessage(channelID string, title string, text string)
	SendMessage(channel string, message string)
}

// GetCommandEntry accepts a tokenized parameter slice and locates the
// associated config.CommandEntry. If the number of matching commands is != 1,
// an error is returned.
func GetCommandEntry(tokens []string) (config.CommandEntry, error) {
	entries, err := config.FindCommandEntry(tokens[0])
	if err != nil {
		return config.CommandEntry{}, err
	}

	if len(entries) == 0 {
		return config.CommandEntry{}, fmt.Errorf("No such bundle:command: %s", tokens[0])
	}

	if len(entries) > 1 {
		return config.CommandEntry{},
			fmt.Errorf("Multiple commands found: %s:%s vs %s:%s",
				entries[0].Bundle.Name, entries[0].Command.Name,
				entries[1].Bundle.Name, entries[1].Command.Name,
			)
	}

	return entries[0], nil
}

// TokenizeParameters splits a command string into parameter tokens. The
// trigger character, if any, is trimmed. Quotes aren't yet respected, but
// they will be eventually.
func TokenizeParameters(commandString string) []string {
	if commandString[0] == '!' {
		commandString = commandString[1:]
	}

	return strings.Split(commandString, " ")
}

// SpawnWorker receives a CommandEntry and a slice of command parameters
// strings, and constructs a new worker.Worker.
func SpawnWorker(command config.CommandEntry, cmdParameters []string) (*worker.Worker, error) {
	image := command.Bundle.Docker.Image
	tag := command.Bundle.Docker.Tag
	entrypoint := command.Command.Executable

	return worker.NewWorker(image, tag, entrypoint, cmdParameters...)
}

// StartListening instructs all relays to establish connections, receives all
// events from all relays, and forwards them to the various On* handler functions.
func StartListening() {
	allEvents := make(chan *ProviderEvent)
	relays := make(map[string]bool)

	for _, sp := range config.GetSlackProviders() {
		if relays[sp.Name] {
			log.Panicf("Relay name collision: %s\n", sp.Name)
		}

		relay := NewSlackRelay(sp)
		relays[sp.Name] = true

		go func() {
			for event := range relay.Listen() {
				allEvents <- event
			}
		}()
	}

	go func() {
		for event := range allEvents {
			switch ev := event.Data.(type) {
			case *ConnectedEvent:
				OnConnected(event, ev)

			case *ChannelMessageEvent:
				OnChannelMessage(event, ev)

			case *DirectMessageEvent:
				OnDirectMessage(event, ev)

			default:
				log.Fatalf("Unknown data type: %T\n", ev)
			}
		}
	}()
}

// OnConnected handles ConnectedEvent events.
func OnConnected(event *ProviderEvent, data *ConnectedEvent) {
	log.Printf(
		"Connection established to %s provider %s. I am @%s!\n",
		event.Info.Provider.Type,
		event.Info.Provider.Name,
		event.Info.User.Name,
	)

	channels, err := event.Relay.GetPresentChannels(event.Info.User.ID)
	if err != nil {
		log.Printf("Failed to get channels list for %s: %s", event.Info.Provider.Name, err.Error())
		return
	}

	for _, c := range channels {
		message := fmt.Sprintf("Cog2 version %s is online. Hello, %s!", context.CogVersion, c.Name)
		event.Relay.SendMessage(c.ID, message)
	}
}

// OnChannelMessage handles ChannelMessageEvent events.
func OnChannelMessage(event *ProviderEvent, data *ChannelMessageEvent) {
	channel, err := event.Relay.GetChannelInfo(data.Channel)
	if err != nil {
		log.Printf("Could not find channel: " + err.Error())
		return
	}

	userinfo, err := event.Relay.GetUserInfo(data.User)
	if err != nil {
		log.Printf("Could not find user: " + err.Error())
		return
	}

	text := strings.TrimSpace(data.Text)

	log.Printf("Message from @%s in %s: %s\n",
		userinfo.DisplayNameNormalized,
		channel.Name,
		text,
	)

	if len(text) < 2 || text[0] != '!' {
		return
	}

	// Remove the "trigger character" (!)
	text = text[1:]

	err = TriggerCommand(text, event.Relay, data.Channel)
	if err != nil {
		log.Printf(err.Error())
	}
}

// OnDirectMessage handles DirectMessageEvent events.
func OnDirectMessage(event *ProviderEvent, data *DirectMessageEvent) {
	userinfo, err := event.Relay.GetUserInfo(data.User)
	if err != nil {
		log.Printf("Could not find user: " + err.Error())
		return
	}

	log.Printf("Direct message from @%s: %s\n",
		userinfo.DisplayNameNormalized,
		data.Text,
	)

	text := strings.TrimSpace(data.Text)

	if len(text) < 1 {
		return
	}

	err = TriggerCommand(text, event.Relay, data.User)
	if err != nil {
		log.Printf(err.Error())
	}
}

// TriggerCommand is called by OnChannelMessage or OnDirectMessage when a
// valid command trigger is identified.
func TriggerCommand(rawCommand string, relay Relay, channelID string) error {
	params := TokenizeParameters(rawCommand)

	command, err := GetCommandEntry(params)
	if err != nil {
		return err
	}

	log.Printf("Found matching command: %s:%s\n", command.Bundle.Name, command.Command.Name)

	worker, err := SpawnWorker(command, params[1:])
	if err != nil {
		return fmt.Errorf("Error spawning %s: %s", command.Bundle.Docker.Image, err.Error())
	}

	output, err := worker.Start()
	if err != nil {
		return fmt.Errorf("Error starting worker %s: %s", command.Bundle.Docker.Image, err.Error())
	}

	go func() {
		outputText := ""

		for str := range output {
			outputText += fmt.Sprintf("%s\n", str)
		}

		status := <-worker.ExitStatus
		log.Printf("Worker exited with status %d\n", status)

		// TODO THIS IS NOT GENERALIZED! Replace Slack-specific ``` with templates!!! Eventually.
		if status == 0 {
			relay.SendMessage(channelID, "```"+outputText+"```")
		} else {
			relay.SendErrorMessage(
				channelID,
				"Command Error",
				generateCommandErrorMessage(command, params[1:], outputText))
		}
	}()

	return nil
}

func generateCommandErrorMessage(command config.CommandEntry, params []string, output string) string {
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
