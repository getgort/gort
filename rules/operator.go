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

	"github.com/getgort/gort/types"
)

type Operator func(a, b types.Value) (bool, error)

func Equals(a, b types.Value) (bool, error) {
	return a.Equals(b)
}

func NotEquals(a, b types.Value) (bool, error) {
	eq, err := a.Equals(b)
	return !eq, err
}

func LessThan(a, b types.Value) (bool, error) {
	return a.LessThan(b)
}

func LessThanOrEqualTo(a, b types.Value) (bool, error) {
	lt, err := a.LessThan(b)
	if err != nil {
		return false, err
	}

	eq, err := a.Equals(b)
	if err != nil {
		return false, err
	}

	return lt || eq, nil
}

func GreaterThan(a, b types.Value) (bool, error) {
	lt, err := a.LessThan(b)
	if err != nil {
		return false, err
	}

	eq, err := a.Equals(b)
	if err != nil {
		return false, err
	}

	return !(lt || eq), nil
}

func GreaterThanOrEqualTo(a, b types.Value) (bool, error) {
	lt, err := a.LessThan(b)
	return !lt, err
}

func In(a, b types.Value) (bool, error) {
	return false, fmt.Errorf("not implemented")
}
