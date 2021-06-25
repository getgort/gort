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

import "github.com/getgort/gort/types"

type LogicalOperator int

const (
	Undefined LogicalOperator = iota
	And
	Or
)

// Expression describes a single.
// Condition should be Undefined for the first element, but defined for each subsequent element.
type Expression struct {
	A, B      types.Value
	Operator  Operator
	Condition LogicalOperator
}

type EvaluationEnvironment map[string]interface{}

func (e Expression) Evaluate(env EvaluationEnvironment) bool {
	e.A = define(e.A, env)
	e.B = define(e.B, env)

	return e.Operator(e.A, e.B)
}

func define(v types.Value, env EvaluationEnvironment) types.Value {
	switch o := v.(type) {
	case types.ListValue:
		i, exists := env[o.Name]
		if !exists {
			return v
		}

		c, ok := i.([]types.Value)
		if !ok {
			return v
		}

		if o.Index < 0 || o.Index >= len(c) {
			return v
		}

		o.V = c
		return o

	case types.MapValue:
		i, exists := env[o.Name]
		if !exists {
			return v
		}

		c, ok := i.(map[string]types.Value)
		if !ok {
			return v
		}

		o.V = c
		return o
	}

	return v
}
