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

package adapter

import (
	"strings"
)

// TokenizeParameters splits a command string into parameter tokens. Any
// trigger characters (i.e., !) are expected to have already been removed.
func TokenizeParameters(commandString string) []string {
	commandString = strings.TrimSpace(commandString)

	tokens := make([]string, 0)
	inDoubleQuote := false
	inSingleQuote := false
	currentToken := ""
	escapeNextRune := false

	for _, char := range commandString {
		if escapeNextRune {
			switch char {
			case 'a':
				currentToken += "\a"
			case 'b':
				currentToken += "\b"
			case 'f':
				currentToken += "\f"
			case 'n':
				currentToken += "\n"
			case 'r':
				currentToken += "\r"
			case 't':
				currentToken += "\t"
			case 'v':
				currentToken += "\v"
			case '\\':
				currentToken += "\\"
			case '\'':
				currentToken += "'"
			case '“': // Smart left quote
				fallthrough
			case '”': // Smart right quote
				fallthrough
			case '"':
				currentToken += "\""
			case '?':
				currentToken += "?"
			}

			escapeNextRune = false
		} else {
			switch char {
			case ' ':
				if inDoubleQuote {
					currentToken += string(char)
				} else if len(currentToken) > 0 {
					tokens = append(tokens, currentToken)
					currentToken = ""
				}
			case '\\':
				escapeNextRune = true
			case '“': // Smart left quote
				fallthrough
			case '”': // Smart right quote
				fallthrough
			case '"':
				if inSingleQuote {
					currentToken += "\""
				} else {
					inDoubleQuote = !inDoubleQuote
				}
			case '\'':
				if inDoubleQuote {
					currentToken += "'"
				} else {
					inSingleQuote = !inSingleQuote
				}
			default:
				currentToken += string(char)
			}
		}
	}

	if len(currentToken) > 0 {
		tokens = append(tokens, currentToken)
	}

	return tokens
}
