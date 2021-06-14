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
	groupListUse   = "list"
	groupListShort = "List all existing groups"
	groupListLong  = "List all existing groups."
)

// GetGroupListCmd is a command
func GetGroupListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   groupListUse,
		Short: groupListShort,
		Long:  groupListLong,
		RunE:  groupListCmd,
	}

	return cmd
}

func groupListCmd(cmd *cobra.Command, args []string) error {
	const format = "%s\n"

	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	groups, err := gortClient.GroupList()
	if err != nil {
		return err
	}

	fmt.Printf(format, "GROUP NAME")
	for _, g := range groups {
		fmt.Printf(format, g.Name)
	}

	return nil
}
