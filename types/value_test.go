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
	type Expected struct {
		Result   bool
		HasError bool
	}

	tests := map[Test]Expected{
		{false, BoolValue{Value: false}}:      {true, false},
		{false, BoolValue{Value: true}}:       {false, false},
		{true, BoolValue{Value: false}}:       {false, false},
		{false, FloatValue{Value: 0.0}}:       {false, true},
		{false, IntValue{Value: 0}}:           {true, false},
		{false, RegexValue{Value: `^false$`}}: {true, false},
		{false, StringValue{Value: "false"}}:  {true, false},
		{true, BoolValue{Value: true}}:        {true, false},
		{true, FloatValue{Value: 0.0}}:        {false, true},
		{true, IntValue{Value: 1}}:            {true, false},
		{true, RegexValue{Value: `^true$`}}:   {true, false},
		{true, StringValue{Value: "true"}}:    {true, false},
		{true, IntValue{Value: 2}}:            {false, true},
		{false, StringValue{Value: "foo"}}:    {false, true},
	}

	for test, expected := range tests {
		input := BoolValue{Value: test.Input}
		comparedTo := test.ComparedTo

		result, err := input.Equals(comparedTo)
		if expected.HasError && !assert.Error(t, err, msg(test.Input, test.ComparedTo)) {
			continue
		}
		if !expected.HasError && !assert.NoError(t, err, msg(test.Input, test.ComparedTo)) {
			continue
		}

		assert.Equal(t, expected.Result, result, msg(test.Input, test.ComparedTo))

		result, err = comparedTo.Equals(input)
		if expected.HasError && !assert.Error(t, err, msg(test.Input, test.ComparedTo)) {
			continue
		}
		if !expected.HasError && !assert.NoError(t, err, msg(test.Input, test.ComparedTo)) {
			continue
		}

		assert.Equal(t, expected.Result, result, msg(test.Input, test.ComparedTo))
	}
}

func TestFloatValueEquals(t *testing.T) {
	type Test struct {
		Input      float64
		ComparedTo Value
	}

	type Expected struct {
		Result   bool
		HasError bool
	}

	tests := map[Test]Expected{
		{0.0, BoolValue{Value: false}}:       {false, true},
		{0.0, FloatValue{Value: 0.0}}:        {true, false},
		{0.0, FloatValue{Value: 1.0}}:        {false, false},
		{0.0, FloatValue{Value: -1.0}}:       {false, false},
		{0.0, IntValue{Value: 0}}:            {true, false},
		{0.0, RegexValue{Value: `^0.0*$`}}:   {true, false},
		{0.0, StringValue{Value: "0.0"}}:     {false, true},
		{1.0, BoolValue{Value: true}}:        {false, true},
		{1.0, FloatValue{Value: 1.0}}:        {true, false},
		{1.0, IntValue{Value: 1}}:            {true, false},
		{1.0, RegexValue{Value: `^1.0*$`}}:   {true, false},
		{1.0, StringValue{Value: "1.0"}}:     {false, true},
		{-1.0, BoolValue{Value: false}}:      {false, true},
		{-1.0, FloatValue{Value: -1.0}}:      {true, false},
		{-1.0, IntValue{Value: -1}}:          {true, false},
		{-1.0, RegexValue{Value: `^-1.0*$`}}: {true, false},
		{-1.0, StringValue{Value: "-1.0"}}:   {false, true},
	}

	for test, expected := range tests {
		input := FloatValue{Value: test.Input}
		comparedTo := test.ComparedTo

		result, err := input.Equals(comparedTo)
		if expected.HasError && !assert.Error(t, err, msg(test.Input, test.ComparedTo)) {
			continue
		}
		if !expected.HasError && !assert.NoError(t, err, msg(test.Input, test.ComparedTo)) {
			continue
		}

		assert.Equal(t, expected.Result, result, msg(test.Input, test.ComparedTo))

		result, err = comparedTo.Equals(input)
		if expected.HasError && !assert.Error(t, err, msg(test.Input, test.ComparedTo)) {
			continue
		}
		if !expected.HasError && !assert.NoError(t, err, msg(test.Input, test.ComparedTo)) {
			continue
		}

		assert.Equal(t, expected.Result, result, msg(test.Input, test.ComparedTo))
	}
}

