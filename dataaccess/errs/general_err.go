package errs

import (
	"errors"
)

// ErrDataAccessCantInitialize indicates that the data layer can't be initialized.
var ErrDataAccessCantInitialize = errors.New("can't initialize data access layer")

// ErrDataAccessCantConnect indicates that an error has been reported by the data store.
var ErrDataAccessCantConnect = errors.New("can't connect to the data store")

// ErrDataAccessindicates that an error has been reported by the data store.
var ErrDataAccess = errors.New("error reported by the data store")

// ErrNotImplemented indicates that the DAL implementation isn't complete for
// the method.
var ErrNotImplemented = errors.New("method not implemented")

// ErrAdminUndeletable ...
var ErrAdminUndeletable = errors.New("admin can't be deleted")
