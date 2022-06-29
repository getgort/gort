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
	"strings"

	"github.com/spf13/cobra"

	"github.com/getgort/gort/client"
)

// $ cogctl group info --help
// Usage: cogctl group info [OPTIONS] GROUP
//
//   Show info on a specific group.
//
// Options:
//   --help  Show this message and exit.

const (
	groupInfoUse   = "info"
	groupInfoShort = "Show info on a specific group"
	groupInfoLong  = "Show info on a specific group."
	groupInfoUsage = `Usage:
  gort group info [flags] group_name

Flags:
  -h, --help   Show this message and exit

Global Flags:
  -P, --profile string   The Gort profile within the config file to use
`
)

// GetGroupInfoCmd is a command
func GetGroupInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   groupInfoUse,
		Short: groupInfoShort,
		Long:  groupInfoLong,
		RunE:  groupInfoCmd,
		Args:  cobra.ExactArgs(1),
	}

	cmd.SetUsageTemplate(groupInfoUsage)

	return cmd
}

func groupInfoCmd(cmd *cobra.Command, args []string) error {
	groupname := args[0]

	gortClient, err := client.Connect(FlagGortProfile, FlagConfigBaseDir)
	if err != nil {
		return err
	}

	//
	// TODO Maybe multiplex the following queries with gofuncs?
	//

	users, err := gortClient.GroupMemberList(groupname)
	if err != nil {
		return err
	}

	roles, err := gortClient.GroupRoleList(groupname)
	if err != nil {
		return err
	}

	const format = `Name   %s
Users  %s
Roles  %s
`

	fmt.Printf(
		format,
		groupname,
		strings.Join(userNames(users), ", "),
		strings.Join(roleNames(roles), ", "),
	)

	return nil
}
