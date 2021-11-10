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

	"github.com/getgort/gort/client"
)

const (
	bootstrapUse   = "bootstrap"
	bootstrapShort = "Bootstrap a Gort server"
	bootstrapLong  = `Bootstrap can be used on a brand new server, addressable at the specified URL,
to create the administrative user ("admin"). Bootstrapping will only work on
an instance that doesn't yet have any users defined.

Following a successful bootstrapping, the returned password and user
information are added to gort's profiles file as a new profile. If this
is the first profile to be added to this configuration file, it will be marked
as the default.

By default, the new profile will be named for the hostname of the server being
bootstrapped. This can be overridden using the -P or --profile flags.`

	bootstrapUsage = `Usage:
  gort bootstrap [flags] [URL]

Flags:
  -i, --allow-insecure    Permit http URLs to be used
  -F, --force-overwrite   Overwrite the profile if it already exists
  -h, --help              help for bootstrap

Global Flags:
  -P, --profile string   The Gort profile within the config file to use
`
)

var (
	flagBootstrapAllowInsecure    bool
	flagBootstrapOverwriteProfile bool
)

// GetBootstrapCmd bootstrap
func GetBootstrapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   bootstrapUse,
		Short: bootstrapShort,
		Long:  bootstrapLong,
		RunE:  bootstrapCmd,
		Args:  cobra.ExactArgs(1),
	}

	cmd.Flags().BoolVarP(&flagBootstrapAllowInsecure, "allow-insecure", "i", false, "Permit http URLs to be used")
	cmd.Flags().BoolVarP(&flagBootstrapOverwriteProfile, "force-overwrite", "F", false, "Overwrite the profile if it already exists")

	cmd.SetUsageTemplate(bootstrapUsage)

	return cmd
}

func bootstrapCmd(cmd *cobra.Command, args []string) error {
	entry := client.ProfileEntry{
		Name:          FlagGortProfile,
		URLString:     args[0],
		AllowInsecure: flagBootstrapAllowInsecure,
	}

	gortClient, err := client.ConnectWithNewProfile(entry)
	if err != nil {
		return err
	}

	// Client Bootstrap will create the gort config if necessary, and append
	// the new credentials to it.
	user, err := gortClient.Bootstrap(flagBootstrapOverwriteProfile)
	if err != nil {
		return err
	}

	o := struct {
		User string
	}{
		User: user.Username,
	}

	template := `User {{ .User | quote }} created and credentials appended to Gort config.`
	if err := Output(FlagGortFormat, o, template); err != nil {
		return err
	}

	return nil
}
