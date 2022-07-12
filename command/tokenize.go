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

	"golang.org/x/text/unicode/rangetable"
)

var (
	singleQuotes = rangetable.New('\'', '‘', '’')

	doubleQuotes = rangetable.New('"', '“', '”', '„')
)

const RuneNull = rune(0)

// Tokenize takes an input string and splits it into tokens. Any control
// character sequences ("\n", "\t", etc), pass through in their original
// form.
// Examples:
//    echo -n foo bar -> {"echo", "-n", "foo", "bar"}
//    echo -n "foo bar" -> {"echo", "-n", "foo bar"}
func Tokenize(input string) ([]string, error) {
	b := strings.Builder{}
	tokens := []string{}

	input = strings.TrimSpace(input)

	var quote *unicode.RangeTable = nil
	quoteStart := 0

	control := false

	for i, ch := range input {
		switch {
		// Backslash turns on the control flag.
		case ch == '\\':
			// If we aren't in double-quotes, append the literal backslash.
			if quote != doubleQuotes {
				b.WriteRune(ch)
			}
			control = true

		// If the control flag is set, and we aren't in double-quotes, append
		// the entire control character to the token.
		case control && quote != doubleQuotes:
			b.WriteRune(ch)
			control = false

		// If the control flag is set, and we are in double-quotes, append the
		// unescaped control character to the token.
		case control && quote == doubleQuotes:
			r, err := unescape(ch)
			if err != nil {
				return tokens, TokenizeError{
					Text:     err.Error(),
					Position: i,
				}
			}
			b.WriteRune(r)
			control = false

		// Spaces outside of quotes are token delimiters.
		case unicode.IsSpace(ch) && quote == nil:
			if t := b.String(); len(t) > 0 {
				tokens = append(tokens, t)
			}
			b.Reset()

		// Everything inside a pair of quotes is added to the same token.
		case quote != nil && isMatchingQuotationMark(ch, quote):
			b.WriteRune(getBasicQuote(quote))
			tokens = append(tokens, b.String())
			quote = nil
			b.Reset()

		// Turn quote-mode on and off.
		case quotationMarkCategory(ch) != nil:
			b.WriteRune(getBasicQuote(quotationMarkCategory(ch)))
			if quote == nil {
				quote = quotationMarkCategory(ch)
				quoteStart = i
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

	if quote != nil {
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

// getBasicQuote accepts either the singleQuote or doubleQuote *RangeTable, and
// returns the ascii representative of that category (either ' or ").
func getBasicQuote(table *unicode.RangeTable) rune {
	switch table {
	case singleQuotes:
		return '\''
	case doubleQuotes:
		return '"'
	default:
		return rune(0)
	}
}

// isMatchingQuotationMark returns whether a quotation mark is in the given
// *RangeTable, either singleQuote or doubleQuote.
func isMatchingQuotationMark(r rune, other *unicode.RangeTable) bool {
	return unicode.In(r, other)
}

// quotationMarkCategory returns a pointer to the range of unicode characters
// equivalent to the given quote, or nil if the given character isn't a
// quotation mark.
func quotationMarkCategory(r rune) *unicode.RangeTable {
	switch {
	case unicode.In(r, singleQuotes):
		return singleQuotes
	case unicode.In(r, doubleQuotes):
		return doubleQuotes
	default:
		return nil
	}
}

// unescape translates a given printable character rune into it's corresponding
// control character.
func unescape(r rune) (rune, error) {
	switch r {
	case 'a':
		return '\a', nil
	case 'b':
		return '\b', nil
	case 'f':
		return '\f', nil
	case 'n':
		return '\n', nil
	case 'r':
		return '\r', nil
	case 't':
		return '\t', nil
	case 'v':
		return '\v', nil
	case '\'':
		return '\'', nil
	case '"':
		return '"', nil
	case '?':
		return '?', nil
	case '\\':
		return '\\', nil
	default:
		return RuneNull, fmt.Errorf("undefined control character '%c'", r)
	}
}
