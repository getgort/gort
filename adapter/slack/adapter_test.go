package slack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScrubMarkdown(t *testing.T) {
	tests := map[string]string{
		"<https://google.com>":                                       "https://google.com",
		"<http://very-serio.us|very-serio.us>":                       "very-serio.us",
		"<mailto:matthew.titmus@gmail.com|matthew.titmus@gmail.com>": "matthew.titmus@gmail.com",
		"curl -I <http://very-serio.us>":                             "curl -I http://very-serio.us",
		"curl -I <http://very-serio.us|very-serio.us>":               "curl -I very-serio.us",
	}

	for test, expected := range tests {
		assert.Equal(t, expected, ScrubMarkdown(test))
	}
}
