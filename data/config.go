package data

// CogConfig is the top-level configuration object
type CogConfig struct {
	GlobalConfigs  GlobalConfigs   `yaml:"global,omitempty"`
	DockerConfigs  DockerConfigs   `yaml:"docker,omitempty"`
	SlackProviders []SlackProvider `yaml:"slack,omitempty"`
	BundleConfigs  []Bundle        `yaml:"bundles,omitempty"`
}

// GlobalConfigs is the data wrapper for the "global" section
type GlobalConfigs struct {
	CommandTimeoutSeconds int `yaml:"command-timeout-seconds,omitempty"`
}

// DockerConfigs is the data wrapper for the "docker" section. Small now, but more will come.
type DockerConfigs struct {
	DockerHost string `yaml:"host,omitempty"`
}
