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
	"reflect"
	"strings"

	io2 "github.com/getgort/gort/data/io"

	"github.com/getgort/gort/adapter"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/telemetry"
	"github.com/getgort/gort/templates"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

var _ adapter.Adapter = &ClassicAdapter{}

// ClassicAdapter is the Slack provider implementation of a relay, which knows how
// to receive events from the Slack API, translate them into Gort events, and
// forward them along.
type ClassicAdapter struct {
	client   *slack.Client
	provider data.SlackProvider
	rtm      *slack.RTM
}

func (s *ClassicAdapter) React(ctx context.Context, message adapter.MessageRef, emoji adapter.Emoji) error {
	return s.client.AddReactionContext(ctx, emoji.Shortname(), slack.ItemRef{
		Channel:   message.ChannelID,
		Timestamp: message.Timestamp,
	})
}

func (s *ClassicAdapter) Reply(ctx context.Context, message adapter.MessageRef, content string) error {
	_, _, _, err := s.client.SendMessageContext(ctx, message.ChannelID, slack.MsgOptionTS(message.Timestamp), slack.MsgOptionText(content, false))
	return err
}

// GetChannelInfo returns the ChannelInfo for a requested channel.
func (s ClassicAdapter) GetChannelInfo(channelID string) (*io2.ChannelInfo, error) {
	ch, err := s.rtm.GetConversationInfo(channelID, false)
	if err != nil {
		return nil, err
	}

	return newChannelInfoFromSlackChannel(ch), nil
}

// GetName returns this adapter's configured name
func (s ClassicAdapter) GetName() string {
	return s.provider.Name
}

// GetPresentChannels returns a slice of channel ID strings that the Adapter
// is present in. This is expensive. Don't use it often.
func (s ClassicAdapter) GetPresentChannels() ([]*io2.ChannelInfo, error) {
	allChannels, _, err := s.rtm.GetConversations(&slack.GetConversationsParameters{})
	if err != nil {
		return nil, err
	}

	channels := make([]*io2.ChannelInfo, 0)
	for _, ch := range allChannels {
		if ch.IsMember {
			channels = append(channels, newChannelInfoFromSlackChannel(&ch))
		}
	}

	return channels, nil
}

// GetUserInfo returns the UserInfo for a requested user.
func (s ClassicAdapter) GetUserInfo(userID string) (*io2.UserInfo, error) {
	u, err := s.rtm.GetUserInfo(userID)
	if err != nil {
		return nil, err
	}

	return newUserInfoFromSlackUser(u), nil
}

