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

func testGroupAccess(t *testing.T) {
	t.Run("testGroupUserAdd", testGroupUserAdd)
	t.Run("testGroupUserList", testGroupUserList)
	t.Run("testGroupCreate", testGroupCreate)
	t.Run("testGroupDelete", testGroupDelete)
	t.Run("testGroupExists", testGroupExists)
	t.Run("testGroupGet", testGroupGet)
	t.Run("testGroupRoleAdd", testGroupRoleAdd)
	t.Run("testGroupPermissionList", testGroupPermissionList)
	t.Run("testGroupList", testGroupList)
	t.Run("testGroupRoleList", testGroupRoleList)
	t.Run("testGroupUserDelete", testGroupUserDelete)
}

func testGroupUserAdd(t *testing.T) {
	var (
		groupname = "group-test-group-user-add"
		username  = "user-test-group-user-add"
		useremail = "user@foo.bar"
	)

	err := da.GroupUserAdd(ctx, groupname, username)
	assert.Error(t, err, errs.ErrNoSuchGroup)

	da.GroupCreate(ctx, rest.Group{Name: groupname})
	defer da.GroupDelete(ctx, groupname)

	err = da.GroupUserAdd(ctx, groupname, username)
	assert.Error(t, err, errs.ErrNoSuchUser)

	da.UserCreate(ctx, rest.User{Username: username, Email: useremail})
	defer da.UserDelete(ctx, username)

	err = da.GroupUserAdd(ctx, groupname, username)
	assert.NoError(t, err)

	group, _ := da.GroupGet(ctx, groupname)

	if !assert.Len(t, group.Users, 1) {
		t.FailNow()
	}

	assert.Equal(t, group.Users[0].Username, username)
	assert.Equal(t, group.Users[0].Email, useremail)
}

func testGroupUserList(t *testing.T) {
	var (
		groupname = "group-test-group-user-list"
		expected  = []rest.User{
			{Username: "user-test-group-user-list-0", Email: "user-test-group-user-list-0@email.com", Mappings: map[string]string{}},
			{Username: "user-test-group-user-list-1", Email: "user-test-group-user-list-1@email.com", Mappings: map[string]string{}},
		}
	)

	_, err := da.GroupUserList(ctx, groupname)
	assert.Error(t, err)
	assert.ErrorIs(t, err, errs.ErrNoSuchGroup)

	da.GroupCreate(ctx, rest.Group{Name: groupname})
	defer da.GroupDelete(ctx, groupname)

	da.UserCreate(ctx, expected[0])
	defer da.UserDelete(ctx, expected[0].Username)
	da.UserCreate(ctx, expected[1])
	defer da.UserDelete(ctx, expected[1].Username)

	da.GroupUserAdd(ctx, groupname, expected[0].Username)
	da.GroupUserAdd(ctx, groupname, expected[1].Username)

	actual, err := da.GroupUserList(ctx, groupname)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.Len(t, actual, 2) {
		t.FailNow()
	}

	assert.Equal(t, expected, actual)
}

func testGroupCreate(t *testing.T) {
	var err error
	var group rest.Group

	// Expect an error
	err = da.GroupCreate(ctx, group)
	assert.Error(t, err, errs.ErrEmptyGroupName)

	// Expect no error
	err = da.GroupCreate(ctx, rest.Group{Name: "test-create"})
	defer da.GroupDelete(ctx, "test-create")
	assert.NoError(t, err)

	// Expect an error
	err = da.GroupCreate(ctx, rest.Group{Name: "test-create"})
	assert.Error(t, err, errs.ErrGroupExists)
}

func testGroupDelete(t *testing.T) {
	// Delete blank group
	err := da.GroupDelete(ctx, "")
	assert.Error(t, err, errs.ErrEmptyGroupName)

	// Delete group that doesn't exist
	err = da.GroupDelete(ctx, "no-such-group")
	assert.Error(t, err, errs.ErrNoSuchGroup)

	da.GroupCreate(ctx, rest.Group{Name: "test-delete"}) // This has its own test
	defer da.GroupDelete(ctx, "test-delete")

	err = da.GroupDelete(ctx, "test-delete")
	assert.NoError(t, err)

	exists, _ := da.GroupExists(ctx, "test-delete")
	if exists {
		t.Error("Shouldn't exist anymore!")
		t.FailNow()
	}
}

func testGroupExists(t *testing.T) {
	var exists bool

	exists, _ = da.GroupExists(ctx, "test-exists")
	if exists {
		t.Error("Group should not exist now")
		t.FailNow()
	}

	// Now we add a group to find.
	da.GroupCreate(ctx, rest.Group{Name: "test-exists"})
	defer da.GroupDelete(ctx, "test-exists")

	exists, _ = da.GroupExists(ctx, "test-exists")
	if !exists {
		t.Error("Group should exist now")
		t.FailNow()
	}
}

func testGroupGet(t *testing.T) {
	groupname := "group-test-group-get"

	var err error
	var group rest.Group

	// Expect an error
	_, err = da.GroupGet(ctx, "")
	assert.ErrorIs(t, err, errs.ErrEmptyGroupName)

	// Expect an error
	_, err = da.GroupGet(ctx, groupname)
	assert.ErrorIs(t, err, errs.ErrNoSuchGroup)

	da.GroupCreate(ctx, rest.Group{Name: groupname})
	defer da.GroupDelete(ctx, groupname)

	// da.Group ctx, should exist now
	exists, _ := da.GroupExists(ctx, groupname)
	if !assert.True(t, exists) {
		t.FailNow()
	}

	// Expect no error
	group, err = da.GroupGet(ctx, groupname)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.Equal(t, groupname, group.Name) {
		t.FailNow()
	}
}

