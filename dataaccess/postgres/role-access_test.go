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

package postgres

import (
	"testing"

	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/dataaccess/errs"
	"github.com/stretchr/testify/assert"
)

func testRoleAccess(t *testing.T) {
	t.Run("testRoleCreate", testRoleCreate)
	t.Run("testRoleExists", testRoleExists)
	t.Run("testRoleDelete", testRoleDelete)
	t.Run("testRoleGet", testRoleGet)
}

func testRoleCreate(t *testing.T) {
	var err error

	// Expect an error
	err = da.RoleCreate(ctx, "")
	assert.Error(t, err, errs.ErrEmptyRoleName)

	// Expect no error
	err = da.RoleCreate(ctx, "test-create")
	defer da.RoleDelete(ctx, "test-create")
	assert.NoError(t, err)

	// Expect an error
	err = da.RoleCreate(ctx, "test-create")
	assert.Error(t, err, errs.ErrRoleExists)
}

func testRoleDelete(t *testing.T) {
	// Delete blank group
	err := da.RoleDelete(ctx, "")
	assert.Error(t, err, errs.ErrEmptyRoleName)

	// Delete group that doesn't exist
	err = da.RoleDelete(ctx, "no-such-group")
	assert.Error(t, err, errs.ErrNoSuchRole)

	da.RoleCreate(ctx, "test-delete") // This has its own test
	defer da.RoleDelete(ctx, "test-delete")

	err = da.RoleDelete(ctx, "test-delete")
	assert.NoError(t, err)

	exists, _ := da.RoleExists(ctx, "test-delete")
	if exists {
		t.Error("Shouldn't exist anymore!")
		t.FailNow()
	}
}

func testRoleExists(t *testing.T) {
	var exists bool

	exists, _ = da.RoleExists(ctx, "test-exists")
	if exists {
		t.Error("Role should not exist now")
		t.FailNow()
	}

	// Now we add a group to find.
	da.RoleCreate(ctx, "test-exists")
	defer da.RoleDelete(ctx, "test-exists")

	exists, _ = da.RoleExists(ctx, "test-exists")
	if !exists {
		t.Error("Role should exist now")
		t.FailNow()
	}
}

func testRoleGet(t *testing.T) {
	var err error
	var role rest.Role

	// Expect an error
	_, err = da.RoleGet(ctx, "")
	assert.Error(t, err, errs.ErrEmptyRoleName)

	// Expect an error
	_, err = da.RoleGet(ctx, "test-get")
	assert.Error(t, err, errs.ErrNoSuchRole)

	da.RoleCreate(ctx, "test-get")
	defer da.RoleDelete(ctx, "test-get")

	// da.Role ctx, should exist now
	exists, _ := da.RoleExists(ctx, "test-get")
	if !exists {
		t.Error("Role should exist now")
		t.FailNow()
	}

	err = da.RoleGrantPermission(ctx, "test-get", "foo", "bar")
	assert.NoError(t, err)

	err = da.RoleGrantPermission(ctx, "test-get", "foo", "bat")
	assert.NoError(t, err)

	err = da.RoleRevokePermission(ctx, "test-get", "foo", "bat")
	assert.NoError(t, err)

	expected := rest.Role{
		Name:        "test-get",
		Permissions: []rest.RolePermission{{BundleName: "foo", Permission: "bar"}},
	}

	// Expect no error
	role, err = da.RoleGet(ctx, "test-get")
	assert.NoError(t, err)
	assert.Equal(t, expected, role)
}
