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

package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDynamicConfigValidate(t *testing.T) {
	tests := []struct {
		layer string
		err   bool
	}{
		{layer: "bundle", err: false},
		{layer: "room", err: false},
		{layer: "group", err: false},
		{layer: "user", err: false},
		{layer: "Bundle", err: false},
		{layer: "", err: true},
		{layer: "foo", err: true},
	}

	const msg = "layer=%q"

	for _, test := range tests {
		layer := ConfigurationLayer(test.layer)

		if test.err {
			assert.Error(t, layer.Validate(), msg, test.layer)
		} else {
			assert.NoError(t, layer.Validate(), msg, test.layer)
		}
	}
}
