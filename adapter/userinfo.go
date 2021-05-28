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

// UserInfo contains the basic information for a single user in any chat provider.
type UserInfo struct {
	ID                    string
	Name                  string
	DisplayName           string
	DisplayNameNormalized string
	Email                 string
	FirstName             string
	LastName              string
	RealName              string
	RealNameNormalized    string
}
