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
	"fmt"
	"strings"
	"unicode"
)

// Tokenize takes an input string and splits it into tokens. Any control
// character sequences ("\n", "\t", etc), pass through in their original
// form.
// Examples:
//    echo -n foo bar -> {"echo", "-n", "foo", "bar"}
//    echo -n "foo bar" -> {"echo", "-n", "foo bar"}
//    echo "What's" "\"this\"?" -> {"echo", "What's", "\"this\"?"}
func Tokenize(input string) ([]string, error) {
	const RuneNull = rune(0)

	b := strings.Builder{}
	tokens := []string{}

	input = strings.Trim(input, " \t\n\v\f\r\u0085\u00A0")

	quote := RuneNull
	quoteStart := 0

	control := false

	for i, ch := range input {
		switch {

		// Backslash turns on the control flag.
		case ch == '\\':
			b.WriteRune(ch)
			control = true

		// If the control flag is set, append the entire control character to the token.
		case control:
			b.WriteRune(ch)
			control = false

		// Spaces outside of quotes are token delimitters.
		case unicode.IsSpace(ch) && quote == RuneNull:
			if t := b.String(); len(t) > 0 {
				tokens = append(tokens, t)
			}
			b.Reset()

		// Everything inside a pair of quotes is added to the same token.
		case ch == quote:
			tokens = append(tokens, b.String())
			quote = RuneNull
			b.Reset()

		// Turn quote-mode on and off.
		case ch == '"':
			fallthrough

		case ch == '\'':
			if quote == RuneNull {
				quote = ch
				quoteStart = i
			} else {
				b.WriteRune(ch)
			}

		// Anything else gets appended to the current token.
		default:
			b.WriteRune(ch)
		}
	}

	// Grab that last token
	if t := b.String(); len(t) > 0 {
		tokens = append(tokens, t)
	}

	if control {
		return tokens, TokenizeError{"unterminated control character at %d", len(input)}
	}

	if quote != RuneNull {
		return tokens, TokenizeError{"unterminated quote at %d", quoteStart + 1}
	}

	return tokens, nil
}

type TokenizeError struct {
	Text     string
	Position int
}

func (e TokenizeError) Error() string {
	return fmt.Sprintf(e.Text, e.Position)
}
