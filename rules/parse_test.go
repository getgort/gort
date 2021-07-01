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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/getgort/gort/types"
)

func TestParse(t *testing.T) {
	inputs := map[string]Rule{
		`foo:bar allow`: {Command: "foo:bar", Conditions: []Expression{}, Permissions: []Permission{}},
		`foo:bar with option['delete'] == /^.*$/ must have foo:destroy`: {
			Command:     "foo:bar",
			Conditions:  []Expression{{A: types.MapElementValue{V: types.MapValue{Name: "option"}, Key: "delete"}, B: types.RegexValue{V: `^.*$`}, Operator: Equals, Condition: Undefined}},
			Permissions: []Permission{{Name: "foo:destroy"}}},
		`foo:bar with option['delete'] == true must have foo:destroy`: {
			Command:     "foo:bar",
			Conditions:  []Expression{{A: types.MapElementValue{V: types.MapValue{Name: "option"}, Key: "delete"}, B: types.BoolValue{V: true}, Operator: Equals, Condition: Undefined}},
			Permissions: []Permission{{Name: "foo:destroy"}}},
		`foo:bar with option['delete'] == true and arg[0] == false must have foo:destroy`: {
			Command: "foo:bar",
			Conditions: []Expression{
				{A: types.MapElementValue{V: types.MapValue{Name: "option"}, Key: "delete"}, B: types.BoolValue{V: true}, Operator: Equals, Condition: Undefined},
				{A: types.ListElementValue{V: types.ListValue{Name: "arg"}, Index: 0}, B: types.BoolValue{V: false}, Operator: Equals, Condition: And},
			},
			Permissions: []Permission{{Name: "foo:destroy"}}},
		`foo:bar with any arg in ['wubba'] must have foo:read`: {
			Command: "foo:bar",
			Conditions: []Expression{{
				A:        types.UnknownValue{V: "arg"},
				B:        types.ListValue{V: []types.Value{types.StringValue{V: "wubba", Quote: '\''}}},
				Operator: In,
				Modifier: CollAny,
			}},
			Permissions: []Permission{{Name: "foo:read"}},
		},
		`foo:bar with any arg in ['wubba'] must have foo:read and foo:write`: {
			Command: "foo:bar",
			Conditions: []Expression{{
				A:        types.UnknownValue{V: "arg"},
				B:        types.ListValue{V: []types.Value{types.StringValue{V: "wubba", Quote: '\''}}},
				Operator: In,
				Modifier: CollAny,
			}},
			Permissions: []Permission{{Name: "foo:read"}, {"foo:write", And}},
		},
		`foo:bar with any arg in ['wubba'] must have foo:read and foo:write or foo:destroy`: {
			Command: "foo:bar",
			Conditions: []Expression{{
				A:        types.UnknownValue{V: "arg"},
				B:        types.ListValue{V: []types.Value{types.StringValue{V: "wubba", Quote: '\''}}},
				Operator: In,
				Modifier: CollAny,
			}},
			Permissions: []Permission{{Name: "foo:read"}, {"foo:write", And}, {"foo:destroy", Or}},
		},
	}

	for in, expected := range inputs {
		rt, err := Tokenize(in)
		if !assert.NoError(t, err, in) {
			continue
		}

		rule, err := Parse(rt)
		if !assert.NoError(t, err, in) {
			continue
		}

		assert.Equal(t, expected.Command, rule.Command, in)

		for i, e := range expected.Conditions {
			assert.EqualValues(t, e.A, rule.Conditions[i].A)
			assert.EqualValues(t, e.B, rule.Conditions[i].B)
			assert.Equal(t, fmt.Sprintf("%v", e.Operator), fmt.Sprintf("%v", rule.Conditions[i].Operator))
		}

		assert.Equal(t, expected.Permissions, rule.Permissions, in)
	}
}

