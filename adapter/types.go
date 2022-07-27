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

package adapter

type MessageRef struct {
	ID        string
	ChannelID string
	Timestamp string
	Adapter   string
}

type Emoji struct {
	shortname string
	unicode   rune
}

func (e Emoji) Shortname() string {
	return e.shortname
}

func (e Emoji) Unicode() rune {
	return e.unicode
}

func EmojiFrom(e string) Emoji {
	return Emoji{
		shortname: e,
		unicode:   rune(e[0]),
	}
}
