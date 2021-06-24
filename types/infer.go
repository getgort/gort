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
		return BoolValue{Value: value}, err

	case reFloat.MatchString(str):
		value, err := strconv.ParseFloat(str, 64)
		return FloatValue{Value: value}, err

	case reInt.MatchString(str):
		value, err := strconv.Atoi(str)
		return IntValue{Value: value}, err

	case !basicTypes && reRegex.MatchString(str):
		value := reRegexTrim.ReplaceAllString(str, "")
		return RegexValue{Value: value}, nil

	case reString.MatchString(str):
		quoteFlavor := str[0]
		value := reStringTrim.ReplaceAllString(str, "")
		return StringValue{Value: value, Quote: rune(quoteFlavor)}, nil

	default:
		if !strictStrings {
			return StringValue{Value: str, Quote: '\u0000'}, nil
		}

		return nil, fmt.Errorf("unknown type: %s", str)
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
