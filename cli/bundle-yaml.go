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

package cli

import (
	"fmt"

	"github.com/getgort/gort/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	bundleYamlUse   = "yaml"
	bundleYamlShort = "Retrieve the raw YAML for a bundle."
	bundleYamlLong  = `Retrieve the raw YAML for a bundle.`
	bundleYamlUsage = `Usage:
  gort bundle yaml [flags] bundle_name version

Flags:
  -h, --help   Show this message and exit

Global Flags:
  -P, --profile string   The Gort profile within the config file to use
`
)

// GetBundleYamlCmd is a command
func GetBundleYamlCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   bundleYamlUse,
		Short: bundleYamlShort,
		Long:  bundleYamlLong,
		RunE:  bundleYamlCmd,
		Args:  cobra.ExactArgs(2),
	}

	cmd.SetUsageTemplate(bundleYamlUsage)

	return cmd
}

func bundleYamlCmd(cmd *cobra.Command, args []string) error {
	name := args[0]
	version := args[1]

	// TODO Implement that no specified version returns enabled version.

	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	bundle, err := gortClient.BundleGet(name, version)
	if err != nil {
		return err
	}

	bytes, err := yaml.Marshal(bundle)
	if err != nil {
		return err
	}

	fmt.Println(string(bytes))

	return nil
}
