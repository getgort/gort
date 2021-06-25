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

var (
	reBool       = regexp.MustCompile(`^(true|True|TRUE|false|False|FALSE)$`)
	reFloat      = regexp.MustCompile(`^-?[0-9]*\.[0-9]+$`)
	reInt        = regexp.MustCompile(`^-?[0-9]+$`)
	reRegex      = regexp.MustCompile(`^[\"\']?/.*/[\"\']?$`)
	reRegexTrim  = regexp.MustCompile(`(^[\"\']?/|/[\"\']?$)`)
	reString     = regexp.MustCompile(`^[\"\'].*[\"\']$`)
	reStringTrim = regexp.MustCompile(`(^[\"\']?|[\"\']?$)`)
	reCollection = regexp.MustCompile(`^([A-Za-z0-9_]*)\[(.*)\]$`)
)

// Infer accepts a string, attempts to determine its type, and based
// on the outcome returns an appropriate Value value. if strictStrings
// is true unquoted values that aren't obviously another type will return an
// error; if not then they will be treated as strings with a "quote flavor"
// of null ('\u0000). If basicsTypes is set then only the "basic" types (bool,
// float, int, string) will be returned.
func Infer(str string, basicTypes, strictStrings bool) (Value, error) {
	switch {
	case reBool.MatchString(str):
		value, err := strconv.ParseBool(str)
		return BoolValue{V: value}, err

	case reFloat.MatchString(str):
		value, err := strconv.ParseFloat(str, 64)
		return FloatValue{V: value}, err

	case reInt.MatchString(str):
		value, err := strconv.Atoi(str)
		return IntValue{V: value}, err

	case !basicTypes && reRegex.MatchString(str):
		value := reRegexTrim.ReplaceAllString(str, "")
		return RegexValue{V: value}, nil

	case reString.MatchString(str):
		quoteFlavor := str[0]
		value := reStringTrim.ReplaceAllString(str, "")
		return StringValue{V: value, Quote: rune(quoteFlavor)}, nil

	case !basicTypes && reCollection.MatchString(str):
		subs := reCollection.FindStringSubmatch(str)
		name, param := subs[1], subs[2]
		paramValue, err := Infer(param, true, true)
		if err != nil {
			return NullValue{}, err
		}

		// Determine the collection type by the type of argument.
		// IntValue -> ListValue
		// StringValue -> MapValue
		// Anything else -> error
		switch v := paramValue.(type) {
		case IntValue:
			return ListValue{Name: name, Index: v.Value().(int)}, nil
		case StringValue:
			return MapValue{Name: name, Key: v.Value().(string)}, nil
		default:
			return NullValue{}, fmt.Errorf("invalid collection parameter: %T", v)
		}

	default:
		if strictStrings {
			return UnknownValue{V: str}, nil
		}

		return StringValue{V: str}, nil
	}
}

func InferAll(strs []string, basicTypes, strictStrings bool) ([]Value, error) {
	values := []Value{}

	for _, s := range strs {
		v, err := Infer(s, basicTypes, strictStrings)
		if err != nil {
			return nil, err
		}

		values = append(values, v)
	}

	return values, nil
}