func TestParseExpression(t *testing.T) {
	type Expected struct {
		a, b string
		o    Operator
		m    CollectionOperationModifier
	}

	inputs := map[string][]Expected{
		`foo:bar with option[foo] in ["foo", "bar"] allow`:                                                  {{a: `option[foo]`, b: `["foo", "bar"]`, o: In}},
		`foo:bar with option['delete'] == true must have foo:destroy`:                                       {{a: `option['delete']`, b: `true`, o: Equals}},
		`foo:set with option['set'] == /.*/ must have foo:baz-set`:                                          {{a: `option['set']`, b: `/.*/`, o: Equals}},
		`foo:qux with arg[0] == 'status' must have foo:view`:                                                {{a: `arg[0]`, b: `'status'`, o: Equals}},
		`foo:barqux with option['delete'] == true and arg[0] > 5 must have foo:destroy`:                     {{a: `option['delete']`, b: `true`, o: Equals}, {a: `arg[0]`, b: `5`, o: GreaterThan}},
		`foo:bar with any arg in ['wubba'] must have foo:read`:                                              {{a: `arg`, b: `['wubba']`, o: In, m: CollAny}},
		`foo:bar with any arg in ['wubba', /^f.*/, 10] must have foo:read`:                                  {{a: `arg`, b: `['wubba', /^f.*/, 10]`, o: In, m: CollAny}},
		`foo:bar with all arg in [10, 'baz', 'wubba'] must have foo:read`:                                   {{a: `arg`, b: `[10, 'baz', 'wubba']`, o: In, m: CollAll}},
		`foo:bar with arg[0] in ['baz', false, 100] must have foo:read`:                                     {{a: `arg[0]`, b: `['baz', false, 100]`, o: In}},
		`foo:bar with any option != /^prod.*/ must have foo:read`:                                           {{a: `option`, b: `/^prod.*/`, o: NotEquals, m: CollAny}},
		`foo:bar with all option == 10 must have foo:read`:                                                  {{a: `option`, b: `10`, o: Equals, m: CollAll}},
		`foo:bar with all option < 10 must have foo:read`:                                                   {{a: `option`, b: `10`, o: LessThan, m: CollAll}},
		`foo:bar with all option <= 10 must have foo:read`:                                                  {{a: `option`, b: `10`, o: LessThanOrEqualTo, m: CollAll}},
		`foo:bar with all option > 10 must have foo:read`:                                                   {{a: `option`, b: `10`, o: GreaterThan, m: CollAll}},
		`foo:bar with all option >= 10 must have foo:read`:                                                  {{a: `option`, b: `10`, o: GreaterThanOrEqualTo, m: CollAll}},
		`foo:bar with all option != 10 must have foo:read`:                                                  {{a: `option`, b: `10`, o: NotEquals, m: CollAll}},
		`foo:deploy with option["environment"] == 'prod' must have all in [site:it, site:prod, foo:deploy]`: {{a: `option["environment"]`, b: `'prod'`, o: Equals}},
		`foo:patch must have all in [foo:patch, site:it]
			or all in [site:qa, site:test, foo:patch]
			or all in [site:eng, site:stage, foo:patch]`: {{}},
		`foo:bar
		    with option['delete'] == true
			   must have foo:destroy`: {{a: `option['delete']`, b: `true`, o: Equals}},
	}

	for in, expected := range inputs {
		rt, err := Tokenize(in)
		if !assert.NoError(t, err, in) {
			continue
		}

		for i := 0; i < len(rt.Conditions); i += 2 {
			expr := rt.Conditions[i]
			a, b, o, m, err := ParseExpression(expr)

			// Workaround for function comparison
			os := fmt.Sprintf("%v", o)
			eos := fmt.Sprintf("%v", expected[i/2].o)

			if !assert.NoError(t, err, in) ||
				!assert.Equal(t, expected[i/2].a, a, in) ||
				!assert.Equal(t, expected[i/2].b, b, in) ||
				!assert.Equal(t, expected[i/2].m, m, in) ||
				!assert.Equal(t, eos, os, in) {

				t.Logf("Erroneous expression: %q", expr)
			}
		}
	}
}
