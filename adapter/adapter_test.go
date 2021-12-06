package adapter

import (
	"context"
	"testing"

	"github.com/getgort/gort/config"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/dataaccess"
	"github.com/getgort/gort/service"
	"github.com/getgort/gort/templates"
)

func TestChannelMessageWithTrigger(t *testing.T) {
	setupGort(t)

	result, err := OnChannelMessage(
		context.Background(),
		&ProviderEvent{
			EventType: EventChannelMessage,
			Data:      nil,
			Info: &Info{
				Provider: &ProviderInfo{
					Type: "test",
					Name: "provider",
				},
			},
			Adapter: &testAdapter{},
		},
		&ChannelMessageEvent{
			ChannelID: "mychannel",
			Text:      "command",
			UserID:    "user",
		},
	)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	expected := "test:cmd "
	if result.String() != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func setupGort(t *testing.T) {
	// Init Gort
	err := config.Initialize("../testing/config/no-database.yml")
	if err != nil {
		t.Fatalf("%v", err)
	}

	da, err := dataaccess.Get()
	if err != nil {
		t.Fatalf("%v", err)
	}

	da.Initialize(context.Background())
	service.DoBootstrap(context.Background(), rest.User{
		Email:    "user@getgort.io",
		Username: "user",
	})

	err = da.BundleCreate(context.Background(), testBundle)
	if err != nil {
		t.Fatalf("%v", err)
	}
}

var testBundle = data.Bundle{
	GortBundleVersion: 1,
	Name:              "test",
	Version:           "1.0.0",
	Description:       "a test bundle",
	Enabled:           true,
	Commands: map[string]*data.BundleCommand{
		"cmd": {
			Name:    "cmd",
			Trigger: "com+and",
			Rules:   []string{"allow"},
		},
	},
}

var _ Adapter = &testAdapter{}

type testAdapter struct{}

// GetChannelInfo provides info on a specific provider channel accessible
// to the adapter.
func (t *testAdapter) GetChannelInfo(channelID string) (*ChannelInfo, error) {
	return &ChannelInfo{
		ID:      channelID,
		Members: nil,
		Name:    channelID,
	}, nil
}

// GetName provides the name of this adapter as per the configuration.
func (t *testAdapter) GetName() string {
	return "testAdapter"
}

// GetPresentChannels returns a slice of channels that the adapter is present in.
func (t *testAdapter) GetPresentChannels() ([]*ChannelInfo, error) {
	panic("not implemented") // TODO: Implement
}

// GetUserInfo provides info on a specific provider user accessible
// to the adapter.
func (t *testAdapter) GetUserInfo(userID string) (*UserInfo, error) {
	return &UserInfo{
		ID:          userID,
		Name:        userID,
		DisplayName: userID,
	}, nil
}

// Listen causes the Adapter to initiate a connection to its provider and
// begin relaying back events (including errors) via the returned channel.
func (t *testAdapter) Listen(ctx context.Context) <-chan *ProviderEvent {
	panic("not implemented") // TODO: Implement
}

// Send sends the contents of a response envelope to a
// specified channel. If channelID is empty the value of
// envelope.Request.ChannelID will be used.
func (t *testAdapter) Send(ctx context.Context, channelID string, elements templates.OutputElements) error {
	return nil
}

// SendText sends a simple text message to the specified channel.
func (t *testAdapter) SendText(ctx context.Context, channelID string, message string) error {
	panic("not implemented") // TODO: Implement
}

// SendError is a break-glass error message function that's used when the
// templating function fails somehow. Obviously, it does not utilize the
// templating engine.
func (t *testAdapter) SendError(ctx context.Context, channelID string, title string, err error) error {
	panic("not implemented") // TODO: Implement
}
