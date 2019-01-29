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

	if !ErrEquals(newErr, nestedErr) {
		t.Error("ErrEquals false when expected true")
	}

	if ErrEquals(newErr, wrappedErr) {
		t.Error("ErrEquals true when expected false")
	}
}
