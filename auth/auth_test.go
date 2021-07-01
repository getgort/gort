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

package auth

import (
	"fmt"
	"testing"

	"github.com/getgort/gort/bundles"
	"github.com/getgort/gort/command"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/rules"
	"github.com/getgort/gort/types"

	"github.com/stretchr/testify/assert"
)

func TestEvaluateCommandEntry1(t *testing.T) {
	b, err := bundles.LoadBundle("../testing/test-bundle-foo.yml")
	if err != nil {
		t.Error(err.Error())
	}
	cmd := data.CommandEntry{Bundle: b, Command: *b.Commands["foo"]}
	envFooTrue := rules.EvaluationEnvironment{"option": map[string]types.Value{"foo": types.BoolValue{V: true}}}
	envFooFalse := rules.EvaluationEnvironment{"option": map[string]types.Value{}}

	// Command:   "test:foo"
	// Perms:     None
	// Condition: with option["foo"] == false allow
	// Expected:  TRUE
	result, err := EvaluateCommandEntry([]string{}, cmd, envFooFalse)
	assert.NoError(t, err)
	assert.True(t, result)

	// Command:   "test:foo"
	// Perms:     "test:foo"
	// Condition: with option["foo"] == false allow
	// Expected:  TRUE
	result, err = EvaluateCommandEntry([]string{"test:foo"}, cmd, envFooFalse)
	assert.NoError(t, err)
	assert.True(t, result)

	// Command:   "test:foo --foo"
	// Perms:     NONE
	// Condition: with option["foo"] == true must have test:foo
	// Expected:  FALSE
	result, err = EvaluateCommandEntry([]string{}, cmd, envFooTrue)
	assert.NoError(t, err)
	assert.False(t, result)

	// Command:   "test:foo --foo"
	// Perms:     "test:foo"
	// Condition: with option["foo"] == true must have test:foo
	// Expected:  TRUE
	result, err = EvaluateCommandEntry([]string{"test:foo"}, cmd, envFooTrue)
	assert.NoError(t, err)
	assert.True(t, result)
}

func TestEvaluateCommandEntry2(t *testing.T) {
	b, err := bundles.LoadBundle("../testing/test-default.yml")
	if err != nil {
		t.Error(err.Error())
	}

	cmd, env, err := parse("gort:gort user --help")
	assert.NoError(t, err)
	cmdE := data.CommandEntry{Bundle: b, Command: *b.Commands[cmd.Command]}

	result, err := EvaluateCommandEntry([]string{"test:foo", "gort:manage_users"}, cmdE, env)
	assert.NoError(t, err)
	assert.True(t, result)

	result, err = EvaluateCommandEntry([]string{"test:foo"}, cmdE, env)
	assert.NoError(t, err)
	assert.False(t, result)
}

func parse(s string) (command.Command, rules.EvaluationEnvironment, error) {
	cmd, err := command.TokenizeAndParse(s)
	if err != nil {
		return command.Command{}, nil, err
	}

	env := rules.EvaluationEnvironment{}
	env["option"] = cmd.OptionsValues()
	env["arg"] = cmd.Parameters

	return cmd, env, nil
}

func TestParseCommandEntry(t *testing.T) {
	b, err := bundles.LoadBundle("../testing/test-bundle-foo.yml")
	if err != nil {
		t.Error(err.Error())
	}

	cmd := data.CommandEntry{Bundle: b, Command: *b.Commands["foo"]}

	expected := []rules.Rule{
		{
			Command:     "test:foo",
			Conditions:  []rules.Expression{{A: types.MapElementValue{V: types.MapValue{Name: "option"}, Key: "foo"}, B: types.BoolValue{V: false}, Operator: rules.Equals}},
			Permissions: []rules.Permission{},
		}, {
			Command:     "test:foo",
			Conditions:  []rules.Expression{{A: types.MapElementValue{V: types.MapValue{Name: "option"}, Key: "foo"}, B: types.BoolValue{V: true}, Operator: rules.Equals}},
			Permissions: []rules.Permission{{Name: "test:foo"}},
		},
	}

	rules, err := ParseCommandEntry(cmd)
	assert.NoError(t, err)

	for i, rule := range rules {
		assert.Equal(t, expected[i].Command, rule.Command)

		for i, e := range expected[i].Conditions {
			assert.EqualValues(t, e.A, rule.Conditions[i].A)
			assert.EqualValues(t, e.B, rule.Conditions[i].B)
			assert.Equal(t, fmt.Sprintf("%v", e.Operator), fmt.Sprintf("%v", rule.Conditions[i].Operator))
		}

		assert.Equal(t, expected[i].Permissions, rule.Permissions)
	}
}
