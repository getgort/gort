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

package config

import (
	"context"

	"github.com/getgort/gort/bundles"
	"github.com/getgort/gort/data"
)

// ConfigCommandEntryFinder just has a FindCommandEntry() function so that
// config can provide functionality that's compliant with the
// bundles.CommandEntryFinder interface.
type ConfigCommandEntryFinder struct{}

// FindCommandEntry looks for a command in the configuration. It assumes that
// command character(s) have already been removed, and expects a string in the
// format "bundle:command" or "command"; the latter can return multiple values
// if a similarly-named command is found in multiple bundles.
func (c ConfigCommandEntryFinder) FindCommandEntry(ctx context.Context, bundleName, commandName string) ([]data.CommandEntry, error) {
	entries := make([]data.CommandEntry, 0)

	for _, bundle := range GetBundleConfigs() {
		if bundleName != bundle.Name && bundleName != "" {
			continue
		}

		for name, command := range bundle.Commands {
			if name == commandName {
				command.Name = name
				entries = append(entries, data.CommandEntry{
					Bundle:  bundle,
					Command: *command,
				})
			}
		}
	}

	return entries, nil
}

// CommandEntryFinder returns a bundles.CommandEntryFinder implementation that
// allows interrogation of the bundles described in the config in a way that's
// compliant with the bundles.CommandEntryFinder interface.
func CommandEntryFinder() bundles.CommandEntryFinder {
	return ConfigCommandEntryFinder{}
}
