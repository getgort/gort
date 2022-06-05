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

// $ cogctl group remove --help
// Usage: cogctl group remove [OPTIONS] GROUP USERNAMES...
//
//   Remove one or more users from a group.
//
// Options:
//   --yes   Confirm the action without prompting.
//   --help  Show this message and exit.

const (
	groupRemoveUse   = "remove"
	groupRemoveShort = "Remove a user from an existing group"
	groupRemoveLong  = "Remove a user from an existing group."
	groupRemoveUsage = `Usage:
  gort group remove [flags] group_name user_name...

Flags:
  -h, --help   Show this message and exit

Global Flags:
  -P, --profile string   The Gort profile within the config file to use
`
)

// GetGroupRemoveCmd is a command
func GetGroupRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   groupRemoveUse,
		Short: groupRemoveShort,
		Long:  groupRemoveLong,
		RunE:  groupRemoveCmd,
		Args:  cobra.MinimumNArgs(2),
	}

	cmd.SetUsageTemplate(groupRemoveUsage)

	return cmd
}

func groupRemoveCmd(cmd *cobra.Command, args []string) error {
	groupname := args[0]
	usernames := args[1:]

	gortClient, err := client.Connect(FlagGortProfile, FlagConfigBaseDir)
	if err != nil {
		return err
	}

	var errs int

	for _, name := range usernames {
		var output string

		if err := gortClient.GroupMemberDelete(groupname, name); err != nil {
			output = fmt.Sprintf("User NOT removed from %s: %s (%s)", groupname, name, err.Error())
			errs++
		} else {
			output = fmt.Sprintf("User removed from %s: %s", groupname, name)
		}

		fmt.Println(output)
	}

	fmt.Printf("%d user(s) removed; %d not removed.\n", len(usernames)-errs, errs)

	return nil
}
