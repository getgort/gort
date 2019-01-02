package relay

import (
	"fmt"
	"log"
	"regexp"

	"github.com/clockworksoul/cog2/config"
	"github.com/nlopes/slack"
)

var (
	linkMarkdownRegex = regexp.MustCompile("<([a-zA-Z0-9]*://[a-zA-Z0-9]*\\.[a-zA-Z0-9]*)\\|([a-zA-Z0-9]*\\.[a-zA-Z0-9]*)>")
)

type SlackRelay struct {
	Relay

	client   *slack.Client
	provider config.SlackProvider
	rtm      *slack.RTM
}

func NewSlackRelay(provider config.SlackProvider) SlackRelay {
	client := slack.New(provider.SlackAPIToken)
	rtm := client.NewRTM()

	return SlackRelay{
		client:   client,
		provider: provider,
		rtm:      rtm,
	}
}

func (s SlackRelay) GetChannelInfo(channelID string) (*ChannelInfo, error) {
	ch, err := s.rtm.GetChannelInfo(channelID)
	if err != nil {
		return nil, err
	}

	return newChannelInfoFromSlackChannel(ch), nil
}

func (s SlackRelay) GetUserInfo(userID string) (*UserInfo, error) {
	u, err := s.rtm.GetUserInfo(userID)
	if err != nil {
		return nil, err
	}

	return newUserInfoFromSlackUser(u), nil
}

// Channels returns a slice of channel ID strings that the Relay is present in.
// This is expensive. Don't use it often.
func (s SlackRelay) GetPresentChannels(userID string) ([]*ChannelInfo, error) {
	allChannels, err := s.rtm.GetChannels(true)
	if err != nil {
		return nil, err
	}

	channels := make([]*ChannelInfo, 0)

	// A nested loop. It's terrible. It's hacky. I know.
	for _, ch := range allChannels {
		members := ch.Members

	inner:
		for _, memberID := range members {
			if userID == memberID {
				channels = append(
					channels,
					newChannelInfoFromSlackChannel(&ch),
				)
				break inner
			}
		}
	}

	return channels, nil
}

func (s SlackRelay) Listen() <-chan *ProviderEvent {
	events := make(chan *ProviderEvent)

	log.Printf("Connecting to Slack provider %s...\n", s.provider.Name)

	go s.rtm.ManageConnection()

	go func() {
		info := &Info{
			Provider: NewProviderInfoFromConfig(s.provider),
			User:     &UserInfo{},
		}

	eventLoop:
		for msg := range s.rtm.IncomingEvents {
			switch ev := msg.Data.(type) {
			case *slack.ConnectedEvent:
				suser, err := s.rtm.GetUserInfo(ev.Info.User.ID)
				if err != nil {
					log.Printf("Error finding user %s on connect: %s\n",
						ev.Info.User.ID,
						err.Error())
					continue eventLoop
				}

				info.User.setFromSlackUser(suser)

				events <- s.OnConnected(ev, info)

			case *slack.MessageEvent:
				providerEvent := s.OnMessage(ev, info)
				if providerEvent.EventType != "" {
					events <- providerEvent
				}

			case *slack.RTMError:
				events <- s.OnError(ev, info)

			case *slack.InvalidAuthEvent:
				events <- s.OnAuthenticationError(ev, info)
				break eventLoop

			default:
				// Ignore other events..
			}
		}

		close(events)
	}()

	return events
}

func (s *SlackRelay) OnAuthenticationError(event *slack.InvalidAuthEvent, info *Info) *ProviderEvent {
	return s.wrapEvent(
		"authentication_error",
		info,
		&AuthenticationErrorEvent{
			Msg: fmt.Sprintf("Connection failed to %s: invalid credentials", s.provider.Name),
		},
	)
}

func (s *SlackRelay) OnConnected(event *slack.ConnectedEvent, info *Info) *ProviderEvent {
	return s.wrapEvent(
		"connected",
		info,
		&ConnectedEvent{},
	)
}

func (s *SlackRelay) OnError(event *slack.RTMError, info *Info) *ProviderEvent {
	return s.wrapEvent(
		"error",
		info,
		&ErrorEvent{
			Code: event.Code,
			Msg:  event.Msg,
		},
	)
}

func (s *SlackRelay) OnChannelMessage(event *slack.MessageEvent, info *Info) *ProviderEvent {
	return s.wrapEvent(
		"channel_message",
		info,
		&ChannelMessageEvent{
			Channel: event.Channel,
			Text:    ScrubMarkdown(event.Msg.Text),
			User:    event.Msg.User,
		},
	)
}

func (s *SlackRelay) OnDirectMessage(event *slack.MessageEvent, info *Info) *ProviderEvent {
	return s.wrapEvent(
		"direct_message",
		info,
		&DirectMessageEvent{
			Text: ScrubMarkdown(event.Msg.Text),
			User: event.Msg.User,
		},
	)
}

func (s *SlackRelay) OnMessage(event *slack.MessageEvent, info *Info) *ProviderEvent {
	switch event.Msg.SubType {
	case "": // Just a plain message. Handle accordingly.
		if event.Channel[0] == 'D' {
			return s.OnDirectMessage(event, info)
		} else {
			return s.OnChannelMessage(event, info)
		}
	case "message_changed":
		// Note here for later; ignore for now.
		return &ProviderEvent{}
	case "message_deleted":
		// Note here for later; ignore for now.
		return &ProviderEvent{}
	case "bot_message":
		// Note here for later; ignore for now.
		return &ProviderEvent{}
	default:
		log.Printf("Received unknown submessage type (%s)", event.Msg.SubType)
		return &ProviderEvent{}
	}
}

func (s SlackRelay) SendMessage(channel string, message string) {
	s.rtm.PostMessage(
		channel,
		slack.MsgOptionDisableMarkdown(),
		slack.MsgOptionAsUser(false),
		slack.MsgOptionUsername(s.provider.BotName),
		slack.MsgOptionText(message, true),
		slack.MsgOptionPostMessageParameters(slack.PostMessageParameters{
			IconURL: s.provider.IconURL,
		}),
	)
}

// Wrap event creates a new ProviderEvent instance with metadata and the Event data attached.
func (s *SlackRelay) wrapEvent(eventType string, info *Info, data interface{}) *ProviderEvent {
	return &ProviderEvent{
		EventType: eventType,
		Data:      data,
		Info:      info,
		Relay:     s,
	}
}

func ScrubMarkdown(text string) string {
	indices := linkMarkdownRegex.FindAllSubmatchIndex([]byte(text), -1)
	last := 0

	if len(indices) > 0 {
		newStr := ""

		for _, z := range indices {
			newStr += text[last:z[0]]
			newStr += text[z[4]:z[5]]
			last = z[5] + 1
		}

		newStr += text[indices[len(indices)-1][5]+1:]
		text = newStr
	}

	return text
}
