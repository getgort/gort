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
	"github.com/spf13/cobra"
)

// $ cogctl profile --help
// Usage: cogctl profile [OPTIONS] COMMAND [ARGS]...

// Manage Cog profiles.

// If invoked without a subcommand, lists all the profiles in the config
// file.

// Options:
// --help  Show this message and exit.

// Commands:
// create   Add a new profile to a the configuration...
// default  Sets the default profile in the configuration...

const (
	profileUse   = "profile"
	profileShort = "Manage Gort profiles"
	profileLong  = "Manage Gort profiles."
)

// GetProfileCmd profile
func GetProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   profileUse,
		Short: profileShort,
		Long:  profileLong,
	}

	cmd.AddCommand(GetProfileCreateCmd())
	cmd.AddCommand(GetProfileDefaultCmd())
	cmd.AddCommand(GetProfileDeleteCmd())
	cmd.AddCommand(GetProfileListCmd())

	return cmd
}
