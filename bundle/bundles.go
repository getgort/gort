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

package bundle

import (
	"io/ioutil"

	"github.com/clockworksoul/gort/data"
	gorterr "github.com/clockworksoul/gort/errors"
	yaml "gopkg.in/yaml.v3"
)

// LoadBundle is called by ...
func LoadBundle(file string) (data.Bundle, error) {
	// Read file as a byte slice
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		return data.Bundle{}, gorterr.Wrap(gorterr.ErrIO, err)
	}

	var bun data.Bundle

	err = yaml.Unmarshal(dat, &bun)
	if err != nil {
		return data.Bundle{}, gorterr.Wrap(gorterr.ErrUnmarshal, err)
	}

	// Ensure that the command name is propagated from the map key.
	for n := range bun.Commands {
		(bun.Commands[n]).Name = n
	}

	return bun, nil
}