// Listen instructs the relay to begin listening to the provider that it's attached to.
// It exits immediately, returning a channel that emits ProviderEvents.
func (s ClassicAdapter) Listen(ctx context.Context) <-chan *adapter.ProviderEvent {
	le := log.WithField("adapter", s.GetName())
	events := make(chan *adapter.ProviderEvent)

	le.WithField("provider", s.provider.Name).Info("Connecting to Slack provider")

	go s.rtm.ManageConnection()

	go func() {
		info := &adapter.Info{
			Provider: data.NewProviderInfoFromConfig(s.provider),
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
				e.WithField("attempt", ev.ConnectionCount).
					WithField("info.team.id", ev.Info.Team.ID).
					WithField("info.team.domain", ev.Info.Team.Domain).
					WithField("info.user.id", ev.Info.User.ID).
					WithField("info.user.name", ev.Info.User.Name).
					Trace("Slack event: connected")

				events <- s.onConnected(ev, info)

			case *slack.ConnectionErrorEvent:
				e.WithError(ev.ErrorObj).
					WithField("attempt", ev.Attempt).
					WithField("backoff", ev.Backoff).
					Error("Slack event: connection error -- backing off")
				telemetry.Errors().WithError(ev.ErrorObj).Commit(ctx)

				events <- s.onConnectionError(ev, info)

			case *slack.DisconnectedEvent:
				e.WithError(ev.Cause).
					WithField("intentional", ev.Intentional).
					Debug("Slack event: disconnected")
				telemetry.Errors().WithError(ev.Cause).Commit(ctx)

				events <- s.onDisconnected(ev, info)

			case *slack.InvalidAuthEvent:
				e.Debug("Slack event: invalid auth")

				events <- s.onInvalidAuth(ev, info)

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
				providerEvent := s.onMessage(ev, info)
				if providerEvent != nil && providerEvent.EventType != "" {
					events <- providerEvent
				}

			case *slack.RTMError:
				e.WithError(ev).
					WithField("code", ev.Code).
					WithField("msg", ev.Msg).
					Debug("Slack event: RTM error")
				telemetry.Errors().WithError(ev).Commit(ctx)

				events <- s.onRTMError(ev, info)

			case *slack.AckErrorEvent:
				e.WithError(ev.ErrorObj).
					WithField("replyto", ev.ReplyTo).
					Error("Slack event: ACK event error")
				telemetry.Errors().WithError(ev.ErrorObj).Commit(ctx)

			case *slack.ConnectingEvent:
				e.WithField("attempt", ev.Attempt).
					WithField("connection.count", ev.ConnectionCount).
					Debug("Slack event: connecting")

			case *slack.IncomingEventError:
				e.WithError(ev).
					Debug("Slack event: error receiving incoming event")
				telemetry.Errors().WithError(ev).Commit(ctx)

			case *slack.OutgoingErrorEvent:
				e.WithError(ev.ErrorObj).
					WithField("message.channel", ev.Message.Channel).
					WithField("message.id", ev.Message.ID).
					WithField("message.text", ev.Message.Text).
					WithField("message.type", ev.Message.Type).
					Error("Slack event: outgoing message error")
				telemetry.Errors().WithError(ev.ErrorObj).Commit(ctx)

			case *slack.RateLimitedError:
				e.WithError(ev).
					WithField("retryafter", ev.RetryAfter).
					Error("Slack event: API rate limited error")
				telemetry.Errors().WithError(ev).Commit(ctx)

			case *slack.UnmarshallingErrorEvent:
				e.WithError(ev.ErrorObj).
					Error("Slack event: failed to deconstruct Slack response")
				telemetry.Errors().WithError(ev.ErrorObj).Commit(ctx)

			case *slack.UserTypingEvent:
				e.WithField("user", ev.User).
					WithField("channel", ev.Channel).
					Trace("Slack event: user is typing")

			case *slack.HelloEvent:
				// Do nothing (for now).

			default:
				// Report and ignore other events..
				e.WithField("message.data", msg.Data).
					WithField("type", fmt.Sprintf("%T", ev)).
					Debug("Slack event: unhandled event type")
			}
		}

		close(events)
	}()

	return events
}

// Send the contents of a response envelope to a specified channel. If
// channelID is empty the value of envelope.Request.ChannelID will be used.
func (s *ClassicAdapter) Send(ctx context.Context, channelID string, elements templates.OutputElements) error {
	return Send(ctx, s.client, s, channelID, elements)
}

// SendText sends a simple text message to the specified channel.
func (s *ClassicAdapter) SendText(ctx context.Context, channelID string, message string) error {
	return SendText(ctx, s.client, s, channelID, message)
}

// SendError is a break-glass error message function that's used when the
// templating function fails somehow. Obviously, it does not utilize the
// templating engine.
func (s *ClassicAdapter) SendError(ctx context.Context, channelID string, title string, err error) error {
	return SendError(ctx, s.client, channelID, title, err)
}

