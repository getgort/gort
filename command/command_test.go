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

package command

import (
	"testing"

	"github.com/getgort/gort/types"
	"github.com/stretchr/testify/assert"
)

func TestCommandParseEmpty(t *testing.T) {
	_, err := Parse([]string{})
	assert.Error(t, err)
}

func TestCommandParseDefaults(t *testing.T) {
	tests := map[string]Command{
		`curl localhost`:              {"curl", map[string]CommandOption{}, []string{"localhost"}},
		`curl -Ik localhost`:          {"curl", map[string]CommandOption{"I": {"I", types.BoolValue{Value: true}}, "k": {"k", types.BoolValue{Value: true}}}, []string{"localhost"}},
		`curl --ssl localhost`:        {"curl", map[string]CommandOption{"ssl": {"ssl", types.BoolValue{Value: true}}}, []string{"localhost"}},
		`curl -Ik -- --ssl localhost`: {"curl", map[string]CommandOption{"I": {"I", types.BoolValue{Value: true}}, "k": {"k", types.BoolValue{Value: true}}}, []string{"--ssl", "localhost"}},
		`echo -n foo bar`:             {"echo", map[string]CommandOption{"n": {"n", types.BoolValue{Value: true}}}, []string{"foo", "bar"}},
		`echo -n foo -E bar`:          {"echo", map[string]CommandOption{"n": {"n", types.BoolValue{Value: true}}}, []string{"foo", "-E", "bar"}},
		`echo -n "foo bar"`:           {"echo", map[string]CommandOption{"n": {"n", types.BoolValue{Value: true}}}, []string{"foo bar"}},
	}

	for test, expected := range tests {
		tokens, err := Tokenize(test)
		assert.NoError(t, err, test)

		actual, err := Parse(tokens)
		assert.NoError(t, err, test)

		assert.Equal(t, expected, actual, test)
	}
}

func TestCommandTypes(t *testing.T) {
	test := `test --flag --int 10 --float 0.1 --notregex "/^foo$/" --string str this is text`

	expected := Command{"test",
		map[string]CommandOption{
			"flag":     {"flag", types.BoolValue{Value: true}},
			"int":      {"int", types.IntValue{Value: 10}},
			"float":    {"float", types.FloatValue{Value: 0.1}},
			"notregex": {"notregex", types.StringValue{Value: `/^foo$/`}},
			"string":   {"string", types.StringValue{Value: "str"}},
		},
		[]string{"this", "is", "text"},
	}

	options := []ParseOption{ParseAssumeOptionArguments(true)}

	tokens, err := Tokenize(test)
	assert.NoError(t, err, test)

	actual, err := Parse(tokens, options...)
	assert.NoError(t, err, test)

	assert.Equal(t, expected, actual, test)
}

func TestCommandParseBareFlagsAreTrue(t *testing.T) {
	tv := types.BoolValue{Value: true}

	tests := map[string]Command{
		`curl -Ik localhost`: {"curl", map[string]CommandOption{"I": {"I", tv}, "k": {"k", tv}}, []string{"localhost"}},
		`echo -n foo -E bar`: {"echo", map[string]CommandOption{"n": {"n", tv}}, []string{"foo", "-E", "bar"}},
		`echo -n "foo bar"`:  {"echo", map[string]CommandOption{"n": {"n", tv}}, []string{"foo bar"}},
	}

	for test, expected := range tests {
		tokens, err := Tokenize(test)
		assert.NoError(t, err, test)

		actual, err := Parse(tokens)
		assert.NoError(t, err, test)

		assert.Equal(t, expected, actual, test)
	}
}

