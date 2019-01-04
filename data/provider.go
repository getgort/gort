package data

//// The wrappers for the "slack" section.
//// Other providers will eventually get their own sections

// Provider is the general interface for all providers.
// Currently only Slack is supported, with HipChat coming in time.
type Provider interface{}

// SlackProvider is the data wrapper for a Slack provider.
type SlackProvider struct {
	Provider      `yaml:"-"`
	BotName       string `yaml:"bot-name,omitempty"`
	IconURL       string `yaml:"icon-url,omitempty"`
	Name          string `yaml:"name,omitempty"`
	SlackAPIToken string `yaml:"api-token,omitempty"`
}
