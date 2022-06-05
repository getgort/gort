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
	"github.com/getgort/gort/data"
	"github.com/spf13/cobra"
)

// $ cogctl bundle --help
// Usage: cogctl bundle [OPTIONS] COMMAND [ARGS]...
//
// Manage command bundles and their config.
//
// If no subcommand is given, lists all bundles installed, and their
// currently enabled version, if any.
//
// Options:
// -e, --enabled   List only enabled bundles
// -d, --disabled  List only disabled bundles
// -v, --verbose   Display additional bundle details
// --help          Show this message and exit.

const (
	bundleListUse   = "list"
	bundleListShort = "List all bundles installed"
	bundleListLong  = "Lists all bundles installed, and their currently enabled version, if any."
	bundleListUsage = `Usage:
  gort bundle list [flags]

Flags:
  -d, --disabled   List only disabled bundles
  -e, --enabled    List only enabled bundles
  -h, --help       help for list
  -v, --verbose    Display additional bundle details

Global Flags:
  -P, --profile string   The Gort profile within the config file to use
`
)

var (
	flagBundleListEnabled  bool
	flagBundleListDisabled bool
	flagBundleListVerbose  bool
)

// GetBundleListCmd is a command
func GetBundleListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   bundleListUse,
		Short: bundleListShort,
		Long:  bundleListLong,
		RunE:  bundleListCmd,
	}

	cmd.Flags().BoolVarP(&flagBundleListEnabled, "enabled", "e", false, "List only enabled bundles")
	cmd.Flags().BoolVarP(&flagBundleListDisabled, "disabled", "d", false, "List only disabled bundles")
	cmd.Flags().BoolVarP(&flagBundleListVerbose, "verbose", "v", false, "Display additional bundle details")

	cmd.SetUsageTemplate(bundleListUsage)

	return cmd
}

type bundleData struct {
	name           string
	enabled        bool
	enabledVersion string
	versions       []string
}

func bundleListCmd(cmd *cobra.Command, args []string) error {
	if flagBundleListEnabled && flagBundleListDisabled {
		return fmt.Errorf("--enabled and --disabled flags are mutually exclusive")
	}

	gortClient, err := client.Connect(FlagGortProfile, FlagConfigBaseDir)
	if err != nil {
		return err
	}

	bundles, err := gortClient.BundleList()
	if err != nil {
		return err
	}

	metadata := getBundleData(bundles)

	switch {
	case flagBundleListEnabled:
		metadata = filterBundleData(metadata, func(b bundleData) bool {
			return b.enabled
		})
	case flagBundleListDisabled:
		metadata = filterBundleData(metadata, func(b bundleData) bool {
			return !b.enabled
		})
	}

	c := &Columnizer{}
	c.StringColumn("BUNDLE", func(i int) string { return metadata[i].name })
	c.StringColumn("ENABLED", func(i int) string {
		version := metadata[i].enabledVersion
		if version == "" {
			version = "-"
		}
		return version
	})

	if flagBundleListVerbose {
		c.StringColumn("INSTALLED VERSIONS", func(i int) string {
			return strings.Join(metadata[i].versions, ", ")
		})
	}

	c.Print(metadata)

	return nil
}

func getBundleData(bundles []data.Bundle) []bundleData {
	m := map[string]bundleData{}
	for _, b := range bundles {
		d := m[b.Name]
		d.name = b.Name
		if b.Enabled {
			d.enabled = true
			d.enabledVersion = b.Version
		}
		d.versions = append(d.versions, b.Version)
		m[b.Name] = d
	}

	var bd []bundleData
	for _, b := range m {
		bd = append(bd, b)
	}

	sort.Slice(bd, func(i, j int) bool { return bd[i].name < bd[j].name })

	return bd
}

// If filter(ss[i]) returns false for an element, that element is filtered out.
func filterBundleData(in []bundleData, filter func(bundleData) bool) []bundleData {
	var out []bundleData

	for _, b := range in {
		if filter(b) {
			out = append(out, b)
		}
	}

	return out
}
