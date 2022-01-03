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

package main

import (
	"github.com/spf13/cobra"

	"github.com/getgort/gort/cli"
)

const (
	rootUse   = "gort"
	rootShort = "Bringing the power of the command line to chat"
	rootLong  = "Bringing the power of the command line to chat."
)

// GetRootCmd root
func GetRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:          rootUse,
		Short:        rootShort,
		Long:         rootLong,
		SilenceUsage: true,
	}

	root.AddCommand(GetStartCmd())
	root.AddCommand(cli.GetBootstrapCmd())
	root.AddCommand(cli.GetBundleCmd())
	root.AddCommand(cli.GetConfigCmd())
	root.AddCommand(cli.GetGroupCmd())
	root.AddCommand(cli.GetHiddenCmd())
	root.AddCommand(cli.GetPermissionCmd())
	root.AddCommand(cli.GetProfileCmd())
	root.AddCommand(cli.GetRoleCmd())
	root.AddCommand(cli.GetUserCmd())
	root.AddCommand(cli.GetVersionCmd())

	root.PersistentFlags().StringVarP(&cli.FlagGortProfile, "profile", "P", "", "The Gort profile within the config file to use")

	return root
}
