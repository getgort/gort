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
	"strings"

	"github.com/getgort/gort/client"
	"github.com/spf13/cobra"
)

const (
	roleInfoUse   = "info"
	roleInfoShort = "Retrieve information about an existing role"
	roleInfoLong  = "Retrieve information about an existing role."
	roleInfoUsage = `Usage:
  gort role info [flags] role_name [version]

Flags:
  -h, --help   Show this message and exit

Global Flags:
  -P, --profile string   The Gort profile within the config file to use
`
)

// GetRoleInfoCmd is a command
func GetRoleInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   roleInfoUse,
		Short: roleInfoShort,
		Long:  roleInfoLong,
		RunE:  roleInfoCmd,
		Args:  cobra.ExactArgs(1),
	}

	cmd.SetUsageTemplate(roleInfoUsage)

	return cmd
}

func roleInfoCmd(cmd *cobra.Command, args []string) error {
	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	rolename := args[0]

	role, err := gortClient.RoleGet(rolename)
	if err != nil {
		return err
	}

	const format = `Name         %s
Permissions  %s
Groups       %s
`

	fmt.Printf(format,
		role.Name,
		strings.Join(role.Permissions.Strings(), ", "),
		strings.Join(groupNames(role.Groups), ", "))

	return nil
}
