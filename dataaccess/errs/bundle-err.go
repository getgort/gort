package errs

import (
	"errors"
)

// ErrNoSuchBundle indicates...
var ErrNoSuchBundle = errors.New("no such bundle")

// ErrEmptyBundleName indicates...
var ErrEmptyBundleName = errors.New("bundle name is empty")

// ErrEmptyBundleVersion indicates...
var ErrEmptyBundleVersion = errors.New("bundle version is empty")

// ErrBundleExists TBD
var ErrBundleExists = errors.New("bundle already exists")
