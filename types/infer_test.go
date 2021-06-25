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

func TestInfer(t *testing.T) {
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
		`arg[0]`:        ListValue{Name: "arg", Index: 0},
		`option["foo"]`: MapValue{Name: "option", Key: "foo"},
		`arg`:           UnknownValue{"arg"},
		`option`:        UnknownValue{"option"},
		`arbitrary`:     UnknownValue{"arbitrary"},
	}

	for input, expected := range tests {
		actual, err := Infer(input, false, true)
		if !assert.NoError(t, err, input) {
			continue
		}

		assert.Equal(t, expected, actual)
	}
}

func TestInferInvalid(t *testing.T) {
	tests := []string{
		`option[option[foo]]`,
	}

	for _, input := range tests {
		actual, err := Infer(input, false, true)
		assert.Error(t, err, input)
		assert.Equal(t, NullValue{}, actual, input)
	}
}

func TestGuessTypesValueStrict(t *testing.T) {
	tests := map[string]interface{}{
		`"testing"`: StringValue{"testing", '"'},
		`'testing'`: StringValue{"testing", '\''},
		`""`:        StringValue{"", '"'},
		`''`:        StringValue{"", '\''},
		`'/.*/'`:    RegexValue{`.*`},
		`arbitrary`: UnknownValue{`arbitrary`},
	}

	for input, expected := range tests {
		actual, err := Infer(input, false, true)
		if !assert.NoError(t, err, input) {
			continue
		}

		assert.Equal(t, expected, actual)
	}
}
