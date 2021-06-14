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

package cmd

import (
	"fmt"

	"github.com/getgort/gort/bundles"
	"github.com/getgort/gort/client"
	"github.com/spf13/cobra"
)

const (
	bundleInstallUse   = "install"
	bundleInstallShort = "Install a bundle"
	bundleInstallLong  = "Install a bundle."
)

// GetBundleInstallCmd is a command
func GetBundleInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   bundleInstallUse,
		Short: bundleInstallShort,
		Long:  bundleInstallLong,
		RunE:  bundleInstallCmd,
		Args:  cobra.ExactArgs(1),
	}

	return cmd
}

func bundleInstallCmd(cmd *cobra.Command, args []string) error {
	bundlefile := args[0]

	c, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	bundle, err := bundles.LoadBundle(bundlefile)
	if err != nil {
		return err
	}

	err = c.BundleInstall(bundle)
	if err != nil {
		return err
	}

	fmt.Printf("Bundle %q installed.\n", bundle.Name)

	return nil
}
