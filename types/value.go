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
	String() string
	Resolve() (interface{}, error)
	Equals(Value) (bool, error)
	LessThan(Value) (bool, error)
}

// BoolValue is a literal boolean value.
type BoolValue struct {
	Value bool
}

func (v BoolValue) Equals(q Value) (bool, error) {
	switch o := q.(type) {
	case BoolValue:
		return v.Value == o.Value, nil
	case IntValue:
		switch o.Value {
		case 0:
			return !v.Value, nil
		case 1:
			return v.Value, nil
		default:
			return false, fmt.Errorf("cannot compare integer %d to bool", o.Value)
		}
	case StringValue:
		b, err := strconv.ParseBool(o.Value)
		if err != nil {
			return false, err
		}
		return v.Value == b, nil
	case RegexValue:
		return o.Equals(v)
	}

	return false, fmt.Errorf("cannot compare %T and %T", v, q)
}

func (v BoolValue) LessThan(q Value) (bool, error) {
	return false, fmt.Errorf("%T is not a scalar type", v)
}

func (v BoolValue) Resolve() (interface{}, error) {
	return v.Value, nil
}

func (v BoolValue) String() string {
	return fmt.Sprintf("%t", v.Value)
}

// FloatValue is a literal floating point value.
type FloatValue struct {
	Value float64
}

func (v FloatValue) Equals(q Value) (bool, error) {
	switch o := q.(type) {
	case FloatValue:
		return v.Value == o.Value, nil
	case IntValue:
		asFloat := float64(o.Value)
		return asFloat == v.Value, nil
	case RegexValue:
		return o.Equals(v)
	}

	return false, fmt.Errorf("cannot compare %T and %T", v, q)
}

func (v FloatValue) LessThan(q Value) (bool, error) {
	switch o := q.(type) {
	case FloatValue:
		return v.Value < o.Value, nil
	case IntValue:
		asFloat := float64(o.Value)
		return asFloat < v.Value, nil
	}

	return false, fmt.Errorf("cannot compare %T and %T", v, q)
}

func (v FloatValue) Resolve() (interface{}, error) {
	return v.Value, nil
}

func (v FloatValue) String() string {
	return fmt.Sprintf("%f", v.Value)
}

// IntValue is a literal integer value.
type IntValue struct {
	Value int
}

func (v IntValue) Equals(q Value) (bool, error) {
	switch o := q.(type) {
	case BoolValue:
		return o.Equals(v)
	case FloatValue:
		return o.Equals(v)
	case IntValue:
		return v.Value == o.Value, nil
	case RegexValue:
		return o.Equals(v)
	}

	return false, fmt.Errorf("cannot compare %T and %T", v, q)
}

func (v IntValue) LessThan(q Value) (bool, error) {
	switch o := q.(type) {
	case IntValue:
		return v.Value < o.Value, nil
	case FloatValue:
		asFloat := float64(v.Value)
		return asFloat < o.Value, nil
	}

	return false, fmt.Errorf("cannot compare %T and %T", v, q)
}

func (v IntValue) Resolve() (interface{}, error) {
	return v.Value, nil
}

func (v IntValue) String() string {
	return fmt.Sprintf("%d", v.Value)
}

// RegexValue describes a regular expression. Its Resolve() function returns
// the product of `regexp.CompilePOSIX(v.Value)`.
type RegexValue struct {
	Value string
}

func (v RegexValue) Equals(q Value) (bool, error) {
	switch o := q.(type) {
	case RegexValue:
		return false, fmt.Errorf("cannot compare %T and %T", v, q)
	default:
		resolve, err := v.Resolve()
		if err != nil {
			return false, err
		}

		re := resolve.(*regexp.Regexp)

		return re.MatchString(o.String()), nil
	}
}

func (v RegexValue) LessThan(q Value) (bool, error) {
	return false, fmt.Errorf("%T is not a scalar type", v)
}

func (v RegexValue) Resolve() (interface{}, error) {
	return regexp.CompilePOSIX(v.Value)
}

func (v RegexValue) String() string {
	return v.Value
}

// StringValue is a literal string value.
type StringValue struct {
	Value string
	Quote rune
}

func (v StringValue) Equals(q Value) (bool, error) {
	switch o := q.(type) {
	case BoolValue:
		return o.Equals(v)
	case RegexValue:
		return o.Equals(v)
	case StringValue:
		return v.Value == o.Value, nil
	}

	return false, fmt.Errorf("cannot compare %T and %T", v, q)
}

func (v StringValue) LessThan(q Value) (bool, error) {
	return false, fmt.Errorf("%T is not a scalar type", v)
}

func (v StringValue) Resolve() (interface{}, error) {
	return v.Value, nil
}

func (v StringValue) String() string {
	return v.Value
}
