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

func (da DataAccessTester) testGroupAccess(t *testing.T) {
	t.Run("testGroupUserAdd", da.testGroupUserAdd)
	t.Run("testGroupUserList", da.testGroupUserList)
	t.Run("testGroupCreate", da.testGroupCreate)
	t.Run("testGroupDelete", da.testGroupDelete)
	t.Run("testGroupExists", da.testGroupExists)
	t.Run("testGroupGet", da.testGroupGet)
	t.Run("testGroupRoleAdd", da.testGroupRoleAdd)
	t.Run("testGroupPermissionList", da.testGroupPermissionList)
	t.Run("testGroupList", da.testGroupList)
	t.Run("testGroupRoleList", da.testGroupRoleList)
	t.Run("testGroupUserDelete", da.testGroupUserDelete)
}

func (da DataAccessTester) testGroupUserAdd(t *testing.T) {
	var (
		groupname = "group-test-group-user-add"
		username  = "user-test-group-user-add"
		useremail = "user@foo.bar"
	)

	err := da.GroupUserAdd(da.ctx, groupname, username)
	assert.Error(t, err, errs.ErrNoSuchGroup)

	da.GroupCreate(da.ctx, rest.Group{Name: groupname})
	defer da.GroupDelete(da.ctx, groupname)

	err = da.GroupUserAdd(da.ctx, groupname, username)
	assert.Error(t, err, errs.ErrNoSuchUser)

	da.UserCreate(da.ctx, rest.User{Username: username, Email: useremail})
	defer da.UserDelete(da.ctx, username)

	err = da.GroupUserAdd(da.ctx, groupname, username)
	assert.NoError(t, err)

	group, _ := da.GroupGet(da.ctx, groupname)
	require.Len(t, group.Users, 1)

	assert.Equal(t, group.Users[0].Username, username)
	assert.Equal(t, group.Users[0].Email, useremail)
}

func (da DataAccessTester) testGroupUserList(t *testing.T) {
	var (
		groupname = "group-test-group-user-list"
		expected  = []rest.User{
			{Username: "user-test-group-user-list-0", Email: "user-test-group-user-list-0@email.com", Mappings: map[string]string{}},
			{Username: "user-test-group-user-list-1", Email: "user-test-group-user-list-1@email.com", Mappings: map[string]string{}},
		}
	)

	_, err := da.GroupUserList(da.ctx, groupname)
	assert.Error(t, err)
	assert.ErrorIs(t, err, errs.ErrNoSuchGroup)

	da.GroupCreate(da.ctx, rest.Group{Name: groupname})
	defer da.GroupDelete(da.ctx, groupname)

	da.UserCreate(da.ctx, expected[0])
	defer da.UserDelete(da.ctx, expected[0].Username)
	da.UserCreate(da.ctx, expected[1])
	defer da.UserDelete(da.ctx, expected[1].Username)

	da.GroupUserAdd(da.ctx, groupname, expected[0].Username)
	da.GroupUserAdd(da.ctx, groupname, expected[1].Username)

	actual, err := da.GroupUserList(da.ctx, groupname)
	require.NoError(t, err)
	require.Len(t, actual, 2)

	assert.Equal(t, expected, actual)
}

func (da DataAccessTester) testGroupCreate(t *testing.T) {
	var err error
	var group rest.Group

	// Expect an error
	err = da.GroupCreate(da.ctx, group)
	assert.Error(t, err, errs.ErrEmptyGroupName)

	// Expect no error
	err = da.GroupCreate(da.ctx, rest.Group{Name: "test-create"})
	defer da.GroupDelete(da.ctx, "test-create")
	assert.NoError(t, err)

	// Expect an error
	err = da.GroupCreate(da.ctx, rest.Group{Name: "test-create"})
	assert.Error(t, err, errs.ErrGroupExists)
}

