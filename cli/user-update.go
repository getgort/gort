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
	userUpdateUse   = "update"
	userUpdateShort = "Update an existing user"
	userUpdateLong  = "Update an existing user."
)

var (
	flagUserUpdateEmail    string
	flagUserUpdateName     string
	flagUserUpdatePassword string
	flagUserUpdateProfile  string
)

// GetUserUpdateCmd is a command
func GetUserUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   userUpdateUse,
		Short: userUpdateShort,
		Long:  userUpdateLong,
		RunE:  userUpdateCmd,
		Args:  cobra.ExactArgs(1),
	}

	cmd.Flags().StringVarP(&flagUserUpdateEmail, "email", "e", "", "Email for the user")
	cmd.Flags().StringVarP(&flagUserUpdateName, "name", "n", "", "Full name of the user")
	cmd.Flags().StringVarP(&flagUserUpdatePassword, "password", "p", "", "Password for user")

	return cmd
}

func userUpdateCmd(cmd *cobra.Command, args []string) error {
	username := args[0]

	c, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	// Only allow this operation if the user already exists.
	exists, err := c.UserExists(username)
	if err != nil {
		return err
	}
	if !exists {
		return client.ErrResourceNotFound
	}

	// Empty fields will not be overwritten.
	user := rest.User{
		Email:    flagUserUpdateEmail,
		FullName: flagUserUpdateName,
		Password: flagUserUpdatePassword,
		Username: username,
	}

	err = c.UserSave(user)
	if err != nil {
		return err
	}

	fmt.Printf("User %q updated.\n", user.Username)

	return nil
}
