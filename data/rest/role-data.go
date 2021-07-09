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

import "fmt"

type Role struct {
	Name        string
	Permissions RolePermissionList
	Groups      []Group
}

type RolePermission struct {
	BundleName string
	Permission string
}

func (p RolePermission) String() string {
	return fmt.Sprintf("%s:%s", p.BundleName, p.Permission)
}

type RolePermissionList []RolePermission

func (l RolePermissionList) Strings() []string {
	s := make([]string, len(l))

	for i, p := range l {
		s[i] = p.String()
	}

	return s
}
