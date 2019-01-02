package config

import (
	"fmt"
	"testing"

	yaml "gopkg.in/yaml.v1"
)

func Test(t *testing.T) {
	config := CogConfig{}

	config.BundleConfigs = make([]BundleConfig, 1)

	config.BundleConfigs[0] = BundleConfig{
		Name:        "test",
		Description: "A description",
	}

	config.BundleConfigs[0].Docker = BundleDockerConfig{
		Image: "clockworksoul/foo",
		Tag:   "latest",
	}

	config.BundleConfigs[0].Commands = make([]BundleCommandConfig, 3)

	config.BundleConfigs[0].Commands[0] = BundleCommandConfig{
		Command:     "splitecho",
		Description: "Echos back anything sent to it, one parameter at a time.",
		Executable:  []string{"/opt/app/splitecho.sh"},
	}

	config.BundleConfigs[0].Commands[1] = BundleCommandConfig{
		Command:     "curl",
		Description: "The official curl command",
		Executable:  []string{"/usr/bin/curl"},
	}

	config.BundleConfigs[0].Commands[2] = BundleCommandConfig{
		Command:     "echo",
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
