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
	"errors"
	"strings"

	"github.com/clockworksoul/gort/data"
)

var (
	// ErrInvalidBundleCommandPair is returned by FindCommandEntry when the
	// command entry string doesn't look like  "command" or "bundle:command".
	ErrInvalidBundleCommandPair = errors.New("invalid bundle:comand pair")
)

// FindCommandEntry looks for a command in the configuration. It assumes that
// command character(s) have already been removed, and expects a string in the
// format "bundle:command" or "command"; the latter can return multiple values
// if a similarly-named command is found in multiple bundles
func FindCommandEntry(name string) ([]data.CommandEntry, error) {
	var bundleName string
	var commandName string

	split := strings.Split(name, ":")

	switch len(split) {
	case 1:
		bundleName = "*"
		commandName = split[0]
	case 2:
		bundleName = split[0]
		commandName = split[1]
	default:
		return nil, ErrInvalidBundleCommandPair
	}

	entries := make([]data.CommandEntry, 0)

	for _, bundle := range GetBundleConfigs() {
		if bundleName != bundle.Name && bundleName != "*" {
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
