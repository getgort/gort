package relay

import "fmt"

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

	// The relay that generated the event
	Relay Relay
}

// Failed to authenticate
type AuthenticationErrorEvent struct {
	Msg string
}

// Received a message via a public or private channel (message.channels)
type ChannelMessageEvent struct {
	Channel string
	Text    string
	User    string
}

// The client has successfully connected to the provider server (hello)
type ConnectedEvent struct {
}

// You joined a channel
type ChannelJoinedEvent struct {
	Channel string
}

// Received a message from a user (message.im)
type DirectMessageEvent struct {
	Text string
	User string
}

// Error reported by the provider
type ErrorEvent struct {
	Code int
	Msg  string
}

func (e ErrorEvent) Error() string {
	return fmt.Sprintf("Code %d - %s", e.Code, e.Msg)
}
