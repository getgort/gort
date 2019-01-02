package config

// CogConfig is the top-level configuration object
type CogConfig struct {
	GlobalConfigs  GlobalConfigs   `json:"global"`
	DockerConfigs  DockerConfigs   `json:"docker"`
	SlackProviders []SlackProvider `json:"slack"`
	BundleConfigs  []BundleConfig  `json:"bundles"`
}

// GlobalConfigs is the data wrapper for the "global" section
type GlobalConfigs struct {
	CommandTimeoutSeconds int `json:"command-timeout-seconds"`
}

// DockerConfigs is the data wrapper for the "docker" section. Small now, but more will come.
type DockerConfigs struct {
	DockerHost string `json:"host"`
}

//// The wrappers for the "slack" section.
//// Other providers will eventually get their own sections

// Provider is the general interface for all providers.
// Currently only Slack is supported, with HipChat coming in time.
type Provider interface{}

// AbstractProvider is the data wrapper that all providers have in common.
type AbstractProvider struct {
	Provider
	BotName string `json:"bot-name"`
	Name    string `json:"name"`
}

// SlackProvider is the data wrapper for a Slack provider.
type SlackProvider struct {
	AbstractProvider
	IconURL       string `json:"icon-url"`
	SlackAPIToken string `json:"api-token"`
}

//// The wrappers for the "bundles" section
//

// BundleConfig represents an element in the "bundles" section
type BundleConfig struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Docker      BundleDockerConfig    `json:"docker"`
	Commands    []BundleCommandConfig `json:"commands"`
}

// BundleDockerConfig represents the "bundles/docker" subsection of the config doc
type BundleDockerConfig struct {
	Image string `json:"image"`
	Tag   string `json:"tag"`
}

// BundleCommandConfig represents the "bundles/commands" subsection of the config doc
type BundleCommandConfig struct {
	Command     string   `json:"command"`
	Description string   `json:"description"`
	Executable  []string `json:"executable"`
}
