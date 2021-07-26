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
	"sort"

	"github.com/getgort/gort/client"
	"github.com/spf13/cobra"
)

const (
	roleListUse   = "list"
	roleListShort = "List all existing roles"
	roleListLong  = "List all existing roles."
	roleListUsage = `Usage:
  gort role list [flags]

Flags:
  -h, --help   Show this message and exit

Global Flags:
  -P, --profile string   The Gort profile within the config file to use
`
)

// GetRoleListCmd is a command
func GetRoleListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   roleListUse,
		Short: roleListShort,
		Long:  roleListLong,
		RunE:  roleListCmd,
	}

	cmd.SetUsageTemplate(roleListUsage)

	return cmd
}

func roleListCmd(cmd *cobra.Command, args []string) error {
	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	roles, err := gortClient.RoleList()
	if err != nil {
		return err
	}

	// Sort by name, for presentation purposes.
	sort.Slice(roles, func(i, j int) bool { return roles[i].Name < roles[j].Name })

	c := &Columnizer{}
	c.StringColumn("ROLE NAME", func(i int) string { return roles[i].Name })
	c.Print(roles)

	return nil
}
