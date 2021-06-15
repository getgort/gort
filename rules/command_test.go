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

func TestParseEmpty(t *testing.T) {
	_, err := Parse([]string{})
	assert.Error(t, err)
}

func TestParseDefaults(t *testing.T) {
	tests := map[string]Command{
		`curl localhost`:              {"curl", map[string]CommandOption{}, []string{"localhost"}},
		`curl -Ik localhost`:          {"curl", map[string]CommandOption{"I": {Name: "I"}, "k": {Name: "k"}}, []string{"localhost"}},
		`curl --ssl localhost`:        {"curl", map[string]CommandOption{"ssl": {Name: "ssl"}}, []string{"localhost"}},
		`curl -Ik -- --ssl localhost`: {"curl", map[string]CommandOption{"I": {Name: "I"}, "k": {Name: "k"}}, []string{"--ssl", "localhost"}},
		`echo -n foo bar`:             {"echo", map[string]CommandOption{"n": {Name: "n"}}, []string{"foo", "bar"}},
		`echo -n foo -E bar`:          {"echo", map[string]CommandOption{"n": {Name: "n"}}, []string{"foo", "-E", "bar"}},
		`echo -n "foo bar"`:           {"echo", map[string]CommandOption{"n": {Name: "n"}}, []string{"foo bar"}},
	}

	for test, expected := range tests {
		tokens, err := Tokenize(test)
		assert.NoError(t, err, test)

		actual, err := Parse(tokens)
		assert.NoError(t, err, test)

		assert.Equal(t, expected, actual, test)
	}
}

func TestParseAgnosticDashesTrue(t *testing.T) {
	tests := map[string]Command{
		`curl localhost`:              {"curl", map[string]CommandOption{}, []string{"localhost"}},
		`curl -Ik localhost`:          {"curl", map[string]CommandOption{"Ik": {Name: "Ik"}}, []string{"localhost"}},
		`curl --ssl localhost`:        {"curl", map[string]CommandOption{"ssl": {Name: "ssl"}}, []string{"localhost"}},
		`curl -Ik --ssl localhost`:    {"curl", map[string]CommandOption{"Ik": {Name: "Ik"}, "ssl": {Name: "ssl"}}, []string{"localhost"}},
		`curl -Ik -- --ssl localhost`: {"curl", map[string]CommandOption{"Ik": {Name: "Ik"}}, []string{"--ssl", "localhost"}},
	}

	options := []ParseOption{ParseAgnosticDashes(true)}

	for test, expected := range tests {
		tokens, err := Tokenize(test)
		assert.NoError(t, err, test)

		actual, err := Parse(tokens, options...)
		assert.NoError(t, err, test)

		assert.Equal(t, expected, actual, test)
	}
}

func TestParseAssumeOptionArgumentsTrue(t *testing.T) {
	tests := map[string]Command{
		`curl localhost`:              {"curl", map[string]CommandOption{}, []string{"localhost"}},
		`curl -Ik localhost`:          {"curl", map[string]CommandOption{"I": {Name: "I"}, "k": {Name: "k", Value: "localhost"}}, []string{}},
		`curl --ssl localhost`:        {"curl", map[string]CommandOption{"ssl": {Name: "ssl", Value: "localhost"}}, []string{}},
		`curl -Ik --ssl localhost`:    {"curl", map[string]CommandOption{"I": {Name: "I"}, "k": {Name: "k"}, "ssl": {Name: "ssl", Value: "localhost"}}, []string{}},
		`curl -Ik -- --ssl localhost`: {"curl", map[string]CommandOption{"I": {Name: "I"}, "k": {Name: "k"}}, []string{"--ssl", "localhost"}},
		`echo -n foo bar`:             {"echo", map[string]CommandOption{"n": {Name: "n", Value: "foo"}}, []string{"bar"}},
		`echo -n "foo bar"`:           {"echo", map[string]CommandOption{"n": {Name: "n", Value: "foo bar"}}, []string{}},
	}

	options := []ParseOption{ParseAssumeOptionArguments(true)}

	for test, expected := range tests {
		tokens, err := Tokenize(test)
		assert.NoError(t, err, test)

		actual, err := Parse(tokens, options...)
		assert.NoError(t, err, test)

		assert.Equal(t, expected, actual, test)
	}
}

