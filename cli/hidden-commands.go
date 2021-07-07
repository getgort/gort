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

	"github.com/getgort/gort/client"
	"github.com/spf13/cobra"
)

const (
	hiddenCommandsUse   = "commands"
	hiddenCommandsShort = "List all available commands"
	hiddenCommandsLong  = "Lists all available commands."
)

// GetHiddenCommandsCmd is a command
func GetHiddenCommandsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   hiddenCommandsUse,
		Short: hiddenCommandsShort,
		Long:  hiddenCommandsLong,
		RunE:  hiddenCommandsCmd,
	}

	return cmd
}

func hiddenCommandsCmd(cmd *cobra.Command, args []string) error {
	gortClient, err := client.Connect(FlagGortProfile)
	if err != nil {
		return err
	}

	bundles, err := gortClient.BundleList()
	if err != nil {
		return err
	}

	fmt.Printf("I know about these commands:\n\n")

	cmds := []string{}

	for _, b := range bundles {
		for k, _ := range b.Commands {
			cmds = append(cmds, fmt.Sprintf("%s:%s", b.Name, k))
		}
	}

	sort.Strings(cmds)

	for _, c := range cmds {
		fmt.Println(c)
	}

	return nil
}
