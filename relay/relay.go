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

func FindImage(commandString string) (string, error) {
	return "clockworksoul/echotest", nil
}

func TokenizeParameters(commandString string) []string {
	return strings.Split(commandString, " ")
}

func SpawnWorker(imageName string, cmdParameters []string) (<-chan string, error) {
	worker, err := worker.NewWorker(imageName, cmdParameters...)
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
			log.Printf(
				"provider=%s(%s) event=%s data=%v (type %T)\n",
				event.Info.Provider.Name,
				event.Info.Provider.Type,
				event.EventType,
				event.Data,
				event.Data,
			)

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
		data.Text,
	)

	go func() {
		params := TokenizeParameters(data.Text)
		image, _ := FindImage(data.Text)
		output, _ := SpawnWorker(image, params)

		for str := range output {
			event.Relay.SendMessage(channel.ID, str)
		}
	}()
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
