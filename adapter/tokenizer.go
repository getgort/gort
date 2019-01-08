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
