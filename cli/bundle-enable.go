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

const (
	bundleEnableUse   = "enable"
	bundleEnableShort = "Enable the specified version of the bundle"
	bundleEnableLong  = `Enable the specified version of the bundle.

If no version is given, the latest installed version (by standard semantic
version ordering) will be enabled.

If any version of this bundle is currently enabled, it will be disabled in
the process.`
	bundleEnableUsage = `Usage:
  gort bundle enable [flags] bundle_name [version]

Flags:
  --help  Show this message and exit.

Global Flags:
  -P, --profile string   The Gort profile within the config file to use
`
)

// GetBundleEnableCmd is a command
func GetBundleEnableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   bundleEnableUse,
		Short: bundleEnableShort,
		Long:  bundleEnableLong,
		RunE:  bundleEnableCmd,
		Args:  cobra.RangeArgs(1, 2),
	}

	cmd.SetUsageTemplate(bundleEnableUsage)

	return cmd
}

func bundleEnableCmd(cmd *cobra.Command, args []string) error {
	var bundleName = args[0]
	var bundleVersion string

	c, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	if len(args) > 1 {
		bundleVersion = args[1]
	} else {
		bundleVersion, err = findLatestVersion(c, bundleName)
		if err != nil {
			return err
		}
	}

	err = c.BundleEnable(bundleName, bundleVersion)
	if err != nil {
		return err
	}

	fmt.Printf("Bundle \"%s\" version %s enabled.\n", bundleName, bundleVersion)

	return nil
}

func findLatestVersion(c *client.GortClient, bundleName string) (string, error) {
	bb, err := c.BundleListVersions(bundleName)
	if err != nil {
		return "", err
	}

	if len(bb) == 0 {
		return "", nil
	}

	return bb[len(bb)-1].Version, nil
}
