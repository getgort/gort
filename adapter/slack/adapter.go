package slack

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/clockworksoul/gort/adapter"
	"github.com/clockworksoul/gort/data"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

var (
	linkMarkdownRegexShort = regexp.MustCompile(`\<([^|:]*://[^:|]*)\>`)
	linkMarkdownRegexLong  = regexp.MustCompile(`\<[^|:]*://[^:|]*\|([^:|]*)\>`)
)

// SlackAdapter is the Slack provider implementation of a relay, which knows how
// to receive events from the Slack API, translate them into Gort events, and
// forward them along.
type SlackAdapter struct {
	client   *slack.Client
	provider data.SlackProvider
	rtm      *slack.RTM
}

// GetChannelInfo returns the ChannelInfo for a requested channel.
func (s SlackAdapter) GetChannelInfo(channelID string) (*adapter.ChannelInfo, error) {
	ch, err := s.rtm.GetConversationInfo(channelID, false)
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
func (s SlackAdapter) GetPresentChannels(userID string) ([]*adapter.ChannelInfo, error) {
	allChannels, _, err := s.rtm.GetConversations(&slack.GetConversationsParameters{})
	if err != nil {
		return nil, err
	}

	channels := make([]*adapter.ChannelInfo, 0)

	// A nested loop. It's terrible. It's hacky. I know.
	for _, ch := range allChannels {
		members := ch.Members

	inner:
		for _, memberID := range members {
			if userID == memberID {
				channels = append(channels, newChannelInfoFromSlackChannel(&ch))
				break inner
			}
		}
	}

	return channels, nil
}

// GetUserInfo returns the UserInfo for a requested user.
func (s SlackAdapter) GetUserInfo(userID string) (*adapter.UserInfo, error) {
	u, err := s.rtm.GetUserInfo(userID)
	if err != nil {
		return nil, err
	}

	return newUserInfoFromSlackUser(u), nil
}

func getFields(name string, i interface{}) map[string]interface{} {
	merge := func(m, n map[string]interface{}) map[string]interface{} {
		for k, v := range n {
			m[k] = v
		}
		return m
	}

	value := reflect.ValueOf(i)

	if value.IsZero() {
		return map[string]interface{}{}
	}

	switch value.Kind() {
	case reflect.Interface:
		fallthrough

	case reflect.Ptr:
		if value.IsNil() {
			return map[string]interface{}{}
		}

		v := value.Elem()
		if !v.CanInterface() {
			return map[string]interface{}{}
		}

		return getFields(name, v.Interface())

	case reflect.Struct:
		m := map[string]interface{}{}
		t := value.Type()

		fmt.Println("GOT TYPE:", t.Name())

		for i := 0; i < t.NumField(); i++ {
			fv := value.Field(i)
			fn := t.Field(i).Name

			if !fv.CanInterface() {
				continue
			}

			m = merge(m, getFields(strings.ToLower(fn), fv.Interface()))
		}

		return m

	default:
		return map[string]interface{}{name: i}
	}
}

// Listen instructs the relay to begin listening to the provider that it's attached to.
// It exits immediately, returning a channel that emits ProviderEvents.
func (s SlackAdapter) Listen() <-chan *adapter.ProviderEvent {
	le := log.WithField("adapter", s.GetName())
	events := make(chan *adapter.ProviderEvent)

	le.WithField("provider", s.provider.Name).Info("Connecting to Slack provider")

	go s.rtm.ManageConnection()

	go func() {
		info := &adapter.Info{
			Provider: adapter.NewProviderInfoFromConfig(s.provider),
			User:     &adapter.UserInfo{},
		}

	eventLoop:
		for msg := range s.rtm.IncomingEvents {
			e := le.WithField("message.type", msg.Type)

			if log.IsLevelEnabled(log.TraceLevel) {
				fields := getFields("", msg.Data)

				for k, v := range fields {
					e = e.WithField(k, v)
				}
			}

			e.Trace("Incoming message")

			switch ev := msg.Data.(type) {
			case *slack.ConnectedEvent:
				suser, err := s.rtm.GetUserInfo(ev.Info.User.ID)
				if err != nil {
					e.WithError(err).
						WithField("user.id", ev.Info.User.ID).
						Error("Error finding user on connect")

					continue eventLoop
				}

				info.User = newUserInfoFromSlackUser(suser)

				e.WithField("attempt", ev.ConnectionCount).
					WithField("info.team.id", ev.Info.Team.ID).
					WithField("info.team.domain", ev.Info.Team.Domain).
					WithField("info.user.id", ev.Info.User.ID).
					WithField("info.user.name", ev.Info.User.Name).
					Info("Slack event: connected")

				events <- s.OnConnected(ev, info)

			case *slack.ConnectionErrorEvent:
				e.WithError(ev.ErrorObj).
					WithField("attempt", ev.Attempt).
					WithField("backoff", ev.Backoff).
					Error("Slack event: connection error -- backing off")

				events <- s.OnConnectionError(ev, info)

			case *slack.DisconnectedEvent:
				e.WithError(ev.Cause).
					WithField("intentional", ev.Intentional).
					Error("Slack event: disconnected")

				events <- s.OnDisconnected(ev, info)

			case *slack.InvalidAuthEvent:
				e.Fatal("Slack event: invalid auth")

				events <- s.OnInvalidAuth(ev, info)

				break eventLoop

			case *slack.LatencyReport:
				le := e.WithField("latency", ev.Value)
				millis := ev.Value.Milliseconds()

				switch {
				case millis >= 1000 && millis < 1500:
					le.Debug("Slack event: high latency detected")
				case millis >= 1500 && millis < 2000:
					le.Info("Slack event: high latency detected")
				case millis >= 2000:
					le.Warn("Slack event: high latency detected")
				}

			case *slack.MessageEvent:
				providerEvent := s.OnMessage(ev, info)
				if providerEvent != nil && providerEvent.EventType != "" {
					events <- providerEvent
				}

			case *slack.RTMError:
				e.WithError(ev).
					WithField("code", ev.Code).
					WithField("msg", ev.Msg).
					Error("Slack event: RTM error")

				events <- s.OnRTMError(ev, info)

			case *slack.AckErrorEvent:
				e.WithError(ev.ErrorObj).
					WithField("replyto", ev.ReplyTo).
					Error("Slack event: ACK event error")

			case *slack.ConnectingEvent:
				e.WithField("attempt", ev.Attempt).
					WithField("connection.count", ev.ConnectionCount).
					Info("Slack event: connecting")

			case *slack.IncomingEventError:
				e.WithError(ev).
					Error("Slack event: error receiving incoming event")

			case *slack.OutgoingErrorEvent:
				e.WithError(ev.ErrorObj).
					WithField("message.channel", ev.Message.Channel).
					WithField("message.id", ev.Message.ID).
					WithField("message.text", ev.Message.Text).
					WithField("message.type", ev.Message.Type).
					Error("Slack event: outgoing message error")

			case *slack.RateLimitedError:
				e.WithError(ev).
					WithField("retryafter", ev.RetryAfter).
					Error("Slack event: API rate limited error")

			case *slack.UnmarshallingErrorEvent:
				e.WithError(ev.ErrorObj).
					Error("Slack event: failed to deconstruct Slack response")

			default:
				// Report and ignore other events..
				e.WithField("message.data", msg.Data).
					WithField("type", fmt.Sprintf("%T", ev)).
					Info("Slack event: unhandled event type")
			}
		}

		close(events)
	}()

	return events
}

// OnChannelMessage is called when the Slack API emits an MessageEvent for a message in a channel.
func (s *SlackAdapter) OnChannelMessage(event *slack.MessageEvent, info *adapter.Info) *adapter.ProviderEvent {
	return s.wrapEvent(
		"channel_message",
		info,
		&adapter.ChannelMessageEvent{
			ChannelID: event.Channel,
			Text:      ScrubMarkdown(event.Msg.Text),
			UserID:    event.Msg.User,
		},
	)
}

// OnConnected is called when the Slack API emits a ConnectedEvent.
func (s *SlackAdapter) OnConnected(event *slack.ConnectedEvent, info *adapter.Info) *adapter.ProviderEvent {
	return s.wrapEvent(
		"connected",
		info,
		&adapter.ConnectedEvent{},
	)
}

// OnConnectionError is called when the Slack API emits an ConnectionErrorEvent.
func (s *SlackAdapter) OnConnectionError(event *slack.ConnectionErrorEvent, info *adapter.Info) *adapter.ProviderEvent {
	return s.wrapEvent(
		"connection_error",
		info,
		&adapter.ErrorEvent{Msg: event.Error()},
	)
}

// OnDirectMessage is called when the Slack API emits an MessageEvent for a direct message.
func (s *SlackAdapter) OnDirectMessage(event *slack.MessageEvent, info *adapter.Info) *adapter.ProviderEvent {
	return s.wrapEvent(
		"direct_message",
		info,
		&adapter.DirectMessageEvent{
			ChannelID: event.Channel,
			Text:      ScrubMarkdown(event.Msg.Text),
			UserID:    event.Msg.User,
		},
	)
}

// OnDisconnected is called when the Slack API emits a DisconnectedEvent.
func (s *SlackAdapter) OnDisconnected(event *slack.DisconnectedEvent, info *adapter.Info) *adapter.ProviderEvent {
	return s.wrapEvent(
		"disconnected",
		info,
		&adapter.DisconnectedEvent{Intentional: event.Intentional},
	)
}

// OnInvalidAuth is called when the Slack API emits an InvalidAuthEvent.
func (s *SlackAdapter) OnInvalidAuth(event *slack.InvalidAuthEvent, info *adapter.Info) *adapter.ProviderEvent {
	return s.wrapEvent(
		"authentication_error",
		info,
		&adapter.AuthenticationErrorEvent{
			Msg: fmt.Sprintf("Connection failed to %s: invalid credentials", s.provider.Name),
		},
	)
}

// OnMessage is called when the Slack API emits a MessageEvent.
func (s *SlackAdapter) OnMessage(event *slack.MessageEvent, info *adapter.Info) *adapter.ProviderEvent {
	switch event.Msg.SubType {
	case "": // Just a plain message. Handle accordingly.
		if event.Channel[0] == 'D' {
			return s.OnDirectMessage(event, info)
		}

		return s.OnChannelMessage(event, info)
	case "message_changed":
		// Note here for later; ignore for now.
		return nil
	case "message_deleted":
		// Note here for later; ignore for now.
		return nil
	case "bot_message":
		// Note here for later; ignore for now.
		return nil
	default:
		log.WithField("subtype", event.Msg.SubType).
			Warn("Received message subtype")
		return nil
	}
}

// OnRTMError is called when the Slack API emits an RTMError.
func (s *SlackAdapter) OnRTMError(event *slack.RTMError, info *adapter.Info) *adapter.ProviderEvent {
	return s.wrapEvent(
		"error",
		info,
		&adapter.ErrorEvent{
			Code: event.Code,
			Msg:  event.Msg,
		},
	)
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
func (s *SlackAdapter) wrapEvent(eventType string, info *adapter.Info, data interface{}) *adapter.ProviderEvent {
	return &adapter.ProviderEvent{
		EventType: eventType,
		Data:      data,
		Info:      info,
		Adapter:   s,
	}
}

// NewAdapter will construct a SlackAdapter instance for a given provider configuration.
func NewAdapter(provider data.SlackProvider) SlackAdapter {
	client := slack.New(provider.APIToken)
	rtm := client.NewRTM()

	return SlackAdapter{
		client:   client,
		provider: provider,
		rtm:      rtm,
	}
}

// ScrubMarkdown removes unnecessary/undesirable Slack markdown (of links, of
// example) from text recieved from Slack.
func ScrubMarkdown(text string) string {
	// Remove links of the format "<https://google.com>"
	//
	if index := linkMarkdownRegexShort.FindStringIndex(text); index != nil {
		submatch := linkMarkdownRegexShort.FindStringSubmatch(text)
		text = text[:index[0]] + submatch[1] + text[index[1]:]
	}

	// Remove links of the format "<http://google.com|google.com>"
	//
	if index := linkMarkdownRegexLong.FindStringIndex(text); index != nil {
		submatch := linkMarkdownRegexLong.FindStringSubmatch(text)
		text = text[:index[0]] + submatch[1] + text[index[1]:]
	}

	return text
}
