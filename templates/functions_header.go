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
	"fmt"
	"strconv"
	"strings"
)

type Header struct {
	Tag

	// Color must be expressed in RGB hex as "#123456"
	Color string `json:",omitempty"`

	Title string `json:",omitempty"`
}

func (o *Header) String() string {
	return encodeTag(*o)
}

func (f *Functions) HeaderFunction() *Header {
	return &Header{}
}

func (f *Functions) HeaderColorFunction(s string, t *Header) (*Header, error) {
	s = strings.Replace(s, "#", "", 1)

	v, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		return nil, fmt.Errorf("colors should be expressed in RGB hex format: #123456")
	}

	t.Color = fmt.Sprintf("#%X", v)
	return t, nil
}

func (f *Functions) HeaderTitleFunction(s string, t *Header) *Header {
	t.Title = s
	return t
}

type HeaderEnd struct {
	Tag
}

func (o *HeaderEnd) String() string {
	return encodeTag(*o)
}

func (f *Functions) HeaderEndFunction() *HeaderEnd {
	o := &HeaderEnd{}
	return o
}
