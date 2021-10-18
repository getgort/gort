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
	"github.com/stretchr/testify/require"
)

func testUserAccess(t *testing.T) {
	t.Run("testUserAuthenticate", testUserAuthenticate)
	t.Run("testUserCreate", testUserCreate)
	t.Run("testUserDelete", testUserDelete)
	t.Run("testUserExists", testUserExists)
	t.Run("testUserGet", testUserGet)
	t.Run("testUserGetByEmail", testUserGetByEmail)
	t.Run("testUserGetByID", testUserGetByID)
	t.Run("testUserGetNoMappings", testUserGetNoMappings)
	t.Run("testUserGroupList", testUserGroupList)
	t.Run("testUserList", testUserList)
	t.Run("testUserNotExists", testUserNotExists)
	t.Run("testUserPermissionList", testUserPermissionList)
	t.Run("testUserUpdate", testUserUpdate)
}

func testUserAuthenticate(t *testing.T) {
	var err error

	authenticated, err := da.UserAuthenticate(ctx, "test-auth", "no-match")
	assert.Error(t, err, errs.ErrNoSuchUser)
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

func testUserCreate(t *testing.T) {
	var err error
	var user rest.User

	// Expect an error
	err = da.UserCreate(ctx, user)
	assert.Error(t, err, errs.ErrEmptyUserName)

	// Expect no error
	err = da.UserCreate(ctx, rest.User{Username: "test-create", Email: "test-create@bar.com"})
	defer da.UserDelete(ctx, "test-create")
	assert.NoError(t, err)

	// Expect an error
	err = da.UserCreate(ctx, rest.User{Username: "test-create", Email: "test-create@bar.com"})
	assert.Error(t, err, errs.ErrUserExists)
}

func testUserDelete(t *testing.T) {
	// Delete blank user
	err := da.UserDelete(ctx, "")
	assert.Error(t, err, errs.ErrEmptyUserName)

	// Delete admin user
	err = da.UserDelete(ctx, "admin")
	assert.Error(t, err, errs.ErrAdminUndeletable)

	// Delete user that doesn't exist
	err = da.UserDelete(ctx, "no-such-user")
	assert.Error(t, err, errs.ErrNoSuchUser)

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

func testUserGet(t *testing.T) {
	const userName = "test-get"
	const userEmail = "test-get@foo.com"
	const userAdapter = "slack-get"
	const userAdapterID = "U12345-get"

	var err error
	var user rest.User

	// Expect an error
	_, err = da.UserGet(ctx, "")
	assert.EqualError(t, err, errs.ErrEmptyUserName.Error())

	// Expect an error
	_, err = da.UserGet(ctx, userName)
	assert.EqualError(t, err, errs.ErrNoSuchUser.Error())

	// Create the test user
	err = da.UserCreate(ctx, rest.User{
		Username: userName,
		Email:    userEmail,
		Mappings: map[string]string{userAdapter: userAdapterID},
	})
	defer da.UserDelete(ctx, userName)
	require.NoError(t, err)

	// User should exist now
	exists, err := da.UserExists(ctx, userName)
	require.NoError(t, err)
	require.True(t, exists)

	// Expect no error
	user, err = da.UserGet(ctx, userName)
	require.NoError(t, err)
	require.Equal(t, user.Username, userName)
	require.Equal(t, user.Email, userEmail)
	require.NotNil(t, user.Mappings)
	require.Equal(t, userAdapterID, user.Mappings[userAdapter])
}

func testUserGetNoMappings(t *testing.T) {
	const userName = "test-get-no-mappings"
	const userEmail = "test-get-no-mappings@foo.com"

	var err error
	var user rest.User

	// Create the test user
	err = da.UserCreate(ctx, rest.User{Username: userName, Email: userEmail})
	defer da.UserDelete(ctx, userName)
	require.NoError(t, err)

	// Expect no error
	user, err = da.UserGet(ctx, userName)
	require.NoError(t, err)
	require.Equal(t, user.Username, userName)
	require.Equal(t, user.Email, userEmail)
	require.NotNil(t, user.Mappings)
}

func testUserGetByEmail(t *testing.T) {
	const userName = "test-get-by-email"
	const userEmail = "test-get-by-email@foo.com"
	const userAdapter = "slack-get-by-email"
	const userAdapterID = "U12345-get-by-email"

	var err error
	var user rest.User

	// Expect an error
	_, err = da.UserGetByEmail(ctx, "")
	assert.EqualError(t, err, errs.ErrEmptyUserEmail.Error())

	// Expect an error
	_, err = da.UserGetByEmail(ctx, userEmail)
	assert.EqualError(t, err, errs.ErrNoSuchUser.Error())

	// Create the test user
	err = da.UserCreate(ctx, rest.User{
		Username: userName,
		Email:    userEmail,
		Mappings: map[string]string{userAdapter: userAdapterID},
	})
	defer da.UserDelete(ctx, userName)
	require.NoError(t, err)

	// User should exist now
	exists, err := da.UserExists(ctx, userName)
	require.NoError(t, err)
	require.True(t, exists)

	// Expect no error
	user, err = da.UserGetByEmail(ctx, userEmail)
	require.NoError(t, err)
	require.Equal(t, user.Username, userName)
	require.Equal(t, user.Email, userEmail)
	require.NotNil(t, user.Mappings)
	require.Equal(t, userAdapterID, user.Mappings[userAdapter])
}

func testUserGetByID(t *testing.T) {
	const userName = "test-get-by-id"
	const userEmail = "test-get-by-id@foo.com"
	const userAdapter = "slack-get-by-id"
	const userAdapterID = "U12345-get-by-id"

	var err error
	var user rest.User

	// Expect errors
	_, err = da.UserGetByID(ctx, "", userAdapterID)
	assert.EqualError(t, err, errs.ErrEmptyUserAdapter.Error())

	_, err = da.UserGetByID(ctx, userAdapter, "")
	assert.EqualError(t, err, errs.ErrEmptyUserID.Error())

	_, err = da.UserGetByID(ctx, userAdapter, userAdapterID)
	assert.EqualError(t, err, errs.ErrNoSuchUser.Error())

	// Create the test user
	err = da.UserCreate(ctx, rest.User{
		Username: userName,
		Email:    userEmail,
		Mappings: map[string]string{userAdapter: userAdapterID},
	})
	defer da.UserDelete(ctx, userName)
	require.NoError(t, err)

	// User should exist now
	exists, err := da.UserExists(ctx, userName)
	require.NoError(t, err)
	require.True(t, exists)

	// Expect no error
	user, err = da.UserGetByID(ctx, userAdapter, userAdapterID)
	require.NoError(t, err)
	require.Equal(t, user.Username, userName)
	require.Equal(t, user.Email, userEmail)
	require.NotNil(t, user.Mappings)
	require.Equal(t, userAdapterID, user.Mappings[userAdapter])
}

func testUserGroupList(t *testing.T) {
	da.GroupCreate(ctx, rest.Group{Name: "group-test-user-group-list-0"})
	defer da.GroupDelete(ctx, "group-test-user-group-list-0")

	da.GroupCreate(ctx, rest.Group{Name: "group-test-user-group-list-1"})
	defer da.GroupDelete(ctx, "group-test-user-group-list-1")

	da.UserCreate(ctx, rest.User{Username: "user-test-user-group-list"})
	defer da.UserDelete(ctx, "user-test-user-group-list")

	da.GroupUserAdd(ctx, "group-test-user-group-list-0", "user-test-user-group-list")

	expected := []rest.Group{{Name: "group-test-user-group-list-0", Users: nil}}

	actual, err := da.UserGroupList(ctx, "user-test-user-group-list")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, expected, actual)
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

func testUserPermissionList(t *testing.T) {
	var err error

	err = da.GroupCreate(ctx, rest.Group{Name: "test-perms"})
	defer da.GroupDelete(ctx, "test-perms")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	err = da.UserCreate(ctx, rest.User{Username: "test-perms", Password: "password0!", Email: "test-perms"})
	defer da.UserDelete(ctx, "test-perms")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	err = da.GroupUserAdd(ctx, "test-perms", "test-perms")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	da.RoleCreate(ctx, "test-perms")
	defer da.RoleDelete(ctx, "test-perms")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	err = da.GroupRoleAdd(ctx, "test-perms", "test-perms")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	err = da.RolePermissionAdd(ctx, "test-perms", "test", "test-perms-1")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	err = da.RolePermissionAdd(ctx, "test-perms", "test", "test-perms-2")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	err = da.RolePermissionAdd(ctx, "test-perms", "test", "test-perms-0")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// Expected: a sorted list of strings
	expected := []string{"test:test-perms-0", "test:test-perms-1", "test:test-perms-2"}

	actual, err := da.UserPermissionList(ctx, "test-perms")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, expected, actual.Strings())
}

func testUserUpdate(t *testing.T) {
	// Update blank user
	err := da.UserUpdate(ctx, rest.User{})
	assert.Error(t, err, errs.ErrEmptyUserName)

	// Update user that doesn't exist
	err = da.UserUpdate(ctx, rest.User{Username: "no-such-user"})
	assert.Error(t, err, errs.ErrNoSuchUser)

	userA := rest.User{Username: "test-update", Email: "foo1.example.com"}
	da.UserCreate(ctx, userA)
	defer da.UserDelete(ctx, "test-update")

	// Get the user we just added. Emails should match.
	user1, _ := da.UserGet(ctx, "test-update")
	if userA.Email != user1.Email {
		t.Errorf("Email mismatch: %q vs %q", userA.Email, user1.Email)
		t.FailNow()
	}

	// Do the update
	userB := rest.User{Username: "test-update", Email: "foo2.example.com"}
	err = da.UserUpdate(ctx, userB)
	assert.NoError(t, err)

	// Get the user we just updated. Emails should match.
	user2, _ := da.UserGet(ctx, "test-update")
	if userB.Email != user2.Email {
		t.Errorf("Email mismatch: %q vs %q", userB.Email, user2.Email)
		t.FailNow()
	}
}
