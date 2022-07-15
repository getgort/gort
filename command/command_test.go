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

	"github.com/stretchr/testify/assert"

	. "github.com/getgort/gort/types"
)

func TestCommandParseEmpty(t *testing.T) {
	_, err := Parse([]string{})
	assert.Error(t, err)
}

func TestCommandParseDefaults(t *testing.T) {
	tests := map[string]Command{
		`foo:curl localhost`:              {`foo`, `curl`, map[string]CommandOption{}, []Value{stringValue("localhost")}, "foo:curl localhost"},
		`foo:curl -Ik localhost`:          {`foo`, `curl`, map[string]CommandOption{"I": {"I", BoolValue{V: true}}, "k": {"k", BoolValue{V: true}}}, []Value{stringValue("localhost")}, "foo:curl -Ik localhost"},
		`foo:curl --ssl localhost`:        {`foo`, `curl`, map[string]CommandOption{"ssl": {"ssl", BoolValue{V: true}}}, []Value{stringValue("localhost")}, "foo:curl --ssl localhost"},
		`foo:curl -Ik -- --ssl localhost`: {`foo`, `curl`, map[string]CommandOption{"I": {"I", BoolValue{V: true}}, "k": {"k", BoolValue{V: true}}}, []Value{stringValue("--ssl"), stringValue("localhost")}, "foo:curl -Ik -- --ssl localhost"},
		`bar:echo -n foo bar`:             {`bar`, `echo`, map[string]CommandOption{"n": {"n", BoolValue{V: true}}}, []Value{stringValue("foo"), stringValue("bar")}, "bar:echo -n foo bar"},
		`bar:echo -n foo -E bar`:          {`bar`, `echo`, map[string]CommandOption{"n": {"n", BoolValue{V: true}}}, []Value{stringValue("foo"), stringValue("-E"), stringValue("bar")}, "bar:echo -n foo -E bar"},
		`bar:echo -n "foo bar"`:           {`bar`, `echo`, map[string]CommandOption{"n": {"n", BoolValue{V: true}}}, []Value{StringValue{V: "foo bar", Quote: '"'}}, "bar:echo -n \"foo bar\""},
	}

	for test, expected := range tests {
		tokens, err := Tokenize(test)
		assert.NoError(t, err, test)

		actual, err := Parse(tokens)
		assert.NoError(t, err, test)
		actual.original = test

		assert.Equal(t, expected, actual, test)
	}
}

func TestCommandOptionTypes(t *testing.T) {
	test := `test --flag --int 10 --float 0.1 --notregex "/^foo$/" --string str this is text`

	expected := Command{"", "test",
		map[string]CommandOption{
			"flag":     {"flag", BoolValue{V: true}},
			"int":      {"int", IntValue{V: 10}},
			"float":    {"float", FloatValue{V: 0.1}},
			"notregex": {"notregex", StringValue{V: `/^foo$/`, Quote: '"'}},
			"string":   {"string", StringValue{V: "str"}},
		},
		[]Value{stringValue("this"), stringValue("is"), stringValue("text")},
		test,
	}

	options := []ParseOption{ParseAssumeOptionArguments(true)}

	tokens, err := Tokenize(test)
	assert.NoError(t, err, test)

	actual, err := Parse(tokens, options...)
	assert.NoError(t, err, test)
	actual.original = test

	assert.Equal(t, expected, actual, test)
}

func TestCommandParameterTypes(t *testing.T) {
	test := `test string 10.0 42 false "foo bar" /^.*$/`

	expected := Command{"", "test",
		map[string]CommandOption{},
		[]Value{
			StringValue{V: "string", Quote: '\u0000'},
			FloatValue{V: 10.0},
			IntValue{V: 42},
			BoolValue{V: false},
			StringValue{V: "foo bar", Quote: '"'},
			StringValue{V: "/^.*$/", Quote: '\u0000'},
		},
		test,
	}

	tokens, err := Tokenize(test)

	assert.NoError(t, err, test)

	actual, err := Parse(tokens)
	assert.NoError(t, err, test)
	actual.original = test

	assert.Equal(t, expected, actual, test)
}

func TestCommandParseBareFlagsAreTrue(t *testing.T) {
	tv := BoolValue{V: true}

	tests := map[string]Command{
		`foo:curl -Ik localhost`: {`foo`, `curl`, map[string]CommandOption{"I": {"I", tv}, "k": {"k", tv}}, []Value{stringValue("localhost")}, `foo:curl -Ik localhost`},
		`bar:echo -n foo -E bar`: {`bar`, `echo`, map[string]CommandOption{"n": {"n", tv}}, []Value{stringValue("foo"), stringValue("-E"), stringValue("bar")}, `bar:echo -n foo -E bar`},
		`bar:echo -n "foo bar"`:  {`bar`, `echo`, map[string]CommandOption{"n": {"n", tv}}, []Value{StringValue{V: "foo bar", Quote: '"'}}, `bar:echo -n "foo bar""`},
	}

	for test, expected := range tests {
		tokens, err := Tokenize(test)
		assert.NoError(t, err, test)

		actual, err := Parse(tokens)
		assert.NoError(t, err, test)
		actual.original = test

		assert.Equal(t, expected, actual, test)
	}
}

