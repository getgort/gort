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

package slack

import (
	"context"
	"fmt"

	"github.com/getgort/gort/adapter"
	"github.com/getgort/gort/data"
	io2 "github.com/getgort/gort/data/io"
	"github.com/getgort/gort/telemetry"
	"github.com/getgort/gort/templates"

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

func (s *SocketModeAdapter) React(ctx context.Context, message adapter.MessageRef, emoji adapter.Emoji) error {
	return s.client.AddReactionContext(ctx, emoji.Shortname(), slack.ItemRef{
		Channel:   message.ChannelID,
		Timestamp: message.Timestamp,
	})
}

func (s *SocketModeAdapter) Reply(ctx context.Context, message adapter.MessageRef, content string) error {
	_, _, _, err := s.client.SendMessageContext(ctx, message.ChannelID, slack.MsgOptionTS(message.Timestamp), slack.MsgOptionText(content, false))
	return err
}

// GetChannelInfo provides info on a specific provider channel accessible
// to the adapter.
func (s *SocketModeAdapter) GetChannelInfo(channelID string) (*io2.ChannelInfo, error) {
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
func (s *SocketModeAdapter) GetPresentChannels() ([]*io2.ChannelInfo, error) {
	allChannels, _, err := s.client.GetConversations(&slack.GetConversationsParameters{})
	if err != nil {
		return nil, err
	}

	channels := make([]*io2.ChannelInfo, 0)
	for _, ch := range allChannels {
		// Is this user in this channel?
		if ch.IsMember {
			channels = append(channels, newChannelInfoFromSlackChannel(&ch))
		}
	}

	return channels, nil
}

// GetUserInfo provides info on a specific provider user accessible
// to the adapter.
func (s *SocketModeAdapter) GetUserInfo(userID string) (*io2.UserInfo, error) {
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
		Provider: data.NewProviderInfoFromConfig(s.provider),
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

				events <- s.onConnected(info)
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
						// Ignore messages from bots
						if ev.BotID != "" {
							e.WithField("message.data", fmt.Sprintf("%+v", evt.Data)).
								WithField("bot_id", ev.BotID).
								Debug("Slack event: ignoring message from bot")
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
				// Do nothing for now
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

// Send the contents of a response envelope to a specified channel. If
// channelID is empty the value of envelope.Request.ChannelID will be used.
func (s *SocketModeAdapter) Send(ctx context.Context, channelID string, elements templates.OutputElements) error {
	return Send(ctx, s.client, s, channelID, elements)
}

// SendText sends a simple text message to the specified channel.
func (s *SocketModeAdapter) SendText(ctx context.Context, channelID string, message string) error {
	return SendText(ctx, s.client, s, channelID, message)
}

// SendError is a break-glass error message function that's used when the
// templating function fails somehow. Obviously, it does not utilize the
// templating engine.
func (s *SocketModeAdapter) SendError(ctx context.Context, channelID string, title string, err error) error {
	return SendError(ctx, s.client, channelID, title, err)
}

// onChannelMessage is called when the Slack API emits an MessageEvent for a message in a channel.
func (s *SocketModeAdapter) onChannelMessage(event *slackevents.MessageEvent, info *adapter.Info) *adapter.ProviderEvent {
	mr := adapter.MessageRef{
		ID:        "",
		ChannelID: event.Channel,
		Timestamp: event.TimeStamp,
		Adapter:   s.GetName(),
	}
	return s.wrapEvent(
		adapter.EventChannelMessage,
		info,
		&adapter.ChannelMessageEvent{
			ChannelID:  event.Channel,
			Text:       ScrubMarkdown(event.Text),
			UserID:     event.User,
			MessageRef: mr,
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
	mr := adapter.MessageRef{
		ID:        "",
		ChannelID: event.Channel,
		Timestamp: event.TimeStamp,
		Adapter:   s.GetName(),
	}
	return s.wrapEvent(
		adapter.EventDirectMessage,
		info,
		&adapter.DirectMessageEvent{
			ChannelID:  event.Channel,
			Text:       ScrubMarkdown(event.Text),
			UserID:     event.User,
			MessageRef: mr,
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
