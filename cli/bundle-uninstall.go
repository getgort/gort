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
)

// $ cogctl bundle uninstall --help
// Usage: cogctl bundle uninstall [OPTIONS] NAME [VERSION]
//
//   Uninstall bundles.
//
// Options:
//   -c, --clean         Uninstall all disabled bundle versions
//   -x, --incompatible  Uninstall all incompatible versions of the bundle
//   -a, --all           Uninstall all versions of the bundle
//   --help              Show this message and exit.

const (
	bundleUninstallUse   = "uninstall"
	bundleUninstallShort = "Uninstall bundles"
	bundleUninstallLong  = `Uninstall bundles.`
	bundleUninstallUsage = `Usage:
  gort bundle uninstall [flags] bundle_name version

Flags:
  -h, --help   Show this message and exit

Global Flags:
  -P, --profile string   The Gort profile within the config file to use
`
)

// GetBundleUninstallCmd is a command
func GetBundleUninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   bundleUninstallUse,
		Short: bundleUninstallShort,
		Long:  bundleUninstallLong,
		RunE:  bundleUninstallCmd,
		Args:  cobra.ExactArgs(2),
	}

	cmd.SetUsageTemplate(bundleUninstallUsage)

	return cmd
}

func bundleUninstallCmd(cmd *cobra.Command, args []string) error {
	bundleName := args[0]
	bundleVersion := args[1]

	c, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	err = c.BundleUninstall(bundleName, bundleVersion)
	if err != nil {
		return err
	}

	fmt.Printf("Bundle %q uninstalled.\n", bundleName)

	return nil
}
