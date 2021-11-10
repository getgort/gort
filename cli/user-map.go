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
can only be mapped to one Gort user.`

	userMapUsage = `Usage:
  gort user map [flags] username adapter_name [chat_user_id]

Flags:
  -D, --delete   Delete a mapping instead of creating
  -h, --help     Show this message and exit

Global Flags:
  -o, --output string    The output format: text (default), json, yaml
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
	o := struct {
		*CommandResult
		Action   string
		Adapter  string
		Username string
		ChatID   string `json:",omitempty" yaml:",omitempty"`
	}{
		CommandResult: &CommandResult{},
		Adapter:       args[1],
		Username:      args[0],
	}

	if flagUserMapDelete {
		o.Action = "delete"
	} else {
		o.Action = "add"
	}

	var tmpl string

	if len(args) == 2 && !flagUserMapDelete {
		return OutputError(cmd, o, fmt.Errorf("chat provider user ID is missing"))
	}

	c, err := client.Connect(FlagGortProfile)
	if err != nil {
		return OutputError(cmd, o, err)
	}

	user, err := c.UserGet(o.Username)
	if err != nil {
		return OutputError(cmd, o, err)
	}

	if user.Mappings == nil {
		user.Mappings = map[string]string{}
	}

	if flagUserMapDelete {
		tmpl = `user {{ .Username | quote }} unmapped from {{ .Adapter | quote }}.`

		if user.Mappings[o.Adapter] == "" {
			return OutputError(cmd, o, fmt.Errorf("user %q doesn't have a mapping for provider %q", user.Username, o.Adapter))
		}

		delete(user.Mappings, o.Adapter)

		if err := c.UserSave(user); err != nil {
			return OutputError(cmd, o, err)
		}
	} else {
		o.ChatID = args[2]
		tmpl = `user {{ .Username | quote }} unmapped from {{ printf "%s:%s" .Adapter .ChatID | quote }}.`

		user.Mappings[o.Adapter] = o.ChatID

		if err := c.UserSave(user); err != nil {
			return OutputError(cmd, o, err)
		}
	}

	return OutputSuccess(cmd, o, tmpl)
}
