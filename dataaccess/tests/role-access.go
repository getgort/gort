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
)

func (da DataAccessTester) testRoleAccess(t *testing.T) {
	t.Run("testRoleCreate", da.testRoleCreate)
	t.Run("testRoleList", da.testRoleList)
	t.Run("testRoleExists", da.testRoleExists)
	t.Run("testRoleDelete", da.testRoleDelete)
	t.Run("testRoleGet", da.testRoleGet)
	t.Run("testRoleGroupAdd", da.testRoleGroupAdd)
	t.Run("testRoleGroupDelete", da.testRoleGroupDelete)
	t.Run("testRoleGroupExists", da.testRoleGroupExists)
	t.Run("testRoleGroupList", da.testRoleGroupList)
	t.Run("testRolePermissionExists", da.testRolePermissionExists)
	t.Run("testRolePermissionAdd", da.testRolePermissionAdd)
	t.Run("testRolePermissionList", da.testRolePermissionList)
}

func (da DataAccessTester) testRoleCreate(t *testing.T) {
	var err error

	// Expect an error
	err = da.RoleCreate(da.ctx, "")
	assert.Error(t, err, errs.ErrEmptyRoleName)

	// Expect no error
	err = da.RoleCreate(da.ctx, "test-create")
	defer da.RoleDelete(da.ctx, "test-create")
	assert.NoError(t, err)

	// Expect an error
	err = da.RoleCreate(da.ctx, "test-create")
	assert.Error(t, err, errs.ErrRoleExists)
}

func (da DataAccessTester) testRoleList(t *testing.T) {
	var err error

	// Get initial set of roles
	roles, err := da.RoleList(da.ctx)
	assert.NoError(t, err)
	startingRoles := len(roles)

	// Create and populate role
	rolename := "test-role-list"
	bundle := "test-bundle-list"
	permission := "test-permission-list"
	err = da.RoleCreate(da.ctx, rolename)
	defer da.RoleDelete(da.ctx, rolename)
	assert.NoError(t, err)

	err = da.RolePermissionAdd(da.ctx, rolename, bundle, permission)
	assert.NoError(t, err)

	// Expect 1 new role
	roles, err = da.RoleList(da.ctx)
	assert.NoError(t, err)
	if assert.Equal(t, startingRoles+1, len(roles)) {
		assert.Equal(t, rolename, roles[startingRoles].Name)
		assert.Equal(t, bundle, roles[startingRoles].Permissions[0].BundleName)
		assert.Equal(t, permission, roles[startingRoles].Permissions[0].Permission)
	}
}

func (da DataAccessTester) testRoleDelete(t *testing.T) {
	// Delete blank group
	err := da.RoleDelete(da.ctx, "")
	assert.Error(t, err, errs.ErrEmptyRoleName)

	// Delete group that doesn't exist
	err = da.RoleDelete(da.ctx, "no-such-group")
	assert.Error(t, err, errs.ErrNoSuchRole)

	da.RoleCreate(da.ctx, "test-delete") // This has its own test
	defer da.RoleDelete(da.ctx, "test-delete")

	err = da.RoleDelete(da.ctx, "test-delete")
	assert.NoError(t, err)

	exists, _ := da.RoleExists(da.ctx, "test-delete")
	if exists {
		t.Error("Shouldn't exist anymore!")
		t.FailNow()
	}
}

func (da DataAccessTester) testRoleExists(t *testing.T) {
	var exists bool

	exists, _ = da.RoleExists(da.ctx, "test-exists")
	if exists {
		t.Error("Role should not exist now")
		t.FailNow()
	}

	// Now we add a group to find.
	da.RoleCreate(da.ctx, "test-exists")
	defer da.RoleDelete(da.ctx, "test-exists")

	exists, _ = da.RoleExists(da.ctx, "test-exists")
	if !exists {
		t.Error("Role should exist now")
		t.FailNow()
	}
}

func (da DataAccessTester) testRoleGet(t *testing.T) {
	var err error
	var role rest.Role

	// Expect an error
	_, err = da.RoleGet(da.ctx, "")
	assert.Error(t, err, errs.ErrEmptyRoleName)

	// Expect an error
	_, err = da.RoleGet(da.ctx, "test-get")
	assert.Error(t, err, errs.ErrNoSuchRole)

	da.RoleCreate(da.ctx, "test-get")
	defer da.RoleDelete(da.ctx, "test-get")

	// da.Role da.ctx, should exist now
	exists, _ := da.RoleExists(da.ctx, "test-get")
	if !exists {
		t.Error("Role should exist now")
		t.FailNow()
	}

	err = da.RolePermissionAdd(da.ctx, "test-get", "foo", "bar")
	assert.NoError(t, err)

	err = da.RolePermissionAdd(da.ctx, "test-get", "foo", "bat")
	assert.NoError(t, err)

	err = da.RolePermissionDelete(da.ctx, "test-get", "foo", "bat")
	assert.NoError(t, err)

	expected := rest.Role{
		Name:        "test-get",
		Permissions: []rest.RolePermission{{BundleName: "foo", Permission: "bar"}},
	}

	// Expect no error
	role, err = da.RoleGet(da.ctx, "test-get")
	assert.NoError(t, err)
	assert.Equal(t, expected, role)
}

