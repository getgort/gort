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

// $ cogctl role create --help
// Usage: cogctl role create [OPTIONS] NAME
//
//   Create a role
//
// Options:
//   --help  Show this message and exit.

const (
	roleCreateUse   = "create"
	roleCreateShort = "Create a role"
	roleCreateLong  = "Create a role."
	roleCreateUsage = `Usage:
  gort role create [flags] role_name

Flags:
  -h, --help   Show this message and exit

Global Flags:
  -P, --profile string   The Gort profile within the config file to use
`
)

// GetRoleCreateCmd is a command
func GetRoleCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   roleCreateUse,
		Short: roleCreateShort,
		Long:  roleCreateLong,
		RunE:  roleCreateCmd,
		Args:  cobra.ExactArgs(1),
	}

	cmd.SetUsageTemplate(roleCreateUsage)

	return cmd
}

func roleCreateCmd(cmd *cobra.Command, args []string) error {
	rolename := args[0]

	c, err := client.Connect(FlagGortProfile, FlagConfigBaseDir)
	if err != nil {
		return err
	}

	// Only allow this operation if the role doesn't already exist.
	exists, err := c.RoleExists(rolename)
	if err != nil {
		return err
	}
	if exists {
		return client.ErrResourceExists
	}

	// Client roleCreate will create the gort config if necessary, and append
	// the new credentials to it.
	err = c.RoleCreate(rolename)
	if err != nil {
		return err
	}

	fmt.Printf("role %q created.\n", rolename)

	return nil
}
