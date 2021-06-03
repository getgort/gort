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

package bundles

import (
	"errors"
	"io/ioutil"
	"strings"

	"github.com/getgort/gort/data"
	gerrs "github.com/getgort/gort/errors"
	yaml "gopkg.in/yaml.v3"
)

var (
	// ErrInvalidBundleCommandPair is returned by FindCommandEntry when the
	// command entry string doesn't look like  "command" or "bundle:command".
	ErrInvalidBundleCommandPair = errors.New("invalid bundle:comand pair")
)

// LoadBundle is called by ...
func LoadBundle(file string) (data.Bundle, error) {
	// Read file as a byte slice
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		return data.Bundle{}, gerrs.Wrap(gerrs.ErrIO, err)
	}

	var bun data.Bundle

	err = yaml.Unmarshal(dat, &bun)
	if err != nil {
		return data.Bundle{}, gerrs.Wrap(gerrs.ErrUnmarshal, err)
	}

	// Ensure that the command name is propagated from the map key.
	for n := range bun.Commands {
		(bun.Commands[n]).Name = n
	}

	return bun, nil
}

// SplitCommand accepts a string in the style of "bundle:command" or "command"
// and returns the bundle and command as a pair of strings. If there's no
// indicated bundle, the bundle string (the first string) will be empty. If
// there's more than one colon, an error will be returned.
func SplitCommand(name string) (bundle, command string, err error) {
	split := strings.Split(name, ":")

	switch len(split) {
	case 1:
		command = split[0]
	case 2:
		bundle = split[0]
		command = split[1]
	default:
		err = ErrInvalidBundleCommandPair
	}

	return
}