func TestCommandParseAgnosticDashesTrue(t *testing.T) {
	tests := map[string]Command{
		`curl localhost`:              {"curl", map[string]CommandOption{}, []string{"localhost"}},
		`curl -Ik localhost`:          {"curl", map[string]CommandOption{"Ik": {"Ik", types.BoolValue{Value: true}}}, []string{"localhost"}},
		`curl --ssl localhost`:        {"curl", map[string]CommandOption{"ssl": {"ssl", types.BoolValue{Value: true}}}, []string{"localhost"}},
		`curl -Ik --ssl localhost`:    {"curl", map[string]CommandOption{"Ik": {"Ik", types.BoolValue{Value: true}}, "ssl": {"ssl", types.BoolValue{Value: true}}}, []string{"localhost"}},
		`curl -Ik -- --ssl localhost`: {"curl", map[string]CommandOption{"Ik": {"Ik", types.BoolValue{Value: true}}}, []string{"--ssl", "localhost"}},
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

func TestCommandParseAssumeOptionArgumentsTrue(t *testing.T) {
	tests := map[string]Command{
		`curl localhost`:              {"curl", map[string]CommandOption{}, []string{"localhost"}},
		`curl -Ik localhost`:          {"curl", map[string]CommandOption{"I": {"I", types.BoolValue{Value: true}}, "k": {"k", types.StringValue{Value: "localhost", Quote: '\u0000'}}}, []string{}},
		`curl --ssl localhost`:        {"curl", map[string]CommandOption{"ssl": {"ssl", types.StringValue{Value: "localhost", Quote: '\u0000'}}}, []string{}},
		`curl -Ik --ssl localhost`:    {"curl", map[string]CommandOption{"I": {"I", types.BoolValue{Value: true}}, "k": {"k", types.BoolValue{Value: true}}, "ssl": {"ssl", types.StringValue{Value: "localhost", Quote: '\u0000'}}}, []string{}},
		`curl -Ik -- --ssl localhost`: {"curl", map[string]CommandOption{"I": {"I", types.BoolValue{Value: true}}, "k": {"k", types.BoolValue{Value: true}}}, []string{"--ssl", "localhost"}},
		`echo -n foo bar`:             {"echo", map[string]CommandOption{"n": {"n", types.StringValue{Value: "foo", Quote: '\u0000'}}}, []string{"bar"}},
		`echo -n "foo bar"`:           {"echo", map[string]CommandOption{"n": {"n", types.StringValue{Value: "foo bar", Quote: '\u0000'}}}, []string{}},
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

func TestCommandParseAssumeOptionArgumentsFalse(t *testing.T) {
	tests := map[string]Command{
		`curl localhost`:              {"curl", map[string]CommandOption{}, []string{"localhost"}},
		`curl -Ik localhost`:          {"curl", map[string]CommandOption{"I": {"I", types.BoolValue{Value: true}}, "k": {"k", types.BoolValue{Value: true}}}, []string{"localhost"}},
		`curl --ssl localhost`:        {"curl", map[string]CommandOption{"ssl": {"ssl", types.BoolValue{Value: true}}}, []string{"localhost"}},
		`curl -Ik --ssl localhost`:    {"curl", map[string]CommandOption{"I": {"I", types.BoolValue{Value: true}}, "k": {"k", types.BoolValue{Value: true}}, "ssl": {"ssl", types.BoolValue{Value: true}}}, []string{"localhost"}},
		`curl -Ik -- --ssl localhost`: {"curl", map[string]CommandOption{"I": {"I", types.BoolValue{Value: true}}, "k": {"k", types.BoolValue{Value: true}}}, []string{"--ssl", "localhost"}},
		`echo -n foo bar`:             {"echo", map[string]CommandOption{"n": {"n", types.BoolValue{Value: true}}}, []string{"foo", "bar"}},
		`echo -n "foo bar"`:           {"echo", map[string]CommandOption{"n": {"n", types.BoolValue{Value: true}}}, []string{"foo bar"}},
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

func TestCommandParseOptionHasArgument(t *testing.T) {
	tests := map[string]Command{
		`curl localhost`:                 {"curl", map[string]CommandOption{}, []string{"localhost"}},
		`curl -Ik localhost`:             {"curl", map[string]CommandOption{"I": {"I", types.BoolValue{Value: true}}, "k": {"k", types.BoolValue{Value: true}}}, []string{"localhost"}},
		`curl --ssl localhost`:           {"curl", map[string]CommandOption{"ssl": {"ssl", types.BoolValue{Value: true}}}, []string{"localhost"}},
		`curl -Ik --cert file localhost`: {"curl", map[string]CommandOption{"I": {"I", types.BoolValue{Value: true}}, "k": {"k", types.BoolValue{Value: true}}, "cert": {"cert", types.StringValue{Value: "file", Quote: '\u0000'}}}, []string{"localhost"}},
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

func TestCommandParseOptionAlias(t *testing.T) {
	tests := map[string]Command{
		`curl localhost`:                 {"curl", map[string]CommandOption{}, []string{"localhost"}},
		`curl -Ik localhost`:             {"curl", map[string]CommandOption{"I": {"I", types.BoolValue{Value: true}}, "k": {"k", types.BoolValue{Value: true}}}, []string{"localhost"}},
		`curl --ssl localhost`:           {"curl", map[string]CommandOption{"ssl": {"ssl", types.BoolValue{Value: true}}}, []string{"localhost"}},
		`curl -Ik --cert file localhost`: {"curl", map[string]CommandOption{"I": {"I", types.BoolValue{Value: true}}, "k": {"k", types.BoolValue{Value: true}}, "cert": {"cert", types.StringValue{Value: "file", Quote: '\u0000'}}}, []string{"localhost"}},
		`curl -E file localhost`:         {"curl", map[string]CommandOption{"cert": {"cert", types.StringValue{Value: "file", Quote: '\u0000'}}}, []string{"localhost"}},
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
