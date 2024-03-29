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

package tests

import (
	"testing"

	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/dataaccess/errs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (da DataAccessTester) testUserAccess(t *testing.T) {
	t.Run("testUserAuthenticate", da.testUserAuthenticate)
	t.Run("testUserCreate", da.testUserCreate)
	t.Run("testUserDelete", da.testUserDelete)
	t.Run("testUserExists", da.testUserExists)
	t.Run("testUserGet", da.testUserGet)
	t.Run("testUserGetByEmail", da.testUserGetByEmail)
	t.Run("testUserGetByID", da.testUserGetByID)
	t.Run("testUserGetNoMappings", da.testUserGetNoMappings)
	t.Run("testUserGroupList", da.testUserGroupList)
	t.Run("testUserList", da.testUserList)
	t.Run("testUserNotExists", da.testUserNotExists)
	t.Run("testUserPermissionList", da.testUserPermissionList)
	t.Run("testUserUpdate", da.testUserUpdate)
}

func (da DataAccessTester) testUserAuthenticate(t *testing.T) {
	var err error

	authenticated, err := da.UserAuthenticate(da.ctx, "test-auth", "no-match")
	assert.Error(t, err, errs.ErrNoSuchUser)
	require.False(t, authenticated)

	// Expect no error
	err = da.UserCreate(da.ctx, rest.User{
		Username: "test-auth",
		Email:    "test-auth@bar.com",
		Password: "password",
	})
	defer da.UserDelete(da.ctx, "test-auth")
	assert.NoError(t, err)

	authenticated, err = da.UserAuthenticate(da.ctx, "test-auth", "no-match")
	assert.NoError(t, err)
	require.False(t, authenticated)

	authenticated, err = da.UserAuthenticate(da.ctx, "test-auth", "password")
	assert.NoError(t, err)
	require.True(t, authenticated)
}

func (da DataAccessTester) testUserCreate(t *testing.T) {
	var err error
	var user rest.User

	// Expect an error
	err = da.UserCreate(da.ctx, user)
	assert.Error(t, err, errs.ErrEmptyUserName)

	// Expect no error
	err = da.UserCreate(da.ctx, rest.User{Username: "test-create", Email: "test-create@bar.com"})
	defer da.UserDelete(da.ctx, "test-create")
	assert.NoError(t, err)

	// Expect an error
	err = da.UserCreate(da.ctx, rest.User{Username: "test-create", Email: "test-create@bar.com"})
	assert.Error(t, err, errs.ErrUserExists)
}

func (da DataAccessTester) testUserDelete(t *testing.T) {
	// Delete blank user
	err := da.UserDelete(da.ctx, "")
	assert.Error(t, err, errs.ErrEmptyUserName)

	// Delete admin user
	err = da.UserDelete(da.ctx, "admin")
	assert.Error(t, err, errs.ErrAdminUndeletable)

	// Delete user that doesn't exist
	err = da.UserDelete(da.ctx, "no-such-user")
	assert.Error(t, err, errs.ErrNoSuchUser)

	user := rest.User{Username: "test-delete", Email: "foo1.example.com"}
	da.UserCreate(da.ctx, user) // This has its own test
	defer da.UserDelete(da.ctx, "test-delete")

	err = da.UserDelete(da.ctx, "test-delete")
	assert.NoError(t, err)

	exists, _ := da.UserExists(da.ctx, "test-delete")
	require.False(t, exists)
}

func (da DataAccessTester) testUserExists(t *testing.T) {
	var exists bool

	exists, _ = da.UserExists(da.ctx, "test-exists")
	if exists {
		t.Error("User should not exist now")
		t.FailNow()
	}

	// Now we add a user to find.
	err := da.UserCreate(da.ctx, rest.User{Username: "test-exists", Email: "test-exists@bar.com"})
	defer da.UserDelete(da.ctx, "test-exists")
	assert.NoError(t, err)

	exists, _ = da.UserExists(da.ctx, "test-exists")
	require.True(t, exists)
}

func (da DataAccessTester) testUserGet(t *testing.T) {
	const userName = "test-get"
	const userEmail = "test-get@foo.com"
	const userAdapter = "slack-get"
	const userAdapterID = "U12345-get"

	var err error
	var user rest.User

	// Expect an error
	_, err = da.UserGet(da.ctx, "")
	assert.EqualError(t, err, errs.ErrEmptyUserName.Error())

	// Expect an error
	_, err = da.UserGet(da.ctx, userName)
	assert.EqualError(t, err, errs.ErrNoSuchUser.Error())

	// Create the test user
	err = da.UserCreate(da.ctx, rest.User{
		Username: userName,
		Email:    userEmail,
		Mappings: map[string]string{userAdapter: userAdapterID},
	})
	defer da.UserDelete(da.ctx, userName)
	require.NoError(t, err)

	// User should exist now
	exists, err := da.UserExists(da.ctx, userName)
	require.NoError(t, err)
	require.True(t, exists)

	// Expect no error
	user, err = da.UserGet(da.ctx, userName)
	require.NoError(t, err)
	require.Equal(t, user.Username, userName)
	require.Equal(t, user.Email, userEmail)
	require.NotNil(t, user.Mappings)
	require.Equal(t, userAdapterID, user.Mappings[userAdapter])
}

func (da DataAccessTester) testUserGetNoMappings(t *testing.T) {
	const userName = "test-get-no-mappings"
	const userEmail = "test-get-no-mappings@foo.com"

	var err error
	var user rest.User

	// Create the test user
	err = da.UserCreate(da.ctx, rest.User{Username: userName, Email: userEmail})
	defer da.UserDelete(da.ctx, userName)
	require.NoError(t, err)

	// Expect no error
	user, err = da.UserGet(da.ctx, userName)
	require.NoError(t, err)
	require.Equal(t, user.Username, userName)
	require.Equal(t, user.Email, userEmail)
	require.NotNil(t, user.Mappings)
}

func (da DataAccessTester) testUserGetByEmail(t *testing.T) {
	const userName = "test-get-by-email"
	const userEmail = "test-get-by-email@foo.com"
	const userAdapter = "slack-get-by-email"
	const userAdapterID = "U12345-get-by-email"

	var err error
	var user rest.User

	// Expect an error
	_, err = da.UserGetByEmail(da.ctx, "")
	assert.EqualError(t, err, errs.ErrEmptyUserEmail.Error())

	// Expect an error
	_, err = da.UserGetByEmail(da.ctx, userEmail)
	assert.EqualError(t, err, errs.ErrNoSuchUser.Error())

	// Create the test user
	err = da.UserCreate(da.ctx, rest.User{
		Username: userName,
		Email:    userEmail,
		Mappings: map[string]string{userAdapter: userAdapterID},
	})
	defer da.UserDelete(da.ctx, userName)
	require.NoError(t, err)

	// User should exist now
	exists, err := da.UserExists(da.ctx, userName)
	require.NoError(t, err)
	require.True(t, exists)

	// Expect no error
	user, err = da.UserGetByEmail(da.ctx, userEmail)
	require.NoError(t, err)
	require.Equal(t, user.Username, userName)
	require.Equal(t, user.Email, userEmail)
	require.NotNil(t, user.Mappings)
	require.Equal(t, userAdapterID, user.Mappings[userAdapter])
}

func (da DataAccessTester) testUserGetByID(t *testing.T) {
	const userName = "test-get-by-id"
	const userEmail = "test-get-by-id@foo.com"
	const userAdapter = "slack-get-by-id"
	const userAdapterID = "U12345-get-by-id"

	var err error
	var user rest.User

	// Expect errors
	_, err = da.UserGetByID(da.ctx, "", userAdapterID)
	assert.EqualError(t, err, errs.ErrEmptyUserAdapter.Error())

	_, err = da.UserGetByID(da.ctx, userAdapter, "")
	assert.EqualError(t, err, errs.ErrEmptyUserID.Error())

	_, err = da.UserGetByID(da.ctx, userAdapter, userAdapterID)
	assert.EqualError(t, err, errs.ErrNoSuchUser.Error())

	// Create the test user
	err = da.UserCreate(da.ctx, rest.User{
		Username: userName,
		Email:    userEmail,
		Mappings: map[string]string{userAdapter: userAdapterID},
	})
	defer da.UserDelete(da.ctx, userName)
	require.NoError(t, err)

	// User should exist now
	exists, err := da.UserExists(da.ctx, userName)
	require.NoError(t, err)
	require.True(t, exists)

	// Expect no error
	user, err = da.UserGetByID(da.ctx, userAdapter, userAdapterID)
	require.NoError(t, err)
	require.Equal(t, user.Username, userName)
	require.Equal(t, user.Email, userEmail)
	require.NotNil(t, user.Mappings)
	require.Equal(t, userAdapterID, user.Mappings[userAdapter])
}

func (da DataAccessTester) testUserGroupList(t *testing.T) {
	da.GroupCreate(da.ctx, rest.Group{Name: "group-test-user-group-list-0"})
	defer da.GroupDelete(da.ctx, "group-test-user-group-list-0")

	da.GroupCreate(da.ctx, rest.Group{Name: "group-test-user-group-list-1"})
	defer da.GroupDelete(da.ctx, "group-test-user-group-list-1")

	da.UserCreate(da.ctx, rest.User{Username: "user-test-user-group-list"})
	defer da.UserDelete(da.ctx, "user-test-user-group-list")

	da.GroupUserAdd(da.ctx, "group-test-user-group-list-0", "user-test-user-group-list")

	expected := []rest.Group{{Name: "group-test-user-group-list-0", Users: nil}}

	actual, err := da.UserGroupList(da.ctx, "user-test-user-group-list")
	require.NoError(t, err)

	assert.Equal(t, expected, actual)
}

func (da DataAccessTester) testUserList(t *testing.T) {
	da.UserCreate(da.ctx, rest.User{Username: "test-list-0", Password: "password0!", Email: "test-list-0"})
	defer da.UserDelete(da.ctx, "test-list-0")
	da.UserCreate(da.ctx, rest.User{Username: "test-list-1", Password: "password1!", Email: "test-list-1"})
	defer da.UserDelete(da.ctx, "test-list-1")
	da.UserCreate(da.ctx, rest.User{Username: "test-list-2", Password: "password2!", Email: "test-list-2"})
	defer da.UserDelete(da.ctx, "test-list-2")
	da.UserCreate(da.ctx, rest.User{Username: "test-list-3", Password: "password3!", Email: "test-list-3"})
	defer da.UserDelete(da.ctx, "test-list-3")

	users, err := da.UserList(da.ctx)
	assert.NoError(t, err)

	require.Len(t, users, 4)

	for _, u := range users {
		require.Empty(t, u.Password)
		require.NotEmpty(t, u.Username)
	}
}

func (da DataAccessTester) testUserNotExists(t *testing.T) {
	var exists bool

	err := da.Initialize(da.ctx)
	assert.NoError(t, err)

	exists, _ = da.UserExists(da.ctx, "test-not-exists")
	require.False(t, exists)
}

func (da DataAccessTester) testUserPermissionList(t *testing.T) {
	var err error

	err = da.GroupCreate(da.ctx, rest.Group{Name: "test-perms"})
	defer da.GroupDelete(da.ctx, "test-perms")
	require.NoError(t, err)

	err = da.UserCreate(da.ctx, rest.User{Username: "test-perms", Password: "password0!", Email: "test-perms"})
	defer da.UserDelete(da.ctx, "test-perms")
	require.NoError(t, err)
	err = da.GroupUserAdd(da.ctx, "test-perms", "test-perms")
	require.NoError(t, err)

	da.RoleCreate(da.ctx, "test-perms")
	defer da.RoleDelete(da.ctx, "test-perms")
	require.NoError(t, err)
	err = da.GroupRoleAdd(da.ctx, "test-perms", "test-perms")
	require.NoError(t, err)

	err = da.RolePermissionAdd(da.ctx, "test-perms", "test", "test-perms-1")
	require.NoError(t, err)
	err = da.RolePermissionAdd(da.ctx, "test-perms", "test", "test-perms-2")
	require.NoError(t, err)
	err = da.RolePermissionAdd(da.ctx, "test-perms", "test", "test-perms-0")
	require.NoError(t, err)

	// Expected: a sorted list of strings
	expected := []string{"test:test-perms-0", "test:test-perms-1", "test:test-perms-2"}

	actual, err := da.UserPermissionList(da.ctx, "test-perms")
	require.NoError(t, err)

	assert.Equal(t, expected, actual.Strings())
}

func (da DataAccessTester) testUserUpdate(t *testing.T) {
	// Update blank user
	err := da.UserUpdate(da.ctx, rest.User{})
	assert.Error(t, err, errs.ErrEmptyUserName)

	// Update user that doesn't exist
	err = da.UserUpdate(da.ctx, rest.User{Username: "no-such-user"})
	assert.Error(t, err, errs.ErrNoSuchUser)

	userA := rest.User{Username: "test-update", Email: "foo1.example.com"}
	da.UserCreate(da.ctx, userA)
	defer da.UserDelete(da.ctx, "test-update")

	// Get the user we just added. Emails should match.
	user1, _ := da.UserGet(da.ctx, "test-update")
	require.Equal(t, userA.Email, user1.Email)

	// Do the update
	userB := rest.User{Username: "test-update", Email: "foo2.example.com"}
	err = da.UserUpdate(da.ctx, userB)
	assert.NoError(t, err)

	// Get the user we just updated. Emails should match.
	user2, _ := da.UserGet(da.ctx, "test-update")
	require.Equal(t, userB.Email, user2.Email)
}
