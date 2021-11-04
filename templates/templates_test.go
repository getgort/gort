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

package templates

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/getgort/gort/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testStructuredEnvelope = data.CommandResponseEnvelope{
	Request: data.CommandRequest{
		CommandEntry: data.CommandEntry{
			Bundle: data.Bundle{
				Name: "gort",
			},
			Command: data.BundleCommand{
				Name:       "echo",
				Executable: []string{"echo"},
			},
		},
		Parameters: []string{"foo", "bar"},
	},
	Response: data.CommandResponse{
		Lines:      []string{"foo bar"},
		Out:        "foo bar",
		Title:      "Error",
		Structured: false,
	},
}

const payloadJSON = `{
  "User": "Assistant to the Regional Manager Dwight",
  "Requestor": "Michael Scott",
  "Company": "Dunder Mifflin",
  "Results": [
    {
      "Name": "Farmhouse Thai Cuisine",
      "Reviews": 1234,
      "Description": "Awesome",
      "Stars": 4,
      "Image": "https://s3-media3.fl.yelpcdn.com/bphoto/c7ed05m9lC2EmA3Aruue7A/o.jpg"
    }
  ]
}`

// const testTemplate = `{{ text | emoji false | monospace false }}
// Hello, {{ .Payload.User }}!

// *{{ .Payload.Requestor }}* wants to know where you'd like to take the {{ .Payload.Company }} investors to dinner tonight.

// *Please select a restaurant:*
// {{ endtext }}

// {{ divider }}

// {{ range $index, $result := .Payload.Results }}
// 	{{ $stars := int $result.Stars }}
// 	{{ section }}
// 		{{ text }}
// 			*{{ $result.Name }}*
// 			{{ repeat $stars "::star::" }} {{ $result.Reviews }} reviews
// 			{{ $result.Description }}
// 		{{ endtext }}
// 		{{ image $result.Image }}
// 	{{ endsection }}
// {{ end }}
// `

func TestMain(m *testing.M) {
	json.Unmarshal([]byte(payloadJSON), &testStructuredEnvelope.Payload)
	m.Run()
}

func TestCalcLine(t *testing.T) {
	text := "This is line 1.\n" +
		"This is line 2.\n" +
		"\n" +
		"This is line 4."

	// Out of bounds should return -1
	assert.Equal(t, -1, calculateLineNumber(text, -1))
	assert.Equal(t, -1, calculateLineNumber(text, 1000))

	for i, line := range strings.Split(text, "\n") {
		if len(line) == 0 {
			continue
		}

		start := strings.Index(text, line)
		end := start + len(line)
		assert.Equal(t, i+1, calculateLineNumber(text, start))
		assert.Equal(t, i+1, calculateLineNumber(text, end))
	}
}

func TestNextTag(t *testing.T) {
	text := `<<Text|{"Foo":"Bar"}>>This is text.<<TextEnd|{}>>`

	tag, json, first, last := nextTag(text, 0)
	assert.Equal(t, "Text", tag)
	assert.Equal(t, `{"Foo":"Bar"}`, json)
	assert.Equal(t, 0, first)
	assert.Equal(t, 21, last)

	tag, json, first, last = nextTag(text, last)
	assert.Equal(t, "TextEnd", tag)
	assert.Equal(t, `{}`, json)
	assert.Equal(t, 35, first)
	assert.Equal(t, 48, last)

	tag, json, first, last = nextTag(text, last)
	assert.Equal(t, "", tag)
	assert.Equal(t, "", json)
	assert.Equal(t, -1, first)
	assert.Equal(t, -1, last)
}

