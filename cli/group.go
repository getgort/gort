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
	groupUse   = "group"
	groupShort = "Manage Cog user groups"
	groupLong  = "Manage Cog user groups."
)

// GetGroupCmd group
func GetGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   groupUse,
		Short: groupShort,
		Long:  groupLong,
	}

	cmd.AddCommand(GetGroupAddCmd())
	cmd.AddCommand(GetGroupCreateCmd())
	cmd.AddCommand(GetGroupDeleteCmd())
	cmd.AddCommand(GetGroupGrantCmd())
	cmd.AddCommand(GetGroupInfoCmd())
	cmd.AddCommand(GetGroupListCmd())
	cmd.AddCommand(GetGroupRemoveCmd())
	cmd.AddCommand(GetGroupRevokeCmd())

	return cmd
}
