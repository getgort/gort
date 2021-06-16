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
	"github.com/getgort/gort/data/rest"
	"github.com/spf13/cobra"
)

const (
	groupCreateUse   = "create"
	groupCreateShort = "Create an existing group"
	groupCreateLong  = "Create an existing group."
)

// GetGroupCreateCmd is a command
func GetGroupCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   groupCreateUse,
		Short: groupCreateShort,
		Long:  groupCreateLong,
		RunE:  groupCreateCmd,
		Args:  cobra.ExactArgs(1),
	}

	return cmd
}

func groupCreateCmd(cmd *cobra.Command, args []string) error {
	groupname := args[0]

	c, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	// Only allow this operation if the group doesn't already exist.
	exists, err := c.GroupExists(groupname)
	if err != nil {
		return err
	}
	if exists {
		return client.ErrResourceExists
	}

	group := rest.Group{Name: groupname}

	// Client GroupCreate will create the gort config if necessary, and append
	// the new credentials to it.
	err = c.GroupSave(group)
	if err != nil {
		return err
	}

	fmt.Printf("Group %q created.\n", group.Name)

	return nil
}