// onChannelMessage is called when the Slack API emits an MessageEvent for a message in a channel.
func (s *ClassicAdapter) onChannelMessage(event *slack.MessageEvent, info *adapter.Info) *adapter.ProviderEvent {
	mr := adapter.MessageRef{
		ID:        "",
		ChannelID: event.Channel,
		Timestamp: event.Timestamp,
		Adapter:   s.GetName(),
	}
	return s.wrapEvent(
		adapter.EventChannelMessage,
		info,
		&adapter.ChannelMessageEvent{
			ChannelID:  event.Channel,
			Text:       ScrubMarkdown(event.Msg.Text),
			UserID:     event.Msg.User,
			MessageRef: mr,
		},
	)
}

// onConnected is called when the Slack API emits a ConnectedEvent.
func (s *ClassicAdapter) onConnected(event *slack.ConnectedEvent, info *adapter.Info) *adapter.ProviderEvent {
	return s.wrapEvent(
		adapter.EventConnected,
		info,
		&adapter.ConnectedEvent{},
	)
}

// onConnectionError is called when the Slack API emits an ConnectionErrorEvent.
func (s *ClassicAdapter) onConnectionError(event *slack.ConnectionErrorEvent, info *adapter.Info) *adapter.ProviderEvent {
	return s.wrapEvent(
		adapter.EventConnectionError,
		info,
		&adapter.ErrorEvent{Msg: event.Error()},
	)
}

// onDirectMessage is called when the Slack API emits an MessageEvent for a direct message.
func (s *ClassicAdapter) onDirectMessage(event *slack.MessageEvent, info *adapter.Info) *adapter.ProviderEvent {
	mr := adapter.MessageRef{
		ID:        "",
		ChannelID: event.Channel,
		Timestamp: event.Timestamp,
		Adapter:   s.GetName(),
	}
	return s.wrapEvent(
		adapter.EventDirectMessage,
		info,
		&adapter.DirectMessageEvent{
			ChannelID:  event.Channel,
			Text:       ScrubMarkdown(event.Msg.Text),
			UserID:     event.Msg.User,
			MessageRef: mr,
		},
	)
}

// onDisconnected is called when the Slack API emits a DisconnectedEvent.
func (s *ClassicAdapter) onDisconnected(event *slack.DisconnectedEvent, info *adapter.Info) *adapter.ProviderEvent {
	return s.wrapEvent(
		adapter.EventDisconnected,
		info,
		&adapter.DisconnectedEvent{Intentional: event.Intentional},
	)
}

// onInvalidAuth is called when the Slack API emits an InvalidAuthEvent.
func (s *ClassicAdapter) onInvalidAuth(event *slack.InvalidAuthEvent, info *adapter.Info) *adapter.ProviderEvent {
	return s.wrapEvent(
		adapter.EventAuthenticationError,
		info,
		&adapter.AuthenticationErrorEvent{
			Msg: fmt.Sprintf("Connection failed to %s: invalid credentials", s.provider.Name),
		},
	)
}

// onMessage is called when the Slack API emits a MessageEvent.
func (s *ClassicAdapter) onMessage(event *slack.MessageEvent, info *adapter.Info) *adapter.ProviderEvent {
	switch event.Msg.SubType {
	case "": // Just a plain message. Handle accordingly.
		if event.Channel[0] == 'D' {
			return s.onDirectMessage(event, info)
		}

		return s.onChannelMessage(event, info)
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

// onRTMError is called when the Slack API emits an RTMError.
func (s *ClassicAdapter) onRTMError(event *slack.RTMError, info *adapter.Info) *adapter.ProviderEvent {
	return s.wrapEvent(
		adapter.EventError,
		info,
		&adapter.ErrorEvent{
			Code: event.Code,
			Msg:  event.Msg,
		},
	)
}

// wrapEvent creates a new ProviderEvent instance with metadata and the Event data attached.
func (s *ClassicAdapter) wrapEvent(eventType adapter.EventType, info *adapter.Info, data interface{}) *adapter.ProviderEvent {
	return &adapter.ProviderEvent{
		EventType: eventType,
		Data:      data,
		Info:      info,
		Adapter:   s,
	}
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
