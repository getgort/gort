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

type Image struct {
	Tag
	Height    int    `json:",omitempty"`
	Width     int    `json:",omitempty"`
	URL       string `json:",omitempty"`
	Thumbnail bool   `json:",omitempty"`
}

func (o *Image) String() string {
	return encodeTag(*o)
}

func (o *Image) Alt() string {
	return o.URL
}

func (f *Functions) ImageFunction(url string) *Image {
	return &Image{URL: url}
}

func (f *Functions) ImageThumbnailunction(b bool, i *Image) *Image {
	i.Thumbnail = b
	return i
}
