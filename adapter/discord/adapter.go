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

package discord

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/getgort/gort/adapter"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/templates"
	"github.com/prometheus/common/log"

	"github.com/bwmarrin/discordgo"
)

const ZeroWidthSpace = "\u200b"

// NewAdapter will construct a DiscordAdapter instance for a given provider configuration.
func NewAdapter(provider data.DiscordProvider) (adapter.Adapter, error) {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + provider.BotToken)
	if err != nil {
		return nil, err
	}

	return &Adapter{
		provider: provider,
		session:  dg,
	}, nil
}

var _ adapter.Adapter = &Adapter{}

// Adapter is the Discord provider implementation of a relay, which knows how
// to receive events from the Discord API, translate them into Gort events, and
// forward them along.
type Adapter struct {
	session  *discordgo.Session
	provider data.DiscordProvider
	events   chan *adapter.ProviderEvent
}

// GetChannelInfo provides info on a specific provider channel accessible
// to the adapter.
func (s *Adapter) GetChannelInfo(channelID string) (*adapter.ChannelInfo, error) {
	channel, err := s.session.Channel(channelID)
	if err != nil {
		return nil, err
	}
	return newChannelInfoFromDiscordChannel(channel), nil
}

func newChannelInfoFromDiscordChannel(channel *discordgo.Channel) *adapter.ChannelInfo {
	out := &adapter.ChannelInfo{
		ID:   channel.ID,
		Name: channel.Name,
	}
	for _, r := range channel.Recipients {
		out.Members = append(out.Members, r.Username)
	}
	return out
}

// GetName provides the name of this adapter as per the configuration.
func (s *Adapter) GetName() string {
	return s.provider.Name
}

// GetPresentChannels returns a slice of channels that a user is present in.
func (s *Adapter) GetPresentChannels() ([]*adapter.ChannelInfo, error) {
	allChannels, err := s.session.UserChannels()
	if err != nil {
		return nil, err
	}

	channels := make([]*adapter.ChannelInfo, 0)
	for _, ch := range allChannels {
		channels = append(channels, newChannelInfoFromDiscordChannel(ch))
	}

	return channels, nil
}

// GetUserInfo provides info on a specific provider user accessible
// to the adapter.
func (s *Adapter) GetUserInfo(userID string) (*adapter.UserInfo, error) {
	u, err := s.session.User(userID)
	if err != nil {
		return nil, err
	}

	return newUserInfoFromDiscordUser(u), nil
}

