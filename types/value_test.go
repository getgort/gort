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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoolValueEquals(t *testing.T) {
	type Test struct {
		Input      bool
		ComparedTo Value
	}

	tests := map[Test]bool{
		{false, BoolValue{V: false}}:      true,
		{false, BoolValue{V: true}}:       false,
		{true, BoolValue{V: false}}:       false,
		{false, FloatValue{V: 0.0}}:       false,
		{false, IntValue{V: 0}}:           true,
		{false, RegexValue{V: `^false$`}}: true,
		{false, StringValue{V: "false"}}:  true,
		{true, BoolValue{V: true}}:        true,
		{true, FloatValue{V: 0.0}}:        false,
		{true, IntValue{V: 1}}:            true,
		{true, RegexValue{V: `^true$`}}:   true,
		{true, StringValue{V: "true"}}:    true,
		{true, IntValue{V: 2}}:            false,
		{false, StringValue{V: "foo"}}:    false,
	}

	for test, expected := range tests {
		input := BoolValue{V: test.Input}
		comparedTo := test.ComparedTo

		result := input.Equals(comparedTo)
		assert.Equal(t, expected, result, msg(test.Input, test.ComparedTo))

		result = comparedTo.Equals(input)
		assert.Equal(t, expected, result, msg(test.Input, test.ComparedTo))
	}
}

func TestFloatValueEquals(t *testing.T) {
	type Test struct {
		Input      float64
		ComparedTo Value
	}

	tests := map[Test]bool{
		{0.0, BoolValue{V: false}}:          false,
		{0.0, FloatValue{V: 0.0}}:           true,
		{0.0, FloatValue{V: 1.0}}:           false,
		{0.0, FloatValue{V: -1.0}}:          false,
		{0.0, IntValue{V: 0}}:               true,
		{0.0, RegexValue{V: `^0(.0*)?$`}}:   true,
		{0.0, StringValue{V: "0.0"}}:        false,
		{1.0, BoolValue{V: true}}:           false,
		{1.0, FloatValue{V: 1.0}}:           true,
		{1.0, IntValue{V: 1}}:               true,
		{1.0, RegexValue{V: `^1(.0*)?$`}}:   true,
		{1.0, StringValue{V: "1.0"}}:        false,
		{-1.0, BoolValue{V: false}}:         false,
		{-1.0, FloatValue{V: -1.0}}:         true,
		{-1.0, IntValue{V: -1}}:             true,
		{-1.0, RegexValue{V: `^-1(.0*)?$`}}: true,
		{-1.0, StringValue{V: "-1.0"}}:      false,
	}

	for test, expected := range tests {
		input := FloatValue{V: test.Input}
		comparedTo := test.ComparedTo

		result := input.Equals(comparedTo)
		assert.Equal(t, expected, result, msg(test.Input, test.ComparedTo))

		result = comparedTo.Equals(input)
		assert.Equal(t, expected, result, msg(test.Input, test.ComparedTo))
	}
}

func TestIntValueEquals(t *testing.T) {
	type Test struct {
		Input      int
		ComparedTo Value
	}

	tests := map[Test]bool{
		{0, BoolValue{V: false}}:     true,
		{0, FloatValue{V: 0.0}}:      true,
		{0, IntValue{V: 0}}:          true,
		{0, RegexValue{V: `^0*$`}}:   true,
		{0, StringValue{V: "0"}}:     false,
		{1, BoolValue{V: true}}:      true,
		{1, FloatValue{V: 1.0}}:      true,
		{1, IntValue{V: 1}}:          true,
		{1, RegexValue{V: `^1*$`}}:   true,
		{1, StringValue{V: "1"}}:     false,
		{-1, BoolValue{V: false}}:    false,
		{-1, FloatValue{V: -1.0}}:    true,
		{-1, IntValue{V: -1}}:        true,
		{-1, RegexValue{V: `^-1*$`}}: true,
		{-1, StringValue{V: "-1"}}:   false,
	}

	for test, expected := range tests {
		input := IntValue{V: test.Input}
		comparedTo := test.ComparedTo

		result := input.Equals(comparedTo)
		assert.Equal(t, expected, result, msg(test.Input, test.ComparedTo))

		result = comparedTo.Equals(input)
		assert.Equal(t, expected, result, msg(test.Input, test.ComparedTo))
	}
}

func TestRegexValueEquals(t *testing.T) {
	type Test struct {
		Input      string
		ComparedTo Value
	}

	tests := map[Test]bool{
		{`^false$`, BoolValue{V: false}}:  true,
		{`^false$`, BoolValue{V: true}}:   false,
		{`^0(.0*)?$`, FloatValue{V: 0.0}}: true,
		{`^0(.0*)?$`, FloatValue{V: 1.0}}: false,
		{`^0$`, IntValue{V: 0}}:           true,
		{`^0$`, IntValue{V: 1}}:           false,
		{`^.*$`, RegexValue{V: `^.*$`}}:   true,
		{`^foo$`, StringValue{V: `foo`}}:  true,
		{`^foo$`, StringValue{V: "bar"}}:  false,
	}

	for test, expected := range tests {
		input := RegexValue{V: test.Input}
		comparedTo := test.ComparedTo

		result := input.Equals(comparedTo)
		assert.Equal(t, expected, result, msg(test.Input, test.ComparedTo))

		result = comparedTo.Equals(input)
		assert.Equal(t, expected, result, msg(test.Input, test.ComparedTo))
	}
}

func TestStringValueEquals(t *testing.T) {
	type Test struct {
		Input      string
		ComparedTo Value
	}

	tests := map[Test]bool{
		{"foo", BoolValue{V: false}}:   false,
		{"foo", FloatValue{V: 0.0}}:    false,
		{"foo", IntValue{V: 0}}:        false,
		{"foo", RegexValue{V: `^.*$`}}: true,
		{"foo", StringValue{V: "foo"}}: true,
		{"foo", StringValue{V: "0"}}:   false,
	}

	for test, expected := range tests {
		input := StringValue{V: test.Input}
		comparedTo := test.ComparedTo

		result := input.Equals(comparedTo)
		assert.Equal(t, expected, result, msg(test.Input, test.ComparedTo))

		result = comparedTo.Equals(input)
		assert.Equal(t, expected, result, msg(test.Input, test.ComparedTo))
	}
}

func msg(input interface{}, comparedTo Value) string {
	return fmt.Sprintf("Input=%v (%T) ComparedTo=%v (%T)", input, input, comparedTo, comparedTo)
}
