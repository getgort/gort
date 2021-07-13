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
	"net/url"

	"github.com/getgort/gort/client"
	"github.com/spf13/cobra"
)

// $ cogctl profile create --help
// Usage: cogctl profile create [OPTIONS] NAME URL USER PASSWORD
//
// Add a new profile to a the configuration file.
//
// Options:
// --help  Show this message and exit.

const (
	profileCreateUse   = "create"
	profileCreateShort = "Create a new Gort user profile"
	profileCreateLong  = "Adds a new profile with the given name for the specified Gort server."
	profileCreateUsage = `Usage:
  gort profile create [flags] profile_name url user password

Flags:
  -h, --help   Show this message and exit
`
)

// GetProfileCreateCmd is a command
func GetProfileCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   profileCreateUse,
		Short: profileCreateShort,
		Long:  profileCreateLong,
		RunE:  profileCreateCmd,
		Args:  cobra.ExactArgs(4),
	}

	cmd.SetUsageTemplate(profileCreateUsage)

	return cmd
}

func profileCreateCmd(cmd *cobra.Command, args []string) error {
	profile, err := client.LoadClientProfile()
	if err != nil {
		fmt.Println("Failed to load existing profiles:", err)
		return nil
	}

	if len(profile.Profiles) == 0 {
		fmt.Println("No profile file found. Creating.")
	}

	name := args[0]
	urlstring := args[1]
	user := args[2]
	password := args[3]

	if _, exists := profile.Profiles[name]; exists {
		fmt.Printf("Profile '%s' already exists.\n", name)
		return nil
	}

	furl, err := url.Parse(urlstring)
	if err != nil {
		fmt.Printf("Failed to parse URL '%s': %s\n", urlstring, err.Error())
		return nil
	}

	pe := client.ProfileEntry{
		Name:      name,
		URLString: furl.String(),
		Password:  password,
		Username:  user,
	}

	profile.Profiles[name] = pe

	if profile.Defaults.Profile == "" {
		profile.Defaults.Profile = pe.Name
	}

	err = client.SaveClientProfile(profile)
	if err != nil {
		fmt.Printf("Failed to update profile: %s\n", err.Error())
		return nil
	}

	fmt.Printf("Profile '%s' (%s@%s) created.\n", pe.Name, pe.Username, pe.URLString)

	return nil
}
