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
	configSetShort = "Set or update a configuration value"
	configSetLong  = "Set or update a configuration value."

// 	configSetUsage = `Usage:
//    gort group set [flags] group_name user_name...

//  Flags:
//    -h, --help   Show this message and exit

//  Global Flags:
//    -P, --profile string   The Gort profile within the config file to use
//  `
)

var flagGortConfigSecret bool

// GetConfigSetCmd is a command
func GetConfigSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   configSetUse,
		Short: configSetShort,
		Long:  configSetLong,
		RunE:  configSetCmd,
		Args:  cobra.ExactArgs(1),
	}

	// cmd.SetUsageTemplate(configSetUsage)

	cmd.Flags().StringVarP(&flagGortConfigLayer, "layer", "l", "bundle", "One of: [bundle room group user]")
	cmd.Flags().StringVarP(&flagGortConfigBundle, "bundle", "b", "", "The bundle to configure")
	cmd.Flags().StringVarP(&flagGortConfigOwner, "owner", "o", "", "The owning room, group, or user")
	cmd.Flags().StringVarP(&flagGortConfigKey, "key", "k", "", "The name of the configuration")
	cmd.Flags().BoolVarP(&flagGortConfigSecret, "secret", "s", false, "Makes a configuration value secret")

	return cmd
}

func configSetCmd(cmd *cobra.Command, args []string) error {
	value := args[0]

	dc := data.DynamicConfiguration{
		Bundle: flagGortConfigBundle,
		Layer:  data.ConfigurationLayer(flagGortConfigLayer),
		Owner:  flagGortConfigOwner,
		Key:    flagGortConfigKey,
		Value:  value,
		Secret: flagGortConfigSecret,
	}

	switch {
	case dc.Bundle == "":
		return fmt.Errorf("dynamic configuration bundle is required")
	case dc.Layer == data.ConfigurationLayer(""):
		return fmt.Errorf("dynamic configuration layer is required")
	case dc.Layer.Validate() != nil:
		return dc.Layer.Validate()
	case dc.Owner == "" && dc.Layer != data.LayerBundle:
		return fmt.Errorf("dynamic configuration owner is required for layer %s", dc.Layer)
	case dc.Key == "":
		return fmt.Errorf("dynamic configuration key is required")
	}

	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	vs, err := gortClient.BundleListVersions(dc.Bundle)
	if err != nil {
		return err
	}
	if len(vs) == 0 {
		return fmt.Errorf("no such bundle installed: %s", dc.Bundle)
	}

	err = gortClient.DynamicConfigurationSave(dc)
	if err != nil {
		return err
	}

	fmt.Printf("Configuration set: bundle=%q layer=%q owner=%q key=%q\n",
		flagGortConfigBundle,
		flagGortConfigLayer,
		flagGortConfigOwner,
		flagGortConfigKey)

	return err
}
