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
		"curl -I <https://very-serio.us|https://very-serio.us>":      "curl -I https://very-serio.us",
	}

	for test, expected := range tests {
		assert.Equal(t, expected, ScrubMarkdown(test))
	}
}
