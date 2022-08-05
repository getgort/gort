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

package io

import (
	"testing"

	"github.com/getgort/gort/command"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NewCommandInfo creates a new CommandInfo from a command.Command value.
func TestNewCommandInfo(t *testing.T) {
	tests := []struct {
		Text     string
		Expected CommandInfo
	}{
		{
			Text: `foo:curl localhost`,
			Expected: CommandInfo{
				Bundle:     `foo`,
				Command:    `curl`,
				Options:    map[string]string{},
				Parameters: []string{"localhost"},
			},
		},
		{
			Text: `foo:curl -Ik localhost`,
			Expected: CommandInfo{
				Bundle:     `foo`,
				Command:    `curl`,
				Options:    map[string]string{"I": "true", "k": "true"},
				Parameters: []string{"localhost"},
			},
		},
		{
			Text: `foo:curl --ssl localhost`,
			Expected: CommandInfo{
				Bundle:     `foo`,
				Command:    `curl`,
				Options:    map[string]string{"ssl": "true"},
				Parameters: []string{"localhost"},
			},
		},
		{
			Text: `foo:curl -Ik -- --ssl localhost`,
			Expected: CommandInfo{
				Bundle:     `foo`,
				Command:    `curl`,
				Options:    map[string]string{"I": "true", "k": "true"},
				Parameters: []string{"--ssl", "localhost"},
			},
		},
		{
			Text: `bar:echo -n foo bar`,
			Expected: CommandInfo{
				Bundle:     `bar`,
				Command:    `echo`,
				Options:    map[string]string{"n": "true"},
				Parameters: []string{"foo", "bar"},
			},
		},
		{
			Text: `bar:echo -n foo -E bar`,
			Expected: CommandInfo{
				Bundle:     `bar`,
				Command:    `echo`,
				Options:    map[string]string{"n": "true"},
				Parameters: []string{"foo", "-E", "bar"},
			},
		},
		{
			Text: `bar:echo -n "foo bar"`,
			Expected: CommandInfo{
				Bundle:     `bar`,
				Command:    `echo`,
				Options:    map[string]string{"n": "true"},
				Parameters: []string{"foo bar"},
			},
		},
	}

	for i, test := range tests {
		tokens, err := command.Tokenize(test.Text)
		require.NoError(t, err, "(%d) failed to tokenize: %s", i, test.Text)

		command, err := command.Parse(tokens)
		require.NoError(t, err, "(%d) failed to parse: %s", i, test.Text)

		actual := NewCommandInfo(command)

		assert.Equal(t, test.Expected, actual, "(%d) mismatch: %s", i, test.Text)
	}
}
