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

package cmd

import (
	"github.com/spf13/cobra"
)

const (
	bundleUse   = "bundle"
	bundleShort = "Perform operations on bundles"
	bundleLong  = "Allows you to perform bundle administration."
)

// $ gort bundle --help
// Usage: gort bundle [OPTIONS] COMMAND [ARGS]...

//   Manage command bundles and their config.

//   If no subcommand is given, lists all bundles installed, and their
//   currently enabled version, if any.

// Options:
//   -e, --enabled   List only enabled bundles
//   -d, --disabled  List only disabled bundles
//   -v, --verbose   Display additional bundle details
//   --help          Show this message and exit.

// Commands:
//   config     Manage dynamic configuration layers.
//   disable    Disable a bundle by name.
//   enable     Enable the specified version of the bundle.
//   info       Display bundle information.
//   install    Install a bundle.
//   uninstall  Uninstall bundles.
//   versions   List installed bundle versions.

// GetBundleCmd bundle
func GetBundleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   bundleUse,
		Short: bundleShort,
		Long:  bundleLong,
	}

	cmd.AddCommand(GetBundleDisableCmd())
	cmd.AddCommand(GetBundleEnableCmd())
	cmd.AddCommand(GetBundleInfoCmd())
	cmd.AddCommand(GetBundleInstallCmd())
	cmd.AddCommand(GetBundleListCmd())
	cmd.AddCommand(GetBundleUninstallCmd())
	cmd.AddCommand(GetBundleYamlCmd())

	return cmd
}
