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
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTokenize tests that various valid rule constructions resolve to
// produce the expected data structures.
func TestTokenize(t *testing.T) {
	inputs := map[string]RuleTokens{
		`foo:bar allow`: {`foo:bar`, []string{}, []string{}},
		`foo:bar with option[foo] in ["foo", "bar"] allow`:                                                  {`foo:bar`, []string{`option[foo] in ["foo", "bar"]`}, []string{}},
		`foo:bar with option['delete'] == true must have foo:destroy`:                                       {`foo:bar`, []string{`option['delete'] == true`}, []string{`foo:destroy`}},
		`foo:set with option['set'] == /.*/ must have foo:baz-set`:                                          {`foo:set`, []string{`option['set'] == /.*/`}, []string{`foo:baz-set`}},
		`foo:qux with arg[0] == 'status' must have foo:view`:                                                {`foo:qux`, []string{`arg[0] == 'status'`}, []string{`foo:view`}},
		`foo:barqux with option['delete'] == true and arg[0] > 5 must have foo:destroy`:                     {`foo:barqux`, []string{`option['delete'] == true`, `and`, `arg[0] > 5`}, []string{`foo:destroy`}},
		`foo:bar with any arg in ['wubba'] must have foo:read`:                                              {`foo:bar`, []string{`any arg in ['wubba']`}, []string{`foo:read`}},
		`foo:bar with any arg in ['wubba', /^f.*/, 10] must have foo:read`:                                  {`foo:bar`, []string{`any arg in ['wubba', /^f.*/, 10]`}, []string{`foo:read`}},
		`foo:bar with all arg in [10, 'baz', 'wubba'] must have foo:read`:                                   {`foo:bar`, []string{`all arg in [10, 'baz', 'wubba']`}, []string{`foo:read`}},
		`foo:bar with arg[0] in ['baz', false, 100] must have foo:read`:                                     {`foo:bar`, []string{`arg[0] in ['baz', false, 100]`}, []string{`foo:read`}},
		`foo:bar with any option == /^prod.*/ must have foo:read`:                                           {`foo:bar`, []string{`any option == /^prod.*/`}, []string{`foo:read`}},
		`foo:bar with all option < 10 must have foo:read`:                                                   {`foo:bar`, []string{`all option < 10`}, []string{`foo:read`}},
		`foo:bar with all option in ['staging', 'list'] must have foo:read`:                                 {`foo:bar`, []string{`all option in ['staging', 'list']`}, []string{`foo:read`}},
		`foo:deploy with option["environment"] == 'prod' must have all in [site:it, site:prod, foo:deploy]`: {`foo:deploy`, []string{`option["environment"] == 'prod'`}, []string{`all in [site:it, site:prod, foo:deploy]`}},
		`foo:deploy with option["environment"] == 'qa' must have site:test and foo:deploy`:                  {`foo:deploy`, []string{`option["environment"] == 'qa'`}, []string{`site:test`, `and`, `foo:deploy`}},
		`foo:deploy with option["environment"] == 'stage' must have site:stage and foo:deploy`:              {`foo:deploy`, []string{`option["environment"] == 'stage'`}, []string{`site:stage`, `and`, `foo:deploy`}},
		`foo:patch must have all in [foo:patch, site:it]
			or all in [site:qa, site:test, foo:patch]
			or all in [site:eng, site:stage, foo:patch]`: {`foo:patch`, []string{}, []string{`all in [foo:patch, site:it]`, `or`, `all in [site:qa, site:test, foo:patch]`, `or`, `all in [site:eng, site:stage, foo:patch]`}},
		`foo:bar
		    with option['delete'] == true
			   must have foo:destroy`: {`foo:bar`, []string{`option['delete'] == true`}, []string{`foo:destroy`}},
	}

	for str, expected := range inputs {
		actual, err := Tokenize(str)
		if !assert.NoError(t, err, str) {
			continue
		}

		assert.Equal(t, expected, actual, str)
	}
}

// TestTokenizeErrors tests that various invalid rule constructions generate
// an error.
func TestTokenizeErrors(t *testing.T) {
	inputs := []string{
		``,
		`foo:bar`,
		`foobar allow`,
		`foo:bar allow foo`,
		`foo:bar with allow`,
		`foo:bar with option['delete'] == true`,
		`foo:bar with with option['delete'] == true allow`,
		`foo:bar with option['delete'] == true allow foo`,
		`foo:bar with option['delete'] == true must`,
		`foo:bar with option['delete'] == true must allow`,
		`foo:bar with option['delete'] == true must have`,
		`foo:bar with option['delete'] == true must have and`,
		`foo:bar with option['delete'] == true must have allow`,
		`with option['delete'] == true allow`,
		`must have foo:read`,
		`allow`,
	}

	for _, str := range inputs {
		_, err := Tokenize(str)
		assert.Error(t, err, str)
	}
}
