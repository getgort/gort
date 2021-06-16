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
	bundleListUse   = "list"
	bundleListShort = "List all existing bundles"
	bundleListLong  = "List all existing bundles."
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

// GetBundleListCmd is a command
func GetBundleListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   bundleListUse,
		Short: bundleListShort,
		Long:  bundleListLong,
		RunE:  bundleListCmd,
	}

	return cmd
}

func bundleListCmd(cmd *cobra.Command, args []string) error {
	const format = "%-12s%-12s%-12s%s\n"

	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	bundles, err := gortClient.BundleList()
	if err != nil {
		return err
	}

	fmt.Printf(format, "BUNDLE", "VERSION", "TYPE", "STATUS")

	for _, b := range bundles {
		if b.Version == "" {
			b.Version = "-"
		}

		status := "Disabled"
		if b.Enabled {
			status = "Enabled"
		}

		kind := "Explicit"
		if b.Default {
			kind = "Default"
		}

		fmt.Printf(format, b.Name, b.Version, kind, status)
	}

	return nil
}
