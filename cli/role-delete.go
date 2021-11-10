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
	"github.com/getgort/gort/data/rest"

	"github.com/spf13/cobra"
)

const (
	roleDeleteUse   = "delete"
	roleDeleteShort = "Delete an existing role"
	roleDeleteLong  = "Delete an existing role."
	roleDeleteUsage = `Usage:
  gort role delete [flags] role_name

Flags:
  -h, --help   Show this message and exit

Global Flags:
  -o, --output string    The output format: text (default), json, yaml
  -P, --profile string   The Gort profile within the config file to use
`
)

// GetroleDeleteCmd is a command
func GetRoleDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   roleDeleteUse,
		Short: roleDeleteShort,
		Long:  roleDeleteLong,
		RunE:  roleDeleteCmd,
		Args:  cobra.ExactArgs(1),
	}

	cmd.SetUsageTemplate(roleDeleteUsage)

	return cmd
}

func roleDeleteCmd(cmd *cobra.Command, args []string) error {
	o := struct {
		*CommandResult
		Role rest.Role `json:",omitempty" yaml:",omitempty"`
	}{CommandResult: &CommandResult{}}

	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	rolename := args[0]

	o.Role, err = gortClient.RoleGet(rolename)
	if err != nil {
		return OutputError(cmd, o, err)
	}

	err = gortClient.RoleDelete(o.Role.Name)
	if err != nil {
		return OutputError(cmd, o, err)
	}

	var tmpl = `Deleted user {{ .Role.Name | quote }}.`
	return OutputSuccess(cmd, o, tmpl)
}
