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

	"github.com/getgort/gort/bundles"
	"github.com/getgort/gort/client"
	"github.com/spf13/cobra"
)

// cogctl bundle install --help
// Usage: cogctl bundle install [OPTIONS] BUNDLE_OR_PATH [VERSION]
//
//   Install a bundle.
//
//   Bundles may be installed from either a file (i.e., the `config.yaml` file
//   of a bundle), or from Operable's Warehouse bundle registry
//   (https://warehouse.operable.io).
//
//   When installing from a file, you may either give the path to the file, as
//   in:
//
//       cogctl bundle install /path/to/my/bundle/config.yaml
//
//   or you may give the path as `-`, in which case standard input is used:
//
//       cat config.yaml | cogctl bundle install -
//
//   When installing from the bundle registry, you should instead provide the
//   name of the bundle, as well as an optional version to install. No version
//   means the latest will be installed.
//
//       cogctl bundle install cfn
//
//       cogctl bundle install cfn 0.5.13
//
// Options:
//   -e, --enable               Automatically enable a bundle after installing?
//                              [default: False]
//   -f, --force                Install even if a bundle with the same version is
//                              already installed. Applies only to bundles
//                              installed from a file, and not from the Warehouse
//                              bundle registry. Use this to shorten iteration
//                              cycles in bundle development.  [default: False]
//   -r, --relay-group TEXT     Relay group to assign the bundle to. Can be
//                              specified multiple times.
//   -t, --templates DIRECTORY  Path to templates directory. Template bodies will
//                              be inserted into the bundle configuration prior
//                              to uploading. This makes it easier to manage
//                              complex templates.
//   --help                     Show this message and exit.

const (
	bundleInstallUse   = "install"
	bundleInstallShort = "Install a bundle"
	bundleInstallLong  = `Install a bundle from a bundle file.

When using this command, you must provide the path to the file, as follows:

  gort bundle install /path/to/my/bundle/config.yaml
`
	bundleInstallUsage = `Usage:
  gort bundle install [flags] config_path

Flags:
  -h, --help			Show this message and exit
  -e, --enable			Automatically enable a bundle after installing?
						[default: False]
  -f, --force			Install even if a bundle with the same version is
						already installed. Applies only to bundles
						installed from a file, and not from the Warehouse
						bundle registry. Use this to shorten iteration
						cycles in bundle development.  [default: False]

Global Flags:
  -P, --profile string   The Gort profile within the config file to use
`
)

var (
	flagBundleInstallEnable bool
	flagBundleInstallForce  bool
)

// GetBundleInstallCmd is a command
func GetBundleInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   bundleInstallUse,
		Short: bundleInstallShort,
		Long:  bundleInstallLong,
		RunE:  bundleInstallCmd,
		Args:  cobra.ExactArgs(1),
	}

	cmd.SetUsageTemplate(bundleInstallUsage)
	cmd.Flags().BoolVarP(&flagBundleInstallEnable, "enable", "e", false, "Automatically enable a bundle after installing")
	cmd.Flags().BoolVarP(&flagBundleInstallForce, "force", "f", false, "Install even if a bundle with the same version is already installed.")

	return cmd
}

func bundleInstallCmd(cmd *cobra.Command, args []string) error {
	bundlefile := args[0]

	c, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	bundle, err := bundles.LoadBundle(bundlefile)
	if err != nil {
		return err
	}

	// Check for existing instances of this bundle, allowing forced replacement.
	exists, err := c.BundleExists(bundle.Name, bundle.Version)
	if err != nil {
		return err
	}
	if exists {
		if !flagBundleInstallForce {
			return fmt.Errorf("bundle %q already exists at this version. Use --force to force install from your file", bundle.Name)
		}
		err = c.BundleUninstall(bundle.Name, bundle.Version)
		if err != nil {
			return err
		}
	}

	err = c.BundleInstall(bundle)
	if err != nil {
		return err
	}

	if flagBundleInstallEnable {
		err = c.BundleEnable(bundle.Name, bundle.Version)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Bundle %q installed.\n", bundle.Name)

	return nil
}
