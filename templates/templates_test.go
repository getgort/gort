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
	"bytes"
	"encoding/json"
	"testing"
	"text/template"

	"github.com/getgort/gort/data"
	"github.com/stretchr/testify/require"
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
Hello, {{ .Response.Payload.User }}!

*{{ .Response.Payload.Requestor }}* wants to know where you'd like to take the {{ .Response.Payload.Company }} investors to dinner tonight.

*Please select a restaurant:*
{{ endtext }}

{{ divider }}

{{ range $index, $result := .Response.Payload.Results }}
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

func TestAll(t *testing.T) {
	json.Unmarshal([]byte(payloadJSON), &testStructuredEnvelope.Response.Payload)

	// TODO(mtitmus) Cache this result somewhere?
	tpl := template.New("envelope")

	tpl, err := tpl.Funcs(FunctionMap()).Parse(testTemplate)
	require.NoError(t, err)

	b := new(bytes.Buffer)
	err = tpl.Execute(b, testStructuredEnvelope)
	require.NoError(t, err)

	Parse(b.String())
}
