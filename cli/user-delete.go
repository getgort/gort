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
	userDeleteUse   = "delete"
	userDeleteShort = "Deletes an existing user"
	userDeleteLong  = "Deletes an existing user."
	userDeleteUsage = `Usage:
  gort user delete [flags] user_name

Flags:
  -h, --help   Show this message and exit

Global Flags:
  -o, --output string    The output format: text (default), json, yaml
  -P, --profile string   The Gort profile within the config file to use
`
)

// GetUserDeleteCmd is a command
func GetUserDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   userDeleteUse,
		Short: userDeleteShort,
		Long:  userDeleteLong,
		RunE:  userDeleteCmd,
		Args:  cobra.ExactArgs(1),
	}

	cmd.SetUsageTemplate(userDeleteUsage)

	return cmd
}

func userDeleteCmd(cmd *cobra.Command, args []string) error {
	o := struct {
		*CommandResult
		User rest.User `json:",omitempty" yaml:",omitempty"`
	}{CommandResult: &CommandResult{}}

	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return OutputError(cmd, o, err)
	}

	o.User, err = gortClient.UserGet(args[0])
	if err != nil {
		return OutputError(cmd, o, err)
	}

	// This isn't strictly necessary, but you can't be too careful.
	o.User.Password = ""

	err = gortClient.UserDelete(o.User.Username)
	if err != nil {
		return OutputError(cmd, o, err)
	}

	var tmpl = `Deleted user {{ .User.Username | quote }} ({{ .Email }}).`
	return OutputSuccess(cmd, o, tmpl)
}
