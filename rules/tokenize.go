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
	"strings"
)

var reSplit = regexp.MustCompile(`\s+`)

// RuleTokens represents a tokenized Gort rule of the form "COMMAND [when
// CONDITION (and|or)]? [allow|must have PERMISSION (and|or)]".
type RuleTokens struct {
	Command     string
	Conditions  []string
	Permissions []string
}

// Tokenize accepts a raw Gort rule of the form "COMMAND [when CONDITION
// (and|or)]? [allow|must have PERMISSION (and|or)]", and returns a RuleTokens
// value. A parsing error will produce a non-nil error. The RuleTokens' Command
// value should always be non-empty; Conditions and Permissions can both be
// empty (but non-nil). Empty Conditions always match the command. Empty
// Permissions indicating the use of the "allow" keyword and always pass.
func Tokenize(s string) (RuleTokens, error) {
	const (
		StateCommand int = iota
		StateConditions
		StatePermissionsMust
		StatePermissionsHave
		StateEnd
	)

	rt := RuleTokens{Conditions: []string{}, Permissions: []string{}}

	if s == "" {
		return rt, fmt.Errorf("empty rule")
	}

	// This is a primitive state machine. Regex just wasn't powerful enough.
	// Sorry.

	currentState := StateCommand
	b := &strings.Builder{}

	for _, s := range reSplit.Split(s, -1) {
		switch currentState {
		case StateCommand:
			switch s {
			case "with":
				if b.Len() == 0 && len(rt.Conditions) == 0 {
					return rt, fmt.Errorf("expected command; got '%s'", s)
				}

				rt.Command = b.String()
				b.Reset()
				currentState = StateConditions
			case "must":
				if b.Len() == 0 && len(rt.Conditions) == 0 {
					return rt, fmt.Errorf("expected command; got '%s'", s)
				}

				rt.Command = b.String()
				b.Reset()
				currentState = StatePermissionsMust
			case "allow":
				if b.Len() == 0 && len(rt.Conditions) == 0 {
					return rt, fmt.Errorf("expected command; got '%s'", s)
				}

				rt.Command = b.String()
				b.Reset()
				currentState = StateEnd
			case "and":
				fallthrough
			case "or":
				fallthrough
			case "have":
				return rt, fmt.Errorf("expected command; got '%s'", s)
			default:
				bappend(b, s)
			}

		case StateConditions:
			switch s {
			case "and":
				fallthrough
			case "or":
				rt.Conditions = append(rt.Conditions, b.String(), s)
				b.Reset()
			case "must":
				rt.Conditions = append(rt.Conditions, b.String())
				b.Reset()
				currentState = StatePermissionsMust
			case "allow":
				if b.Len() == 0 && len(rt.Conditions) == 0 {
					return rt, fmt.Errorf("'with' missing conditions")
				}

				rt.Conditions = append(rt.Conditions, b.String())
				b.Reset()
				currentState = StateEnd
			case "with":
				fallthrough
			case "have":
				return rt, fmt.Errorf("unexpected keyword '%s'", s)
			default:
				bappend(b, s)
			}

		case StatePermissionsMust:
			switch s {
			case "have":
				currentState = StatePermissionsHave
			default:
				return rt, fmt.Errorf("expected have; got %s", s)
			}

		case StatePermissionsHave:
			switch s {
			case "and":
				fallthrough
			case "or":
				if b.Len() == 0 && len(rt.Permissions) == 0 {
					return rt, fmt.Errorf("expected permission; got '%s'", s)
				}

				rt.Permissions = append(rt.Permissions, b.String(), s)
				b.Reset()
			case "allow":
				fallthrough
			case "with":
				fallthrough
			case "must":
				fallthrough
			case "have":
				return rt, fmt.Errorf("unexpected keyword '%s'", s)
			default:
				bappend(b, s)
			}

		case StateEnd:
			return rt, fmt.Errorf("unexpected text after allow")
		}
	}

	switch currentState {
	case StateCommand:
		return rt, fmt.Errorf("missing conditions and permissions clauses")
	case StateConditions:
		return rt, fmt.Errorf("missing permissions clause")
	case StatePermissionsMust:
		return rt, fmt.Errorf("incomplete permissions clause")
	case StatePermissionsHave:
		if b.Len() == 0 && len(rt.Permissions) == 0 {
			return rt, fmt.Errorf("'must have' missing permissions")
		}

		rt.Permissions = append(rt.Permissions, b.String())
	}

	return rt, nil
}

func bappend(b *strings.Builder, s string) {
	if b.Len() != 0 {
		b.WriteRune(' ')
	}
	b.WriteString(s)
}
