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
	"errors"
	"fmt"

	"github.com/getgort/gort/command"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/dataaccess"
	gerrs "github.com/getgort/gort/errors"
	"github.com/getgort/gort/rules"
)

const (
	// TODO This is just here to facilitate testing. Remove it when it's no longer useful.
	commandsRequireAtLeastOneRule = true
)

var (
	ErrRuleLoadError  = fmt.Errorf("rule load failure")
	ErrNoRulesDefined = fmt.Errorf("command has no rules")

	// ErrNotAllowed is thrown when checking user permissions for a command if
	// the user does not have the appropriate permissions to use the command.
	ErrNotAllowed = errors.New("user not allowed to use command")
)

// EvaluateRules returns true if the provided permissions meet the requirements
// defined by the given rules and EvaluationEnvironment. It returns an error if
// there isn't at least one rule in the Rule slice.
func EvaluateRules(perms []string, r []rules.Rule, env rules.EvaluationEnvironment) (bool, error) {
	if commandsRequireAtLeastOneRule && len(r) == 0 {
		return false, ErrNoRulesDefined
	}

	allowed := false

	// Loop over the rules and evaluate them one-by-one.
	for _, r := range r {
		// If the rule's conditions don't evaluate to true, ignore it.
		if !r.Matches(env) {
			continue
		}

		if allowed = r.Allowed(perms); !allowed {
			return false, nil
		}
	}

	return allowed, nil
}

// EvaluateCommandEntry is equivalent to EvaluateRules, except that it accepts
// a data.Command entry from which it builds its complete rules set, upon which
// it calls EvaluateRules. It returns true if the provided permissions meet the
// requirements derived by the CommandEntry and EvaluationEnvironment. It
// returns an error if there isn't at least one rule in the Rule slice.
func EvaluateCommandEntry(perms []string, ce data.CommandEntry, env rules.EvaluationEnvironment) (bool, error) {
	// Retrieve the command's rules as a []rules.Rule so that we can
	// evaluate against them.
	r, err := ParseCommandEntry(ce)
	if err != nil {
		return false, gerrs.Wrap(ErrRuleLoadError, err)
	}

	return EvaluateRules(perms, r, env)
}

// ParseCommandEntry is a helper function that accepts a fully-constructed
// data.CommandEntry, tokenizes and parses all of the command's rule strings,
// and returns a []Rules value.
func ParseCommandEntry(ce data.CommandEntry) ([]rules.Rule, error) {
	rr := []rules.Rule{}

	for i, r := range ce.Command.Rules {
		s := fmt.Sprintf("%s:%s %s", ce.Bundle.Name, ce.Command.Name, r)

		rule, err := rules.TokenizeAndParse(s)
		if err != nil {
			return rr, fmt.Errorf("cannot parse rule %s:%s rule %d (%s): %w", ce.Bundle.Name, ce.Command.Name, i+1, r, err)
		}

		rr = append(rr, rule)
	}

	return rr, nil
}

// CheckPermissions errors if the given user does not have permission to run the given command.
func CheckPermissions(ctx context.Context, userName string, cmdInput command.Command, cmdEntry data.CommandEntry) error {
	da, err := dataaccess.Get()
	if err != nil {
		return err
	}

	perms, err := da.UserPermissionList(ctx, userName)
	if err != nil {
		return err
	}

	allowed, err := EvaluateCommandEntry(
		perms.Strings(),
		cmdEntry,
		rules.EvaluationEnvironment{
			"option": cmdInput.OptionsValues(),
			"arg":    cmdInput.Parameters,
		},
	)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrNotAllowed
	}
	return nil
}
