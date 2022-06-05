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
	bundleDisableUse   = "disable"
	bundleDisableShort = "Disable a bundle by name"
	bundleDisableLong  = "Disable a bundle by name."
	bundleDisableUsage = `Usage:
  gort bundle disable [flags] bundle_name

Flags:
  --help  Show this message and exit.

Global Flags:
  -P, --profile string   The Gort profile within the config file to use
`
)

// GetBundleDisableCmd is a command
func GetBundleDisableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   bundleDisableUse,
		Short: bundleDisableShort,
		Long:  bundleDisableLong,
		RunE:  bundleDisableCmd,
		Args:  cobra.ExactArgs(1),
	}

	cmd.SetUsageTemplate(bundleDisableUsage)

	return cmd
}

func bundleDisableCmd(cmd *cobra.Command, args []string) error {
	bundleName := args[0]

	c, err := client.Connect(FlagGortProfile, FlagConfigBaseDir)
	if err != nil {
		return err
	}

	err = c.BundleDisable(bundleName)
	if err != nil {
		return err
	}

	fmt.Printf("Bundle \"%s\" disabled.\n", bundleName)

	return nil
}
