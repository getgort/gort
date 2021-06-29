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

package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInferrerBuilders(t *testing.T) {
	i := Inferrer{}
	assertCanInferCollection(t, i, false)
	assertCanInferLiteralLists(t, i, false)
	assertCanInferRegularExpressions(t, i, false)
	assertStrictStrings(t, i, false)

	i = Inferrer{}.CollectionReferences(true)
	assertCanInferCollection(t, i, true)
	assertCanInferLiteralLists(t, i, false)
	assertCanInferRegularExpressions(t, i, false)
	assertStrictStrings(t, i, false)

	i = Inferrer{}.ComplexTypes(true)
	assertCanInferCollection(t, i, true)
	assertCanInferLiteralLists(t, i, true)
	assertCanInferRegularExpressions(t, i, true)
	assertStrictStrings(t, i, false)

	i = Inferrer{}.LiteralLists(true)
	assertCanInferCollection(t, i, false)
	assertCanInferLiteralLists(t, i, true)
	assertCanInferRegularExpressions(t, i, false)
	assertStrictStrings(t, i, false)

	i = Inferrer{}.RegularExpressions(true)
	assertCanInferCollection(t, i, false)
	assertCanInferLiteralLists(t, i, false)
	assertCanInferRegularExpressions(t, i, true)
	assertStrictStrings(t, i, false)

	i = Inferrer{}.StrictStrings(true)
	assertCanInferCollection(t, i, false)
	assertCanInferLiteralLists(t, i, false)
	assertCanInferRegularExpressions(t, i, false)
	assertStrictStrings(t, i, true)
}

func assertCanInferCollection(t *testing.T, i Inferrer, enabled bool) {
	t.Helper()
	s := `arg[0]`
	v, err := i.Infer(s)

	if enabled {
		_, ok := v.(ListElementValue)
		assert.True(t, ok, "Got: %T", v)
	} else {
		sv, sok := v.(StringValue)
		_, uok := v.(UnknownValue)
		assert.True(t, err != nil || uok || (sok && sv.Quote == '\u0000'), "Got: %T", v)
	}
}

func assertCanInferLiteralLists(t *testing.T, i Inferrer, enabled bool) {
	t.Helper()
	s := `[ "foo" ]`
	v, err := i.Infer(s)

	if enabled {
		_, ok := v.(ListValue)
		assert.True(t, ok, "Got: %T", v)
	} else {
		sv, sok := v.(StringValue)
		_, uok := v.(UnknownValue)
		assert.True(t, err != nil || uok || (sok && sv.Quote == '\u0000'), "Got: %T", v)
	}
}

func assertCanInferRegularExpressions(t *testing.T, i Inferrer, enabled bool) {
	t.Helper()
	s := `/^foo$/`
	v, err := i.Infer(s)

	if enabled {
		_, ok := v.(RegexValue)
		assert.True(t, ok, "Got: %T", v)
	} else {
		sv, sok := v.(StringValue)
		_, uok := v.(UnknownValue)
		assert.True(t, err != nil || uok || (sok && sv.Quote == '\u0000'), "Got: %T", v)
	}
}

func assertStrictStrings(t *testing.T, i Inferrer, enabled bool) {
	t.Helper()
	s := `arbitrary`
	v, err := i.Infer(s)

	if enabled {
		_, ok := v.(UnknownValue)
		assert.True(t, err != nil || ok, "Expected UnknownValue: %T %v", v, v)
	} else {
		v, ok := v.(StringValue)
		assert.True(t, ok && v.Quote == '\u0000', "Expected StringValue: %T %v", v, v)
	}
}

func TestIsCollection(t *testing.T) {
	type Test struct {
		Value    Value
		Expected bool
	}

	tests := []Test{
		{BoolValue{}, false},
		{FloatValue{}, false},
		{IntValue{}, false},
		{ListElementValue{}, false},
		{ListValue{}, true},
		{MapElementValue{}, false},
		{MapValue{}, true},
		{NullValue{}, false},
		{RegexValue{}, false},
		{StringValue{}, false},
		{UnknownValue{}, false},
	}

	for _, test := range tests {
		_, ok := test.Value.(CollectionValue)
		assert.Equal(t, test.Expected, ok, "%T", test.Value)
	}
}

