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

type Text struct {
	Tag
	Emoji     bool   `json:",omitempty"`
	Inline    bool   `json:",omitempty"`
	Markdown  bool   `json:",omitempty"`
	Monospace bool   `json:",omitempty"`
	Title     string `json:",omitempty"`
	Text      string `json:",omitempty"`
}

func (o *Text) String() string {
	return encodeTag(*o)
}

func (o *Text) Alt() string {
	return o.Text
}

func (f *Functions) TextFunction() *Text {
	return &Text{
		Emoji:    true,
		Markdown: true,
	}
}

func (f *Functions) TextEmojiFunction(b bool, t *Text) *Text {
	t.Emoji = b
	return t
}

func (f *Functions) TextInlineFunction(b bool, t *Text) *Text {
	t.Inline = b
	return t
}

func (f *Functions) TextMarkdownFunction(b bool, t *Text) *Text {
	t.Markdown = b
	return t
}

func (f *Functions) TextMonospaceFunction(b bool, t *Text) *Text {
	t.Monospace = b
	return t
}

type TextEnd struct {
	Tag
}

func (o *TextEnd) String() string {
	return encodeTag(*o)
}

func (o *TextEnd) Alt() string {
	return ""
}

func (f *Functions) TextEndFunction() *TextEnd {
	o := &TextEnd{}
	return o
}