func TestCommandParseAgnosticDashesTrue(t *testing.T) {
	tests := map[string]Command{
		`foo:curl localhost`:              {`foo`, `curl`, map[string]CommandOption{}, []Value{stringValue("localhost")}, `foo:curl localhost`},
		`foo:curl -Ik localhost`:          {`foo`, `curl`, map[string]CommandOption{"Ik": {"Ik", BoolValue{V: true}}}, []Value{stringValue("localhost")}, `foo:curl -Ik localhost`},
		`foo:curl --ssl localhost`:        {`foo`, `curl`, map[string]CommandOption{"ssl": {"ssl", BoolValue{V: true}}}, []Value{stringValue("localhost")}, `foo:curl --ssl localhost`},
		`foo:curl -Ik --ssl localhost`:    {`foo`, `curl`, map[string]CommandOption{"Ik": {"Ik", BoolValue{V: true}}, "ssl": {"ssl", BoolValue{V: true}}}, []Value{stringValue("localhost")}, `foo:curl -Ik --ssl localhost`},
		`foo:curl -Ik -- --ssl localhost`: {`foo`, `curl`, map[string]CommandOption{"Ik": {"Ik", BoolValue{V: true}}}, []Value{stringValue("--ssl"), stringValue("localhost")}, `foo:curl -Ik -- --ssl localhost`},
	}

	options := []ParseOption{ParseAgnosticDashes(true)}

	for test, expected := range tests {
		tokens, err := Tokenize(test)
		assert.NoError(t, err, test)

		actual, err := Parse(tokens, options...)
		assert.NoError(t, err, test)
		actual.original = test

		assert.Equal(t, expected, actual, test)
	}
}

func TestCommandParseAssumeOptionArgumentsTrue(t *testing.T) {
	tests := map[string]Command{
		`foo:curl localhost`:              {`foo`, `curl`, map[string]CommandOption{}, []Value{stringValue("localhost")}, `foo:curl localhost`},
		`foo:curl -Ik localhost`:          {`foo`, `curl`, map[string]CommandOption{"I": {"I", BoolValue{V: true}}, "k": {"k", StringValue{V: "localhost", Quote: '\u0000'}}}, []Value{}, `foo:curl -Ik localhost`},
		`foo:curl --ssl localhost`:        {`foo`, `curl`, map[string]CommandOption{"ssl": {"ssl", StringValue{V: "localhost", Quote: '\u0000'}}}, []Value{}, `foo:curl --ssl localhost`},
		`foo:curl -Ik --ssl localhost`:    {`foo`, `curl`, map[string]CommandOption{"I": {"I", BoolValue{V: true}}, "k": {"k", BoolValue{V: true}}, "ssl": {"ssl", StringValue{V: "localhost", Quote: '\u0000'}}}, []Value{}, `foo:curl -Ik --ssl localhost`},
		`foo:curl -Ik -- --ssl localhost`: {`foo`, `curl`, map[string]CommandOption{"I": {"I", BoolValue{V: true}}, "k": {"k", BoolValue{V: true}}}, []Value{stringValue("--ssl"), stringValue("localhost")}, `foo:curl -Ik -- --ssl localhost`},
		`bar:echo -n foo bar`:             {`bar`, `echo`, map[string]CommandOption{"n": {"n", StringValue{V: "foo", Quote: '\u0000'}}}, []Value{stringValue("bar")}, `bar:echo -n foo bar`},
		`bar:echo -n "foo bar"`:           {`bar`, `echo`, map[string]CommandOption{"n": {"n", StringValue{V: "foo bar", Quote: '"'}}}, []Value{}, `bar:echo -n "foo bar"`},
	}

	options := []ParseOption{ParseAssumeOptionArguments(true)}

	for test, expected := range tests {
		tokens, err := Tokenize(test)
		assert.NoError(t, err, test)

		actual, err := Parse(tokens, options...)
		assert.NoError(t, err, test)
		actual.original = test

		assert.Equal(t, expected, actual, test)
	}
}