func (da DataAccessTester) testRoleGroupAdd(t *testing.T) {
	var err error

	rolename := "role-test-role-group-add"
	groupnames := []string{
		"perm-test-role-group-add-0",
		"perm-test-role-group-add-1",
	}

	// No such group yet
	err = da.RoleGroupAdd(da.ctx, rolename, groupnames[1])
	assert.ErrorIs(t, err, errs.ErrNoSuchGroup)

	da.GroupCreate(da.ctx, rest.Group{Name: groupnames[0]})
	defer da.GroupDelete(da.ctx, groupnames[0])
	da.GroupCreate(da.ctx, rest.Group{Name: groupnames[1]})
	defer da.GroupDelete(da.ctx, groupnames[1])

	// Groups exist now, but the role doesn't
	err = da.RoleGroupAdd(da.ctx, rolename, groupnames[1])
	assert.ErrorIs(t, err, errs.ErrNoSuchRole)

	da.RoleCreate(da.ctx, rolename)
	defer da.RoleDelete(da.ctx, rolename)

	for _, groupname := range groupnames {
		err = da.RoleGroupAdd(da.ctx, rolename, groupname)
		assert.NoError(t, err)
	}

	for _, groupname := range groupnames {
		exists, _ := da.RoleGroupExists(da.ctx, rolename, groupname)
		assert.True(t, exists, groupname)
	}
}

func (da DataAccessTester) testRoleGroupDelete(t *testing.T) {

}

