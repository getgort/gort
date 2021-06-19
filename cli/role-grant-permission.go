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
	roleGrantUse   = "grant-permission role bundle permission"
	roleGrantShort = "grant a permission to an existing role"
	roleGrantLong  = "grant a permission to an existing role."
)

// GetRoleGrantPermissionCmd is a command
func GetRoleGrantPermissionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   roleGrantUse,
		Short: roleGrantShort,
		Long:  roleGrantLong,
		RunE:  roleGrantCmd,
		Args:  cobra.ExactArgs(3),
	}

	return cmd
}

func roleGrantCmd(cmd *cobra.Command, args []string) error {
	rolename := args[0]
	bundlename := args[1]
	permissionname := args[2]

	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	err = gortClient.RolePermissionGrant(rolename, bundlename, permissionname)
	if err != nil {
		return err
	}

	fmt.Printf("permission granted to %s: %s\n", rolename, permissionname)

	return nil
}
