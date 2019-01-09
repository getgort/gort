package data

// CogConfig is the top-level configuration object
type CogConfig struct {
	CogServerConfigs CogServerConfigs `yaml:"cog,omitempty"`
	GlobalConfigs    GlobalConfigs    `yaml:"global,omitempty"`
	DockerConfigs    DockerConfigs    `yaml:"docker,omitempty"`
	SlackProviders   []SlackProvider  `yaml:"slack,omitempty"`
	BundleConfigs    []Bundle         `yaml:"bundles,omitempty"`
}

// CogServerConfigs is the data wrapper for the "cog" section.
type CogServerConfigs struct {
	AllowSelfRegistration bool   `yaml:"allow_self_registration,omitempty"`
	APIAddress            string `yaml:"api_address,omitempty"`
	APIURLBase            string `yaml:"api_url_base,omitempty"`
	EnableSpokenCommands  bool   `yaml:"enable_spoken_commands,omitempty"`
}

// GlobalConfigs is the data wrapper for the "global" section
type GlobalConfigs struct {
	CommandTimeoutSeconds int `yaml:"command_timeout_seconds,omitempty"`
}

// DockerConfigs is the data wrapper for the "docker" section.
// This will move into the relay config(s) eventually.
type DockerConfigs struct {
	DockerHost string `yaml:"host,omitempty"`
}
