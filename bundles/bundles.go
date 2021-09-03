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
	"io"
	"io/ioutil"
	"os"

	"github.com/getgort/gort/data"
	gerrs "github.com/getgort/gort/errors"
	yaml "gopkg.in/yaml.v3"
)

var (
	// ErrInvalidBundleCommandPair is returned by FindCommandEntry when the
	// command entry string doesn't look like  "command" or "bundle:command".
	ErrInvalidBundleCommandPair = errors.New("invalid bundle:comand pair")
)

func LoadBundleFromFile(file string) (data.Bundle, error) {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return data.Bundle{}, err
	}
	return LoadBundle(f)
}

// LoadBundle is called by ...
func LoadBundle(r io.Reader) (data.Bundle, error) {
	dat, err := ioutil.ReadAll(r)
	if err != nil {
		return data.Bundle{}, gerrs.Wrap(gerrs.ErrIO, err)
	}

	return unmarshal(dat)
}

func unmarshal(dat []byte) (data.Bundle, error) {
	var bun data.Bundle

	err := yaml.Unmarshal(dat, &bun)
	if err != nil {
		return data.Bundle{}, gerrs.Wrap(gerrs.ErrUnmarshal, err)
	}

	// Ensure that the command name is propagated from the map key.
	for n := range bun.Commands {
		(bun.Commands[n]).Name = n
	}

	return bun, nil
}
