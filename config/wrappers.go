package config

// CogConfig is the top-level configuration object
type CogConfig struct {
	GlobalConfigs  GlobalConfigs   `yaml:"global"`
	DockerConfigs  DockerConfigs   `yaml:"docker"`
	SlackProviders []SlackProvider `yaml:"slack"`
	BundleConfigs  []BundleConfig  `yaml:"bundles"`
}

// GlobalConfigs is the data wrapper for the "global" section
type GlobalConfigs struct {
	CommandTimeoutSeconds int `yaml:"command-timeout-seconds"`
}

// DockerConfigs is the data wrapper for the "docker" section. Small now, but more will come.
type DockerConfigs struct {
	DockerHost string `yaml:"host"`
}

//// The wrappers for the "slack" section.
//// Other providers will eventually get their own sections

// Provider is the general interface for all providers.
// Currently only Slack is supported, with HipChat coming in time.
type Provider interface{}

// SlackProvider is the data wrapper for a Slack provider.
type SlackProvider struct {
	Provider      `yaml:"-"`
	BotName       string `yaml:"bot-name"`
	IconURL       string `yaml:"icon-url"`
	Name          string `yaml:"name"`
	SlackAPIToken string `yaml:"api-token"`
}

//// The wrappers for the "bundles" section

// BundleConfig represents an element in the "bundles" section
type BundleConfig struct {
	Name        string                         `yaml:"name"`
	Description string                         `yaml:"description"`
	Docker      BundleDockerConfig             `yaml:"docker"`
	Commands    map[string]BundleCommandConfig `yaml:"commands"`
}

// BundleDockerConfig represents the "bundles/docker" subsection of the config doc
type BundleDockerConfig struct {
	Image string `yaml:"image"`
	Tag   string `yaml:"tag"`
}

// BundleCommandConfig represents the "bundles/commands" subsection of the config doc
type BundleCommandConfig struct {
	Description string   `yaml:"description"`
	Executable  []string `yaml:"executable"`
	Name        string   `yaml:"-"`
}
