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
	"strings"
)

func Parse(text string) ([]interface{}, error) {
	var lastText *Text
	var lastSection *Section
	var elements []interface{}

	for tag, jsn, first, last := nextTag(text, 0); first != -1; tag, jsn, first, last = nextTag(text, last) {
		switch tag {
		case "": // no-op

		case "DIVIDER":
			elements = append(elements, &Divider{})

		case "IMAGE":
			o := &Image{}
			json.Unmarshal([]byte(jsn), o)
			elements = append(elements, o)

		case "SECTION":
			o := &Section{}
			json.Unmarshal([]byte(jsn), o)
			o.Index = last
			lastSection = o
		case "SECTIONEND":
			if lastSection == nil {
				return nil, fmt.Errorf("unmatched {{sectionend}} tag at index %d", first)
			}
			elements = append(elements, lastSection)
			lastSection = nil

		case "TEXT":
			o := &Text{}
			json.Unmarshal([]byte(jsn), o)
			o.Index = last
			lastText = o
		case "TEXTEND":
			if lastText == nil {
				return nil, fmt.Errorf("unmatched {{textend}} tag at index %d", first)
			}
			elements = append(elements, lastText)
			lastText = nil
		}
	}

	if lastSection == nil {
		return nil, fmt.Errorf("unmatched {{section}} tag at index %d", lastSection.Index)
	}
	if lastText == nil {
		return nil, fmt.Errorf("unmatched {{text}} tag at index %d", lastText.Index)
	}

	return elements, nil
}

// nextTag returns the tag, JSON text, and start and end tags for the next tag
// after start. If no tag is found, the index values are both returned as -1.
func nextTag(text string, start int) (tag string, json string, first int, last int) {
	if start >= len(text) {
		return "", "", -1, -1
	}

	text = text[start:]

	i, j := strings.Index(text, "<<"), strings.Index(text, ">>")
	if i < 0 || j < 0 {
		return "", "", -1, -1
	}

	first = i + start
	last = j + start + 2
	text = text[i+2 : j]

	if m := strings.Index(text, "|{"); m == -1 {
		tag = text
	} else {
		tag = text[:m]
		json = text[m+1:]
	}

	return tag, json, first, last
}
