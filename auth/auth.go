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

package auth

import (
	"context"
	"fmt"

	// "fmt"

	"github.com/getgort/gort/data"
	gerrs "github.com/getgort/gort/errors"
	"github.com/getgort/gort/rules"
)

const (
	CommandsRequireAtLeastOneRule = true
)

var (
	ErrRuleLoadError  = fmt.Errorf("rule load failure")
	ErrNoRulesDefined = fmt.Errorf("command has no rules")
)

func Evaluate(ctx context.Context, permissions []string, cmdEntry data.CommandEntry, env rules.EvaluationEnvironment) (bool, error) {
	// Retrieve the command's rules as a []rules.Rule so that we can
	// evaluate against them.
	ruleList, err := ParseCommandEntry(cmdEntry)
	if err != nil {
		return false, gerrs.Wrap(ErrRuleLoadError, err)
	}

	if CommandsRequireAtLeastOneRule && len(ruleList) == 0 {
		return false, ErrNoRulesDefined
	}

	allowed := false

	// Loop over the rules and evaluate them one-by-one.
	for _, r := range ruleList {
		// If the rule's conditions don't evaluate to true, ignore it.
		if !r.Matches(env) {
			continue
		}

		if allowed = r.Allowed(permissions); !allowed {
			return false, nil
		}
	}

	return allowed, nil
}

// ParseCommandEntry is a helper function that accepts a fully-constructed
// data.CommandEntry, tokenizes and parses all of the command's rule strings,
// and returns a []Rules value.
func ParseCommandEntry(ce data.CommandEntry) ([]rules.Rule, error) {
	rr := []rules.Rule{}

	for i, r := range ce.Command.Rules {
		tokens, err := rules.Tokenize(fmt.Sprintf("%s:%s %s", ce.Bundle.Name, ce.Command.Name, r))
		if err != nil {
			return rr, fmt.Errorf("cannot tokenize %s:%s rule %d (%s): %w", ce.Bundle.Name, ce.Command.Name, i+1, r, err)
		}

		rule, err := rules.Parse(tokens)
		if err != nil {
			return rr, fmt.Errorf("cannot parse rule %s:%s rule %d (%s): %w", ce.Bundle.Name, ce.Command.Name, i+1, r, err)
		}

		rr = append(rr, rule)
	}

	return rr, nil
}
