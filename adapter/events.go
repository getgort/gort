package adapter

import "fmt"

// Info is used by events to wrap user and provider info.
type Info struct {
	User     *UserInfo
	Provider *ProviderInfo
}

// ProviderEvent is the main wrapper. You will find all the other messages
// attached as Data
type ProviderEvent struct {
	// The type of event
	EventType string

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