func (da DataAccessTester) testRoleGroupExists(t *testing.T) {
	var err error

	rolename := "role-test-role-group-exists"
	groupnames := []string{
		"group-test-role-group-exists-0",
		"group-test-role-group-exists-1",
	}
	groupnull := "group-test-role-group-exists-null"

	// No such role yet
	_, err = da.RoleGroupExists(da.ctx, rolename, groupnames[1])
	assert.ErrorIs(t, err, errs.ErrNoSuchRole)

	da.RoleCreate(da.ctx, rolename)
	defer da.RoleDelete(da.ctx, rolename)

	// Groups exist now, but the role doesn't
	_, err = da.RoleGroupExists(da.ctx, rolename, groupnames[1])
	assert.ErrorIs(t, err, errs.ErrNoSuchGroup)

	da.GroupCreate(da.ctx, rest.Group{Name: groupnames[0]})
	defer da.GroupDelete(da.ctx, groupnames[0])
	da.GroupCreate(da.ctx, rest.Group{Name: groupnames[1]})
	defer da.GroupDelete(da.ctx, groupnames[1])
	da.GroupCreate(da.ctx, rest.Group{Name: groupnull})
	defer da.GroupDelete(da.ctx, groupnull)

	for _, groupname := range groupnames {
		da.RoleGroupAdd(da.ctx, rolename, groupname)
	}

	for _, groupname := range groupnames {
		exists, err := da.RoleGroupExists(da.ctx, rolename, groupname)
		assert.NoError(t, err)
		assert.True(t, exists)
	}

	// Null group should NOT exist on the role
	exists, err := da.RoleGroupExists(da.ctx, rolename, groupnull)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func (da DataAccessTester) testRoleGroupList(t *testing.T) {
	var err error

	rolename := "role-test-role-group-list"
	groupnames := []string{
		"group-test-role-group-list-0",
		"group-test-role-group-list-1",
	}
	groupnull := "group-test-role-group-list-null"

	// No such role yet
	_, err = da.RoleGroupList(da.ctx, rolename)
	assert.ErrorIs(t, err, errs.ErrNoSuchRole)

	da.RoleCreate(da.ctx, rolename)
	defer da.RoleDelete(da.ctx, rolename)

	// Groups exist now, but the role doesn't
	groups, err := da.RoleGroupList(da.ctx, rolename)
	assert.NoError(t, err)
	assert.Empty(t, groups)

	da.GroupCreate(da.ctx, rest.Group{Name: groupnames[1]})
	defer da.GroupDelete(da.ctx, groupnames[1])
	da.GroupCreate(da.ctx, rest.Group{Name: groupnames[0]})
	defer da.GroupDelete(da.ctx, groupnames[0])
	da.GroupCreate(da.ctx, rest.Group{Name: groupnull})
	defer da.GroupDelete(da.ctx, groupnull)

	for _, groupname := range groupnames {
		da.RoleGroupAdd(da.ctx, rolename, groupname)
	}

	// Currently the groups are NOT expected to be fully described (i.e.,
	// their roles slices don't have to be complete).
	groups, err = da.RoleGroupList(da.ctx, rolename)
	assert.NoError(t, err)
	assert.Len(t, groups, 2)

	for i, g := range groups {
		assert.Equal(t, groupnames[i], g.Name)
	}
}

func (da DataAccessTester) testRolePermissionAdd(t *testing.T) {
	var exists bool
	var err error

	const rolename = "role-test-role-permission-add"
	const bundlename = "test"
	const permname1 = "perm-test-role-permission-add-0"
	const permname2 = "perm-test-role-permission-add-1"

	da.RoleCreate(da.ctx, rolename)
	defer da.RoleDelete(da.ctx, rolename)

	role, _ := da.RoleGet(da.ctx, rolename)
	if !assert.Len(t, role.Permissions, 0) {
		t.FailNow()
	}

	// First permission
	err = da.RolePermissionAdd(da.ctx, rolename, bundlename, permname1)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer da.RolePermissionDelete(da.ctx, rolename, bundlename, permname1)

	role, _ = da.RoleGet(da.ctx, rolename)
	if !assert.Len(t, role.Permissions, 1) {
		t.FailNow()
	}

	exists, _ = da.RolePermissionExists(da.ctx, rolename, bundlename, permname1)
	if !assert.True(t, exists) {
		t.FailNow()
	}

	// Second permission
	err = da.RolePermissionAdd(da.ctx, rolename, bundlename, permname2)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer da.RolePermissionDelete(da.ctx, rolename, bundlename, permname2)

	role, _ = da.RoleGet(da.ctx, rolename)
	if !assert.Len(t, role.Permissions, 2) {
		t.FailNow()
	}

	exists, _ = da.RolePermissionExists(da.ctx, rolename, bundlename, permname2)
	if !assert.True(t, exists) {
		t.FailNow()
	}
}

func (da DataAccessTester) testRolePermissionExists(t *testing.T) {
	var err error

	da.RoleCreate(da.ctx, "role-test-role-has-permission")
	defer da.RoleDelete(da.ctx, "role-test-role-has-permission")

	has, err := da.RolePermissionExists(da.ctx, "role-test-role-has-permission", "test", "permission-test-role-has-permission-1")
	if !assert.NoError(t, err) || !assert.False(t, has) {
		t.FailNow()
	}

	err = da.RolePermissionAdd(da.ctx, "role-test-role-has-permission", "test", "permission-test-role-has-permission-1")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer da.RolePermissionDelete(da.ctx, "role-test-role-has-permission", "test", "permission-test-role-has-permission-1")

	has, err = da.RolePermissionExists(da.ctx, "role-test-role-has-permission", "test", "permission-test-role-has-permission-1")
	if !assert.NoError(t, err) || !assert.True(t, has) {
		t.FailNow()
	}

	has, err = da.RolePermissionExists(da.ctx, "role-test-role-has-permission", "test", "permission-test-role-has-permission-2")
	if !assert.NoError(t, err) || !assert.False(t, has) {
		t.FailNow()
	}
}

func (da DataAccessTester) testRolePermissionList(t *testing.T) {
	var err error

	da.RoleCreate(da.ctx, "role-test-role-permission-list")
	defer da.RoleDelete(da.ctx, "role-test-role-permission-list")

	err = da.RolePermissionAdd(da.ctx, "role-test-role-permission-list", "test", "permission-test-role-permission-list-1")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer da.RolePermissionDelete(da.ctx, "role-test-role-permission-list", "test", "permission-test-role-permission-list-1")

	err = da.RolePermissionAdd(da.ctx, "role-test-role-permission-list", "test", "permission-test-role-permission-list-3")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer da.RolePermissionDelete(da.ctx, "role-test-role-permission-list", "test", "permission-test-role-permission-list-3")

	err = da.RolePermissionAdd(da.ctx, "role-test-role-permission-list", "test", "permission-test-role-permission-list-2")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer da.RolePermissionDelete(da.ctx, "role-test-role-permission-list", "test", "permission-test-role-permission-list-2")

	// Expect a sorted list!
	expect := rest.RolePermissionList{
		{BundleName: "test", Permission: "permission-test-role-permission-list-1"},
		{BundleName: "test", Permission: "permission-test-role-permission-list-2"},
		{BundleName: "test", Permission: "permission-test-role-permission-list-3"},
	}

	actual, err := da.RolePermissionList(da.ctx, "role-test-role-permission-list")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, expect, actual)
}
