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

package memory

import (
	"context"

	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/dataaccess/errs"
)

// UserAuthenticate authenticates a username/password combination.
func (da *InMemoryDataAccess) UserAuthenticate(ctx context.Context, username string, password string) (bool, error) {
	exists, err := da.UserExists(ctx, username)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, errs.ErrNoSuchUser
	}

	user, err := da.UserGet(ctx, username)
	if err != nil {
		return false, err
	}

	return password == user.Password, nil
}

// UserCreate is used to create a new Gort user in the data store. An error is
// returned if the username is empty or if a user already exists.
func (da *InMemoryDataAccess) UserCreate(ctx context.Context, user rest.User) error {
	if user.Username == "" {
		return errs.ErrEmptyUserName
	}

	exists, err := da.UserExists(ctx, user.Username)
	if err != nil {
		return err
	}
	if exists {
		return errs.ErrUserExists
	}

	da.users[user.Username] = &user

	return nil
}

// UserDelete deletes an existing user from the data store. An error is
// returned if the username parameter is empty of if the user doesn't
// exist.
func (da *InMemoryDataAccess) UserDelete(ctx context.Context, username string) error {
	if username == "" {
		return errs.ErrEmptyUserName
	}

	// Thou Shalt Not Delete Admin
	if username == "admin" {
		return errs.ErrAdminUndeletable
	}

	exists, err := da.UserExists(ctx, username)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchUser
	}

	delete(da.users, username)

	return nil
}

// UserExists is used to determine whether a Gort user with the given username
// exists in the data store.
func (da *InMemoryDataAccess) UserExists(ctx context.Context, username string) (bool, error) {
	_, exists := da.users[username]

	return exists, nil
}

// UserGet returns a user from the data store. An error is returned if the
// username parameter is empty or if the user doesn't exist.
func (da *InMemoryDataAccess) UserGet(ctx context.Context, username string) (rest.User, error) {
	if username == "" {
		return rest.User{}, errs.ErrEmptyUserName
	}

	exists, err := da.UserExists(ctx, username)
	if err != nil {
		return rest.User{}, err
	}
	if !exists {
		return rest.User{}, errs.ErrNoSuchUser
	}

	user := da.users[username]

	return *user, nil
}

// UserGetByEmail returns a user from the data store. An error is returned if
// the email parameter is empty or if the user doesn't exist.
func (da *InMemoryDataAccess) UserGetByEmail(ctx context.Context, email string) (rest.User, error) {
	for _, v := range da.users {
		if v.Email == email {
			return *v, nil
		}
	}

	return rest.User{}, errs.ErrNoSuchUser
}

// UserList returns a list of all known users in the datastore.
// Passwords are not included. Nice try.
func (da *InMemoryDataAccess) UserList(ctx context.Context) ([]rest.User, error) {
	list := make([]rest.User, 0)

	for _, u := range da.users {
		u.Password = ""
		list = append(list, *u)
	}

	return list, nil
}

// UserUpdate is used to update an existing user. An error is returned if the
// username is empty or if the user doesn't exist.
// TODO Should we let this create users that don't exist?
func (da *InMemoryDataAccess) UserUpdate(ctx context.Context, user rest.User) error {
	if user.Username == "" {
		return errs.ErrEmptyUserName
	}

	exists, err := da.UserExists(ctx, user.Username)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchUser
	}

	da.users[user.Username] = &user

	return nil
}

// UserGroupList comments TBD
func (da *InMemoryDataAccess) UserGroupList(ctx context.Context, user string) ([]rest.Group, error) {
	return []rest.Group{}, errs.ErrNotImplemented
}

// UserGroupAdd comments TBD
func (da *InMemoryDataAccess) UserGroupAdd(ctx context.Context, user string, group string) error {
	return errs.ErrNotImplemented
}

// UserGroupDelete comments TBD
func (da *InMemoryDataAccess) UserGroupDelete(ctx context.Context, user string, group string) error {
	return errs.ErrNotImplemented
}
