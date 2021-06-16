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
	"github.com/spf13/cobra"
)

const (
	userDeleteUse   = "delete"
	userDeleteShort = "Delete an existing user"
	userDeleteLong  = "Delete an existing user."
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

	return cmd
}

func userDeleteCmd(cmd *cobra.Command, args []string) error {
	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	username := args[0]

	user, err := gortClient.UserGet(username)
	if err != nil {
		return err
	}

	fmt.Printf("Deleting user %s (%s)... ", user.Username, user.Email)

	err = gortClient.UserDelete(user.Username)
	if err != nil {
		return err
	}

	fmt.Println("Successful.")

	return nil
}
