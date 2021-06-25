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
	"github.com/getgort/gort/types"
)

type Operator func(a, b types.Value) bool

func Equals(a, b types.Value) bool {
	return a.Equals(b)
}

func NotEquals(a, b types.Value) bool {
	return !a.Equals(b)
}

func LessThan(a, b types.Value) bool {
	return a.LessThan(b)
}

func LessThanOrEqualTo(a, b types.Value) bool {
	return a.LessThan(b) || a.Equals(b)
}

func GreaterThan(a, b types.Value) bool {
	return !(a.LessThan(b) || a.Equals(b))
}

func GreaterThanOrEqualTo(a, b types.Value) bool {
	return !a.LessThan(b)
}

func In(a, b types.Value) bool {
	coll, ok := b.(types.CollectionValue)
	if !ok {
		return Equals(a, b)
	}

	return coll.Contains(a)
}