func TestIntValueEquals(t *testing.T) {
	type Test struct {
		Input      int
		ComparedTo Value
	}

	type Expected struct {
		Result   bool
		HasError bool
	}

	tests := map[Test]Expected{
		{0, BoolValue{Value: false}}:     {true, false},
		{0, FloatValue{Value: 0.0}}:      {true, false},
		{0, IntValue{Value: 0}}:          {true, false},
		{0, RegexValue{Value: `^0*$`}}:   {true, false},
		{0, StringValue{Value: "0"}}:     {false, true},
		{1, BoolValue{Value: true}}:      {true, false},
		{1, FloatValue{Value: 1.0}}:      {true, false},
		{1, IntValue{Value: 1}}:          {true, false},
		{1, RegexValue{Value: `^1*$`}}:   {true, false},
		{1, StringValue{Value: "1"}}:     {false, true},
		{-1, BoolValue{Value: false}}:    {false, true},
		{-1, FloatValue{Value: -1.0}}:    {true, false},
		{-1, IntValue{Value: -1}}:        {true, false},
		{-1, RegexValue{Value: `^-1*$`}}: {true, false},
		{-1, StringValue{Value: "-1"}}:   {false, true},
	}

	for test, expected := range tests {
		input := IntValue{Value: test.Input}
		comparedTo := test.ComparedTo

		result, err := input.Equals(comparedTo)
		if expected.HasError && !assert.Error(t, err, msg(test.Input, test.ComparedTo)) {
			continue
		}
		if !expected.HasError && !assert.NoError(t, err, msg(test.Input, test.ComparedTo)) {
			continue
		}

		assert.Equal(t, expected.Result, result, msg(test.Input, test.ComparedTo))

		result, err = comparedTo.Equals(input)
		if expected.HasError && !assert.Error(t, err, msg(test.Input, test.ComparedTo)) {
			continue
		}
		if !expected.HasError && !assert.NoError(t, err, msg(test.Input, test.ComparedTo)) {
			continue
		}

		assert.Equal(t, expected.Result, result, msg(test.Input, test.ComparedTo))
	}
}

func TestRegexValueEquals(t *testing.T) {
	type Test struct {
		Input      string
		ComparedTo Value
	}

	type Expected struct {
		Result   bool
		HasError bool
	}

	tests := map[Test]Expected{
		{`^false$`, BoolValue{Value: false}}: {true, false},
		{`^false$`, BoolValue{Value: true}}:  {false, false},
		{`^0.0*$`, FloatValue{Value: 0.0}}:   {true, false},
		{`^0.0*$`, FloatValue{Value: 1.0}}:   {false, false},
		{`^0$`, IntValue{Value: 0}}:          {true, false},
		{`^0$`, IntValue{Value: 1}}:          {false, false},
		{`^.*$`, RegexValue{Value: `^.*$`}}:  {false, true},
		{`^foo$`, StringValue{Value: `foo`}}: {true, false},
		{`^foo$`, StringValue{Value: "bar"}}: {false, false},
	}

	for test, expected := range tests {
		input := RegexValue{Value: test.Input}
		comparedTo := test.ComparedTo

		result, err := input.Equals(comparedTo)
		if expected.HasError && !assert.Error(t, err, msg(test.Input, test.ComparedTo)) {
			continue
		}
		if !expected.HasError && !assert.NoError(t, err, msg(test.Input, test.ComparedTo)) {
			continue
		}

		assert.Equal(t, expected.Result, result, msg(test.Input, test.ComparedTo))

		result, err = comparedTo.Equals(input)
		if expected.HasError && !assert.Error(t, err, msg(test.Input, test.ComparedTo)) {
			continue
		}
		if !expected.HasError && !assert.NoError(t, err, msg(test.Input, test.ComparedTo)) {
			continue
		}

		assert.Equal(t, expected.Result, result, msg(test.Input, test.ComparedTo))
	}
}

func TestStringValueEquals(t *testing.T) {
	type Test struct {
		Input      string
		ComparedTo Value
	}

	type Expected struct {
		Result   bool
		HasError bool
	}

	tests := map[Test]Expected{
		{"foo", BoolValue{Value: false}}:   {false, true},
		{"foo", FloatValue{Value: 0.0}}:    {false, true},
		{"foo", IntValue{Value: 0}}:        {false, true},
		{"foo", RegexValue{Value: `^.*$`}}: {true, false},
		{"foo", StringValue{Value: "foo"}}: {true, false},
		{"foo", StringValue{Value: "0"}}:   {false, false},
	}

	for test, expected := range tests {
		input := StringValue{Value: test.Input}
		comparedTo := test.ComparedTo

		result, err := input.Equals(comparedTo)
		if expected.HasError && !assert.Error(t, err, msg(test.Input, test.ComparedTo)) {
			continue
		}
		if !expected.HasError && !assert.NoError(t, err, msg(test.Input, test.ComparedTo)) {
			continue
		}

		assert.Equal(t, expected.Result, result, msg(test.Input, test.ComparedTo))

		result, err = comparedTo.Equals(input)
		if expected.HasError && !assert.Error(t, err, msg(test.Input, test.ComparedTo)) {
			continue
		}
		if !expected.HasError && !assert.NoError(t, err, msg(test.Input, test.ComparedTo)) {
			continue
		}

		assert.Equal(t, expected.Result, result, msg(test.Input, test.ComparedTo))
	}
}

func msg(input interface{}, comparedTo Value) string {
	return fmt.Sprintf("Input=%v (%T) ComparedTo=%v (%T)", input, input, comparedTo, comparedTo)
}
