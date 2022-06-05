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

	"github.com/spf13/cobra"

	"github.com/getgort/gort/client"
)

const (
	roleRevokeUse   = "revoke-permission role bundle permission"
	roleRevokeShort = "Revoke a permission from a role"
	roleRevokeLong  = "Revoke a permission from a role."
	roleRevokeUsage = `Usage:
  gort role revoke [flags] role_name permission

Flags:
  -h, --help   Show this message and exit

Global Flags:
  -P, --profile string   The Gort profile within the config file to use
`
)

// GetRoleRevokeCmd is a command
func GetRoleRevokeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   roleRevokeUse,
		Short: roleRevokeShort,
		Long:  roleRevokeLong,
		RunE:  roleRevokeCmd,
		Args:  cobra.ExactArgs(3),
	}

	cmd.SetUsageTemplate(roleRevokeUsage)

	return cmd
}

func roleRevokeCmd(cmd *cobra.Command, args []string) error {
	rolename := args[0]
	bundlename := args[1]
	permissionname := args[2]

	gortClient, err := client.Connect(FlagGortProfile, FlagConfigBaseDir)
	if err != nil {
		return err
	}

	err = gortClient.RolePermissionRevoke(rolename, bundlename, permissionname)
	if err != nil {
		return err
	}

	fmt.Printf("Permission Revoked from %s: %s\n", rolename, permissionname)

	return nil
}