func TestCommandParseAssumeOptionArgumentsFalse(t *testing.T) {
	tests := map[string]Command{
		`foo:curl localhost`:              {`foo`, `curl`, map[string]CommandOption{}, []Value{stringValue("localhost")}, `foo:curl localhost`},
		`foo:curl -Ik localhost`:          {`foo`, `curl`, map[string]CommandOption{"I": {"I", BoolValue{V: true}}, "k": {"k", BoolValue{V: true}}}, []Value{stringValue("localhost")}, `foo:curl -Ik localhost`},
		`foo:curl --ssl localhost`:        {`foo`, `curl`, map[string]CommandOption{"ssl": {"ssl", BoolValue{V: true}}}, []Value{stringValue("localhost")}, `foo:curl --ssl localhost`},
		`foo:curl -Ik --ssl localhost`:    {`foo`, `curl`, map[string]CommandOption{"I": {"I", BoolValue{V: true}}, "k": {"k", BoolValue{V: true}}, "ssl": {"ssl", BoolValue{V: true}}}, []Value{stringValue("localhost")}, `foo:curl -Ik --ssl localhost`},
		`foo:curl -Ik -- --ssl localhost`: {`foo`, `curl`, map[string]CommandOption{"I": {"I", BoolValue{V: true}}, "k": {"k", BoolValue{V: true}}}, []Value{stringValue("--ssl"), stringValue("localhost")}, `foo:curl -Ik -- --ssl localhost`},
		`bar:echo -n foo bar`:             {`bar`, `echo`, map[string]CommandOption{"n": {"n", BoolValue{V: true}}}, []Value{stringValue("foo"), stringValue("bar")}, `bar:echo -n foo bar`},
		`bar:echo -n "foo bar"`:           {`bar`, `echo`, map[string]CommandOption{"n": {"n", BoolValue{V: true}}}, []Value{StringValue{V: "foo bar", Quote: '"'}}, `bar:echo -n "foo bar"`},
	}

	options := []ParseOption{ParseAssumeOptionArguments(false)}

	for test, expected := range tests {
		tokens, err := Tokenize(test)
		assert.NoError(t, err, test)

		actual, err := Parse(tokens, options...)
		assert.NoError(t, err, test)
		actual.original = test

		assert.Equal(t, expected, actual, test)
	}
}

func TestCommandParseOptionHasArgument(t *testing.T) {
	tests := map[string]Command{
		`foo:curl localhost`:                 {`foo`, `curl`, map[string]CommandOption{}, []Value{stringValue("localhost")}, `foo:curl localhost`},
		`foo:curl -Ik localhost`:             {`foo`, `curl`, map[string]CommandOption{"I": {"I", BoolValue{V: true}}, "k": {"k", BoolValue{V: true}}}, []Value{stringValue("localhost")}, `foo:curl -Ik localhost`},
		`foo:curl --ssl localhost`:           {`foo`, `curl`, map[string]CommandOption{"ssl": {"ssl", BoolValue{V: true}}}, []Value{stringValue("localhost")}, `foo:curl --ssl localhost`},
		`foo:curl -Ik --cert file localhost`: {`foo`, `curl`, map[string]CommandOption{"I": {"I", BoolValue{V: true}}, "k": {"k", BoolValue{V: true}}, "cert": {"cert", StringValue{V: "file", Quote: '\u0000'}}}, []Value{stringValue("localhost")}, `foo:curl -Ik --cert file localhost`},
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
		actual.original = test

		assert.Equal(t, expected, actual, test)
	}
}

func TestCommandParseOptionAlias(t *testing.T) {
	tests := map[string]Command{
		`foo:curl localhost`:                 {`foo`, `curl`, map[string]CommandOption{}, []Value{stringValue("localhost")}, `foo:curl localhost`},
		`foo:curl -Ik localhost`:             {`foo`, `curl`, map[string]CommandOption{"I": {"I", BoolValue{V: true}}, "k": {"k", BoolValue{V: true}}}, []Value{stringValue("localhost")}, `foo:curl -Ik localhost`},
		`foo:curl --ssl localhost`:           {`foo`, `curl`, map[string]CommandOption{"ssl": {"ssl", BoolValue{V: true}}}, []Value{stringValue("localhost")}, `foo:curl --ssl localhost`},
		`foo:curl -Ik --cert file localhost`: {`foo`, `curl`, map[string]CommandOption{"I": {"I", BoolValue{V: true}}, "k": {"k", BoolValue{V: true}}, "cert": {"cert", StringValue{V: "file", Quote: '\u0000'}}}, []Value{stringValue("localhost")}, `foo:curl -Ik --cert file localhost`},
		`foo:curl -E file localhost`:         {`foo`, `curl`, map[string]CommandOption{"cert": {"cert", StringValue{V: "file", Quote: '\u0000'}}}, []Value{stringValue("localhost")}, `foo:curl -E file localhost`},
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
		actual.original = test

		assert.Equal(t, expected, actual, test)
	}
}

func TestSplitCommand(t *testing.T) {
	var bundle, command string
	var err error

	bundle, command, err = SplitCommand("foo:bar")
	assert.Equal(t, "foo", bundle)
	assert.Equal(t, "bar", command)
	assert.Nil(t, err)

	bundle, command, err = SplitCommand("bat")
	assert.Equal(t, "", bundle)
	assert.Equal(t, "bat", command)
	assert.Nil(t, err)

	bundle, command, err = SplitCommand("foo:bar:bat")
	assert.Equal(t, "", bundle)
	assert.Equal(t, "", command)
	assert.NotNil(t, err)
}

func stringValue(s string) Value {
	return StringValue{V: s, Quote: '\u0000'}
}
