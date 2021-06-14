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
	"strings"

	"github.com/getgort/gort/client"
	"github.com/spf13/cobra"
)

const (
	groupInfoUse   = "info"
	groupInfoShort = "Retrieve information about an existing group"
	groupInfoLong  = "Retrieve information about an existing group."
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

	return cmd
}

func groupInfoCmd(cmd *cobra.Command, args []string) error {
	groupname := args[0]

	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	//
	// TODO Maybe multiplex the following queries with gofuncs?
	// (when there's more than one)
	//

	users, err := gortClient.GroupMemberList(groupname)
	if err != nil {
		return err
	}

	// TODO Add "roles" here when its supported.

	const format = `Name   %s
Users  %s
`

	fmt.Printf(format, groupname, strings.Join(userNames(users), ", "))

	return nil
}
