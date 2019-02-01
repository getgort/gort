package errs

import (
	"errors"
)

// ErrNoSuchGroup indicates...
var ErrNoSuchGroup = errors.New("no such group")

// ErrEmptyGroupName indicates...
var ErrEmptyGroupName = errors.New("group name is empty")

// ErrGroupExists TBD
var ErrGroupExists = errors.New("group already exists")
