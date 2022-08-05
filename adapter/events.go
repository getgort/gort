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

package adapter

import (
	"fmt"

	"github.com/getgort/gort/data/io"
)

// Info is used by events to wrap user and provider info.
type Info struct {
	Provider *io.ProviderInfo
}

// EventType specifies the kind of event received from an adapter
// in response to an event from the chat server.
type EventType string

const (
	EventChannelMessage      EventType = "channel_message"
	EventConnected           EventType = "connected"
	EventConnectionError     EventType = "connection_error"
	EventDirectMessage       EventType = "direct_message"
	EventDisconnected        EventType = "disconnected"
	EventAuthenticationError EventType = "authentication_error"
	EventError               EventType = "error"
)

// ProviderEvent is the main wrapper. You will find all the other messages
// attached as Data
type ProviderEvent struct {
	// The type of event
	EventType EventType

	// The event instance
	Data interface{}

	// Contextual info (user, provider)
	Info *Info

	// The adapter that generated the event
	Adapter Adapter
}

// AuthenticationErrorEvent indicates failure to authenticate
type AuthenticationErrorEvent struct {
	Msg string
}

// ChannelMessageEvent indicates received a message via a public or private
// channel (message.channels)
type ChannelMessageEvent struct {
	ChannelID string
	Text      string
	UserID    string
	MessageRef
}

// ConnectedEvent indicates the client has successfully connected to
// the provider server
type ConnectedEvent struct {
}

// DisconnectedEvent indicates the client has disconnected from the
// provider server
type DisconnectedEvent struct {
	Intentional bool
}

// ChannelJoinedEvent indicates the bot has joined a channel
type ChannelJoinedEvent struct {
	Channel string
}

// DirectMessageEvent indicates the bot has received a direct message from a
// user (message.im)
type DirectMessageEvent struct {
	ChannelID string
	Text      string
	UserID    string
	MessageRef
}

// ErrorEvent indicates an error reported by the provider. The occurs before a
// successful connection, Code will be unset.
type ErrorEvent struct {
	Code int
	Msg  string
}

func (e ErrorEvent) Error() string {
	if e.Code != 0 {
		return fmt.Sprintf("code %d - %s", e.Code, e.Msg)
	}

	return e.Msg
}
