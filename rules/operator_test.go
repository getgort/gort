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
	"regexp"
	"testing"

	"github.com/getgort/gort/types"
	"github.com/stretchr/testify/assert"
)

func TestOperatorEquals(t *testing.T) {
	evaluate := Equals

	result := evaluate(types.IntValue{V: 42}, types.IntValue{V: 42})
	assert.True(t, result)

	result = evaluate(types.IntValue{V: 42}, types.IntValue{V: 21})
	assert.False(t, result)
}

func TestOperatorRegexpEquals(t *testing.T) {
	re := regexp.MustCompile(`^(.*)\s+([!>=]{1,2}|in)\s+(.*)$`)
	test := "option['delete'] in true"

	subs := re.FindStringSubmatch(test)

	for _, s := range subs {
		t.Log(s)
	}
}

func TestOperatorNotEquals(t *testing.T) {
	evaluate := NotEquals

	result := evaluate(types.IntValue{V: 42}, types.IntValue{V: 42})
	assert.False(t, result)

	result = evaluate(types.IntValue{V: 42}, types.IntValue{V: 21})
	assert.True(t, result)
}

func TestOperatorLessThan(t *testing.T) {
	evaluate := LessThan

	result := evaluate(types.IntValue{V: 21}, types.IntValue{V: 42})
	assert.True(t, result)

	result = evaluate(types.IntValue{V: 42}, types.IntValue{V: 42})
	assert.False(t, result)

	result = evaluate(types.IntValue{V: 42}, types.IntValue{V: 21})
	assert.False(t, result)
}

func TestOperatorLessThanOrEqualTo(t *testing.T) {
	evaluate := LessThanOrEqualTo

	result := evaluate(types.IntValue{V: 21}, types.IntValue{V: 42})
	assert.True(t, result)

	result = evaluate(types.IntValue{V: 42}, types.IntValue{V: 42})
	assert.True(t, result)

	result = evaluate(types.IntValue{V: 42}, types.IntValue{V: 21})
	assert.False(t, result)
}

func TestOperatorGreaterThan(t *testing.T) {
	evaluate := GreaterThan

	result := evaluate(types.IntValue{V: 21}, types.IntValue{V: 42})
	assert.False(t, result)

	result = evaluate(types.IntValue{V: 42}, types.IntValue{V: 42})
	assert.False(t, result)

	result = evaluate(types.IntValue{V: 42}, types.IntValue{V: 21})
	assert.True(t, result)
}

func TestOperatorGreaterThanOrEqualTo(t *testing.T) {
	evaluate := GreaterThanOrEqualTo

	result := evaluate(types.IntValue{V: 21}, types.IntValue{V: 42})
	assert.False(t, result)

	result = evaluate(types.IntValue{V: 42}, types.IntValue{V: 42})
	assert.True(t, result)

	result = evaluate(types.IntValue{V: 42}, types.IntValue{V: 21})
	assert.True(t, result)
}
