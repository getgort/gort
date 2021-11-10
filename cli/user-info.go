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
	"strings"

	"github.com/getgort/gort/client"
	"github.com/getgort/gort/data/rest"
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
	gortClient, err := client.Connect(FlagGortProfile)
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

	o := struct {
		User   rest.User
		Groups []rest.Group `json:"groups"`
	}{
		User:   user,
		Groups: groups,
	}

	tmpl := `Name       {{ if .User.Username }}{{ .User.Username }}{{ else }}<undefined>{{ end }}
Full Name  {{ if .User.FullName }}{{ .User.FullName }}{{ else }}<undefined>{{ end }}
Email      {{ if .User.Email }}{{ .User.Email }}{{ else }}<undefined>{{ end }}
Groups    {{ range $index, $group := .Groups }} {{ $group.Name}}{{ end }}

`

	if len(user.Mappings) == 0 {
		tmpl += "This user has no chat provider mappings. Use 'gort user map' to map a Gort\n" +
			"user to one or more chat provider IDs."
	} else {
		var keys []string
		for k, _ := range user.Mappings {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		c := &Columnizer{}
		c.StringColumn("ADAPTER", func(i int) string { return keys[i] })
		c.StringColumn("ID MAPPING", func(i int) string { return user.Mappings[keys[i]] })
		tmpl += strings.Join(c.Format(keys), "\n")
	}

	if err := Output(FlagGortFormat, o, tmpl); err != nil {
		return err
	}

	return nil
}
