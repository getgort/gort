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

import "fmt"

type Section struct {
	Tag
	Text   *Text           `json:",omitempty"`
	Fields []OutputElement `json:",omitempty"`
}

func (o *Section) String() string {
	return encodeTag(*o)
}

func (o *Section) Alt() string {
	var out string
	if o.Text != nil {
		out = o.Text.Text
	}
	for _, element := range o.Fields {
		out = fmt.Sprintf("%v\n\n%v", out, element.Alt())
	}
	return out
}

func (f *Functions) SectionFunction() *Section {
	return &Section{}
}

type SectionEnd struct {
	Tag
}

func (o *SectionEnd) String() string {
	return encodeTag(*o)
}

func (o *SectionEnd) Alt() string {
	return ""
}

func (f *Functions) SectionEndFunction() *SectionEnd {
	o := &SectionEnd{}
	return o
}
