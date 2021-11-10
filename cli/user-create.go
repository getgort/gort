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
	userCreateUse   = "create"
	userCreateShort = "Create a new user"
	userCreateLong  = "Create a new user."
	userCreateUsage = `Usage:
  gort user create [flags] user_name

Flags:
  -h, --help   Show this message and exit

Global Flags:
  -o, --output string    The output format: text (default), json, yaml
  -P, --profile string   The Gort profile within the config file to use
`
)

var (
	flagUserCreateEmail    string
	flagUserCreateName     string
	flagUserCreatePassword string
)

// GetUserCreateCmd is a command
func GetUserCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   userCreateUse,
		Short: userCreateShort,
		Long:  userCreateLong,
		RunE:  userCreateCmd,
		Args:  cobra.ExactArgs(1),
	}

	cmd.Flags().StringVarP(&flagUserCreateEmail, "email", "e", "", "Email for the user (required)")
	cmd.Flags().StringVarP(&flagUserCreateName, "name", "n", "", "Full name of the user (required)")
	cmd.Flags().StringVarP(&flagUserCreatePassword, "password", "p", "", "Password for user (required)")

	cmd.MarkFlagRequired("email")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("password")

	cmd.SetUsageTemplate(userCreateUsage)

	return cmd
}

func userCreateCmd(cmd *cobra.Command, args []string) error {
	username := args[0]

	o := struct {
		*CommandResult
		Exists bool
		User   rest.User
	}{CommandResult: &CommandResult{}}

	c, err := client.Connect(FlagGortProfile)
	if err != nil {
		return OutputError(cmd, o, err)
	}

	// Only allow this operation if the user doesn't already exist.
	o.Exists, err = c.UserExists(username)
	if err != nil {
		return OutputError(cmd, o, err)
	}
	if o.Exists {
		return OutputError(cmd, o, client.ErrResourceExists)
	}

	o.User = rest.User{
		Email:    flagUserCreateEmail,
		FullName: flagUserCreateName,
		Password: flagUserCreatePassword,
		Username: username,
	}

	err = c.UserSave(o.User)
	if err != nil {
		return OutputError(cmd, o, err)
	}

	if o.User.Password != "" {
		o.User.Password = "(redacted)"
	}

	tmpl := `User {{ .User.Username | quote }} created.`
	return OutputSuccess(cmd, o, tmpl)
}