func (da DataAccessTester) testGroupDelete(t *testing.T) {
	// Delete blank group
	err := da.GroupDelete(da.ctx, "")
	assert.Error(t, err, errs.ErrEmptyGroupName)

	// Delete group that doesn't exist
	err = da.GroupDelete(da.ctx, "no-such-group")
	assert.Error(t, err, errs.ErrNoSuchGroup)

	da.GroupCreate(da.ctx, rest.Group{Name: "test-delete"}) // This has its own test
	defer da.GroupDelete(da.ctx, "test-delete")

	err = da.GroupDelete(da.ctx, "test-delete")
	assert.NoError(t, err)

	exists, _ := da.GroupExists(da.ctx, "test-delete")
	if exists {
		t.Error("Shouldn't exist anymore!")
		t.FailNow()
	}
}

func (da DataAccessTester) testGroupExists(t *testing.T) {
	var exists bool

	exists, _ = da.GroupExists(da.ctx, "test-exists")
	if exists {
		t.Error("Group should not exist now")
		t.FailNow()
	}

	// Now we add a group to find.
	da.GroupCreate(da.ctx, rest.Group{Name: "test-exists"})
	defer da.GroupDelete(da.ctx, "test-exists")

	exists, _ = da.GroupExists(da.ctx, "test-exists")
	if !exists {
		t.Error("Group should exist now")
		t.FailNow()
	}
}

func (da DataAccessTester) testGroupGet(t *testing.T) {
	groupname := "group-test-group-get"

	var err error
	var group rest.Group

	// Expect an error
	_, err = da.GroupGet(da.ctx, "")
	assert.ErrorIs(t, err, errs.ErrEmptyGroupName)

	// Expect an error
	_, err = da.GroupGet(da.ctx, groupname)
	assert.ErrorIs(t, err, errs.ErrNoSuchGroup)

	da.GroupCreate(da.ctx, rest.Group{Name: groupname})
	defer da.GroupDelete(da.ctx, groupname)

	// da.Group ctx, should exist now
	exists, _ := da.GroupExists(da.ctx, groupname)
	require.True(t, exists)

	// Expect no error
	group, err = da.GroupGet(da.ctx, groupname)
	require.NoError(t, err)
	require.Equal(t, groupname, group.Name)
}

func (da DataAccessTester) testGroupPermissionList(t *testing.T) {
	const (
		groupname  = "group-test-group-permission-list"
		rolename   = "role-test-group-permission-list"
		bundlename = "test"
	)

	var expected = rest.RolePermissionList{
		{BundleName: bundlename, Permission: "role-test-group-permission-list-1"},
		{BundleName: bundlename, Permission: "role-test-group-permission-list-2"},
		{BundleName: bundlename, Permission: "role-test-group-permission-list-3"},
	}

	var err error

	da.GroupCreate(da.ctx, rest.Group{Name: groupname})
	defer da.GroupDelete(da.ctx, groupname)

	da.RoleCreate(da.ctx, rolename)
	defer da.RoleDelete(da.ctx, rolename)

	err = da.GroupRoleAdd(da.ctx, groupname, rolename)
	require.NoError(t, err)

	da.RolePermissionAdd(da.ctx, rolename, expected[0].BundleName, expected[0].Permission)
	da.RolePermissionAdd(da.ctx, rolename, expected[1].BundleName, expected[1].Permission)
	da.RolePermissionAdd(da.ctx, rolename, expected[2].BundleName, expected[2].Permission)

	actual, err := da.GroupPermissionList(da.ctx, groupname)
	require.NoError(t, err)

	assert.Equal(t, expected, actual)
}

func (da DataAccessTester) testGroupRoleAdd(t *testing.T) {
	var err error

	groupName := "group-group-grant-role"
	roleName := "role-group-grant-role"
	bundleName := "bundle-group-grant-role"
	permissionName := "perm-group-grant-role"

	da.GroupCreate(da.ctx, rest.Group{Name: groupName})
	defer da.GroupDelete(da.ctx, groupName)

	err = da.RoleCreate(da.ctx, roleName)
	require.NoError(t, err)
	defer da.RoleDelete(da.ctx, roleName)

	err = da.RolePermissionAdd(da.ctx, roleName, bundleName, permissionName)
	require.NoError(t, err)

	err = da.GroupRoleAdd(da.ctx, groupName, roleName)
	require.NoError(t, err)

	expectedRoles := []rest.Role{
		{
			Name:        roleName,
			Permissions: []rest.RolePermission{{BundleName: bundleName, Permission: permissionName}},
		},
	}

	roles, err := da.GroupRoleList(da.ctx, groupName)
	require.NoError(t, err)

	assert.Equal(t, expectedRoles, roles)

	err = da.GroupRoleDelete(da.ctx, groupName, roleName)
	require.NoError(t, err)

	expectedRoles = []rest.Role{}

	roles, err = da.GroupRoleList(da.ctx, groupName)
	require.NoError(t, err)

	assert.Equal(t, expectedRoles, roles)
}

