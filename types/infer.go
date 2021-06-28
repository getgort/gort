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

var (
	reBool                = regexp.MustCompile(`^(true|True|TRUE|false|False|FALSE)$`)
	reFloat               = regexp.MustCompile(`^-?[0-9]*\.[0-9]+$`)
	reInt                 = regexp.MustCompile(`^-?[0-9]+$`)
	reRegex               = regexp.MustCompile(`^[\"\']?/.*/[\"\']?$`)
	reRegexTrim           = regexp.MustCompile(`(^[\"\']?/|/[\"\']?$)`)
	reString              = regexp.MustCompile(`^[“”\"\'].*[“”\"\']$`)
	reStringTrim          = regexp.MustCompile(`(^[“”\"\']?|[“”\"\']?$)`)
	reCollectionReference = regexp.MustCompile(`^([A-Za-z0-9_]*)\[(.*)\]$`)
	reList                = regexp.MustCompile(`^\[(.*)\]$`)
)

// Inferrer is used to infer data types from string representations and
// retrieve the coresponding appropriately-typed Value.
type Inferrer struct {
	literalLists         bool
	collectionReferences bool
	regularExpressions   bool
	strictStrings        bool
}

// Setting ComplexTypes is a helper function that enables the Infer method to
// identify literal lists ([ "foo", "bar" ]), collection references
// (options["foo"]), and regular expressions (/^foo$/).
func (i Inferrer) ComplexTypes(enabled bool) Inferrer {
	i.literalLists = enabled
	i.collectionReferences = enabled
	i.regularExpressions = enabled
	return i
}

// LiteralLists allows the Infer method to infer list literals (["foo, "bar"]).
// Lists may include regular expressions, but may not include other complex
// types.
func (i Inferrer) LiteralLists(enabled bool) Inferrer {
	i.literalLists = enabled
	return i
}

// CollectionReferences allows the Infer method to identify map
// (options["foo"]) and list references (arg[0]), returning MapElementValue and
// ListElementValue values, respectively. An argument that isn't a string or
// integer will cause an error to be returned (unless strict strings is false).
func (i Inferrer) CollectionReferences(enabled bool) Inferrer {
	i.collectionReferences = enabled
	return i
}

// RegularExpressions allows regular expressions (/^foo$/) to be inferred.
func (i Inferrer) RegularExpressions(enabled bool) Inferrer {
	i.regularExpressions = enabled
	return i
}

// StrictStrings requires strings to be wrapped in single or double quotes
// (smart quotes are automatically converted to double quotes), and unknown
// values are returned as an UnknownValue. If not set, unquoted values that
// aren't clearly recognizable as another type are returned as StringValue
// values with a Quote value of \u0000 (null character).
func (i Inferrer) StrictStrings(enabled bool) Inferrer {
	i.strictStrings = enabled
	return i
}

// Infer accepts a string, attempts to determine its type, and based
// on the outcome returns an appropriate Value value. if strictStrings
// is true unquoted values that aren't obviously another type will return an
// error; if not then they will be treated as strings with a "quote flavor"
// of null ('\u0000). If basicsTypes is set then only the "basic" types (bool,
// float, int, string) will be returned.
func (i Inferrer) Infer(str string) (Value, error) {
	subinferrer := Inferrer{}.ComplexTypes(false).RegularExpressions(true).StrictStrings(true)

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

	case i.regularExpressions && reRegex.MatchString(str):
		value := reRegexTrim.ReplaceAllString(str, "")
		return RegexValue{V: value}, nil

	case reString.MatchString(str):
		quoteFlavor := str[0]
		value := reStringTrim.ReplaceAllString(str, "")
		return StringValue{V: value, Quote: rune(quoteFlavor)}, nil

	case i.literalLists && reList.MatchString(str):
		submatches := reList.FindStringSubmatch(str)
		if len(submatches) != 2 {
			return NullValue{}, fmt.Errorf("cannot parse list: %s", str)
		}

		strs := splitListLiteral(submatches[1])
		values, err := subinferrer.InferAll(strs)
		if err != nil {
			return NullValue{}, fmt.Errorf("cannot parse list: %w", err)
		}

		return ListValue{V: values}, nil

	case i.collectionReferences && reCollectionReference.MatchString(str):
		subs := reCollectionReference.FindStringSubmatch(str)
		name, param := subs[1], subs[2]

		paramValue, err := subinferrer.Infer(param)
		if err != nil {
			return NullValue{}, err
		}

		// Determine the collection type by the argument type.
		// [IntValue] -> ListElementValue
		// [StringValue] -> MapElementValue
		// [Anything else] -> error
		switch v := paramValue.(type) {
		case IntValue:
			return ListElementValue{V: ListValue{Name: name}, Index: v.Value().(int)}, nil
		case StringValue:
			return MapElementValue{V: MapValue{Name: name}, Key: v.Value().(string)}, nil
		default:
			return NullValue{}, fmt.Errorf("invalid collection parameter: %T", v)
		}

	default:
		if i.strictStrings {
			return UnknownValue{V: str}, nil
		}

		return StringValue{V: str}, nil
	}
}

func (i Inferrer) InferAll(strs []string) ([]Value, error) {
	values := []Value{}

	for _, s := range strs {
		v, err := i.Infer(s)
		if err != nil {
			return nil, err
		}

		values = append(values, v)
	}

	return values, nil
}

func splitListLiteral(str string) []string {
	str = strings.TrimSpace(str)

	if len(str) == 0 {
		return []string{}
	}

	tokens := make([]string, 0)
	inDoubleQuote := false
	inSingleQuote := false
	inRegex := false
	currentToken := strings.Builder{}

	for _, ch := range str {
		switch ch {
		case '\r':
			fallthrough
		case '\t':
			fallthrough
		case ' ':
			if (inDoubleQuote || inSingleQuote || inRegex) && currentToken.Len() > 0 {
				currentToken.WriteRune(ch)
			}
		case ',':
			if inDoubleQuote || inSingleQuote || inRegex {
				currentToken.WriteRune(',')
			} else {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
		case '“': // Smart left quote
			fallthrough
		case '”': // Smart right quote
			fallthrough
		case '"':
			currentToken.WriteRune('"')
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			}
		case '\'':
			currentToken.WriteRune('\'')
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			}
		case '/':
			currentToken.WriteRune('/')
			if !inDoubleQuote && !inSingleQuote {
				inRegex = !inRegex
			}
		default:
			currentToken.WriteRune(ch)
		}
	}

	tokens = append(tokens, currentToken.String())

	return tokens
}
