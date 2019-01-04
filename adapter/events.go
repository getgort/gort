package adapter

import "fmt"

// Info is used by events to wrap user and provider info.
type Info struct {
	User     *UserInfo
	Provider *ProviderInfo
}

// ProviderEvent is the main wrapper. You will find all the other messages attached as Data
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

// ChannelMessageEvent indicates received a message via a public or private channel (message.channels)
type ChannelMessageEvent struct {
	Channel string
	Text    string
	User    string
}

// ConnectedEvent indicates the client has successfully connected to the provider server (hello)
type ConnectedEvent struct {
}

// ChannelJoinedEvent indicates the bot has joined a channel
type ChannelJoinedEvent struct {
	Channel string
}

// DirectMessageEvent indicates the bot has received a direct message from a user (message.im)
type DirectMessageEvent struct {
	Text string
	User string
}

// ErrorEvent indicates an error reported by the provider
type ErrorEvent struct {
	Code int
	Msg  string
}

func (e ErrorEvent) Error() string {
	return fmt.Sprintf("code %d - %s", e.Code, e.Msg)
}