// Listen causes the Adapter to initiate a connection to its provider and
// begin relaying back events (including errors) via the returned channel.
func (s *Adapter) Listen(ctx context.Context) <-chan *adapter.ProviderEvent {
	s.events = make(chan *adapter.ProviderEvent, 100)

	// Register the messageCreate func as a callback for MessageCreate events.
	s.session.AddHandler(s.messageCreate)
	s.session.AddHandler(s.onConnected)
	s.session.AddHandler(s.onDisconnected)

	go func() {
		// Open a websocket connection to Discord and begin listening.
		err := s.session.Open()
		if err != nil {
			if strings.Contains(err.Error(), "Authentication failed.") {
				s.events <- s.onInvalidAuth()
			} else {
				s.events <- s.onConnectionError(err.Error())
			}
		}
	}()

	return s.events
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func (s *Adapter) messageCreate(sess *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == sess.State.User.ID {
		return
	}
	channel, err := sess.Channel(m.ChannelID)
	if err != nil {
		panic(err)
	}
	if len(channel.Recipients) > 0 {
		s.events <- s.wrapEvent(
			adapter.EventChannelMessage,
			&adapter.DirectMessageEvent{
				ChannelID: m.ChannelID,
				Text:      m.Content,
				UserID:    m.Author.ID,
			},
		)
	} else {
		s.events <- s.wrapEvent(
			adapter.EventChannelMessage,
			&adapter.ChannelMessageEvent{
				ChannelID: m.ChannelID,
				Text:      m.Content,
				UserID:    m.Author.ID,
			},
		)
	}
}

// onConnected is called when the Slack API emits a ConnectedEvent.
func (s *Adapter) onConnected(sess *discordgo.Session, m *discordgo.Connect) *adapter.ProviderEvent {
	return s.wrapEvent(
		adapter.EventConnected,
		&adapter.ConnectedEvent{},
	)
}

// onConnectionError is called when the Slack API emits an ConnectionErrorEvent.
func (s *Adapter) onConnectionError(message string) *adapter.ProviderEvent {
	return s.wrapEvent(
		adapter.EventConnectionError,
		&adapter.ErrorEvent{Msg: message},
	)
}

// onDisconnected is called when the Discord API emits a DisconnectedEvent.
func (s *Adapter) onDisconnected(sess *discordgo.Session, m *discordgo.Disconnect) {
	s.events <- s.wrapEvent(
		adapter.EventDisconnected,
		&adapter.DisconnectedEvent{},
	)
}

// onInvalidAuth is called when the Slack API emits an InvalidAuthEvent.
func (s *Adapter) onInvalidAuth() *adapter.ProviderEvent {
	return s.wrapEvent(
		adapter.EventAuthenticationError,
		&adapter.AuthenticationErrorEvent{
			Msg: fmt.Sprintf("Connection failed to %s: invalid credentials", s.provider.Name),
		},
	)
}

// wrapEvent creates a new ProviderEvent instance with metadata and the Event data attached.
func (s *Adapter) wrapEvent(eventType adapter.EventType, data interface{}) *adapter.ProviderEvent {
	return &adapter.ProviderEvent{
		EventType: eventType,
		Data:      data,
		Info: &adapter.Info{
			Provider: adapter.NewProviderInfoFromConfig(s.provider),
		},
		Adapter: s,
	}
}

// Send the contents of a response envelope to a specified channel. If
// channelID is empty the value of envelope.Request.ChannelID will be used.
func (s *Adapter) Send(ctx context.Context, channelID string, elements templates.OutputElements) error {
	var err error

	log.Debug("Classic Slack apps are deprecated, please upgrade to a Socket mode app.")

	var color uint64
	if elements.Color != "" {
		color, err = strconv.ParseUint(strings.Replace(elements.Color, "#", "", 1), 16, 64)
		if err != nil {
			return fmt.Errorf("badly-formatted color code: %q", elements.Color)
		}
	}

	embed := &discordgo.MessageEmbed{
		Color:     int(color),
		Title:     elements.Title,
		Timestamp: time.Now().Format(time.RFC3339),
		Type:      discordgo.EmbedTypeRich,
	}

	var fields []*discordgo.MessageEmbedField

	for _, e := range elements.Elements {
		switch t := e.(type) {
		case *templates.Divider:
			// Discord dividers are just empty text fields.
			fields = append(fields, &discordgo.MessageEmbedField{
				Name: ZeroWidthSpace, Value: ZeroWidthSpace,
			})

		case *templates.Image:
			img := &discordgo.MessageEmbedImage{URL: t.URL}
			if t.Height != 0 {
				img.Height = t.Height
			}
			if t.Width != 0 {
				img.Width = t.Width
			}
			embed.Image = img

		case *templates.Header:
			elements.Color = strings.TrimPrefix(t.Color, "#")
			elements.Title = t.Title

		case *templates.Section:
			// Ignore sections entirely in Discord.

		case *templates.Text:
			var title = "" // TODO(mtitmus) Implement this.
			var text = t.Text

			if title == "" {
				title = ZeroWidthSpace
			}
			if text == "" {
				text = ZeroWidthSpace
			}
			if t.Monospace {
				text = fmt.Sprintf("```%s```", text)
			}

			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   title,
				Value:  text,
				Inline: false, // TODO(mtitmus) Implement this.
			})

		default:
			return fmt.Errorf("%T fields are not yet supported by Gort for Discord", e)
		}
	}

	if elements.Color != "" {
		c, err := strconv.ParseInt(elements.Color, 16, 64)
		if err != nil {
			return fmt.Errorf("invalid color format (expected '#123456'): %w", err)
		}
		embed.Color = int(c)
	}

	embed.Title = elements.Title
	embed.Fields = fields

	_, err = s.session.ChannelMessageSendEmbed(channelID, embed)

	return err
}

// SendError is a break-glass error message function that's used when the
// templating function fails somehow. Obviously, it does not utilize the
// templating engine.
func (s *Adapter) SendError(ctx context.Context, channelID string, title string, err error) error {
	if title == "" {
		title = "Unhandled Error"
	}

	embed := &discordgo.MessageEmbed{
		Color:     0xFF0000,
		Title:     "Unhandled Error",
		Timestamp: time.Now().Format(time.RFC3339),
		Type:      discordgo.EmbedTypeRich,
		Fields:    []*discordgo.MessageEmbedField{{Name: title, Value: err.Error()}},
	}

	_, err = s.session.ChannelMessageSendEmbed(channelID, embed)

	return err
}

func newUserInfoFromDiscordUser(user *discordgo.User) *adapter.UserInfo {
	u := &adapter.UserInfo{}

	u.ID = user.ID
	u.Name = user.Username
	u.DisplayName = user.Avatar
	u.DisplayNameNormalized = user.Avatar
	u.Email = user.Email
	return u
}
