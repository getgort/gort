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

package types

import (
	"fmt"
	"regexp"
	"strconv"
)

type Value interface {
	Equals(Value) bool
	LessThan(Value) bool
	Value() interface{}
}

type CollectionValue interface {
	Name() string
	Contains(Value) bool
}

// BoolValue is a literal boolean value.
type BoolValue struct {
	V bool
}

func (v BoolValue) Equals(q Value) bool {
	switch o := q.(type) {
	case BoolValue:
		return v.V == o.V
	case IntValue:
		switch o.V {
		case 0:
			return !v.V
		case 1:
			return v.V
		default:
			return false
		}
	case StringValue:
		b, err := strconv.ParseBool(o.V)
		if err != nil {
			return false
		}
		return v.V == b
	case RegexValue:
		return o.Equals(v)
	}

	return false
}

func (v BoolValue) LessThan(q Value) bool {
	return false
}

func (v BoolValue) Value() interface{} {
	return v.V
}

// FloatValue is a literal floating point value.
type FloatValue struct {
	V float64
}

func (v FloatValue) Equals(q Value) bool {
	switch o := q.(type) {
	case FloatValue:
		return v.V == o.V
	case IntValue:
		asFloat := float64(o.V)
		return asFloat == v.V
	case RegexValue:
		return o.Equals(v)
	}

	return false
}

func (v FloatValue) LessThan(q Value) bool {
	switch o := q.(type) {
	case FloatValue:
		return v.V < o.V
	case IntValue:
		asFloat := float64(o.V)
		return asFloat < v.V
	}

	return false
}

func (v FloatValue) Value() interface{} {
	return v.V
}

// IntValue is a literal integer value.
type IntValue struct {
	V int
}

func (v IntValue) Equals(q Value) bool {
	switch o := q.(type) {
	case BoolValue:
		return o.Equals(v)
	case FloatValue:
		return o.Equals(v)
	case IntValue:
		return v.V == o.V
	case RegexValue:
		return o.Equals(v)
	}

	return false
}

func (v IntValue) LessThan(q Value) bool {
	switch o := q.(type) {
	case IntValue:
		return v.V < o.V
	case FloatValue:
		asFloat := float64(v.V)
		return asFloat < o.V
	}

	return false
}

func (v IntValue) Value() interface{} {
	return v.V
}

// NullValue
type NullValue struct{}

func (v NullValue) Equals(q Value) bool {
	switch q.(type) {
	case NullValue:
		return true
	default:
		return false
	}
}

func (v NullValue) LessThan(q Value) bool {
	return false
}

func (v NullValue) Value() interface{} {
	return nil
}

// RegexValue describes a regular expression.
type RegexValue struct {
	V string
}

func (v RegexValue) Pattern() (*regexp.Regexp, error) {
	return regexp.CompilePOSIX(v.V)
}

func (v RegexValue) Equals(q Value) bool {
	re, err := v.Pattern()
	if err != nil {
		return false
	}

	switch o := q.(type) {
	case RegexValue:
		return v.V == o.V
	default:
		return re.MatchString(fmt.Sprintf("%v", q.Value()))
	}
}

func (v RegexValue) LessThan(q Value) bool {
	return false
}

func (v RegexValue) Value() interface{} {
	return v.V
}

// StringValue is a literal string value.
type StringValue struct {
	V     string
	Quote rune
}

func (v StringValue) Equals(q Value) bool {
	switch o := q.(type) {
	case BoolValue:
		return o.Equals(v)
	case RegexValue:
		return o.Equals(v)
	case StringValue:
		return v.V == o.V
	}

	return false
}

func (v StringValue) LessThan(q Value) bool {
	return false
}

func (v StringValue) Value() interface{} {
	return v.V
}

// MapValue
type MapValue struct {
	V    map[string]Value
	Name string
	Key  string
}

func (v MapValue) Contains(q Value) bool {
	key := fmt.Sprintf("%v", q.Value())

	if _, ok := v.V[key]; ok {
		return true
	}

	return false
}

func (v MapValue) Equals(q Value) bool {
	value, exists := v.V[v.Key]

	if bv, ok := q.(BoolValue); ok {
		return bv.V == exists
	}

	if !exists {
		return false
	}

	return value.Equals(q)
}

func (v MapValue) LessThan(q Value) bool {
	value, exists := v.V[v.Key]
	if !exists {
		return false
	}

	return value.LessThan(q)
}

func (v MapValue) Value() interface{} {
	return v.V
}

// ListValue
type ListValue struct {
	V     []Value
	Name  string
	Index int
}

func (v ListValue) Contains(q Value) bool {
	for _, c := range v.V {
		if c.Equals(q) {
			return true
		}
	}

	return false
}

func (v ListValue) Equals(q Value) bool {
	if v.Index < 0 || v.Index >= len(v.V) {
		return false
	}

	return v.V[v.Index].Equals(q)
}

func (v ListValue) LessThan(q Value) bool {
	return false
}

func (v ListValue) Value() interface{} {
	return v.V
}

// UnknownValue is returned by Parse when it can't determine a value type
// solely by looking at it. This could indicate a function, named
// collection(arg, option), or other named entity.
type UnknownValue struct {
	V string
}

func (v UnknownValue) Equals(q Value) bool {
	return false
}

func (v UnknownValue) LessThan(q Value) bool {
	return false
}

func (v UnknownValue) Value() interface{} {
	return v.V
}
