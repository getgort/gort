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

	"github.com/getgort/gort/types"
	"github.com/stretchr/testify/assert"
)

func TestRuleMatches(t *testing.T) {
	options := map[string]types.Value{
		"foo": types.StringValue{V: "bar"},
		"k":   types.BoolValue{V: true},
		"n":   types.IntValue{V: 10},
	}
	args := []types.Value{types.StringValue{V: "foo"}, types.StringValue{V: "bar"}}
	env := EvaluationEnvironment{
		"option": options,
		"arg":    args,
	}

	inputs := map[string]bool{
		`foo:bar allow`:                                                 true,
		`foo:bar with false == false allow`:                             true,
		`foo:bar with true == false allow`:                              false,
		`foo:bar with true == true or true == false allow`:              true,
		`foo:bar with true == true and true == false allow`:             false,
		`foo:bar with option['delete'] == false allow`:                  true,
		`foo:bar with option['delete'] == true allow`:                   false,
		`foo:bar with option['k'] == true allow`:                        true,
		`foo:bar with option['foo'] == "bar" allow`:                     true,
		`foo:bar with option['foo'] == "bat" allow`:                     false,
		`foo:bar with arg[0] == "foo" allow`:                            true,
		`foo:bar with option['foo'] == "bar" and arg[0] == "foo" allow`: true,
		`foo:bar with any arg == /^f.*$/ allow`:                         true,
		`foo:bar with all arg == /^f.*$/ allow`:                         false,
		`foo:bar with all arg in ["foo", "bar"] allow`:                  true,
		`foo:bar with any arg == /^blah.*/ allow`:                       false,
		`foo:bar with arg[0] in ['foo', false, 100] allow`:              true,
		`foo:bar with option["foo"] in ["foo", "bar"] allow`:            true,
		`foo:bar with any option == /^prod.*/ allow`:                    false,
		`foo:bar with any arg in ['wubba'] allow`:                       false,
		`foo:bar with any arg in ['wubba', /^f.*/, 10] allow`:           true,
		`foo:bar with all arg in [10, 'baz', 'wubba'] allow`:            false,
		`foo:bar with all option < 10 allow`:                            false,
		`foo:bar with all option in ['staging', 'list'] allow`:          false,
	}

	for in, expected := range inputs {
		rt, err := Tokenize(in)
		if !assert.NoError(t, err, in) {
			continue
		}

		rule, err := Parse(rt)
		if !assert.NoError(t, err, in) {
			continue
		}

		result := rule.Matches(env)
		assert.Equal(t, expected, result, in)
	}
}
