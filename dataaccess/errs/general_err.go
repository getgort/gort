package errs

import (
	"errors"
)

var (
	// ErrAdminUndeletable is returned when an attempt is made to delete an
	// admin user/account/etc.
	ErrAdminUndeletable = errors.New("admin can't be deleted")

	// ErrDataAccess that an error has been reported by the data store.
	ErrDataAccess = errors.New("error reported by the data store")

	// ErrDataAccessCantConnect indicates that an error has been reported by the
	// data store.
	ErrDataAccessCantConnect = errors.New("can't connect to the data store")

	// ErrDataAccessCantInitialize indicates that the data layer can't be
	// initialized.
	ErrDataAccessCantInitialize = errors.New("can't initialize data access layer")

	// ErrNotImplemented indicates that the DAL implementation isn't complete for
	// the invoked method.
	ErrNotImplemented = errors.New("method not implemented")

	// ErrDataAccessNotInitialized indicates that the data layer has not been
	// initialized.
	ErrDataAccessNotInitialized = errors.New("data access layer nit initialized")
)
