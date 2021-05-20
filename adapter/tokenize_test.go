package adapter

import (
	"strings"
	"testing"
)

func TestTokenizeParameters(t *testing.T) {
	// Each pair is {original string, expected string}
	testStrings := [][]string{
		{"", ""},
		{"foo", "foo"},
		{"foo bar", "foo bar"},
		{"foo\"bar\"", "foobar"},
		{"foo\"bar bat\"", "foobar bat"},
		{"foo\"bar'bat\"", "foobar'bat"},
		{"foo “bar bat”", "foo bar bat"}, // smart quotes
	}

	for _, pair := range testStrings {
		original := pair[0]
		expected := pair[1]

		tokens := TokenizeParameters(original)
		joined := strings.Join(tokens, " ")

		if joined != expected {
			t.Errorf("Expected: %s, Got: %s\n", expected, joined)
		}
	}
}
