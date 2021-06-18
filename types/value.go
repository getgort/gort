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
	"strings"
)

type Value interface {
	forcingFunction()
}

type BoolValue struct {
	Value bool
}

func (v BoolValue) forcingFunction() {}

type IntValue struct {
	Value int
}

func (v IntValue) forcingFunction() {}

type StringValue struct {
	Value       string
	QuoteFlavor rune
}

func (v StringValue) forcingFunction() {}

type FloatValue struct {
	Value float64
}

func (v FloatValue) forcingFunction() {}

var (
	reInt    = regexp.MustCompile(`^-?[0-9]+$`)
	reFloat  = regexp.MustCompile(`^-?[0-9]*\.[0-9]+$`)
	reBool   = regexp.MustCompile(`^(true|false)$`)
	reString = regexp.MustCompile(`^[\"\'].*[\"\']$`)
)

// GuessTypedValue accepts a string, attempts to determine its type, and based
// on the outcome returns an appropriate Value value. if strictStrings
// is true, unquoted values that aren't obviously another type will return an
// error; else they will be treated as strings with a "quote flavor" of null
// ('\u0000); otherwise
func GuessTypedValue(str string, strictStrings bool) (Value, error) {
	switch {
	case reBool.MatchString(str):
		value, err := strconv.ParseBool(str)
		return BoolValue{Value: value}, err

	case reFloat.MatchString(str):
		value, err := strconv.ParseFloat(str, 64)
		return FloatValue{Value: value}, err

	case reInt.MatchString(str):
		value, err := strconv.Atoi(str)
		return IntValue{Value: value}, err

	case reString.MatchString(str):
		quoteFlavor := str[0]
		value := strings.Trim(str, `"'`)
		return StringValue{Value: value, QuoteFlavor: rune(quoteFlavor)}, nil

	default:
		if !strictStrings {
			return StringValue{Value: str, QuoteFlavor: '\u0000'}, nil
		}

		return nil, fmt.Errorf("unknown type: %s", str)
	}
}
