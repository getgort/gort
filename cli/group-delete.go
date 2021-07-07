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

// $ cogctl group delete --help
// Usage: cogctl group delete [OPTIONS] GROUP
//
//   Delete a group.
//
// Options:
//   --help  Show this message and exit.

const (
	groupDeleteUse   = "delete"
	groupDeleteShort = "Delete an existing group"
	groupDeleteLong  = "Delete an existing group."
	groupDeleteUsage = `Usage:
  gort group delete [flags] group_name

Flags:
  -h, --help   Show this message and exit

Global Flags:
  -P, --profile string   The Gort profile within the config file to use
`
)

// GetGroupDeleteCmd is a command
func GetGroupDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   groupDeleteUse,
		Short: groupDeleteShort,
		Long:  groupDeleteLong,
		RunE:  groupDeleteCmd,
		Args:  cobra.ExactArgs(1),
	}

	cmd.SetUsageTemplate(groupDeleteUsage)

	return cmd
}

func groupDeleteCmd(cmd *cobra.Command, args []string) error {
	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	groupname := args[0]

	group, err := gortClient.GroupGet(groupname)
	if err != nil {
		return err
	}

	fmt.Printf("Deleting group %s... ", group.Name)

	err = gortClient.GroupDelete(group.Name)
	if err != nil {
		return err
	}

	fmt.Println("Successful")

	return nil
}
