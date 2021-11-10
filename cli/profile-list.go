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

	"github.com/spf13/cobra"
)

const (
	profileListUse   = "list"
	profileListShort = "List existing Gort user profiles"
	profileListLong  = "List existing Gort user profiles."
	profileListUsage = `Usage:
  gort profile list

Flags:
  -h, --help   Show this message and exit
`
)

// GetProfileListCmd is a command
func GetProfileListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   profileListUse,
		Short: profileListShort,
		Long:  profileListLong,
		RunE:  profileListCmd,
		Args:  cobra.ExactArgs(0),
	}

	cmd.SetUsageTemplate(profileListUsage)

	return cmd
}

func profileListCmd(cmd *cobra.Command, args []string) error {
	o := struct {
		*CommandResult
		Default  string
		Profiles []client.ProfileEntry
	}{CommandResult: &CommandResult{}}

	profile, err := client.LoadClientProfile()
	if err != nil {
		return OutputError(cmd, o, err)
	}

	if len(profile.Profiles) == 0 {
		message := "No profile file found.\nUse 'gort profile create' to create a new profile."
		return OutputErrorMessage(cmd, o, message)
	}

	o.Default = profile.Defaults.Profile
	for name, p := range profile.Profiles {
		p.Name = name
		p.Password = ""
		o.Profiles = append(o.Profiles, p)
	}

	// Sort by name, for presentation purposes.
	sort.Slice(o.Profiles, func(i, j int) bool { return o.Profiles[i].Name < o.Profiles[j].Name })

	c := &Columnizer{}
	c.StringColumn("NAME", func(i int) string { return o.Profiles[i].Name })
	c.StringColumn("USER", func(i int) string { return o.Profiles[i].Username })
	c.StringColumn("URL", func(i int) string { return o.Profiles[i].URL.String() })
	c.StringColumn("DEFAULT", func(i int) string {
		def := ""
		if o.Profiles[i].Name == o.Default {
			def = "   *"
		}
		return def
	})
	tmpl := strings.Join(c.Format(o.Profiles), "\n")

	if o.Default == "" {
		tmpl += "\nWARNING: No default profile set! Use 'gort profile default' to fix."
	}

	return OutputSuccess(cmd, o, tmpl)
}
