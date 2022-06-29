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

// $ cogctl group --help
// Usage: cogctl group [OPTIONS] COMMAND [ARGS]...
//
//   Manage Cog user groups.
//
//   If invoked without a subcommand, lists all user groups.
//
// Options:
//   --help  Show this message and exit.
//
// Commands:
//   add     Add one or more users to a group.
//   create  Create a new user group.
//   delete  Delete a group.
//   grant   Grant one or more roles to a group.
//   info    Show info on a specific group.
//   remove  Remove one or more users from a group.
//   rename  Rename a user group.
//   revoke  Revoke one or more roles from a group.

const (
	groupListUse   = "list"
	groupListShort = "List all existing groups"
	groupListLong  = "List all existing groups."
	groupListUsage = `Usage:
  gort group list [flags]

Flags:
  -h, --help   Show this message and exit

Global Flags:
  -P, --profile string   The Gort profile within the config file to use
`
)

// GetGroupListCmd is a command
func GetGroupListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   groupListUse,
		Short: groupListShort,
		Long:  groupListLong,
		RunE:  groupListCmd,
	}

	cmd.SetUsageTemplate(groupListUsage)

	return cmd
}

func groupListCmd(cmd *cobra.Command, args []string) error {
	gortClient, err := client.Connect(FlagGortProfile, FlagConfigBaseDir)
	if err != nil {
		return err
	}

	groups, err := gortClient.GroupList()
	if err != nil {
		return err
	}

	// Sort by name, for presentation purposes.
	sort.Slice(groups, func(i, j int) bool { return groups[i].Name < groups[j].Name })

	c := &Columnizer{}
	c.StringColumn("GROUP NAME", func(i int) string { return groups[i].Name })
	c.Print(groups)

	return nil
}
