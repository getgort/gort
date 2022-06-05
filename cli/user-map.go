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
	userMapUse   = "map"
	userMapShort = "Map a Gort user to a specific chat user"
	userMapLong  = `Maps a Gort user to a specific user of a chat provider (such as Slack). This
allows the user to execute commands via that chat as the Gort user. Requires
the adapter name for the chat provider as defined in the Gort configuration,
and the provider's user ID.

For Slack this is the "member ID", NOT THE SLACK HANDLE. A user's member ID is
accessible in Slack via View Full Profile -> More -> Copy Member ID. It looks
something like U01234567AB.

A Gort user can only be mapped to one ID per adapter, and each adapter:ID pair
can only be mapped to one Gort user.
`
	userMapUsage = `Usage:
  gort user map [flags] username adapter_name [chat_user_id]

Flags:
  -D, --delete   Delete a mapping instead of creating
  -h, --help     Show this message and exit

Global Flags:
  -P, --profile string   The Gort profile within the config file to use
`
)

var (
	flagUserMapDelete bool
)

// GetUserMapCmd is a command
func GetUserMapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   userMapUse,
		Short: userMapShort,
		Long:  userMapLong,
		RunE:  userMapCmd,
		Args:  cobra.RangeArgs(2, 3),
	}

	cmd.Flags().BoolVarP(&flagUserMapDelete, "delete", "D", false, "Delete a mapping instead of creating")

	cmd.SetUsageTemplate(userMapUsage)

	return cmd
}

func userMapCmd(cmd *cobra.Command, args []string) error {
	username := args[0]
	adapter := args[1]

	if len(args) == 2 && !flagUserMapDelete {
		return fmt.Errorf("chat provider user ID is missing")
	}

	c, err := client.Connect(FlagGortProfile, FlagConfigBaseDir)
	if err != nil {
		return err
	}

	user, err := c.UserGet(username)
	if err != nil {
		return err
	}

	if user.Mappings == nil {
		user.Mappings = map[string]string{}
	}

	if flagUserMapDelete {
		if user.Mappings[adapter] == "" {
			return fmt.Errorf("user %q doesn't have a mapping for provider %q", user.Username, adapter)
		}

		delete(user.Mappings, adapter)

		if err := c.UserSave(user); err != nil {
			return err
		}

		fmt.Printf("User %q unmapped from %q.\n", user.Username, adapter)
	} else {
		chatID := args[2]
		user.Mappings[adapter] = chatID

		if err := c.UserSave(user); err != nil {
			return err
		}

		fmt.Printf("User %q mapped to \"%s:%s\".\n", user.Username, adapter, chatID)
	}

	return nil
}
