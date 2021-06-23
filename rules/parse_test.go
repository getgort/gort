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

package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	inputs := map[string]Rule{
		`foo:bar allow`: {Command: "foo:bar", Conditions: []Expression{}, Permissions: []string{}},
		`foo:bar with option['delete'] == true must have foo:destroy`: {Command: "foo:bar", Conditions: []Expression{}, Permissions: []string{"foo:destroy"}},
	}

	for in, expected := range inputs {
		rt, err := Tokenize(in)
		if !assert.NoError(t, err, in) {
			continue
		}

		for i, c := range rt.Conditions {
			t.Log(i, c)

			// Patterns:
			//
			// 1 each of:
			//
			// $VALUE_A
			// any $SET_A
			// all $SET_A
			//
			// OP $VALUE_B
			// in $SET_B
		}

		rule, err := Parse(rt)
		if !assert.NoError(t, err, in) {
			continue
		}

		assert.Equal(t, expected, rule, in)
	}
}
