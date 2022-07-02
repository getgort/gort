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
	"strings"

	"github.com/getgort/gort/client"
	"github.com/spf13/cobra"
)

const (
	hiddenCommandUse   = "command"
	hiddenCommandShort = "Provides information about a command"
	hiddenCommandLong  = `Provides information about a command.

If no command is specified, this will list all commands installed in Gort.

If a command is specified, this will return information about the specified command.
`
	hiddenCommandUsage = `Usage:
  !gort:help [flags] [command]

Flags:
  -h, --help   Show this message and exit
`
)

// GetHiddenCommandCmd is a command
func GetHiddenCommandCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          hiddenCommandUse,
		Short:        hiddenCommandShort,
		Long:         hiddenCommandLong,
		RunE:         hiddenCommandCmd,
		Args:         cobra.RangeArgs(0, 1),
		SilenceUsage: true,
	}

	cmd.SetUsageTemplate(hiddenCommandUsage)

	return cmd
}

func hiddenCommandCmd(cmd *cobra.Command, args []string) error {
	gortClient, err := client.Connect(FlagGortProfile, FlagConfigBaseDir)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return listAllCommands(gortClient)
	}

	return detailCommand(gortClient, args[0])
}

func detailCommand(gortClient *client.GortClient, command string) error {
	bundles, err := gortClient.BundleList()
	if err != nil {
		return err
	}

	// If the user enters "bundle:command", match both.
	// Otherwise, match only "command"
	var bundleName, cmdName string
	if ss := strings.Split(command, ":"); len(ss) == 2 {
		bundleName = ss[0]
		cmdName = ss[1]
	} else if len(ss) == 1 {
		cmdName = ss[0]
	} else {
		return fmt.Errorf("invalid command syntax: expected <bundle:command> or <command>")
	}

	var found bool

	for _, b := range bundles {
		if !b.Enabled {
			continue
		}

		if bundleName != "" && bundleName != b.Name {
			continue
		}

		// If multiple commands are found, insert a space between them.
		if found {
			fmt.Println()
		}

		for k, v := range b.Commands {
			if cmdName != k {
				continue
			}

			fmt.Printf("%s:%s\n------------------------------\n", b.Name, k)

			if len(v.LongDescription) > 0 {
				fmt.Println(v.LongDescription)
			} else if len(v.Description) > 0 {
				fmt.Println(v.Description)
			}
			fmt.Println()
			fmt.Printf("Type `%s:%s --help` for more information.\n", b.Name, k)
			found = true
		}
	}

	if !found {
		return fmt.Errorf("command not found: %v", command)
	}

	return nil
}

func listAllCommands(gortClient *client.GortClient) error {
	bundles, err := gortClient.BundleList()
	if err != nil {
		return err
	}

	fmt.Printf("I know about these commands:\n\n")

	var cmds []string

	for _, b := range bundles {
		if !b.Enabled {
			continue
		}

		for k := range b.Commands {
			cmds = append(cmds, fmt.Sprintf("- %s:%s", b.Name, k))
		}
	}

	sort.Strings(cmds)

	for _, c := range cmds {
		fmt.Println(c)
	}

	return nil
}
