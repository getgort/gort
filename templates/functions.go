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
		// Section
		"section":    functions.SectionFunction,
		"endsection": functions.SectionEndFunction,

		// Text
		"text":      functions.TextFunction,
		"markup":    functions.TextMarkupFunction,
		"monospace": functions.TextMonospaceFunction,
		"emoji":     functions.TextEmojiFunction,
		"endtext":   functions.TextEndFunction,

		// Simple blocks
		"divider": functions.DividerFunction,
		"image":   functions.ImageFunction,
	}

	for k, f := range sprig.FuncMap() {
		fm[k] = f
	}

	return fm
}

type Tag struct {
	FirstIndex int `json:"-"`
	LastIndex  int `json:"-"`
}

func (t *Tag) SetFirst(i int) {
	t.FirstIndex = i
}

func (t *Tag) First() int {
	return t.FirstIndex
}

func (t *Tag) SetLast(i int) {
	t.LastIndex = i
}

func (t *Tag) Last() int {
	return t.LastIndex
}

func encodeTag(i interface{}) string {
	b, _ := json.Marshal(i)
	return fmt.Sprintf("<<%s|%s>>", reflect.TypeOf(i).Name(), string(b))
}
