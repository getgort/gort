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

	// String returns the value as it would be received by a program to which it
	// is passed as an argument.
	String() string

	// EscapedString returns the value as it would be typed on the command line,
	// including quotation marks and escape sequences, if applicable.
	EscapedString() string
}

type CollectionValue interface {
	Value
	Contains(Value) bool
	Elements() []Value
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

func (v BoolValue) String() string {
	return fmt.Sprintf("%v", v.V)
}

func (v BoolValue) EscapedString() string {
	return v.String()
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

func (v FloatValue) String() string {
	return fmt.Sprintf("%v", v.V)
}

func (v FloatValue) EscapedString() string {
	return v.String()
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

func (v IntValue) String() string {
	return fmt.Sprintf("%v", v.V)
}

func (v IntValue) EscapedString() string {
	return v.String()
}

func (v IntValue) Value() interface{} {
	return v.V
}

// ListValue
type ListValue struct {
	V    []Value
	Name string
}

func (v ListValue) Contains(q Value) bool {
	for _, c := range v.V {
		if q.Equals(c) {
			return true
		}
	}

	return false
}

func (v ListValue) Elements() []Value {
	return v.V
}

func (v ListValue) Equals(q Value) bool {
	list, ok := q.(ListValue)
	if !ok {
		return false
	}

	if len(list.V) != len(v.V) {
		return false
	}

	for i, o1 := range list.V {
		o2 := v.V[i]

		if !o1.Equals(o2) {
			return false
		}
	}

	return true
}

func (v ListValue) String() string {
	return v.Name
}

func (v ListValue) EscapedString() string {
	return v.String()
}

func (v ListValue) LessThan(q Value) bool {
	return false
}

func (v ListValue) Value() interface{} {
	return v.V
}

// ListElementValue
type ListElementValue struct {
	V     ListValue
	Index int
}

func (v ListElementValue) Equals(q Value) bool {
	if v.Index < 0 || v.Index >= len(v.V.V) {
		return false
	}

	return v.V.V[v.Index].Equals(q)
}

func (v ListElementValue) LessThan(q Value) bool {
	if v.Index < 0 || v.Index >= len(v.V.V) {
		return false
	}

	return v.V.V[v.Index].LessThan(q)
}

func (v ListElementValue) String() string {
	return fmt.Sprintf("%s[%d]", v.V.Name, v.Index)
}

func (v ListElementValue) EscapedString() string {
	return v.String()
}

func (v ListElementValue) Value() interface{} {
	return v.V.V[v.Index]
}

// MapValue
type MapValue struct {
	V    map[string]Value
	Name string
}

// Contains returns true if q is a StringValue and that key exists in the map.
func (v MapValue) Contains(q Value) bool {
	str, isStr := q.(StringValue)
	if !isStr {
		return false
	}

	_, exists := v.V[str.V]

	return exists
}

func (v MapValue) Elements() []Value {
	values := []Value{}

	for _, v := range v.V {
		values = append(values, v)
	}

	return values
}

func (v MapValue) Equals(q Value) bool {
	m, ok := q.(MapValue)
	if !ok {
		return false
	}

	if len(m.V) != len(v.V) {
		return false
	}

	for key, o1 := range m.V {
		o2 := v.V[key]

		if !o1.Equals(o2) {
			return false
		}
	}

	return true
}

func (v MapValue) LessThan(q Value) bool {
	return false
}

func (v MapValue) String() string {
	return v.Name
}

func (v MapValue) EscapedString() string {
	return v.String()
}

func (v MapValue) Value() interface{} {
	return v.V
}

// MapElementValue
type MapElementValue struct {
	V   MapValue
	Key string
}

func (v MapElementValue) Equals(q Value) bool {
	if v.Key == "" {
		return false
	}

	value, exists := v.V.V[v.Key]

	if bv, ok := q.(BoolValue); ok {
		return bv.V == exists
	}

	if !exists {
		return false
	}

	return value.Equals(q)
}

func (v MapElementValue) LessThan(q Value) bool {
	if v.Key == "" {
		return false
	}

	value, exists := v.V.V[v.Key]
	if !exists {
		return false
	}

	return value.LessThan(q)
}

func (v MapElementValue) String() string {
	return fmt.Sprintf("%s[\"%s\"]", v.V.Name, v.Key)
}

func (v MapElementValue) EscapedString() string {
	return v.String()
}

func (v MapElementValue) Value() interface{} {
	return v.V.V[v.Key]
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

func (v NullValue) String() string {
	return "NULL"
}

func (v NullValue) EscapedString() string {
	return v.String()
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

func (v RegexValue) String() string {
	return fmt.Sprintf("%v", v.V)
}

func (v RegexValue) EscapedString() string {
	return v.String()
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

func (v StringValue) String() string {
	s := v.V

	return s
}

func (v StringValue) EscapedString() string {
	switch v.Quote {
	case '\'':
		return fmt.Sprintf("'%s'", v.V)
	case '"':
		return strconv.Quote(v.V)
	default:
		return v.V // Theoretically impossible.
	}
}

func (v StringValue) Value() interface{} {
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

func (v UnknownValue) String() string {
	return fmt.Sprintf("??%s??", v.V)
}

func (v UnknownValue) EscapedString() string {
	return v.String()
}

func (v UnknownValue) Value() interface{} {
	return v.V
}
