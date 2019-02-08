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

// ErrEquals compares two errors and returns true if they have the same message.
// If either is a NestedError, only the top-level message is checked.
func ErrEquals(err1 error, err2 error) bool {
	errStr := func(err error) string {
		switch v := err.(type) {
		case NestedError:
			return v.Message
		default:
			return err.Error()
		}
	}

	return errStr(err1) == errStr(err2)
}
