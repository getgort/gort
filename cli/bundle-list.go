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

// $ cogctl bundle --help
// Usage: cogctl bundle [OPTIONS] COMMAND [ARGS]...
//
// Manage command bundles and their config.
//
// If no subcommand is given, lists all bundles installed, and their
// currently enabled version, if any.
//
// Options:
// -e, --enabled   List only enabled bundles
// -d, --disabled  List only disabled bundles
// -v, --verbose   Display additional bundle details
// --help          Show this message and exit.

const (
	bundleListUse   = "list"
	bundleListShort = "List all bundles installed"
	bundleListLong  = "Lists all bundles installed, and their currently enabled version, if any."
	bundleListUsage = `Usage:
  gort bundle list [flags]

Flags:
  -h, --help   Show this message and exit

Global Flags:
  -P, --profile string   The Gort profile within the config file to use
`
)

var (
	flagBundleListEnabled  bool
	flagBundleListDisabled bool
	flagBundleListVerbose  bool
)

// GetBundleListCmd is a command
func GetBundleListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   bundleListUse,
		Short: bundleListShort,
		Long:  bundleListLong,
		RunE:  bundleListCmd,
	}

	cmd.Flags().BoolVarP(&flagBundleListEnabled, "enabled", "e", false, "List only enabled bundles")
	cmd.Flags().BoolVarP(&flagBundleListDisabled, "disabled", "d", false, "List only disabled bundles")
	cmd.Flags().BoolVarP(&flagBundleListVerbose, "verbose", "v", false, "Display additional bundle details")

	cmd.SetUsageTemplate(bundleListUsage)

	return cmd
}

func bundleListCmd(cmd *cobra.Command, args []string) error {
	if flagBundleListEnabled && flagBundleListDisabled {
		return fmt.Errorf("--enabled and --disabled flags are mutually exclusive")
	}

	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	bundles, err := gortClient.BundleList()
	if err != nil {
		return err
	}

	switch {
	case flagBundleListEnabled:
		bundles = filterBundles(bundles, func(b data.Bundle) bool {
			return !b.Enabled
		})
	case flagBundleListDisabled:
		bundles = filterBundles(bundles, func(b data.Bundle) bool {
			return b.Enabled
		})
	}

	c := &Columnizer{}
	c.StringColumn("BUNDLE", func(i int) string { return bundles[i].Name })
	c.StringColumn("VERSION", func(i int) string { return bundles[i].Version })
	c.StringColumn("TYPE", func(i int) string {
		kind := "Explicit"
		if bundles[i].Default {
			kind = "Default"
		}
		return kind
	})
	c.StringColumn("STATUS", func(i int) string {
		status := "Disabled"
		if bundles[i].Enabled {
			status = "Enabled"
		}
		return status
	})
	c.Print(bundles)

	return nil
}

// if filter(ss[i]) returns true, that element is filtered out
func filterBundles(in []data.Bundle, filter func(data.Bundle) bool) []data.Bundle {
	var out []data.Bundle

	for _, b := range in {
		if !filter(b) {
			out = append(out, b)
		}
	}

	return out
}
