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
	o := struct {
		*CommandResult
		Profile client.ProfileEntry `json:",omitempty" yaml:",omitempty"`
	}{
		CommandResult: &CommandResult{},
		Profile: client.ProfileEntry{
			Name:     args[0],
			Password: args[2],
			Username: args[3],
		},
	}

	profile, err := client.LoadClientProfile()
	if err != nil {
		return OutputError(cmd, o, err)
	}

	if _, exists := profile.Profiles[o.Profile.Name]; exists {
		message := "Profile already exists."
		return OutputErrorMessage(cmd, o, message)
	}

	o.Profile.URL, err = url.Parse(args[1])
	if err != nil {
		return OutputError(cmd, o, fmt.Errorf("failed to parse url %q: %w", args[1], err))
	}
	o.Profile.URLString = o.Profile.URL.String()

	profile.Profiles[o.Profile.Name] = o.Profile

	if profile.Defaults.Profile == "" {
		profile.Defaults.Profile = o.Profile.Name
	}

	err = client.SaveClientProfile(profile)
	if err != nil {
		return OutputError(cmd, o, err)
	}

	tmpl := `Profile {{ .Profile | quote }} ({{ .Profile.Username }}@{{ .Profile.URLString }}) created.`
	return OutputSuccess(cmd, o, tmpl)
}
