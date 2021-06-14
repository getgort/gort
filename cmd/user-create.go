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
	"github.com/getgort/gort/data/rest"
	"github.com/spf13/cobra"
)

const (
	userCreateUse   = "create"
	userCreateShort = "Create a new user"
	userCreateLong  = "Create a new user."
)

var (
	flagUserCreateEmail    string
	flagUserCreateName     string
	flagUserCreatePassword string
	flagUserCreateProfile  string
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

	return cmd
}

func userCreateCmd(cmd *cobra.Command, args []string) error {
	username := args[0]

	c, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	// Only allow this operation if the user doesn't already exist.
	exists, err := c.UserExists(username)
	if err != nil {
		return err
	}
	if exists {
		return client.ErrResourceExists
	}

	user := rest.User{
		Email:    flagUserCreateEmail,
		FullName: flagUserCreateName,
		Password: flagUserCreatePassword,
		Username: username,
	}

	err = c.UserSave(user)
	if err != nil {
		return err
	}

	fmt.Printf("User %q created.\n", user.Username)

	return nil
}
