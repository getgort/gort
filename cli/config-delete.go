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
	configDeleteUse   = "delete"
	configDeleteShort = "Delete a configuration value"
	configDeleteLong  = "Delete a configuration value."

	configDeleteUsage = `Delete a configuration value.

 Usage:
 gort config delete [-b bundle] [-l layer] [-o owner] [-k key]

 Flags:
 -b, --bundle string   The bundle to configure (required)
 -h, --help            Show this message and exit
 -k, --key string      The name of the configuration
 -l, --layer string    One of: [bundle room group user] (default "bundle")
 -o, --owner string    The owning room, group, or user

 Global Flags:
 -P, --profile string   The Gort profile within the config file to use`
)

var (
	flagGortConfigDeleteLayer  string
	flagGortConfigDeleteBundle string
	flagGortConfigDeleteOwner  string
	flagGortConfigDeleteKey    string
)

// GetConfigDeleteCmd is a command
func GetConfigDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   configDeleteUse,
		Short: configDeleteShort,
		Long:  configDeleteLong,
		RunE:  configDeleteCmd,
		Args:  cobra.ExactArgs(0),
	}

	cmd.SetUsageTemplate(configDeleteUsage)

	cmd.Flags().StringVarP(&flagGortConfigDeleteLayer, "layer", "l", "bundle", "One of: [bundle room group user]")
	cmd.Flags().StringVarP(&flagGortConfigDeleteBundle, "bundle", "b", "", "The bundle to configure")
	cmd.Flags().StringVarP(&flagGortConfigDeleteOwner, "owner", "o", "", "The owning room, group, or user")
	cmd.Flags().StringVarP(&flagGortConfigDeleteKey, "key", "k", "", "The name of the configuration")

	return cmd
}

func configDeleteCmd(cmd *cobra.Command, args []string) error {
	dc := data.DynamicConfiguration{
		Bundle: flagGortConfigDeleteBundle,
		Layer:  data.ConfigurationLayer(flagGortConfigDeleteLayer),
		Owner:  flagGortConfigDeleteOwner,
		Key:    flagGortConfigDeleteKey,
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

	if exists, err := gortClient.BundleExists(dc.Bundle); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("no such bundle installed: %s", dc.Bundle)
	}

	err = gortClient.DynamicConfigurationDelete(dc.Bundle, dc.Layer, dc.Owner, dc.Key)
	if err != nil {
		return err
	}

	if dc.Layer == data.LayerBundle {
		fmt.Printf("Configuration deleted: bundle=%q layer=%q key=%q\n",
			flagGortConfigDeleteBundle,
			flagGortConfigDeleteLayer,
			flagGortConfigDeleteKey)
	} else {
		fmt.Printf("Configuration deleted: bundle=%q layer=%q owner=%q key=%q\n",
			flagGortConfigDeleteBundle,
			flagGortConfigDeleteLayer,
			flagGortConfigDeleteOwner,
			flagGortConfigDeleteKey)
	}

	return err
}
