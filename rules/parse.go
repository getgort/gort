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
	"fmt"
	"regexp"

	"github.com/getgort/gort/types"
)

func Parse(rt RuleTokens) (Rule, error) {
	r := Rule{
		Command:     rt.Command,
		Conditions:  []Expression{},
		Permissions: rt.Permissions,
	}

	lastCondition := Undefined

	for _, c := range rt.Conditions {
		if c == "and" {
			lastCondition = And
			continue
		}

		if c == "or" {
			lastCondition = Or
			continue
		}

		a, b, o, err := ParseExpression(c)
		if err != nil {
			return r, fmt.Errorf("can't parse condition: %w", err)
		}

		va, err := types.Infer(a, false, true)
		if err != nil {
			return r, fmt.Errorf("can't infer value: %w", err)
		}

		vb, err := types.Infer(b, false, true)
		if err != nil {
			return r, fmt.Errorf("can't infer value: %w", err)
		}

		r.Conditions = append(r.Conditions, Expression{A: va, B: vb, Operator: o, Condition: lastCondition})
	}

	return r, nil
}

var (
	reOperatorParts = regexp.MustCompile(`^(.*)\s+([!<>=]{1,2}|in)\s+(.*)$`)
)

func ParseExpression(expr string) (a, b string, o Operator, err error) {
	subs := reOperatorParts.FindStringSubmatch(expr)

	if len(subs) != 4 {
		err = fmt.Errorf("expression doesn't conform to form A OP B")
		return
	}

	op := subs[2]
	a, b = subs[1], subs[3]

	switch op {
	case "==":
		o = Equals
	case "!=":
		o = NotEquals
	case "<":
		o = LessThan
	case "<=":
		o = LessThanOrEqualTo
	case ">":
		o = GreaterThan
	case ">=":
		o = GreaterThanOrEqualTo
	case "in":
		o = In
	default:
		err = fmt.Errorf("unsupported operator: %s", op)
	}

	return
}
