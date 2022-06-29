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
	profileDeleteUse   = "delete"
	profileDeleteShort = "Delete an existing Gort user profile"
	profileDeleteLong  = "Delete an existing Gort user profile."
	profileDeleteUsage = `Usage:
  gort profile delete [flags] profile_name

Flags:
  -h, --help   Show this message and exit
`
)

// GetProfileDeleteCmd is a command
func GetProfileDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   profileDeleteUse,
		Short: profileDeleteShort,
		Long:  profileDeleteLong,
		RunE:  profileDeleteCmd,
		Args:  cobra.ExactArgs(1),
	}

	cmd.SetUsageTemplate(profileDeleteUsage)

	return cmd
}

func profileDeleteCmd(cmd *cobra.Command, args []string) error {
	profile, err := client.LoadClientProfile(FlagConfigBaseDir)
	if err != nil {
		fmt.Println("Failed to load existing profiles:", err)
		return nil
	}

	if len(profile.Profiles) == 0 {
		fmt.Println("No profile file found.")
		fmt.Println("Use 'gort profile create' to create a new profile.")
		return nil
	}

	name := args[0]

	if _, exists := profile.Profiles[name]; !exists {
		fmt.Printf("Profile '%s' doesn't exist.\n", name)
		return nil
	}

	delete(profile.Profiles, name)

	if profile.Defaults.Profile == name {
		fmt.Println("WARNING: Deleting default profile!")
		profile.Defaults.Profile = ""
	}

	err = client.SaveClientProfile(profile, FlagConfigBaseDir)
	if err != nil {
		fmt.Printf("Failed to update profile: %s\n", err.Error())
		return nil
	}

	fmt.Printf("Profile '%s' deleted.\n", name)

	return nil
}
