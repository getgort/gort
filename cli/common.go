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

package cli

import (
	"github.com/getgort/gort/data/rest"
)

var (
	// FlagGortProfile is a persistent flag specifying the profile name to be used.
	FlagGortProfile string
	// FlagConfigBaseDir is a persistent flag specifying the base directory for storing config
	FlagConfigBaseDir string
)

func groupNames(groups []rest.Group) []string {
	names := make([]string, 0)

	for _, g := range groups {
		names = append(names, g.Name)
	}

	return names
}

func userNames(users []rest.User) []string {
	names := make([]string, 0)

	for _, u := range users {
		names = append(names, u.Username)
	}

	return names
}

func roleNames(roles []rest.Role) []string {
	names := make([]string, 0)

	for _, r := range roles {
		names = append(names, r.Name)
	}

	return names
}