func TestInfer(t *testing.T) {
	infer := Inferrer{}.ComplexTypes(true).StrictStrings(true)

	tests := map[string]Value{
		`true`:          BoolValue{true},
		`false`:         BoolValue{false},
		`0.0`:           FloatValue{0.0},
		`.10`:           FloatValue{0.10},
		`-1.0`:          FloatValue{-1.0},
		`0`:             IntValue{0},
		`10`:            IntValue{10},
		`-1`:            IntValue{-1},
		`/.*/`:          RegexValue{`.*`},
		`/.*//`:         RegexValue{`.*/`},
		`"/\".*\"/"`:    RegexValue{`\".*\"`},
		`"testing"`:     StringValue{"testing", '"'},
		`'testing'`:     StringValue{"testing", '\''},
		`""`:            StringValue{``, '"'},
		`''`:            StringValue{``, '\''},
		`'"'`:           StringValue{`"`, '\''},
		`arg[0]`:        ListElementValue{V: ListValue{Name: "arg"}, Index: 0},
		`option["foo"]`: MapElementValue{V: MapValue{Name: "option"}, Key: "foo"},
		`arg`:           UnknownValue{"arg"},
		`option`:        UnknownValue{"option"},
		`arbitrary`:     UnknownValue{"arbitrary"},
		`["string"]`:    ListValue{V: []Value{StringValue{V: `string`, Quote: '"'}}},
		`["string", 10, false, /.*/]`: ListValue{
			V: []Value{
				StringValue{V: `string`, Quote: '"'},
				IntValue{10},
				BoolValue{false},
				RegexValue{`.*`},
			},
		},
	}

	for input, expected := range tests {
		actual, err := infer.Infer(input)
		if !assert.NoError(t, err, input) {
			continue
		}

		assert.Equal(t, expected, actual, input)
	}
}

func TestInferInvalid(t *testing.T) {
	infer := Inferrer{}.ComplexTypes(true).StrictStrings(false)

	tests := []string{
		`option[option[foo]]`,
		`arg[0.1]`,
	}

	for _, input := range tests {
		actual, err := infer.Infer(input)
		assert.Error(t, err, input)
		assert.Equal(t, NullValue{}, actual, input)
	}
}

func TestGuessTypesValue(t *testing.T) {
	infer := Inferrer{}.ComplexTypes(true).StrictStrings(true)

	tests := map[string]interface{}{
		`"testing"`: StringValue{"testing", '"'},
		`'testing'`: StringValue{"testing", '\''},
		`""`:        StringValue{"", '"'},
		`''`:        StringValue{"", '\''},
		`'/.*/'`:    RegexValue{`.*`},
		`[]`:        ListValue{[]Value{}, ""},
		`arbitrary`: UnknownValue{`arbitrary`},
	}

	for input, expected := range tests {
		actual, err := infer.Infer(input)
		if !assert.NoError(t, err, input) {
			continue
		}

		assert.Equal(t, expected, actual, input)
	}
}

func TestSplitListLiteral(t *testing.T) {
	tests := map[string][]string{
		``:                         {},
		`,`:                        {``, ``},
		`test`:                     {`test`},
		`'test'`:                   {`'test'`},
		`“test”`:                   {`"test"`},
		`foo, bar, bat`:            {`foo`, `bar`, `bat`},
		" foo, \tbar , \t\r  bat ": {`foo`, `bar`, `bat`},
		`foo, 'bar', "bat"`:        {`foo`, `'bar'`, `"bat"`},
		`foo, "bar, bat"`:          {`foo`, `"bar, bat"`},
		`foo, /pattern/`:           {`foo`, `/pattern/`},
		`foo, /foo, bar/`:          {`foo`, `/foo, bar/`},
		`foo, /"foo"/`:             {`foo`, `/"foo"/`},
		`'wubba', /^f.*/, 10`:      {`'wubba'`, `/^f.*/`, `10`},
	}

	for test, expected := range tests {
		result := splitListLiteral(test)
		assert.Equal(t, expected, result, test)
	}
}

func TestInferAll(t *testing.T) {
	infer := Inferrer{}.ComplexTypes(true).StrictStrings(true)

	tests := []string{`"foo"`, `10`, `1.0`, `false`}
	expected := []Value{
		StringValue{V: "foo", Quote: '"'},
		IntValue{10},
		FloatValue{1.0},
		BoolValue{false},
	}

	for i, input := range tests {
		actual, err := infer.Infer(input)
		if !assert.NoError(t, err, input) {
			continue
		}

		assert.Equal(t, expected[i], actual, input)
	}
}
