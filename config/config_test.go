/*
 * Copyright 2021 The Gort Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"fmt"
	"testing"

	yaml "gopkg.in/yaml.v3"
)

func TestInputOutputHumanEyeball(t *testing.T) {
	config, err := loadConfiguration("../config.yml")
	if err != nil {
		t.Error(err.Error())
	}

	y, err := yaml.Marshal(config)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	fmt.Println(string(y))
}

// This is just used to make sure that the YAML will look the way we expect
// based on the struct relationships.
// func TestKindOf(t *testing.T) {
// 	config := data.GortConfig{}

// 	config.SlackProviders = make([]data.SlackProvider, 1)

// 	config.SlackProviders[0] = data.SlackProvider{
// 		BotName:       "Gort",
// 		Name:          "ClockworkSoul",
// 		IconURL:       "https://emoji.slack-edge.com/T025151EM/dragongopher/cdc9b1bd1a7752eb.png",
// 		SlackAPIToken: "SlackAPIToken",
// 	}

// 	config.BundleConfigs = make([]data.Bundle, 1)

// 	config.BundleConfigs[0] = data.Bundle{
// 		Name:        "test",
// 		Description: "A description",
// 	}

// 	config.BundleConfigs[0].Docker = data.BundleDocker{
// 		Image: "clockworksoul/foo",
// 		Tag:   "latest",
// 	}

// 	config.BundleConfigs[0].Commands = make(map[string]data.BundleCommand)

// 	config.BundleConfigs[0].Commands["splitecho"] = data.BundleCommand{
// 		Description: "Echos back anything sent to it, one parameter at a time.",
// 		Executable:  []string{"/opt/app/splitecho.sh"},
// 	}

// 	config.BundleConfigs[0].Commands["curl"] = data.BundleCommand{
// 		Description: "The official curl command",
// 		Executable:  []string{"/usr/bin/curl"},
// 	}

// 	config.BundleConfigs[0].Commands["echo"] = data.BundleCommand{
// 		Description: "Echos back anything sent to it, all at once.",
// 		Executable:  []string{"/bin/echo"},
// 	}

// 	y, err := yaml.Marshal(config)
// 	if err != nil {
// 		fmt.Printf("err: %v\n", err)
// 		return
// 	}

// 	fmt.Println(string(y))
// }
