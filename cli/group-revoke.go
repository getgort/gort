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
	groupRevokeUse   = "revoke"
	groupRevokeShort = "Remove a role from an existing group"
	groupRevokeLong  = "Remove a role from an existing group."
	groupRevokeUsage = `Usage:
  gort group revoke [flags] group_name role_name...

Flags:
  -h, --help   Show this message and exit

Global Flags:
  -P, --profile string   The Gort profile within the config file to use
`
)

// GetGroupRevokeCmd is a command
func GetGroupRevokeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   groupRevokeUse,
		Short: groupRevokeShort,
		Long:  groupRevokeLong,
		RunE:  groupRevokeCmd,
		Args:  cobra.MinimumNArgs(2),
	}

	cmd.SetUsageTemplate(groupRevokeUsage)

	return cmd
}

func groupRevokeCmd(cmd *cobra.Command, args []string) error {
	groupname := args[0]
	rolenames := args[1:]

	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	var errs int

	for _, name := range rolenames {
		var output string

		if err := gortClient.GroupRoleDelete(groupname, name); err != nil {
			output = fmt.Sprintf("Role NOT removed from %s: %s (%s)", groupname, name, err.Error())
			errs++
		} else {
			output = fmt.Sprintf("Role removed from %s: %s", groupname, name)
		}

		fmt.Println(output)
	}

	fmt.Printf("%d role(s) removed; %d not removed.\n", len(rolenames)-errs, errs)

	return nil
}
