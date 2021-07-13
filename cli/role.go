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
	"github.com/spf13/cobra"
)

// Usage: cogctl role [OPTIONS] COMMAND [ARGS]...
//
//   Manage roles and role grants.
//
//   Lists roles when called without a subcommand.
//
// Options:
//   --help  Show this message and exit.
//
// Commands:
//   create  Create a role
//   delete  Delete a role
//   grant   Grant a permission to a role
//   info    Show role details
//   rename  Rename a role
//   revoke  Revoke a permission from a role

const (
	roleUse   = "role"
	roleShort = "Perform operations on roles"
	roleLong  = "Allows you to perform role administration."
)

// GetRoleCmd role
func GetRoleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   roleUse,
		Short: roleShort,
		Long:  roleLong,
	}

	cmd.AddCommand(GetRoleCreateCmd())
	cmd.AddCommand(GetRoleDeleteCmd())
	cmd.AddCommand(GetRoleGrantCmd())
	cmd.AddCommand(GetRoleInfoCmd())
	cmd.AddCommand(GetRoleListCmd())
	cmd.AddCommand(GetRoleRevokeCmd())

	return cmd
}
