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

package rules

import (
	"github.com/getgort/gort/command"
	"github.com/getgort/gort/types"
)

type Rule struct {
	Command     string
	Conditions  []Expression
	Permissions []string
}

// Matches returns true iff the Rule's stated conditions evaluate to true.
func (r Rule) Matches(options map[string]command.CommandOption, args []types.Value) bool {
	// TODO: Create an "arg" and "option" Value type.
	return false
}

// Allowed returns true iff the user has all required permissions (or the rule
// is an "allow" rule).
func (r Rule) Allowed(permissions map[string]interface{}) bool {
	// TODO: Create an "arg" and "option" Value type.
	return false
}
