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

// $ cogctl group add --help
// Usage: cogctl group add [OPTIONS] GROUP USERNAMES...
//
//   Add one or more users to a group.
//
// Options:
//   --help  Show this message and exit.

const (
	groupAddUse   = "add"
	groupAddShort = "Add a user to an existing group"
	groupAddLong  = "Add a user to an existing group."
	groupAddUsage = `Usage:
  gort group add [flags] group_name user_name...

Flags:
  -h, --help   Show this message and exit

Global Flags:
  -P, --profile string   The Gort profile within the config file to use
`
)

// GetGroupAddCmd is a command
func GetGroupAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   groupAddUse,
		Short: groupAddShort,
		Long:  groupAddLong,
		RunE:  groupAddCmd,
		Args:  cobra.MinimumNArgs(2),
	}

	cmd.SetUsageTemplate(groupAddUsage)

	return cmd
}

func groupAddCmd(cmd *cobra.Command, args []string) error {
	groupname := args[0]
	usernames := args[1:]

	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	var errs int

	for _, name := range usernames {
		output := fmt.Sprintf("User added to %s: %s", groupname, name)

		err = gortClient.GroupMemberAdd(groupname, name)
		if err != nil {
			output = fmt.Sprintf("User NOT added to %s: %s (%s)", groupname, name, err.Error())
			errs++
		}

		fmt.Println(output)
	}

	fmt.Printf("%d user(s) added; %d not added.\n", len(usernames)-errs, errs)

	return nil
}
