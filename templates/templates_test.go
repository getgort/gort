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
)

// var testUnstructuredEnvelope = data.CommandResponseEnvelope{
// 	Request: data.CommandRequest{
// 		CommandEntry: data.CommandEntry{
// 			Bundle: data.Bundle{
// 				Name: "gort",
// 			},
// 			Command: data.BundleCommand{
// 				Name:       "echo",
// 				Executable: []string{"echo"},
// 			},
// 		},
// 		Parameters: []string{"foo", "bar"},
// 	},
// 	Response: data.CommandResponse{
// 		Lines:      []string{"foo bar"},
// 		Out:        "foo bar",
// 		Structured: false,
// 		Payload:    "foo bar",
// 	},
// }

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

const testTemplate = `{{ text | emoji false | monospace true }}
Hello, {{ .Payload.User }}!

*{{ .Payload.Requestor }}* wants to know where you'd like to take the {{ .Payload.Company }} investors to dinner tonight.

*Please select a restaurant:*
{{ endtext }}

{{ divider }}

{{ range $index, $result := .Payload.Results }}
	{{ $stars := int $result.Stars }}
	{{ section }}
		{{ text }}
			*{{ $result.Name }}*
			{{ repeat $stars "::star::" }} {{ $result.Reviews }} reviews
			{{ $result.Description }}
		{{ endtext }}
		{{ image $result.Image }}
	{{ endsection }}
{{ end }}
`

func TestMain(m *testing.M) {
	json.Unmarshal([]byte(payloadJSON), &testStructuredEnvelope.Payload)
	m.Run()
}

// func TestAll(t *testing.T) {
// 	s, err := Transform(testTemplate, testStructuredEnvelope)
// 	require.NoError(t, err)

// 	fmt.Println(s)

// 	_, err = EncodeElements(s)
// 	require.NoError(t, err)
// }

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

func TestTransformText(t *testing.T) {
	tests := []struct {
		Template       string
		Transformed    string
		TransformError string
		Encoded        OutputElements
		EncodeError    string
	}{
		{
			Template:    `{{ text | emoji true | markup true | monospace true }}Test`,
			Transformed: `<<Text|{"Emoji":true,"Markup":true,"Monospace":true}>>Test`,
			EncodeError: "unmatched {{text}} tag on line 1",
		},
		{
			Template:    `Test{{ endtext }}`,
			Transformed: `Test<<TextEnd|{}>>`,
			EncodeError: "unmatched {{textend}} tag on line 1",
		},
		{
			Template:    `{{ text | emoji true | markup true | monospace true }}Test{{ endtext }}`,
			Transformed: `<<Text|{"Emoji":true,"Markup":true,"Monospace":true}>>Test<<TextEnd|{}>>`,
			Encoded: OutputElements{
				Elements: []OutputElement{
					&Text{
						Tag:       Tag{FirstIndex: 0, LastIndex: 71},
						Emoji:     true,
						Markup:    true,
						Monospace: true,
						Text:      "Test",
					},
				},
			},
		},
	}

	for _, test := range tests {
		tf, err := Transform(test.Template, testStructuredEnvelope)
		if test.TransformError != "" {
			assert.EqualError(t, err, test.TransformError)
			continue
		} else {
			assert.NoError(t, err)
		}
		assert.Equal(t, test.Transformed, tf)

		enc, err := EncodeElements(tf)
		if test.EncodeError != "" {
			assert.EqualError(t, err, test.EncodeError)
			continue
		} else {
			assert.NoError(t, err)
		}

		assert.Equal(t, test.Encoded, enc)
	}
}
