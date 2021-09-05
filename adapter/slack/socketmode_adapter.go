package slack

import (
	"context"
	"fmt"

	"github.com/getgort/gort/adapter"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/telemetry"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

var _ adapter.Adapter = &SocketModeAdapter{}

// SocketModeAdapter is the Slack provider implementation of a relay, which knows how
// to receive events from the Slack API, translate them into Gort events, and
// forward them along.
type SocketModeAdapter struct {
	client       *slack.Client
	socketClient *socketmode.Client
	provider     data.SlackProvider
}

// GetChannelInfo provides info on a specific provider channel accessible
// to the adapter.
func (s *SocketModeAdapter) GetChannelInfo(channelID string) (*adapter.ChannelInfo, error) {
	channel, err := s.client.GetConversationInfo(channelID, false)
	if err != nil {
		return nil, err
	}
	return newChannelInfoFromSlackChannel(channel), nil
}

// GetName provides the name of this adapter as per the configuration.
func (s *SocketModeAdapter) GetName() string {
	return s.provider.Name
}

// GetPresentChannels returns a slice of channels that a user is present in.
func (s *SocketModeAdapter) GetPresentChannels(userID string) ([]*adapter.ChannelInfo, error) {
	allChannels, _, err := s.client.GetConversations(&slack.GetConversationsParameters{})
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

// GetUserInfo provides info on a specific provider user accessible
// to the adapter.
func (s *SocketModeAdapter) GetUserInfo(userID string) (*adapter.UserInfo, error) {
	u, err := s.client.GetUserInfo(userID)
	if err != nil {
		return nil, err
	}

	return newUserInfoFromSlackUser(u), nil
}

// Listen causes the Adapter to initiate a connection to its provider and
// begin relaying back events (including errors) via the returned channel.
func (s *SocketModeAdapter) Listen(ctx context.Context) <-chan *adapter.ProviderEvent {
	var events = make(chan *adapter.ProviderEvent, 100)

	info := &adapter.Info{
		Provider: adapter.NewProviderInfoFromConfig(s.provider),
		User:     &adapter.UserInfo{},
	}

	go func() {
		le := log.WithField("adapter", s.GetName())
		le.WithField("provider", s.provider.Name).Info("Connecting to Slack provider")

		for evt := range s.socketClient.Events {
			e := le.WithField("message.type", evt.Type)

			if log.IsLevelEnabled(log.TraceLevel) {
				fields := getFields("", evt.Data)

				for k, v := range fields {
					e = e.WithField(k, v)
				}
			}

			switch evt.Type {
			case socketmode.EventTypeConnecting:
				ev := evt.Data.(*slack.ConnectingEvent)
				e.WithField("attempt", ev.Attempt).
					WithField("connection.count", ev.ConnectionCount).
					Debug("Slack event: connecting")

			case socketmode.EventTypeConnected:
				ev := evt.Data.(*socketmode.ConnectedEvent)

				e.WithField("attempt", ev.ConnectionCount).
					Trace("Slack event: connected")

				// Note: the connected event is sent after we identify the user
			case socketmode.EventTypeDisconnect:
				e.Debug("Slack event: disconnected")
				telemetry.Errors().Commit(ctx)
				events <- s.onDisconnected(info)
			case socketmode.EventTypeConnectionError:
				ev := evt.Data.(*slack.ConnectionErrorEvent)

				e.WithError(ev.ErrorObj).
					WithField("attempt", ev.Attempt).
					WithField("backoff", ev.Backoff).
					Error("Slack event: connection error -- backing off")
				telemetry.Errors().WithError(ev.ErrorObj).Commit(ctx)

				events <- s.onConnectionError(ev.Error(), info)
			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					e.WithField("message.data", fmt.Sprintf("%+v", evt.Data)).
						Debug("Slack event: ignored event")
					continue
				}
				s.socketClient.Ack(*evt.Request)

				switch eventsAPIEvent.Type {
				case slackevents.CallbackEvent:
					innerEvent := eventsAPIEvent.InnerEvent
					switch ev := innerEvent.Data.(type) {
					case *slackevents.MessageEvent:
						// Skip events with no message text
						if ev.Text == "" {
							continue
						}
						switch ev.ChannelType {
						case "channel": // Public Channel
							events <- s.onChannelMessage(ev, info)
						case "group": // Private Channel
							events <- s.onChannelMessage(ev, info)
						case "im": // Direct Message
							events <- s.onDirectMessage(ev, info)
						default:
							e.WithField("message.data", fmt.Sprintf("%+v", evt.Data)).
								WithField("channel_type", ev.ChannelType).
								Debug("Slack event: unhandled channel type")
						}
					default:
						e.WithField("message.data", fmt.Sprintf("%+v", evt.Data)).
							WithField("type", eventsAPIEvent.Type).
							Debug("Slack event: unhandled Events API event type")
					}
				}
			case socketmode.EventTypeHello:
				// Identify user
				users, err := s.client.GetUsers()
				if err != nil {
					e.WithError(err).
						Error("Error finding user on connect")
					telemetry.Errors().WithError(err).Commit(ctx)

					continue
				}

				for _, user := range users {
					if user.IsBot && user.Profile.ApiAppID == evt.Request.ConnectionInfo.AppID {
						info.User = newUserInfoFromSlackUser(&user)
					}
				}
				events <- s.onConnected(info)
			default:
				// Report and ignore other events..
				e.WithField("message.data", evt.Data).
					WithField("type", evt.Type).
					Debug("Slack event: unhandled event type")
			}
		}
	}()

	go func() {
		err := s.socketClient.Run()
		if err != nil {
			switch err.Error() {
			case "invalid_auth":
				events <- s.onInvalidAuth(info)
			default:
				events <- s.onConnectionError(err.Error(), info)
			}
		}
	}()

	return events
}

// onChannelMessage is called when the Slack API emits an MessageEvent for a message in a channel.
func (s *SocketModeAdapter) onChannelMessage(event *slackevents.MessageEvent, info *adapter.Info) *adapter.ProviderEvent {
	return s.wrapEvent(
		adapter.EventChannelMessage,
		info,
		&adapter.ChannelMessageEvent{
			ChannelID: event.Channel,
			Text:      ScrubMarkdown(event.Text),
			UserID:    event.User,
		},
	)
}

// onConnected is called when the Slack API emits a ConnectedEvent.
func (s *SocketModeAdapter) onConnected(info *adapter.Info) *adapter.ProviderEvent {
	return s.wrapEvent(
		adapter.EventConnected,
		info,
		&adapter.ConnectedEvent{},
	)
}

// onConnectionError is called when the Slack API emits an ConnectionErrorEvent.
func (s *SocketModeAdapter) onConnectionError(message string, info *adapter.Info) *adapter.ProviderEvent {
	return s.wrapEvent(
		adapter.EventConnectionError,
		info,
		&adapter.ErrorEvent{Msg: message},
	)
}

// onDirectMessage is called when the Slack API emits an MessageEvent for a direct message.
func (s *SocketModeAdapter) onDirectMessage(event *slackevents.MessageEvent, info *adapter.Info) *adapter.ProviderEvent {
	return s.wrapEvent(
		adapter.EventDirectMessage,
		info,
		&adapter.DirectMessageEvent{
			ChannelID: event.Channel,
			Text:      ScrubMarkdown(event.Text),
			UserID:    event.User,
		},
	)
}

// onDisconnected is called when the Slack API emits a DisconnectedEvent.
func (s *SocketModeAdapter) onDisconnected(info *adapter.Info) *adapter.ProviderEvent {
	return s.wrapEvent(
		adapter.EventDisconnected,
		info,
		&adapter.DisconnectedEvent{},
	)
}

// onInvalidAuth is called when the Slack API emits an InvalidAuthEvent.
func (s *SocketModeAdapter) onInvalidAuth(info *adapter.Info) *adapter.ProviderEvent {
	return s.wrapEvent(
		adapter.EventAuthenticationError,
		info,
		&adapter.AuthenticationErrorEvent{
			Msg: fmt.Sprintf("Connection failed to %s: invalid credentials", s.provider.Name),
		},
	)
}

// wrapEvent creates a new ProviderEvent instance with metadata and the Event data attached.
func (s *SocketModeAdapter) wrapEvent(eventType adapter.EventType, info *adapter.Info, data interface{}) *adapter.ProviderEvent {
	return &adapter.ProviderEvent{
		EventType: eventType,
		Data:      data,
		Info:      info,
		Adapter:   s,
	}
}

// SendErrorMessage sends an error message to a specified channel.
// TODO Create a MessageBuilder at some point to replace this.
func (s *SocketModeAdapter) SendErrorMessage(channelID string, title string, text string) error {
	_, _, err := s.client.PostMessage(
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
	if err != nil {
		return err
	}
	return nil
}

// SendMessage sends a standard output message to a specified channel.
// TODO Create a MessageBuilder at some point to replace this.
func (s *SocketModeAdapter) SendMessage(channelID string, message string) error {
	_, _, err := s.client.PostMessage(channelID, slack.MsgOptionDisableMediaUnfurl(),
		slack.MsgOptionAsUser(false),
		slack.MsgOptionUsername(s.provider.BotName),
		slack.MsgOptionText(message, false),
		slack.MsgOptionPostMessageParameters(slack.PostMessageParameters{
			IconURL:  s.provider.IconURL,
			Markdown: true,
		}))
	if err != nil {
		return err
	}
	return nil
}
