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
	bundleVersionsUse   = "versions"
	bundleVersionsShort = "Lists installed bundle versions."
	bundleVersionsLong  = "List all versions of an installed bundle."
	bundleVersionsUsage = `Usage: gort bundle versions [OPTIONS] NAME

	Lists installed versions of a bundle.

	All versions of the specified bundle are listed, along
	with their status ("Enabled", "Disabled", "Incompatible")

  Options:
	--help              Show this message and exit.
`
)

// TODO: Support incompatible flag
// -x, --incompatible  Lists only incompatible bundle versions

// GetBundleVersionsCmd is a command
func GetBundleVersionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   bundleVersionsUse,
		Short: bundleVersionsShort,
		Long:  bundleVersionsLong,
		RunE:  bundleVersionsCmd,
		Args:  cobra.ExactArgs(1),
	}

	cmd.SetUsageTemplate(bundleVersionsUsage)

	return cmd
}

func bundleVersionsCmd(cmd *cobra.Command, args []string) error {
	const format = "%-12s%-12s%-12s\n"

	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	bundles, err := gortClient.BundleListVersions(args[0])
	if err != nil {
		return err
	}

	fmt.Printf(format, "BUNDLE", "VERSION", "STATUS")

	for _, b := range bundles {
		if b.Version == "" {
			b.Version = "-"
		}

		status := "Disabled"
		if b.Enabled {
			status = "Enabled"
		}
		// TODO: Determine whether bundles are incompatible

		fmt.Printf(format, b.Name, b.Version, status)

	}

	return nil
}