func TestTransformAndEncodeText(t *testing.T) {
	tests := []struct {
		Template       string
		Transformed    string
		TransformError string
		Encoded        OutputElements
		EncodeError    string
	}{
		{
			Template:    `{{ divider }}`,
			Transformed: `<<Divider|{}>>`,
			Encoded: OutputElements{
				Elements: []OutputElement{
					&Divider{
						Tag: Tag{FirstIndex: 0, LastIndex: 13},
					},
				},
			},
		},
		{
			Template:    `{{ header }}Error{{ endheader }}`,
			Transformed: `<<Header|{}>>Error<<HeaderEnd|{}>>`,
			Encoded: OutputElements{
				Elements: []OutputElement{
					&Header{
						Tag:   Tag{FirstIndex: 0, LastIndex: 33},
						Title: "Error",
					},
				},
			},
		},
		{
			Template:    `{{ header | color "#FF0000" }}Error{{ endheader }}`,
			Transformed: `<<Header|{"Color":"#FF0000"}>>Error<<HeaderEnd|{}>>`,
			Encoded: OutputElements{
				Elements: []OutputElement{
					&Header{
						Tag:   Tag{FirstIndex: 0, LastIndex: 50},
						Color: "#FF0000",
						Title: "Error",
					},
				},
			},
		},
		{
			Template:    `{{ header | color "FF0000" }}Error{{ endheader }}`,
			Transformed: `<<Header|{"Color":"#FF0000"}>>Error<<HeaderEnd|{}>>`,
			Encoded: OutputElements{
				Elements: []OutputElement{
					&Header{
						Tag:   Tag{FirstIndex: 0, LastIndex: 50},
						Color: "#FF0000",
						Title: "Error",
					},
				},
			},
		},
		{
			Template:       `{{ header | color "FF 00 00" }}Error{{ endheader }}`,
			Transformed:    `<<Header|{"Color":"#FF0000"}>>Error<<HeaderEnd|{}>>`,
			TransformError: `template: gort:echo foo bar:1:12: executing "gort:echo foo bar" at <color "FF 00 00">: error calling color: colors should be expressed in RGB hex format: #123456`,
		},
		{
			Template:    `{{ image "https://example.com/image.jpg" }}`,
			Transformed: `<<Image|{"URL":"https://example.com/image.jpg"}>>`,
			Encoded: OutputElements{
				Elements: []OutputElement{
					&Image{
						Tag: Tag{FirstIndex: 0, LastIndex: 48},
						URL: "https://example.com/image.jpg",
					},
				},
			},
		},
		{
			Template:    `{{ text | emoji true | markdown true | monospace true }}Test`,
			Transformed: `<<Text|{"Emoji":true,"Markdown":true,"Monospace":true}>>Test`,
			EncodeError: "unmatched {{text}} on line 1",
		},
		{
			Template:    `Test{{ endtext }}`,
			Transformed: `Test<<TextEnd|{}>>`,
			EncodeError: "unmatched {{endtext}} on line 1",
		},
		{
			Template:    `{{ text | emoji true | markdown true | monospace true }}Test{{ endtext }}`,
			Transformed: `<<Text|{"Emoji":true,"Markdown":true,"Monospace":true}>>Test<<TextEnd|{}>>`,
			Encoded: OutputElements{
				Elements: []OutputElement{
					&Text{
						Tag:       Tag{FirstIndex: 0, LastIndex: 73},
						Emoji:     true,
						Markdown:  true,
						Monospace: true,
						Text:      "Test",
					},
				},
			},
		},
		{
			Template:    `{{ text | emoji true | markdown true | monospace true }}{{ .Payload.Company }}{{ endtext }}`,
			Transformed: `<<Text|{"Emoji":true,"Markdown":true,"Monospace":true}>>Dunder Mifflin<<TextEnd|{}>>`,
			Encoded: OutputElements{
				Elements: []OutputElement{
					&Text{
						Tag:       Tag{FirstIndex: 0, LastIndex: 83},
						Emoji:     true,
						Markdown:  true,
						Monospace: true,
						Text:      "Dunder Mifflin",
					},
				},
			},
		},
	}

	for _, test := range tests {
		tf, err := Transform(test.Template, testStructuredEnvelope)
		if test.TransformError != "" {
			require.EqualError(t, err, test.TransformError)
			continue
		} else {
			require.NoError(t, err)
		}
		require.Equal(t, test.Transformed, tf)

		enc, err := EncodeElements(tf)
		if test.EncodeError != "" {
			assert.EqualError(t, err, test.EncodeError)
			continue
		} else {
			assert.NoError(t, err)
		}

		assert.Equal(t, test.Encoded, enc, tf)
	}
}
