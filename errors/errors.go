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
	"fmt"
)

// NestedError contains an error inside another error (which may also be nested)
type NestedError struct {
	Message string
	Err     error
}

// Error returns the error message and the message of any wrapped error.
func (e NestedError) Error() string {
	return fmt.Sprintf("%s\n  contains: %s", e.Message, e.Err.Error())
}

// Wrap will return a new error of the specified type, but wrapping the
// specified sub-error.
func Wrap(newErr error, nestedErr error) error {
	return NestedError{Message: newErr.Error(), Err: nestedErr}
}

// Wrap will return a new error of the specified type, but wrapping the
// specified sub-error.
func WrapStr(message string, nestedErr error) error {
	return NestedError{Message: message, Err: nestedErr}
}

// Is compares two errors and returns true if they have the same message.
// If either is a NestedError, only the top-level message is checked.
func Is(err1 error, err2 error) bool {
	errStr := func(err error) string {
		if err == nil {
			return "nil"
		}

		switch v := err.(type) {
		case NestedError:
			return v.Message
		default:
			return err.Error()
		}
	}

	return errStr(err1) == errStr(err2)
}
