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
	profile, err := client.LoadClientProfile()
	if err != nil {
		fmt.Println("Failed to load existing profiles:", err)
		return nil
	}

	if len(profile.Profiles) == 0 {
		fmt.Println("No profile file found.")
		fmt.Println("Use 'gort profile create' to create a new profile.")
		return nil
	}

	profiles := []client.ProfileEntry{}
	for name, p := range profile.Profiles {
		p.Name = name
		profiles = append(profiles, p)

	}

	// Sort by name, for presentation purposes.
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].Name < profiles[j].Name })

	c := &Columnizer{}
	c.StringColumn("NAME", func(i int) string { return profiles[i].Name })
	c.StringColumn("USER", func(i int) string { return profiles[i].Username })
	c.StringColumn("URL", func(i int) string { return profiles[i].URL.String() })
	c.StringColumn("DEFAULT", func(i int) string {
		def := ""
		if profiles[i].Name == profile.Defaults.Profile {
			def = "   *"
		}
		return def
	})
	c.Print(profiles)

	if profile.Defaults.Profile == "" {
		fmt.Println("\nWARNING: No default profile set! Use 'gort profile default' to fix.")
	}

	return nil
}
