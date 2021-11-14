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
	"fmt"
	"reflect"
	"text/template"

	"github.com/Masterminds/sprig"
)

type Functions struct{}

func FunctionMap() template.FuncMap {
	functions := &Functions{}

	fm := map[string]interface{}{
		// Header
		"header": functions.HeaderFunction,
		"color":  functions.HeaderColorFunction,

		// Image
		"image":     functions.ImageFunction,
		"thumbnail": functions.ImageThumbnailunction,

		// Section
		"section":    functions.SectionFunction,
		"endsection": functions.SectionEndFunction,

		// Text
		"text":      functions.TextFunction,
		"inline":    functions.TextInlineFunction,
		"markdown":  functions.TextMarkdownFunction,
		"monospace": functions.TextMonospaceFunction,
		"emoji":     functions.TextEmojiFunction,
		"endtext":   functions.TextEndFunction,

		// Multiform functions
		"title": functions.MultipleTitleFunction,

		// Simple blocks
		"divider": functions.DividerFunction,

		// Alternative text
		"alt": functions.AltFunction,

		// Unimplemented - for testing fallback behavior
		"unimplemented": functions.UnimplementedFunction,
	}

	sprigFuncs := sprig.FuncMap()

	for k, f := range fm {
		sprigFuncs[k] = f
	}

	return template.FuncMap(sprigFuncs)
}

type Tag struct {
	FirstIndex int `json:"-"`
	LastIndex  int `json:"-"`
}

func (t *Tag) First() int {
	return t.FirstIndex
}

func (t *Tag) Last() int {
	return t.LastIndex
}

func encodeTag(i interface{}) string {
	b, _ := json.Marshal(i)
	return fmt.Sprintf("<<%s|%s>>", reflect.TypeOf(i).Name(), string(b))
}
