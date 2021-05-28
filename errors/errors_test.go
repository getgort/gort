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

package errors

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestErrorEquality(t *testing.T) {
	const newErrMsg = "this is an error"
	const wrappedErrMsg = "this is a wrapped error"

	newErr := errors.New(newErrMsg)
	wrappedErr := errors.New(wrappedErrMsg)
	nestedErr := Wrap(newErr, wrappedErr)

	if newErr.Error() != newErrMsg {
		t.Error("Wrong new error message")
	}

	if wrappedErr.Error() != wrappedErrMsg {
		t.Error("Wrong wrapped error message")
	}

	isaNestedError := reflect.TypeOf(nestedErr) == reflect.TypeOf(NestedError{})
	if !isaNestedError {
		t.Error("This really should be a nested error.")
	}

	if !strings.Contains(nestedErr.Error(), newErrMsg) {
		t.Error("Didn't find newErrMsg")
	}

	if !strings.Contains(nestedErr.Error(), wrappedErrMsg) {
		t.Error("Didn't find wrappedErrMsg")
	}

	if !Is(newErr, nestedErr) {
		t.Error("Is false when expected true")
	}

	if Is(newErr, wrappedErr) {
		t.Error("Is true when expected false")
	}
}
