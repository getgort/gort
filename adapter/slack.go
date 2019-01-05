package adapter

import (
	"fmt"
	"regexp"

	"github.com/clockworksoul/cog2/data"
	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
)

var (
	linkMarkdownRegexShort = regexp.MustCompile("<([a-zA-Z0-9]*://[a-zA-Z0-9\\.]*)>")
	linkMarkdownRegexLong  = regexp.MustCompile("<([a-zA-Z0-9]*://[a-zA-Z0-9\\.]*)\\|([a-zA-Z0-9\\.]*)>")
)

// SlackAdapter is the Slack provider implementation of a relay, which knows how
// to receive events from the Slack API, translate them into Cog2 events, and
// forward them along.
type SlackAdapter struct {
	Adapter

	client   *slack.Client
	provider data.SlackProvider
	rtm      *slack.RTM
}

// NewSlackAdapter will construct a SlackAdapter instance for a given provider configuration.
func NewSlackAdapter(provider data.SlackProvider) SlackAdapter {
	client := slack.New(provider.SlackAPIToken)
	rtm := client.NewRTM()

	return SlackAdapter{
		client:   client,
		provider: provider,
		rtm:      rtm,
	}
}

// GetChannelInfo returns the ChannelInfo for a requested channel.
func (s SlackAdapter) GetChannelInfo(channelID string) (*ChannelInfo, error) {
	ch, err := s.rtm.GetChannelInfo(channelID)
	if err != nil {
		return nil, err
	}

	return newChannelInfoFromSlackChannel(ch), nil
}

// GetName returns this adapter's configured name
func (s SlackAdapter) GetName() string {
	return s.provider.Name
}

// GetPresentChannels returns a slice of channel ID strings that the Adapter
// is present in. This is expensive. Don't use it often.
func (s SlackAdapter) GetPresentChannels(userID string) ([]*ChannelInfo, error) {
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

// GetUserInfo returns the UserInfo for a requested user.
func (s SlackAdapter) GetUserInfo(userID string) (*UserInfo, error) {
	u, err := s.rtm.GetUserInfo(userID)
	if err != nil {
		return nil, err
	}

	return newUserInfoFromSlackUser(u), nil
}

// Listen instructs the relay to begin listening to the provider that it's attached to.
// It exits immediately, returning a channel that emits ProviderEvents.
func (s SlackAdapter) Listen() <-chan *ProviderEvent {
	events := make(chan *ProviderEvent)

	log.Infof("[SlackAdapter.Listen] Connecting to Slack provider %s...", s.provider.Name)

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
					log.Errorf("[SlackAdapter.Listen] Error finding user %s on connect: %s",
						ev.Info.User.ID,
						err.Error())

					continue eventLoop
				}

				info.User.setFromSlackUser(suser)

				events <- s.OnConnected(ev, info)

			case *slack.InvalidAuthEvent:
				events <- s.OnInvalidAuth(ev, info)
				break eventLoop

			case *slack.LatencyReport:
				s.OnLatencyReport(ev, info)

			case *slack.MessageEvent:
				providerEvent := s.OnMessage(ev, info)
				if providerEvent.EventType != "" {
					events <- providerEvent
				}

			case *slack.RTMError:
				events <- s.OnError(ev, info)

			default:
				// Ignore other events..
				log.Debugf("[SlackAdapter.Listen] Received (and ignored) event type %T", ev)
			}
		}

		close(events)
	}()

	return events
}

// OnLatencyReport is called when the Slack API emits a LatencyReport.
func (s *SlackAdapter) OnLatencyReport(event *slack.LatencyReport, info *Info) *ProviderEvent {
	millis := event.Value.Nanoseconds() / 1000000
	template := "[SlackAdapter.OnLatencyReport] High latency detected: %s"

	switch {
	case millis > 500 && millis < 1000:
		log.Debugf(template, event.Value.String())
	case millis > 1000 && millis < 1500:
		log.Infof(template, event.Value.String())
	case millis > 1500:
		log.Warnf(template, event.Value.String())
	}

	return nil
}

// OnInvalidAuth is called when the Slack API emits an InvalidAuthEvent.
func (s *SlackAdapter) OnInvalidAuth(event *slack.InvalidAuthEvent, info *Info) *ProviderEvent {
	return s.wrapEvent(
		"authentication_error",
		info,
		&AuthenticationErrorEvent{
			Msg: fmt.Sprintf("Connection failed to %s: invalid credentials", s.provider.Name),
		},
	)
}

