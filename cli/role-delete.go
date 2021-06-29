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
	roleDeleteUse   = "delete"
	roleDeleteShort = "Delete an existing role"
	roleDeleteLong  = "Delete an existing role."
)

// GetroleDeleteCmd is a command
func GetRoleDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   roleDeleteUse,
		Short: roleDeleteShort,
		Long:  roleDeleteLong,
		RunE:  roleDeleteCmd,
		Args:  cobra.ExactArgs(1),
	}

	return cmd
}

func roleDeleteCmd(cmd *cobra.Command, args []string) error {
	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	rolename := args[0]

	role, err := gortClient.RoleGet(rolename)
	if err != nil {
		return err
	}

	fmt.Printf("Deleting role %s... ", role.Name)

	err = gortClient.RoleDelete(role.Name)
	if err != nil {
		return err
	}

	fmt.Println("Successful.")

	return nil
}