func (da DataAccessTester) testGroupList(t *testing.T) {
	da.GroupCreate(da.ctx, rest.Group{Name: "test-list-0"})
	defer da.GroupDelete(da.ctx, "test-list-0")
	da.GroupCreate(da.ctx, rest.Group{Name: "test-list-1"})
	defer da.GroupDelete(da.ctx, "test-list-1")
	da.GroupCreate(da.ctx, rest.Group{Name: "test-list-2"})
	defer da.GroupDelete(da.ctx, "test-list-2")
	da.GroupCreate(da.ctx, rest.Group{Name: "test-list-3"})
	defer da.GroupDelete(da.ctx, "test-list-3")

	groups, err := da.GroupList(da.ctx)
	assert.NoError(t, err)

	if len(groups) != 4 {
		t.Errorf("Expected len(groups) = 4; got %d", len(groups))
		t.FailNow()
	}

	for _, u := range groups {
		if u.Name == "" {
			t.Error("Expected non-empty name")
			t.FailNow()
		}
	}
}

func (da DataAccessTester) testGroupRoleList(t *testing.T) {
	var (
		groupname = "group-test-group-list-roles"
		rolenames = []string{
			"role-test-group-list-roles-0",
			"role-test-group-list-roles-1",
			"role-test-group-list-roles-2",
		}
	)
	da.GroupCreate(da.ctx, rest.Group{Name: groupname})
	defer da.GroupDelete(da.ctx, groupname)

	da.RoleCreate(da.ctx, rolenames[1])
	defer da.RoleDelete(da.ctx, rolenames[1])

	da.RoleCreate(da.ctx, rolenames[0])
	defer da.RoleDelete(da.ctx, rolenames[0])

	da.RoleCreate(da.ctx, rolenames[2])
	defer da.RoleDelete(da.ctx, rolenames[2])

	roles, err := da.GroupRoleList(da.ctx, groupname)
	assert.NoError(t, err)
	require.Empty(t, roles)

	err = da.GroupRoleAdd(da.ctx, groupname, rolenames[1])
	require.NoError(t, err)
	err = da.GroupRoleAdd(da.ctx, groupname, rolenames[0])
	require.NoError(t, err)
	err = da.GroupRoleAdd(da.ctx, groupname, rolenames[2])
	require.NoError(t, err)

	// Note: alphabetically sorted!
	expected := []rest.Role{
		{Name: rolenames[0], Permissions: []rest.RolePermission{}},
		{Name: rolenames[1], Permissions: []rest.RolePermission{}},
		{Name: rolenames[2], Permissions: []rest.RolePermission{}},
	}

	actual, err := da.GroupRoleList(da.ctx, groupname)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func (da DataAccessTester) testGroupUserDelete(t *testing.T) {
	da.GroupCreate(da.ctx, rest.Group{Name: "foo"})
	defer da.GroupDelete(da.ctx, "foo")

	da.UserCreate(da.ctx, rest.User{Username: "bat"})
	defer da.UserDelete(da.ctx, "bat")

	err := da.GroupUserAdd(da.ctx, "foo", "bat")
	assert.NoError(t, err)

	group, err := da.GroupGet(da.ctx, "foo")
	assert.NoError(t, err)

	if len(group.Users) != 1 {
		t.Error("Users list empty")
		t.FailNow()
	}

	if len(group.Users) > 0 && group.Users[0].Username != "bat" {
		t.Error("Wrong user!")
		t.FailNow()
	}

	err = da.GroupUserDelete(da.ctx, "foo", "bat")
	assert.NoError(t, err)

	group, err = da.GroupGet(da.ctx, "foo")
	assert.NoError(t, err)

	if len(group.Users) != 0 {
		t.Error("User not removed")
		t.FailNow()
	}
}
