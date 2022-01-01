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
	"net/http"
	"sort"

	"github.com/getgort/gort/client"
	"github.com/getgort/gort/data"

	"github.com/spf13/cobra"
)

const (
	configGetUse   = "get"
	configGetShort = "Get a non-secret configuration value"
	configGetLong  = "Get a non-secret configuration value."

	configGetUsage = `Get a non-secret configuration value.

 Usage:
 gort config get [-b bundle] [-l layer] [-o owner] [-k key] [flags]

 Flags:
 -b, --bundle string   The bundle to configure (required)
 -h, --help            Show this message and exit
 -k, --key string      The name of the configuration
 -l, --layer string    One of: [bundle room group user] (default "bundle")
 -o, --owner string    The owning room, group, or user
 -s, --secret          Makes a configuration value secret

 Global Flags:
 -P, --profile string   The Gort profile within the config file to use`
)

var (
	flagGortConfigGetLayer  string
	flagGortConfigGetBundle string
	flagGortConfigGetOwner  string
	flagGortConfigGetKey    string
)

// GetConfigGetCmd is a command
func GetConfigGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   configGetUse,
		Short: configGetShort,
		Long:  configGetLong,
		RunE:  configGetCmd,
		Args:  cobra.ExactArgs(0),
	}

	cmd.SetUsageTemplate(configGetUsage)

	cmd.Flags().StringVarP(&flagGortConfigGetBundle, "bundle", "b", "", "The bundle to configure")
	cmd.Flags().StringVarP(&flagGortConfigGetLayer, "layer", "l", "", "One of: [bundle room group user]")
	cmd.Flags().StringVarP(&flagGortConfigGetOwner, "owner", "o", "", "The owning room, group, or user")
	cmd.Flags().StringVarP(&flagGortConfigGetKey, "key", "k", "", "The name of the configuration")

	return cmd
}

func configGetCmd(cmd *cobra.Command, args []string) error {
	if flagGortConfigGetBundle == "" {
		return fmt.Errorf("dynamic configuration bundle is required")
	}

	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	_, err = gortClient.BundleListVersions(flagGortConfigGetBundle)
	if err != nil {
		if cerr, ok := err.(client.Error); ok && cerr.Status() == http.StatusNoContent {
			return fmt.Errorf("no such bundle installed: %s", flagGortConfigGetBundle)
		}
		return err
	}

	layer := data.ConfigurationLayer(flagGortConfigGetLayer)
	if flagGortConfigGetLayer != "" && layer.Validate() != nil {
		return layer.Validate()
	}

	cs, err := gortClient.DynamicConfigurationList(flagGortConfigGetBundle, layer, flagGortConfigGetOwner, flagGortConfigGetKey)
	if err != nil {
		return err
	}

	sort.Sort(SortableConfigurations(cs))

	c := &Columnizer{}
	c.StringColumn("BUNDLE", func(i int) string { return cs[i].Bundle })
	c.StringColumn("LAYER", func(i int) string { return string(cs[i].Layer) })
	c.StringColumn("OWNER", func(i int) string {
		if cs[i].Owner == "" {
			return "-"
		} else {
			return cs[i].Owner
		}
	})
	c.StringColumn("KEY", func(i int) string { return cs[i].Key })
	c.StringColumn("VALUE", func(i int) string {
		if cs[i].Secret {
			return "<secret>"
		} else {
			return cs[i].Value
		}
	})

	c.Print(cs)

	return err
}

type SortableConfigurations []data.DynamicConfiguration

func (s SortableConfigurations) Len() int {
	return len(s)
}

func (s SortableConfigurations) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortableConfigurations) Less(i, j int) bool {
	switch {
	case s[i].Bundle != s[j].Bundle:
		return s[i].Bundle < s[j].Bundle
	case s[i].Layer != s[j].Layer:
		return s[i].Layer < s[j].Layer
	case s[i].Owner != s[j].Owner:
		return s[i].Owner < s[j].Owner
	case s[i].Key != s[j].Key:
		return s[i].Key < s[j].Key
	default:
		return s[i].Value < s[j].Value
	}
}
