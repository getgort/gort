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
	roleListUse   = "list"
	roleListShort = "List all existing roles"
	roleListLong  = "List all existing roles."
)

// GetRoleListCmd is a command
func GetRoleListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   roleListUse,
		Short: roleListShort,
		Long:  roleListLong,
		RunE:  roleListCmd,
	}

	return cmd
}

func roleListCmd(cmd *cobra.Command, args []string) error {
	const format = "%s\n"

	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	roles, err := gortClient.RoleList()
	if err != nil {
		return err
	}

	fmt.Printf(format, "ROLE NAME")
	for _, g := range roles {
		fmt.Printf(format, g.Name)
	}

	return nil
}
