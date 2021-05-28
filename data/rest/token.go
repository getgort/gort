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

package rest

import "time"

// Token contains all of the metadata for an access token.
type Token struct {
	Duration   time.Duration `json:"-"`
	Token      string        `json:",omitempty"`
	User       string        `json:",omitempty"`
	ValidFrom  time.Time     `json:",omitempty"`
	ValidUntil time.Time     `json:",omitempty"`
}

// IsExpired returns true if the token has expired.
func (t Token) IsExpired() bool {
	return time.Now().After(t.ValidUntil)
}
