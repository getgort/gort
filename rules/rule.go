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
	"sort"
)

type Rule struct {
	Command     string
	Conditions  []Expression
	Permissions []string
}

// Allowed returns true iff the user has all required permissions (or the rule
// is an "allow" rule).
func (r Rule) Allowed(permissions []string) bool {
	for _, required := range r.Permissions {
		i := sort.SearchStrings(permissions, required)
		if i >= len(permissions) || permissions[i] != required {
			return false
		}
	}

	return true
}

// Matches returns true iff the Rule's stated conditions evaluate to true.
func (r Rule) Matches(env EvaluationEnvironment) bool {
	// No conditions matches everything
	if len(r.Conditions) == 0 {
		return true
	}

	result := r.Conditions[0].Evaluate(env)

	for i := 1; i < len(r.Conditions); i++ {
		c := r.Conditions[i]

		if c.Condition == And {
			result = (result && c.Evaluate(env))
			continue
		}

		if c.Condition == Or {
			result = (result || c.Evaluate(env))
			continue
		}
	}

	return result
}
