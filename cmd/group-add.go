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
	groupAddUse   = "add"
	groupAddShort = "Add a user to an existing group"
	groupAddLong  = "Add a user to an existing group."
)

// GetGroupAddCmd is a command
func GetGroupAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   groupAddUse,
		Short: groupAddShort,
		Long:  groupAddLong,
		RunE:  groupAddCmd,
		Args:  cobra.ExactArgs(2),
	}

	return cmd
}

func groupAddCmd(cmd *cobra.Command, args []string) error {
	groupname := args[0]
	username := args[1]

	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	err = gortClient.GroupMemberAdd(groupname, username)
	if err != nil {
		return err
	}

	fmt.Printf("User added to %s: %s\n", groupname, username)

	return nil
}
