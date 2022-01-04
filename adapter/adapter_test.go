package adapter

import (
	"context"
	"os"
	"testing"

	"github.com/getgort/gort/config"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/dataaccess"
	"github.com/getgort/gort/service"
	"github.com/getgort/gort/templates"
)

func TestMain(m *testing.M) {
	err := setupGort()
	if err != nil {
		panic(err)
	}
	code := m.Run()
	os.Exit(code)
}

func TestChannelMessage(t *testing.T) {
	var tests = []struct {
		name            string
		message         string
		expected        string
		expectNoRequest bool
		err             bool
	}{
		{
			name:     "can execute command by name with bang",
			message:  "!test:cmd arg1 arg2",
			expected: "test:cmd arg1 arg2",
		},
		{
			name:            "cannot execute command by name without bang",
			message:         "test:cmd arg1 arg2",
			expectNoRequest: true,
		},
		{
			name:     "can execute command by trigger",
			message:  "run this command",
			expected: "test:cmd run this command",
		},
		{
			name:    "error on unknown command with bang",
			message: "!missing:cmd arg1 arg2",
			err:     true,
		},
		{
			name:            "no request on untriggered message without bang",
			message:         "nothing to match",
			expectNoRequest: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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
					Text:      test.message,
					UserID:    "user",
				},
			)
			if err != nil {
				if test.err {
					return
				}
				t.Errorf("%v", err)
				return
			}
			if test.err {
				t.Errorf("expected an error, got %q", result)
				return
			}
			if result == nil {
				if !test.expectNoRequest {
					t.Errorf("expected %q, got %q", test.expected, result)
				}
				return
			}
			if test.expectNoRequest {
				t.Errorf("expected nil, got %q", result)
			}
			if result.String() != test.expected {
				t.Errorf("expected %q, got %q", test.expected, result)
			}
		})
	}

}

func TestDirectMessage(t *testing.T) {
	var tests = []struct {
		name            string
		message         string
		expected        string
		expectNoRequest bool
		err             bool
	}{
		{
			name:     "can execute command by name with bang",
			message:  "test:cmd arg1 arg2",
			expected: "test:cmd arg1 arg2",
		},
		{
			name:     "can execute command by name without bang",
			message:  "test:cmd arg1 arg2",
			expected: "test:cmd arg1 arg2",
		},
		{
			name:     "can execute command with trigger",
			message:  "run this command",
			expected: "test:cmd run this command",
		},
		{
			name:    "error on unknown command with bang",
			message: "!missing:cmd arg1 arg2",
			err:     true,
		},
		{
			name:            "no request on untriggered message without bang",
			message:         "nothing to match",
			expectNoRequest: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := OnDirectMessage(
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
				&DirectMessageEvent{
					ChannelID: "mychannel",
					Text:      test.message,
					UserID:    "user",
				},
			)
			if err != nil {
				if test.err {
					return
				}
				t.Errorf("%v", err)
				return
			}
			if test.err {
				t.Errorf("expected an error, got %q", result)
				return
			}
			if result == nil {
				if !test.expectNoRequest {
					t.Errorf("expected %q, got %q", test.expected, result)
				}
				return
			}
			if test.expectNoRequest {
				t.Errorf("expected nil, got %q", result)
			}
			if result.String() != test.expected {
				t.Errorf("expected %q, got %q", test.expected, result)
			}
		})
	}

}

func setupGort() error {
	// Init Gort
	err := config.Initialize("../testing/config/no-database.yml")
	if err != nil {
		return err
	}

	da, err := dataaccess.Get()
	if err != nil {
		return err
	}

	da.Initialize(context.Background())
	service.DoBootstrap(context.Background(), rest.User{
		Email:    "user@getgort.io",
		Username: "user",
	})

	err = da.BundleCreate(context.Background(), testBundle)
	if err != nil {
		return err
	}

	return nil
}

var testBundle = data.Bundle{
	GortBundleVersion: 1,
	Name:              "test",
	Version:           "1.0.0",
	Description:       "a test bundle",
	Enabled:           true,
	Commands: map[string]*data.BundleCommand{
		"cmd": {
			Name: "cmd",
			Triggers: []data.Trigger{
				{
					Match: "com+and",
				},
			},
			Rules: []string{"allow"},
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