// OnConnected is called when the Slack API emits a ConnectedEvent.
func (s *SlackAdapter) OnConnected(event *slack.ConnectedEvent, info *Info) *ProviderEvent {
	return s.wrapEvent(
		"connected",
		info,
		&ConnectedEvent{},
	)
}

// OnError is called when the Slack API emits an RTMError.
func (s *SlackAdapter) OnError(event *slack.RTMError, info *Info) *ProviderEvent {
	return s.wrapEvent(
		"error",
		info,
		&ErrorEvent{
			Code: event.Code,
			Msg:  event.Msg,
		},
	)
}

// OnChannelMessage is called when the Slack API emits an MessageEvent for a message in a channel.
func (s *SlackAdapter) OnChannelMessage(event *slack.MessageEvent, info *Info) *ProviderEvent {
	return s.wrapEvent(
		"channel_message",
		info,
		&ChannelMessageEvent{
			ChannelID: event.Channel,
			Text:      ScrubMarkdown(event.Msg.Text),
			UserID:    event.Msg.User,
		},
	)
}

// OnDirectMessage is called when the Slack API emits an MessageEvent for a direct message.
func (s *SlackAdapter) OnDirectMessage(event *slack.MessageEvent, info *Info) *ProviderEvent {
	return s.wrapEvent(
		"direct_message",
		info,
		&DirectMessageEvent{
			ChannelID: event.Channel,
			Text:      ScrubMarkdown(event.Msg.Text),
			UserID:    event.Msg.User,
		},
	)
}

// OnMessage is called when the Slack API emits an InvalidAuthEvent.
func (s *SlackAdapter) OnMessage(event *slack.MessageEvent, info *Info) *ProviderEvent {
	switch event.Msg.SubType {
	case "": // Just a plain message. Handle accordingly.
		if event.Channel[0] == 'D' {
			return s.OnDirectMessage(event, info)
		}

		return s.OnChannelMessage(event, info)
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
		log.Warnf("[SlackAdapter.OnMessage] Received unknown submessage type (%s)", event.Msg.SubType)
		return &ProviderEvent{}
	}
}

// SendMessage will send a message (from the bot) into the specified channel.
func (s SlackAdapter) SendMessage(channelID string, text string) error {
	_, _, err := s.rtm.PostMessage(
		channelID,
		slack.MsgOptionDisableMediaUnfurl(),
		slack.MsgOptionAsUser(false),
		slack.MsgOptionUsername(s.provider.BotName),
		slack.MsgOptionText(text, false),
		slack.MsgOptionPostMessageParameters(slack.PostMessageParameters{
			IconURL:  s.provider.IconURL,
			Markdown: true,
		}),
	)

	return err
}

// SendErrorMessage will send a message (from the bot) into the specified channel.
func (s SlackAdapter) SendErrorMessage(channelID string, title string, text string) error {
	_, _, err := s.rtm.PostMessage(
		channelID,
		slack.MsgOptionAttachments(
			slack.Attachment{
				Title:      title,
				Text:       text,
				Color:      "#FF0000",
				MarkdownIn: []string{"text"},
			},
		),
		slack.MsgOptionDisableMediaUnfurl(),
		slack.MsgOptionDisableMarkdown(),
		slack.MsgOptionAsUser(false),
		slack.MsgOptionUsername(s.provider.BotName),
		slack.MsgOptionPostMessageParameters(slack.PostMessageParameters{
			IconURL:  s.provider.IconURL,
			Markdown: true,
		}),
	)

	return err
}

// wrapEvent creates a new ProviderEvent instance with metadata and the Event data attached.
func (s *SlackAdapter) wrapEvent(eventType string, info *Info, data interface{}) *ProviderEvent {
	return &ProviderEvent{
		EventType: eventType,
		Data:      data,
		Info:      info,
		Adapter:   s,
	}
}

// ScrubMarkdown removes unnecessary/undesirable Slack markdown (of links, of
// example) from text recieved from Slack.
func ScrubMarkdown(text string) string {
	var indices [][]int
	var last int

	// Remove links of the format "<https://google.com>"
	//
	indices = linkMarkdownRegexShort.FindAllSubmatchIndex([]byte(text), -1)
	last = 0

	if len(indices) > 0 {
		newStr := ""

		for _, z := range indices {
			newStr += text[last:z[0]]
			newStr += text[z[0]+1 : z[1]-1]
			last = z[1]
		}

		newStr += text[indices[len(indices)-1][1]:]
		text = newStr
	}

	// Remove links of the format "<http://google.com|google.com>"
	//
	indices = linkMarkdownRegexLong.FindAllSubmatchIndex([]byte(text), -1)
	last = 0

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
