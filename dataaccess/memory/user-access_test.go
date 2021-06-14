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
	"testing"

	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/dataaccess/errs"
	"github.com/stretchr/testify/assert"
)

func testUserAccess(t *testing.T) {
	t.Run("testUserNotExists", testUserNotExists)
	t.Run("testUserCreate", testUserCreate)
	t.Run("testUserAuthenticate", testUserAuthenticate)
	t.Run("testUserExists", testUserExists)
	t.Run("testUserDelete", testUserDelete)
	t.Run("testUserGet", testUserGet)
	t.Run("testUserList", testUserList)
	t.Run("testUserUpdate", testUserUpdate)
}

func testUserNotExists(t *testing.T) {
	var exists bool

	err := da.Initialize(ctx)
	assert.NoError(t, err)

	exists, _ = da.UserExists(ctx, "test-not-exists")
	if exists {
		t.Error("User should not exist now")
		t.FailNow()
	}
}

func testUserCreate(t *testing.T) {
	var err error
	var user rest.User

	// Expect an error
	err = da.UserCreate(ctx, user)
	expectErr(t, err, errs.ErrEmptyUserName)

	// Expect no error
	err = da.UserCreate(ctx, rest.User{Username: "test-create", Email: "test-create@bar.com"})
	defer da.UserDelete(ctx, "test-create")
	assert.NoError(t, err)

	// Expect an error
	err = da.UserCreate(ctx, rest.User{Username: "test-create", Email: "test-create@bar.com"})
	expectErr(t, err, errs.ErrUserExists)
}

func testUserAuthenticate(t *testing.T) {
	var err error

	authenticated, err := da.UserAuthenticate(ctx, "test-auth", "no-match")
	expectErr(t, err, errs.ErrNoSuchUser)
	if authenticated {
		t.Error("Expected false")
		t.FailNow()
	}

	// Expect no error
	err = da.UserCreate(ctx, rest.User{
		Username: "test-auth",
		Email:    "test-auth@bar.com",
		Password: "password",
	})
	defer da.UserDelete(ctx, "test-auth")
	assert.NoError(t, err)

	authenticated, err = da.UserAuthenticate(ctx, "test-auth", "no-match")
	assert.NoError(t, err)
	if authenticated {
		t.Error("Expected false")
		t.FailNow()
	}

	authenticated, err = da.UserAuthenticate(ctx, "test-auth", "password")
	assert.NoError(t, err)
	if !authenticated {
		t.Error("Expected true")
		t.FailNow()
	}
}
func testUserExists(t *testing.T) {
	var exists bool

	exists, _ = da.UserExists(ctx, "test-exists")
	if exists {
		t.Error("User should not exist now")
		t.FailNow()
	}

	// Now we add a user to find.
	err := da.UserCreate(ctx, rest.User{Username: "test-exists", Email: "test-exists@bar.com"})
	defer da.UserDelete(ctx, "test-exists")
	assert.NoError(t, err)

	exists, _ = da.UserExists(ctx, "test-exists")
	if !exists {
		t.Error("User should exist now")
		t.FailNow()
	}
}

func testUserDelete(t *testing.T) {
	// Delete blank user
	err := da.UserDelete(ctx, "")
	expectErr(t, err, errs.ErrEmptyUserName)

	// Delete admin user
	err = da.UserDelete(ctx, "admin")
	expectErr(t, err, errs.ErrAdminUndeletable)

	// Delete user that doesn't exist
	err = da.UserDelete(ctx, "no-such-user")
	expectErr(t, err, errs.ErrNoSuchUser)

	user := rest.User{Username: "test-delete", Email: "foo1.example.com"}
	da.UserCreate(ctx, user) // This has its own test
	defer da.UserDelete(ctx, "test-delete")

	err = da.UserDelete(ctx, "test-delete")
	assert.NoError(t, err)

	exists, _ := da.UserExists(ctx, "test-delete")
	if exists {
		t.Error("Shouldn't exist anymore!")
		t.FailNow()
	}
}

func testUserGet(t *testing.T) {
	var err error
	var user rest.User

	// Expect an error
	_, err = da.UserGet(ctx, "")
	expectErr(t, err, errs.ErrEmptyUserName)

	// Expect an error
	_, err = da.UserGet(ctx, "test-get")
	expectErr(t, err, errs.ErrNoSuchUser)

	err = da.UserCreate(ctx, rest.User{Username: "test-get", Email: "test-get@foo.com"})
	defer da.UserDelete(ctx, "test-get")
	assert.NoError(t, err)

	// da.User should ctx, exist now
	exists, _ := da.UserExists(ctx, "test-get")
	if !exists {
		t.Error("User should exist now")
		t.FailNow()
	}

	// Expect no error
	user, err = da.UserGet(ctx, "test-get")
	assert.NoError(t, err)
	if user.Username != "test-get" {
		t.Errorf("User name mismatch: %q is not \"test-get\"", user.Username)
		t.FailNow()
	}
}

func testUserList(t *testing.T) {
	da.UserCreate(ctx, rest.User{Username: "test-list-0", Password: "password0!", Email: "test-list-0"})
	defer da.UserDelete(ctx, "test-list-0")
	da.UserCreate(ctx, rest.User{Username: "test-list-1", Password: "password1!", Email: "test-list-1"})
	defer da.UserDelete(ctx, "test-list-1")
	da.UserCreate(ctx, rest.User{Username: "test-list-2", Password: "password2!", Email: "test-list-2"})
	defer da.UserDelete(ctx, "test-list-2")
	da.UserCreate(ctx, rest.User{Username: "test-list-3", Password: "password3!", Email: "test-list-3"})
	defer da.UserDelete(ctx, "test-list-3")

	users, err := da.UserList(ctx)
	assert.NoError(t, err)

	if len(users) != 4 {
		for i, u := range users {
			t.Logf("User %d: %v\n", i+1, u)
		}

		t.Errorf("Expected len(users) = 4; got %d", len(users))
		t.FailNow()
	}

	for _, u := range users {
		if u.Password != "" {
			t.Error("Expected empty password")
			t.FailNow()
		}

		if u.Username == "" {
			t.Error("Expected non-empty username")
			t.FailNow()
		}
	}
}

func testUserUpdate(t *testing.T) {
	// Update blank user
	err := da.UserUpdate(ctx, rest.User{})
	expectErr(t, err, errs.ErrEmptyUserName)

	// Update user that doesn't exist
	err = da.UserUpdate(ctx, rest.User{Username: "no-such-user"})
	expectErr(t, err, errs.ErrNoSuchUser)

	userA := rest.User{Username: "test-update", Email: "foo1.example.com"}
	da.UserCreate(ctx, userA)
	defer da.UserDelete(ctx, "test-update")

	// Get the user we just added. Emails should match.
	user1, _ := da.UserGet(ctx, "test-update")
	if userA.Email != user1.Email {
		t.Errorf("Email mistatch: %q vs %q", userA.Email, user1.Email)
		t.FailNow()
	}

	// Do the update
	userB := rest.User{Username: "test-update", Email: "foo2.example.com"}
	err = da.UserUpdate(ctx, userB)
	assert.NoError(t, err)

	// Get the user we just updated. Emails should match.
	user2, _ := da.UserGet(ctx, "test-update")
	if userB.Email != user2.Email {
		t.Errorf("Email mistatch: %q vs %q", userB.Email, user2.Email)
		t.FailNow()
	}
}
