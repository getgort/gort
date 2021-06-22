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

	"github.com/getgort/gort/types"
	"github.com/stretchr/testify/assert"
)

func TestOperatorEquals(t *testing.T) {
	evaluate := OperatorEquals

	result, err := evaluate(types.IntValue{Value: 42}, types.IntValue{Value: 42})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.True(t, result) {
		t.FailNow()
	}

	result, err = evaluate(types.IntValue{Value: 42}, types.IntValue{Value: 21})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.False(t, result) {
		t.FailNow()
	}
}

func TestOperatorNotEquals(t *testing.T) {
	evaluate := OperatorNotEquals

	result, err := evaluate(types.IntValue{Value: 42}, types.IntValue{Value: 42})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.False(t, result) {
		t.FailNow()
	}

	result, err = evaluate(types.IntValue{Value: 42}, types.IntValue{Value: 21})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.True(t, result) {
		t.FailNow()
	}
}

func TestOperatorLessThan(t *testing.T) {
	evaluate := OperatorLessThan

	result, err := evaluate(types.IntValue{Value: 21}, types.IntValue{Value: 42})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.True(t, result) {
		t.FailNow()
	}

	result, err = evaluate(types.IntValue{Value: 42}, types.IntValue{Value: 42})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.False(t, result) {
		t.FailNow()
	}

	result, err = evaluate(types.IntValue{Value: 42}, types.IntValue{Value: 21})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.False(t, result) {
		t.FailNow()
	}
}

func TestOperatorLessThanOrEqualTo(t *testing.T) {
	evaluate := OperatorLessThanOrEqualTo

	result, err := evaluate(types.IntValue{Value: 21}, types.IntValue{Value: 42})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.True(t, result) {
		t.FailNow()
	}

	result, err = evaluate(types.IntValue{Value: 42}, types.IntValue{Value: 42})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.True(t, result) {
		t.FailNow()
	}

	result, err = evaluate(types.IntValue{Value: 42}, types.IntValue{Value: 21})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.False(t, result) {
		t.FailNow()
	}
}

func TestOperatorGreaterThan(t *testing.T) {
	evaluate := OperatorGreaterThan

	result, err := evaluate(types.IntValue{Value: 21}, types.IntValue{Value: 42})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.False(t, result) {
		t.FailNow()
	}

	result, err = evaluate(types.IntValue{Value: 42}, types.IntValue{Value: 42})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.False(t, result) {
		t.FailNow()
	}

	result, err = evaluate(types.IntValue{Value: 42}, types.IntValue{Value: 21})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.True(t, result) {
		t.FailNow()
	}
}

func TestOperatorGreaterThanOrEqualTo(t *testing.T) {
	evaluate := OperatorGreaterThanOrEqualTo

	result, err := evaluate(types.IntValue{Value: 21}, types.IntValue{Value: 42})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.False(t, result) {
		t.FailNow()
	}

	result, err = evaluate(types.IntValue{Value: 42}, types.IntValue{Value: 42})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.True(t, result) {
		t.FailNow()
	}

	result, err = evaluate(types.IntValue{Value: 42}, types.IntValue{Value: 21})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.True(t, result) {
		t.FailNow()
	}
}
