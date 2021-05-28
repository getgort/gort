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
	ErrDataAccessNotInitialized = errors.New("data access layer not initialized")

	// ErrFieldRequired is returned by an insert or update when one of the
	// struct's required field values is empty.
	ErrFieldRequired = errors.New("a required field is missing")
)
