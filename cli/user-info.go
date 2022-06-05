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
	"sort"
	"strings"

	"github.com/getgort/gort/client"
	"github.com/spf13/cobra"
)

const (
	userInfoUse   = "info"
	userInfoShort = "Retrieve information about an existing user"
	userInfoLong  = "Retrieve information about an existing user."
	userInfoUsage = `Usage:
  gort user info [flags] user_name [version]

Flags:
  -h, --help   Show this message and exit

Global Flags:
  -P, --profile string   The Gort profile within the config file to use
`
)

// GetUserInfoCmd is a command
func GetUserInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   userInfoUse,
		Short: userInfoShort,
		Long:  userInfoLong,
		RunE:  userInfoCmd,
		Args:  cobra.ExactArgs(1),
	}

	cmd.SetUsageTemplate(userInfoUsage)

	return cmd
}

func userInfoCmd(cmd *cobra.Command, args []string) error {
	gortClient, err := client.Connect(FlagGortProfile, FlagConfigBaseDir)
	if err != nil {
		return err
	}

	//
	// TODO Maybe multiplex the following queries with goroutines?
	//

	username := args[0]

	user, err := gortClient.UserGet(username)
	if err != nil {
		return err
	}

	groups, err := gortClient.UserGroupList(username)
	if err != nil {
		return err
	}

	const format = `Name       %s
Full Name  %s
Email      %s
Groups     %s

`

	fmt.Printf(format,
		process(user.Username),
		process(user.FullName),
		process(user.Email),
		process(groupNames(groups)),
	)

	if len(user.Mappings) == 0 {
		fmt.Println("This user has no chat provider mappings. Use 'gort user map' to map a Gort\n" +
			"user to one or more chat provider IDs.")
	} else {
		var keys []string
		for k := range user.Mappings {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		c := &Columnizer{}
		c.StringColumn("ADAPTER", func(i int) string { return keys[i] })
		c.StringColumn("ID MAPPING", func(i int) string { return user.Mappings[keys[i]] })
		c.Print(keys)
	}

	return nil
}

func process(i interface{}) string {
	switch v := i.(type) {
	case string:
		if v == "" {
			return "<undefined>"
		}

		return v
	case []string:
		if len(v) == 0 {
			return "<undefined>"
		}

		return strings.Join(v, ", ")
	default:
		return fmt.Sprintf("<unexpected type: %T>", i)
	}
}
