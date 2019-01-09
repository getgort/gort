package data

//// The wrappers for the "slack" section.
//// Other providers will eventually get their own sections

// Provider is the general interface for all providers.
// Currently only Slack is supported, with HipChat coming in time.
type Provider interface{}

// AbstractProvider is used to contain the general properties shared by
// all providers.
type AbstractProvider struct {
	BotName string `yaml:"bot_name,omitempty"`
	Name    string `yaml:"name,omitempty"`
}

// SlackProvider is the data wrapper for a Slack provider.
type SlackProvider struct {
	AbstractProvider `yaml:",inline"`
	APIToken         string `yaml:"api_token,omitempty"`
	IconURL          string `yaml:"icon_url,omitempty"`
}
