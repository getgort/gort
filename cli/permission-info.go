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
	permissionInfoUse   = "info"
	permissionInfoShort = "Show info for a specified permission"
	permissionInfoLong  = "Shows info for a specified permission."
	permissionInfoUsage = `Usage:
  gort permission info [flags] permission-name

Flags:
  -h, --help   Show this message and exit

Global Flags:
  -P, --profile string   The Gort profile within the config file to use
`
)

// GetPermissionInfoCmd is a command
func GetPermissionInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   permissionInfoUse,
		Short: permissionInfoShort,
		Long:  permissionInfoLong,
		RunE:  permissionInfoCmd,
		Args:  cobra.ExactArgs(1),
	}

	cmd.SetUsageTemplate(permissionInfoUsage)

	return cmd
}

func permissionInfoCmd(cmd *cobra.Command, args []string) error {
	const format = "%-12s %-12s %-12s\n"

	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	bundles, err := gortClient.BundleList()
	if err != nil {
		return err
	}

	fmt.Printf(format, "BUNDLE", "PERMISSION", "VERSION")

	for _, b := range bundles {
		for _, p := range b.Permissions {
			combinedName := fmt.Sprintf("%v-%v", b.Name, p)
			if p == args[0] || combinedName == args[0] {
				fmt.Printf(format, b.Name, p, b.Version)
			}
		}
	}

	return nil
}
