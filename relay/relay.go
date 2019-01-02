package relay

import (
	"fmt"
	"log"
	"strings"

	"github.com/clockworksoul/cog2/config"
	"github.com/clockworksoul/cog2/context"
	"github.com/clockworksoul/cog2/worker"
)

type Relay interface {
	GetBotUser() *UserInfo
	GetChannelInfo(channelID string) (*ChannelInfo, error)
	GetPresentChannels(userID string) ([]*ChannelInfo, error)
	GetUserInfo(userID string) (*UserInfo, error)
	Listen() <-chan *ProviderEvent
	SendMessage(channel string, message string)
}

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
				entries[0].Bundle.Name, entries[0].Command.Command,
				entries[1].Bundle.Name, entries[1].Command.Command,
			)
	}

	return entries[0], nil
}

func TokenizeParameters(commandString string) []string {
	if commandString[0] == '!' {
		commandString = commandString[1:]
	}

	return strings.Split(commandString, " ")
}

func SpawnWorker(command config.CommandEntry, cmdParameters []string) (<-chan string, error) {
	image := command.Bundle.Docker.Image
	tag := command.Bundle.Docker.Tag
	entrypoint := command.Command.Executable

	worker, err := worker.NewWorker(image, tag, entrypoint, cmdParameters...)
	if err != nil {
		return nil, err
	}

	return worker.Start()
}

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

func OnChannelMessage(event *ProviderEvent, data *ChannelMessageEvent) {
	text := strings.TrimSpace(data.Text)

	if len(text) < 2 || text[0] != '!' {
		return
	}

	// Remove the "trigger character" (!)
	text = text[1:]

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

	log.Printf("Message from @%s in %s: %s\n",
		userinfo.DisplayNameNormalized,
		channel.Name,
		text,
	)

	err = TriggerCommand(text, event.Relay, data.Channel)
	if err != nil {
		log.Printf(err.Error())
	}
}

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
}

func TriggerCommand(text string, relay Relay, channelID string) error {
	params := TokenizeParameters(text)

	command, err := GetCommandEntry(params)
	if err != nil {
		return err
	}

	log.Printf("Found matching command: %s:%s\n", command.Bundle.Name, command.Command.Command)

	output, err := SpawnWorker(command, params[1:])
	if err != nil {
		return fmt.Errorf("Error spawning %s: %s\n", command.Bundle.Docker.Image, err.Error())
	}

	go func() {
		for str := range output {
			relay.SendMessage(channelID, str)
		}
	}()

	return nil
}
