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
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfer(t *testing.T) {
	tests := map[string]Value{
		`true`:       BoolValue{true},
		`false`:      BoolValue{false},
		`0.0`:        FloatValue{0.0},
		`.10`:        FloatValue{0.10},
		`-1.0`:       FloatValue{-1.0},
		`0`:          IntValue{0},
		`10`:         IntValue{10},
		`-1`:         IntValue{-1},
		`"testing"`:  StringValue{"testing", '"'},
		`'testing'`:  StringValue{"testing", '\''},
		`""`:         StringValue{``, '"'},
		`''`:         StringValue{``, '\''},
		`'"'`:        StringValue{`"`, '\''},
		`arbitrary`:  StringValue{"arbitrary", '\u0000'},
		`/.*/`:       RegexValue{`.*`},
		`/.*//`:      RegexValue{`.*/`},
		`"/\".*\"/"`: RegexValue{`\".*\"`},
	}

	for input, expected := range tests {
		actual, err := Infer(input, false, false)
		if !assert.NoError(t, err, input) {
			continue
		}

		assert.Equal(t, expected, actual)
	}
}

func TestValueEvaluation(t *testing.T) {
	type Expected struct {
		Value   Value
		String  string
		Resolve interface{}
	}

	tests := map[string]Expected{
		`true`:       {BoolValue{true}, "true", true},
		`0.0`:        {FloatValue{0.0}, "0.000000", 0.0},
		`-1.0`:       {FloatValue{-1.0}, "-1.000000", -1.0},
		`0`:          {IntValue{0}, "0", 0},
		`10`:         {IntValue{10}, "10", 10},
		`-1`:         {IntValue{-1}, "-1", -1},
		`"testing"`:  {StringValue{"testing", '"'}, "testing", "testing"},
		`'testing'`:  {StringValue{"testing", '\''}, "testing", "testing"},
		`arbitrary`:  {StringValue{"arbitrary", '\u0000'}, "arbitrary", "arbitrary"},
		`/.*/`:       {RegexValue{`.*`}, `.*`, regexp.MustCompilePOSIX(`.*`)},
		`/.*//`:      {RegexValue{`.*/`}, `.*/`, regexp.MustCompilePOSIX(`.*/`)},
		`"/\".*\"/"`: {RegexValue{`\".*\"`}, `\".*\"`, regexp.MustCompilePOSIX(`\".*\"`)},
	}

	for input, expected := range tests {
		term, err := Infer(input, false, false)
		if !assert.NoError(t, err, input) {
			continue
		}

		resolved, err := term.Resolve()
		if !assert.NoError(t, err, input) {
			continue
		}

		assert.Equal(t, expected.Value, term, input)
		assert.Equal(t, expected.String, term.String(), input)
		assert.Equal(t, expected.Resolve, resolved, input)
	}
}

func TestGuessTypesValueStrict(t *testing.T) {
	tests := map[string]interface{}{
		`"testing"`: StringValue{"testing", '"'},
		`'testing'`: StringValue{"testing", '\''},
		`""`:        StringValue{"", '"'},
		`''`:        StringValue{"", '\''},
		`'/.*/'`:    RegexValue{`.*`},
	}

	for input, expected := range tests {
		actual, err := Infer(input, false, true)
		if !assert.NoError(t, err, input) {
			continue
		}

		assert.Equal(t, expected, actual)
	}

	_, err := Infer("arbitrary", false, true)
	assert.Error(t, err, "arbitrary")
}
