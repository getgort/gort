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
	permissionListUse   = "list"
	permissionListShort = "List all permissions installed"
	permissionListLong  = "Lists all permissions installed, and their currently enabled version, if any."
	permissionListUsage = `Usage:
  gort permission list [flags]

Flags:
  -h, --help   Show this message and exit

Global Flags:
  -P, --profile string   The Gort profile within the config file to use
`
)

// GetPermissionListCmd is a command
func GetPermissionListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   permissionListUse,
		Short: permissionListShort,
		Long:  permissionListLong,
		RunE:  permissionListCmd,
	}

	cmd.SetUsageTemplate(permissionListUsage)

	return cmd
}

func permissionListCmd(cmd *cobra.Command, args []string) error {
	const format = "%-12s\n"

	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	bundles, err := gortClient.BundleList()
	if err != nil {
		return err
	}

	fmt.Printf(format, "NAME")

	for _, b := range bundles {
		for _, p := range b.Permissions {
			fmt.Printf(format, fmt.Sprintf("%v:%v", b.Name, p))
		}
	}

	return nil
}
