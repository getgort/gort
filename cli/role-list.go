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
	"sort"

	"github.com/getgort/gort/client"
	"github.com/getgort/gort/data/rest"

	"github.com/spf13/cobra"
)

const (
	roleListUse   = "list"
	roleListShort = "List all existing roles"
	roleListLong  = "List all existing roles."
	roleListUsage = `Usage:
  gort role list [flags]

Flags:
  -h, --help   Show this message and exit

Global Flags:
  -o, --output string    The output format: text (default), json, yaml
  -P, --profile string   The Gort profile within the config file to use
`
)

// GetRoleListCmd is a command
func GetRoleListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   roleListUse,
		Short: roleListShort,
		Long:  roleListLong,
		RunE:  roleListCmd,
	}

	cmd.SetUsageTemplate(roleListUsage)

	return cmd
}

func roleListCmd(cmd *cobra.Command, args []string) error {
	o := struct {
		*CommandResult
		Roles []rest.Role `json:",omitempty" yaml:",omitempty"`
	}{CommandResult: &CommandResult{}}

	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return OutputError(cmd, o, err)
	}

	o.Roles, err = gortClient.RoleList()
	if err != nil {
		return OutputError(cmd, o, err)
	}

	// Sort by name, for presentation purposes.
	sort.Slice(o.Roles, func(i, j int) bool { return o.Roles[i].Name < o.Roles[j].Name })

	tmpl := `ROLE NAME
{{ range $index, $role := .Roles }}{{ $role.Name }}{{end}}
`

	return OutputSuccess(cmd, o, tmpl)
}
