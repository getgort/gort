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
	"github.com/getgort/gort/data"
	"github.com/spf13/cobra"
)

const (
	configSetUse   = "set"
	configSetShort = "Set or update a dynamic configuration value"
	configSetLong  = `Set or update a dynamic configuration value, which can be injected into
commands' environments at execution time with the same name as the key.

Dynamic configuration keys may not start with "GORT_".
`
	configSetUsage = `Usage:
  gort config set [-b bundle] [-l layer] [-o owner] [-k key] [-s secret] [flags] config_value

Flags:
  -b, --bundle string   The bundle to configure (required)
  -h, --help            Show this message and exit
  -k, --key string      The name of the configuration
  -l, --layer string    One of: [bundle room group user] (default "bundle")
  -o, --owner string    The owning room, group, or user
  -s, --secret          Makes a configuration value secret

Global Flags:
  -P, --profile string   The Gort profile within the config file to use
`
)

var (
	flagGortConfigSetLayer  string
	flagGortConfigSetBundle string
	flagGortConfigSetOwner  string
	flagGortConfigSetKey    string
	flagGortConfigSetSecret bool
)

// GetConfigSetCmd is a command
func GetConfigSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   configSetUse,
		Short: configSetShort,
		Long:  configSetLong,
		RunE:  configSetCmd,
		Args:  cobra.ExactArgs(1),
	}

	cmd.SetUsageTemplate(configSetUsage)

	cmd.Flags().StringVarP(&flagGortConfigSetLayer, "layer", "l", "bundle", "One of: [bundle room group user]")
	cmd.Flags().StringVarP(&flagGortConfigSetBundle, "bundle", "b", "", "The bundle to configure")
	cmd.Flags().StringVarP(&flagGortConfigSetOwner, "owner", "o", "", "The owning room, group, or user")
	cmd.Flags().StringVarP(&flagGortConfigSetKey, "key", "k", "", "The name of the configuration")
	cmd.Flags().BoolVarP(&flagGortConfigSetSecret, "secret", "s", false, "Makes a configuration value secret")

	return cmd
}

func configSetCmd(cmd *cobra.Command, args []string) error {
	value := args[0]

	dc := data.DynamicConfiguration{
		Bundle: flagGortConfigSetBundle,
		Layer:  data.ConfigurationLayer(flagGortConfigSetLayer),
		Owner:  flagGortConfigSetOwner,
		Key:    flagGortConfigSetKey,
		Value:  value,
		Secret: flagGortConfigSetSecret,
	}

	switch {
	case dc.Bundle == "":
		return fmt.Errorf("dynamic configuration bundle (--bundle) is required")
	case dc.Layer == data.ConfigurationLayer(""):
		return fmt.Errorf("dynamic configuration layer (--layer) is required")
	case dc.Layer.Validate() != nil:
		return dc.Layer.Validate()
	case dc.Owner == "" && dc.Layer != data.LayerBundle:
		return fmt.Errorf("dynamic configuration owner (--owner) is required for layer %s", dc.Layer)
	case dc.Key == "":
		return fmt.Errorf("dynamic configuration key (--key) is required")
	}

	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	if exists, err := gortClient.BundleExists(dc.Bundle); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("no such bundle installed: %s", dc.Bundle)
	}

	err = gortClient.DynamicConfigurationSave(dc)
	if err != nil {
		return err
	}

	if dc.Layer == data.LayerBundle {
		fmt.Printf("Configuration set: bundle=%q layer=%q key=%q\n",
			flagGortConfigSetBundle,
			flagGortConfigSetLayer,
			flagGortConfigSetKey)
	} else {
		fmt.Printf("Configuration set: bundle=%q layer=%q owner=%q key=%q\n",
			flagGortConfigSetBundle,
			flagGortConfigSetLayer,
			flagGortConfigSetOwner,
			flagGortConfigSetKey)
	}

	return err
}
