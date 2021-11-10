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
	"github.com/getgort/gort/version"

	"github.com/spf13/cobra"
)

const (
	versionUse   = "version"
	versionShort = "Display version and build information"
	versionLong  = "Displays version and build information."
)

var (
	flagVersionShort bool
)

// GetVersionCmd version
func GetVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   versionUse,
		Short: versionShort,
		Long:  versionLong,
		RunE:  versionCmd,
	}

	cmd.Flags().BoolVarP(&flagVersionShort, "short", "s", false, "Print only the version number")

	return cmd
}

func versionCmd(cmd *cobra.Command, args []string) error {
	o := struct {
		*CommandResult
		Version string `json:",omitempty" yaml:",omitempty"`
	}{
		CommandResult: &CommandResult{},
		Version:       version.Version,
	}

	var tmpl string

	if flagVersionShort {
		tmpl = `{{ .Version }}`
	} else {
		tmpl = `Gort ChatOps Engine v{{ .Version }}`
	}

	return OutputSuccess(cmd, o, tmpl)
}
