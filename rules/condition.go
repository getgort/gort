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

type LogicalOperator int

const (
	Undefined LogicalOperator = iota
	And
	Or
)

type CollectionOperationModifier int

const (
	CollOne CollectionOperationModifier = iota
	CollAny
	CollAll
)

// Expression describes a single.
// Condition should be Undefined for the first element, but defined for each subsequent element.
type Expression struct {
	A, B      types.Value
	Operator  Operator
	Modifier  CollectionOperationModifier
	Condition LogicalOperator
}

type EvaluationEnvironment map[string]interface{}

func (e Expression) Evaluate(env EvaluationEnvironment) bool {
	e.A = define(e.A, env)
	e.B = define(e.B, env)
	coll, isColl := e.A.(types.CollectionValue)

	if isColl && e.Modifier == CollAny {
		for _, o := range coll.Elements() {
			if e.Operator(o, e.B) {
				return true
			}
		}
		return false
	}

	if isColl && e.Modifier == CollAll {
		for _, o := range coll.Elements() {
			if !e.Operator(o, e.B) {
				return false
			}
		}
		return true
	}

	return e.Operator(e.A, e.B)
}

func define(v types.Value, env EvaluationEnvironment) types.Value {
	switch o := v.(type) {
	case types.UnknownValue:
		i, exists := env[o.V]
		if !exists {
			return v
		}

		if c, ok := i.([]types.Value); ok {
			return types.ListValue{Name: o.V, V: c}
		}

		if c, ok := i.(command.CommandParameters); ok {
			return types.ListValue{Name: o.V, V: c}
		}

		if m, ok := i.(map[string]types.Value); ok {
			return types.MapValue{Name: o.V, V: m}
		}

		return o

	case types.ListElementValue:
		i, exists := env[o.V.Name]
		if !exists {
			return v
		}

		if c, ok := i.([]types.Value); ok {
			o.V.V = c
		}

		if c, ok := i.(command.CommandParameters); ok {
			o.V.V = c
		}

		return o

	case types.MapElementValue:
		i, exists := env[o.V.Name]
		if !exists {
			return v
		}

		if c, ok := i.(map[string]types.Value); ok {
			o.V.V = c
		}

		return o
	}

	return v
}

type Permission struct {
	Name      string
	Condition LogicalOperator
}