func TestParseAssumeOptionArgumentsFalse(t *testing.T) {
	tests := map[string]Command{
		`curl localhost`:              {"curl", map[string]CommandOption{}, []string{"localhost"}},
		`curl -Ik localhost`:          {"curl", map[string]CommandOption{"I": {Name: "I"}, "k": {Name: "k"}}, []string{"localhost"}},
		`curl --ssl localhost`:        {"curl", map[string]CommandOption{"ssl": {Name: "ssl"}}, []string{"localhost"}},
		`curl -Ik --ssl localhost`:    {"curl", map[string]CommandOption{"I": {Name: "I"}, "k": {Name: "k"}, "ssl": {Name: "ssl"}}, []string{"localhost"}},
		`curl -Ik -- --ssl localhost`: {"curl", map[string]CommandOption{"I": {Name: "I"}, "k": {Name: "k"}}, []string{"--ssl", "localhost"}},
		`echo -n foo bar`:             {"echo", map[string]CommandOption{"n": {Name: "n"}}, []string{"foo", "bar"}},
		`echo -n "foo bar"`:           {"echo", map[string]CommandOption{"n": {Name: "n"}}, []string{"foo bar"}},
	}

	options := []ParseOption{ParseAssumeOptionArguments(false)}

	for test, expected := range tests {
		tokens, err := Tokenize(test)
		assert.NoError(t, err, test)

		actual, err := Parse(tokens, options...)
		assert.NoError(t, err, test)

		assert.Equal(t, expected, actual, test)
	}
}

func TestParseOptionHasArgument(t *testing.T) {
	tests := map[string]Command{
		`curl localhost`:                 {"curl", map[string]CommandOption{}, []string{"localhost"}},
		`curl -Ik localhost`:             {"curl", map[string]CommandOption{"I": {Name: "I"}, "k": {Name: "k"}}, []string{"localhost"}},
		`curl --ssl localhost`:           {"curl", map[string]CommandOption{"ssl": {Name: "ssl"}}, []string{"localhost"}},
		`curl -Ik --cert file localhost`: {"curl", map[string]CommandOption{"I": {Name: "I"}, "k": {Name: "k"}, "cert": {Name: "cert", Value: "file"}}, []string{"localhost"}},
	}

	options := []ParseOption{
		ParseOptionHasArgument("I", false),
		ParseOptionHasArgument("k", false),
		ParseOptionHasArgument("ssl", false),
		ParseOptionHasArgument("cert", true),
	}

	for test, expected := range tests {
		tokens, err := Tokenize(test)
		assert.NoError(t, err, test)

		actual, err := Parse(tokens, options...)
		assert.NoError(t, err, test)

		assert.Equal(t, expected, actual, test)
	}
}

func TestParseOptionAlias(t *testing.T) {
	tests := map[string]Command{
		`curl localhost`:                 {"curl", map[string]CommandOption{}, []string{"localhost"}},
		`curl -Ik localhost`:             {"curl", map[string]CommandOption{"I": {Name: "I"}, "k": {Name: "k"}}, []string{"localhost"}},
		`curl --ssl localhost`:           {"curl", map[string]CommandOption{"ssl": {Name: "ssl"}}, []string{"localhost"}},
		`curl -Ik --cert file localhost`: {"curl", map[string]CommandOption{"I": {Name: "I"}, "k": {Name: "k"}, "cert": {Name: "cert", Value: "file"}}, []string{"localhost"}},
		`curl -E file localhost`:         {"curl", map[string]CommandOption{"cert": {Name: "cert", Value: "file"}}, []string{"localhost"}},
	}

	options := []ParseOption{
		ParseOptionAlias("E", "cert"),
		ParseOptionHasArgument("cert", true),
	}

	for test, expected := range tests {
		tokens, err := Tokenize(test)
		assert.NoError(t, err, test)

		actual, err := Parse(tokens, options...)
		assert.NoError(t, err, test)

		assert.Equal(t, expected, actual, test)
	}
}
