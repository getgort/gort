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

// User is a data struct used to exchange data between a Gort client and a
// Gort controller REST service.
type User struct {
	Email    string `json:"email,omitempty"`
	FullName string `json:"fullname,omitempty"`
	Password string `json:"password,omitempty"`
	Username string `json:"username,omitempty"`

	// Mappings stores this user's identity mappings for each chat service.
	// The key is the adapter name as defined in the config; the value is the
	// associated ID in the service the adapter connects to.
	Mappings map[string]string `json:"mappings,omitempty"`
}
