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
)

func TestTokenize(t *testing.T) {
	inputs := map[string][]string{
		`echo -n foo bar`:             {`echo`, `-n`, `foo`, `bar`},
		`echo -n "foo bar"`:           {`echo`, `-n`, `"foo bar"`},
		`echo "What's" "\"this\"?"`:   {`echo`, `"What's"`, `""this"?"`},
		``:                            {},
		`"" ""`:                       {`""`, `""`},
		`schedule “@every 1m” whoami`: {`schedule`, `"@every 1m"`, `whoami`},
		`echo "hello\nthere"`: {`echo`, `"hello
there"`},
		`echo "hello\nhello"`: {`echo`, "\"hello\nhello\""},
		`echo "hi\tthere"`: {`echo`, `"hi	there"`},
		`echo 'hello\nthere'`: {`echo`, `'hello\nthere'`},
	}

	for in, expected := range inputs {
		token, err := Tokenize(in)
		assert.NoError(t, err, in)
		assert.Equal(t, expected, token, in)
	}
}

func TestTokenizeErrors(t *testing.T) {
	inputs := map[string]string{
		`\`:  "unterminated control character at 1",
		`"`:  "unterminated quote at 1",
		`'`:  "unterminated quote at 1",
		`'"`: "unterminated quote at 1",
	}

	for in, expected := range inputs {
		_, err := Tokenize(in)
		assert.Error(t, err, in)

		msg := err.Error()
		assert.Equal(t, expected, msg, in)

		assert.IsType(t, TokenizeError{}, err, in)
	}
}

func TestIsQuotationMark(t *testing.T) {
	assert.NotNil(t, quotationMarkCategory('"'), "Double universal quotation mark U+0022")
	assert.NotNil(t, quotationMarkCategory('“'), "English left double quotation mark U+201C")
	assert.NotNil(t, quotationMarkCategory('”'), "English right double quotation mark U+201D")
	assert.NotNil(t, quotationMarkCategory('„'), "Double Low-9 quotation mark U+201E")
	//assert.NotNil(t, quotationMarkCategory('«'), "Left-Pointing Double Angle Quotation Mark U+00AB")
	//assert.NotNil(t, quotationMarkCategory('»'), "Right-Pointing Double Angle Quotation Mark U+00BB")
	//assert.NotNil(t, quotationMarkCategory('\''), "Single quote")
}
