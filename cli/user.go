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

// # gort user --help
// Usage: gort user [OPTIONS] COMMAND [ARGS]...
//
//   Manage Gort users.
//
//   If invoked without a subcommand, lists all the users on the server.
//
// Options:
//   --help  Show this message and exit.
//
// Commands:
//   create                  Create a new user.
//   delete                  Deletes a user.
//   info                    Get info about a specific user by username.
//   password-reset          Reset user password with a token.
//   password-reset-request  Request a password reset.
//   update                  Updates an existing user.

const (
	userUse   = "user"
	userShort = "Perform operations on users"
	userLong  = "Allows you to perform user administration."
)

// GetUserCmd user
func GetUserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   userUse,
		Short: userShort,
		Long:  userLong,
	}

	cmd.AddCommand(GetUserCreateCmd())
	cmd.AddCommand(GetUserDeleteCmd())
	cmd.AddCommand(GetUserInfoCmd())
	cmd.AddCommand(GetUserListCmd())
	cmd.AddCommand(GetUserMapCmd())
	cmd.AddCommand(GetUserUpdateCmd())

	return cmd
}
