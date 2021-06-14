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

package cmd

import (
	"fmt"

	"github.com/getgort/gort/client"
	"github.com/spf13/cobra"
)

const (
	groupRemoveUserUse   = "remove-user"
	groupRemoveUserShort = "Remove a user from an existing group"
	groupRemoveUserLong  = "Remove a user from an existing group."
)

// GetGroupRemoveUserCmd is a command
func GetGroupRemoveUserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   groupRemoveUserUse,
		Short: groupRemoveUserShort,
		Long:  groupRemoveUserLong,
		RunE:  groupRemoveUserCmd,
		Args:  cobra.ExactArgs(2),
	}

	return cmd
}

func groupRemoveUserCmd(cmd *cobra.Command, args []string) error {
	groupname := args[0]
	username := args[1]

	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	err = gortClient.GroupMemberDelete(groupname, username)
	if err != nil {
		return err
	}

	fmt.Printf("User removed from %s: %s\n", groupname, username)

	return nil
}
