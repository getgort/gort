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
   gort bundle uninstall [flags] bundle_name [version]

 Flags:
   -a, --all     Uninstall all versions of the bundle
   -c, --clean   Uninstall all disabled bundle versions
   -h, --help    help for uninstall

 Global Flags:
   -P, --profile string   The Gort profile within the config file to use
 `
)

var (
	flagBundleUninstallAll   bool
	flagBundleUninstallClean bool
)

// GetBundleUninstallCmd is a command
func GetBundleUninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   bundleUninstallUse,
		Short: bundleUninstallShort,
		Long:  bundleUninstallLong,
		RunE:  bundleUninstallCmd,
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.Flags().BoolVarP(&flagBundleUninstallAll, "all", "a", false, "Uninstall all versions of the bundle")
	cmd.Flags().BoolVarP(&flagBundleUninstallClean, "clean", "c", false, "Uninstall all disabled bundle versions")

	cmd.SetUsageTemplate(bundleUninstallUsage)

	return cmd
}

func bundleUninstallCmd(cmd *cobra.Command, args []string) error {
	bundleName, bundleVersion := args[0], ""
	if len(args) > 1 {
		bundleVersion = args[1]
	}

	c, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	var uninstall []data.Bundle

	switch {
	case flagBundleUninstallAll:
		uninstall, err = c.BundleListVersions(bundleName)
		if err != nil {
			return err
		}
	case flagBundleUninstallClean:
		bundles, err := c.BundleListVersions(bundleName)
		if err != nil {
			return err
		}

		for _, b := range bundles {
			if !b.Enabled {
				uninstall = append(uninstall, b)
			}
		}
	default:
		if bundleVersion == "" {
			return fmt.Errorf("missing required argument: bundle version")
		}

		uninstall = []data.Bundle{{Name: bundleName, Version: bundleVersion}}
	}

	if len(uninstall) == 0 {
		fmt.Println("No bundles deleted.")
		return nil
	}

	for _, b := range uninstall {
		if err = c.BundleUninstall(b.Name, b.Version); err != nil {
			return err
		}

		fmt.Printf("Bundle %s %s uninstalled.\n", b.Name, b.Version)
	}

	return nil
}
