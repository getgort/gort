package errs

import (
	"errors"
)

// ErrNoSuchUser indicates that the user doesn't exist, or doesn't exist in
// a group.
var ErrNoSuchUser = errors.New("no such user")

// ErrEmptyUserName indicates...
var ErrEmptyUserName = errors.New("user name is empty")

// ErrUserExists TBD
var ErrUserExists = errors.New("user already exists")
