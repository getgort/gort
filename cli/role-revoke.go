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
	"github.com/getgort/gort/client"

	"github.com/spf13/cobra"
)

const (
	roleRevokeUse   = "revoke-permission role bundle permission"
	roleRevokeShort = "Revoke a permission from a role"
	roleRevokeLong  = "Revoke a permission from a role."
	roleRevokeUsage = `Usage:
  gort role revoke [flags] role_name permission

Flags:
  -h, --help   Show this message and exit

Global Flags:
  -o, --output string    The output format: text (default), json, yaml
  -P, --profile string   The Gort profile within the config file to use
`
)

// GetRoleRevokeCmd is a command
func GetRoleRevokeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   roleRevokeUse,
		Short: roleRevokeShort,
		Long:  roleRevokeLong,
		RunE:  roleRevokeCmd,
		Args:  cobra.ExactArgs(3),
	}

	cmd.SetUsageTemplate(roleRevokeUsage)

	return cmd
}

func roleRevokeCmd(cmd *cobra.Command, args []string) error {
	o := struct {
		*CommandResult
		Role       string
		Bundle     string
		Permission string
	}{
		CommandResult: &CommandResult{},
		Role:          args[0],
		Bundle:        args[1],
		Permission:    args[2],
	}

	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return OutputError(cmd, o, err)
	}

	err = gortClient.RolePermissionRevoke(o.Role, o.Bundle, o.Permission)
	if err != nil {
		return OutputError(cmd, o, err)
	}

	var tmpl = `Permission revoked from {{ .Role }}: {{ .Permission }}.`
	return OutputSuccess(cmd, o, tmpl)
}
