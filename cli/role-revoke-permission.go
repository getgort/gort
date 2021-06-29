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
	roleRevokePermissionUse   = "revoke-permission role bundle permission"
	roleRevokePermissionShort = "Revoke a permission from an existing role"
	roleRevokePermissionLong  = "Revoke a permission from an existing role."
)

// GetRoleRevokePermissionCmd is a command
func GetRoleRevokePermissionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   roleRevokePermissionUse,
		Short: roleRevokePermissionShort,
		Long:  roleRevokePermissionLong,
		RunE:  roleRevokePermissionCmd,
		Args:  cobra.ExactArgs(3),
	}

	return cmd
}

func roleRevokePermissionCmd(cmd *cobra.Command, args []string) error {
	rolename := args[0]
	bundlename := args[1]
	permissionname := args[2]

	gortClient, err := client.Connect(FlagGortProfile)
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
