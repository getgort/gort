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
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/getgort/gort/cli"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	rootUse   = "gort"
	rootShort = "Bringing the power of the command line to chat"
	rootLong  = "Bringing the power of the command line to chat."
)

// GetRootCmd root
func GetRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   rootUse,
		Short: rootShort,
		Long:  rootLong,
		// SilenceErrors: true,
		SilenceUsage: true,
	}

	// This makes sure that flag errors are still output.
	root.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	})

	root.AddCommand(GetStartCmd())
	root.AddCommand(cli.GetBootstrapCmd())
	root.AddCommand(cli.GetBundleCmd())
	root.AddCommand(cli.GetGroupCmd())
	root.AddCommand(cli.GetHiddenCmd())
	root.AddCommand(cli.GetPermissionCmd())
	root.AddCommand(cli.GetProfileCmd())
	root.AddCommand(cli.GetRoleCmd())
	root.AddCommand(cli.GetUserCmd())
	root.AddCommand(cli.GetVersionCmd())

	root.PersistentFlags().StringVarP(&cli.FlagGortProfile, "profile", "P", "", "The Gort profile within the config file to use")
	root.PersistentFlags().StringVarP(&cli.FlagGortFormat, "output", "o", "text", "The output format: text (default), json, yaml")

	root.SetErr(errorWriter{})

	return root
}

// errorWriter takes the place of os.Stderr as the root command's error message
// destination. If the command output format is "json" or "yaml", it constructs
// an error value and marshals that into the appropriate format to output.
type errorWriter struct{}

func (w errorWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimSpace(string(p))
	if strings.HasPrefix(msg, "Error: ") {
		msg = strings.Replace(msg, "Error: ", "", 1)
	}

	o := struct{ Error string }{Error: msg}

	var text string

	switch f := strings.ToLower(cli.FlagGortFormat); f {
	case "json":
		b, err := json.MarshalIndent(o, "", "  ")
		if err != nil {
			return 0, fmt.Errorf("failed to marshal as json: %w", err)
		}
		text = string(b)

	case "yaml":
		b, err := yaml.Marshal(o)
		if err != nil {
			return 0, fmt.Errorf("failed to marshal as yaml: %w", err)
		}
		text = string(b)

	case "text":
		text = "Error: " + o.Error

	default:
		return 0, fmt.Errorf("unsupported output format: %s", f)
	}

	return fmt.Fprintln(os.Stderr, strings.TrimSpace(text))
}
