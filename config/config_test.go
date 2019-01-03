package config

import (
	"fmt"
	"testing"

	yaml "gopkg.in/yaml.v1"
)

func Test(t *testing.T) {
	config := CogConfig{}

	config.SlackProviders = make([]SlackProvider, 1)

	config.SlackProviders[0] = SlackProvider{
		BotName:       "Cog2",
		Name:          "ClockworkSoul",
		IconURL:       "https://emoji.slack-edge.com/T025151EM/dragongopher/cdc9b1bd1a7752eb.png",
		SlackAPIToken: "SlackAPIToken",
	}

	config.BundleConfigs = make([]BundleConfig, 1)

	config.BundleConfigs[0] = BundleConfig{
		Name:        "test",
		Description: "A description",
	}

	config.BundleConfigs[0].Docker = BundleDockerConfig{
		Image: "clockworksoul/foo",
		Tag:   "latest",
	}

	config.BundleConfigs[0].Commands = make(map[string]BundleCommandConfig)

	config.BundleConfigs[0].Commands["splitecho"] = BundleCommandConfig{
		Description: "Echos back anything sent to it, one parameter at a time.",
		Executable:  []string{"/opt/app/splitecho.sh"},
	}

	config.BundleConfigs[0].Commands["curl"] = BundleCommandConfig{
		Description: "The official curl command",
		Executable:  []string{"/usr/bin/curl"},
	}

	config.BundleConfigs[0].Commands["echo"] = BundleCommandConfig{
		Description: "Echos back anything sent to it, all at once.",
		Executable:  []string{"/bin/echo"},
	}

	y, err := yaml.Marshal(config)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	fmt.Println(string(y))
}