func testGroupPermissionList(t *testing.T) {
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

	da.GroupCreate(ctx, rest.Group{Name: groupname})
	defer da.GroupDelete(ctx, groupname)

	da.RoleCreate(ctx, rolename)
	defer da.RoleDelete(ctx, rolename)

	err = da.GroupRoleAdd(ctx, groupname, rolename)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	da.RolePermissionAdd(ctx, rolename, expected[0].BundleName, expected[0].Permission)
	da.RolePermissionAdd(ctx, rolename, expected[1].BundleName, expected[1].Permission)
	da.RolePermissionAdd(ctx, rolename, expected[2].BundleName, expected[2].Permission)

	actual, err := da.GroupPermissionList(ctx, groupname)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, expected, actual)
}

func testGroupRoleAdd(t *testing.T) {
	var err error

	groupName := "group-group-grant-role"
	roleName := "role-group-grant-role"
	bundleName := "bundle-group-grant-role"
	permissionName := "perm-group-grant-role"

	da.GroupCreate(ctx, rest.Group{Name: groupName})
	defer da.GroupDelete(ctx, groupName)

	err = da.RoleCreate(ctx, roleName)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer da.RoleDelete(ctx, roleName)

	err = da.RolePermissionAdd(ctx, roleName, bundleName, permissionName)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	err = da.GroupRoleAdd(ctx, groupName, roleName)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	expectedRoles := []rest.Role{
		{
			Name:        roleName,
			Permissions: []rest.RolePermission{{BundleName: bundleName, Permission: permissionName}},
		},
	}

	roles, err := da.GroupRoleList(ctx, groupName)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, expectedRoles, roles)

	err = da.GroupRoleDelete(ctx, groupName, roleName)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	expectedRoles = []rest.Role{}

	roles, err = da.GroupRoleList(ctx, groupName)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, expectedRoles, roles)
}

func testGroupList(t *testing.T) {
	da.GroupCreate(ctx, rest.Group{Name: "test-list-0"})
	defer da.GroupDelete(ctx, "test-list-0")
	da.GroupCreate(ctx, rest.Group{Name: "test-list-1"})
	defer da.GroupDelete(ctx, "test-list-1")
	da.GroupCreate(ctx, rest.Group{Name: "test-list-2"})
	defer da.GroupDelete(ctx, "test-list-2")
	da.GroupCreate(ctx, rest.Group{Name: "test-list-3"})
	defer da.GroupDelete(ctx, "test-list-3")

	groups, err := da.GroupList(ctx)
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

func testGroupRoleList(t *testing.T) {
	var (
		groupname = "group-test-group-list-roles"
		rolenames = []string{
			"role-test-group-list-roles-0",
			"role-test-group-list-roles-1",
			"role-test-group-list-roles-2",
		}
	)
	da.GroupCreate(ctx, rest.Group{Name: groupname})
	defer da.GroupDelete(ctx, groupname)

	da.RoleCreate(ctx, rolenames[1])
	defer da.RoleDelete(ctx, rolenames[1])

	da.RoleCreate(ctx, rolenames[0])
	defer da.RoleDelete(ctx, rolenames[0])

	da.RoleCreate(ctx, rolenames[2])
	defer da.RoleDelete(ctx, rolenames[2])

	roles, err := da.GroupRoleList(ctx, groupname)
	if !assert.NoError(t, err) && !assert.Empty(t, roles) {
		t.FailNow()
	}

	err = da.GroupRoleAdd(ctx, groupname, rolenames[1])
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	err = da.GroupRoleAdd(ctx, groupname, rolenames[0])
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	err = da.GroupRoleAdd(ctx, groupname, rolenames[2])
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// Note: alphabetically sorted!
	expected := []rest.Role{
		{Name: rolenames[0], Permissions: []rest.RolePermission{}},
		{Name: rolenames[1], Permissions: []rest.RolePermission{}},
		{Name: rolenames[2], Permissions: []rest.RolePermission{}},
	}

	actual, err := da.GroupRoleList(ctx, groupname)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, expected, actual)
}

func testGroupUserDelete(t *testing.T) {
	da.GroupCreate(ctx, rest.Group{Name: "foo"})
	defer da.GroupDelete(ctx, "foo")

	da.UserCreate(ctx, rest.User{Username: "bat"})
	defer da.UserDelete(ctx, "bat")

	err := da.GroupUserAdd(ctx, "foo", "bat")
	assert.NoError(t, err)

	group, err := da.GroupGet(ctx, "foo")
	assert.NoError(t, err)

	if len(group.Users) != 1 {
		t.Error("Users list empty")
		t.FailNow()
	}

	if len(group.Users) > 0 && group.Users[0].Username != "bat" {
		t.Error("Wrong user!")
		t.FailNow()
	}

	err = da.GroupUserDelete(ctx, "foo", "bat")
	assert.NoError(t, err)

	group, err = da.GroupGet(ctx, "foo")
	assert.NoError(t, err)

	if len(group.Users) != 0 {
		t.Error("User not removed")
		t.FailNow()
	}
}
