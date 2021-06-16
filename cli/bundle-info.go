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
	"strings"

	"github.com/getgort/gort/client"
	"github.com/getgort/gort/data"
	"github.com/spf13/cobra"
)

const (
	bundleInfoUse   = "info"
	bundleInfoShort = "Info a bundle"
	bundleInfoLong  = `
Display bundle information.

If only a bundle name is provided, information on the bundle as a whole is
presented. If that bundle is also currently enabled, details about the
version that is currently live is also displayed.

If a version is also provided, details on that specific version are
presented, regardless of whether it happens to also be enabled.
`
)

// GetBundleInfoCmd is a command
func GetBundleInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   bundleInfoUse,
		Short: bundleInfoShort,
		Long:  bundleInfoLong,
		RunE:  bundleInfoCmd,
		Args:  cobra.RangeArgs(1, 2),
	}

	return cmd
}

func bundleInfoCmd(cmd *cobra.Command, args []string) error {
	switch len(args) {
	case 1:
		return doBundleInfoAll(args[0])
	case 2:
		return doBundleInfoVersion(args[0], args[1])
	}

	return nil
}

func doBundleInfoAll(name string) error {
	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	bundles, err := gortClient.BundleListVersions(name)
	if err != nil {
		return err
	}

	var enabled *data.Bundle
	var versions = make([]string, 0)

	for _, bundle := range bundles {
		versions = append(versions, bundle.Version)

		if bundle.Enabled {
			enabled = &bundle
		}
	}

	fmt.Printf("Name: %s\n", name)
	fmt.Printf("Versions: %s\n", strings.Join(versions, ", "))

	if enabled != nil {
		fmt.Println("Status: Enabled")
		fmt.Printf("Enabled Version: %s\n", enabled.Version)

		commands := make([]string, 0)
		for name := range enabled.Commands {
			commands = append(commands, name)
		}

		fmt.Printf("Commands: %s\n", strings.Join(commands, ", "))
		fmt.Printf("Permissions: %s\n", strings.Join(enabled.Permissions, ", "))
	} else {
		fmt.Println("Status: Disabled")
	}

	return nil
}

func doBundleInfoVersion(name, version string) error {
	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	bundle, err := gortClient.BundleGet(name, version)
	if err != nil {
		return err
	}

	fmt.Printf("Name: %s\n", bundle.Name)
	fmt.Printf("Version: %s\n", bundle.Version)

	if bundle.Enabled {
		fmt.Println("Status: Enabled")
	} else {
		fmt.Println("Status: Enabled")
	}

	commands := make([]string, 0)
	for name := range bundle.Commands {
		commands = append(commands, name)
	}

	fmt.Printf("Commands: %s\n", strings.Join(commands, ", "))
	fmt.Printf("Permissions: %s\n", strings.Join(bundle.Permissions, ", "))

	return nil
}
