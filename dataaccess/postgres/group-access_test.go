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

func testGroupAccess(t *testing.T) {
	t.Run("testGroupExists", testGroupExists)
	t.Run("testGroupCreate", testGroupCreate)
	t.Run("testGroupDelete", testGroupDelete)
	t.Run("testGroupGet", testGroupGet)
	t.Run("testGroupList", testGroupList)
	t.Run("testGroupAddUser", testGroupAddUser)
	t.Run("testGroupRemoveUser", testGroupRemoveUser)
}

func testGroupExists(t *testing.T) {
	var exists bool

	exists, _ = da.GroupExists("test-exists")
	if exists {
		t.Error("Group should not exist now")
	}

	// Now we add a group to find.
	da.GroupCreate(rest.Group{Name: "test-exists"})
	defer da.GroupDelete("test-exists")

	exists, _ = da.GroupExists("test-exists")
	if !exists {
		t.Error("Group should exist now")
	}
}

func testGroupCreate(t *testing.T) {
	var err error
	var group rest.Group

	// Expect an error
	err = da.GroupCreate(group)
	expectErr(t, err, errs.ErrEmptyGroupName)

	// Expect no error
	err = da.GroupCreate(rest.Group{Name: "test-create"})
	defer da.GroupDelete("test-create")
	assert.NoError(t, err)

	// Expect an error
	err = da.GroupCreate(rest.Group{Name: "test-create"})
	expectErr(t, err, errs.ErrGroupExists)
}

func testGroupDelete(t *testing.T) {
	// Delete blank group
	err := da.GroupDelete("")
	expectErr(t, err, errs.ErrEmptyGroupName)

	// Delete group that doesn't exist
	err = da.GroupDelete("no-such-group")
	expectErr(t, err, errs.ErrNoSuchGroup)

	da.GroupCreate(rest.Group{Name: "test-delete"}) // This has its own test
	defer da.GroupDelete("test-delete")

	err = da.GroupDelete("test-delete")
	assert.NoError(t, err)

	exists, _ := da.GroupExists("test-delete")
	if exists {
		t.Error("Shouldn't exist anymore!")
	}
}

func testGroupGet(t *testing.T) {
	var err error
	var group rest.Group

	// Expect an error
	_, err = da.GroupGet("")
	expectErr(t, err, errs.ErrEmptyGroupName)

	// Expect an error
	_, err = da.GroupGet("test-get")
	expectErr(t, err, errs.ErrNoSuchGroup)

	da.GroupCreate(rest.Group{Name: "test-get"})
	defer da.GroupDelete("test-get")

	// da.Group should exist now
	exists, _ := da.GroupExists("test-get")
	if !exists {
		t.Error("Group should exist now")
	}

	// Expect no error
	group, err = da.GroupGet("test-get")
	assert.NoError(t, err)
	if group.Name != "test-get" {
		t.Errorf("Group name mismatch: %q is not \"test-get\"", group.Name)
	}
}

func testGroupList(t *testing.T) {
	da.GroupCreate(rest.Group{Name: "test-list-0"})
	defer da.GroupDelete("test-list-0")
	da.GroupCreate(rest.Group{Name: "test-list-1"})
	defer da.GroupDelete("test-list-1")
	da.GroupCreate(rest.Group{Name: "test-list-2"})
	defer da.GroupDelete("test-list-2")
	da.GroupCreate(rest.Group{Name: "test-list-3"})
	defer da.GroupDelete("test-list-3")

	groups, err := da.GroupList()
	assert.NoError(t, err)

	if len(groups) != 4 {
		t.Errorf("Expected len(groups) = 4; got %d", len(groups))
	}

	for _, u := range groups {
		if u.Name == "" {
			t.Error("Expected non-empty name")
		}
	}
}

func testGroupAddUser(t *testing.T) {
	err := da.GroupAddUser("foo", "bar")
	expectErr(t, err, errs.ErrNoSuchGroup)

	da.GroupCreate(rest.Group{Name: "foo"})
	defer da.GroupDelete("foo")

	err = da.GroupAddUser("foo", "bar")
	expectErr(t, err, errs.ErrNoSuchUser)

	da.UserCreate(rest.User{Username: "bar", Email: "bar"})
	defer da.UserDelete("bar")

	err = da.GroupAddUser("foo", "bar")
	assert.NoError(t, err)

	group, _ := da.GroupGet("foo")

	if len(group.Users) != 1 {
		t.Error("Users list empty")
	}

	if len(group.Users) > 0 && group.Users[0].Username != "bar" {
		t.Error("Wrong user!")
	}
}

func testGroupRemoveUser(t *testing.T) {
	da.GroupCreate(rest.Group{Name: "foo"})
	defer da.GroupDelete("foo")

	da.UserCreate(rest.User{Username: "bat"})
	defer da.UserDelete("bat")

	err := da.GroupAddUser("foo", "bat")
	assert.NoError(t, err)

	group, err := da.GroupGet("foo")
	assert.NoError(t, err)

	if len(group.Users) != 1 {
		t.Error("Users list empty")
	}

	if len(group.Users) > 0 && group.Users[0].Username != "bat" {
		t.Error("Wrong user!")
	}

	err = da.GroupRemoveUser("foo", "bat")
	assert.NoError(t, err)

	group, err = da.GroupGet("foo")
	assert.NoError(t, err)

	if len(group.Users) != 0 {
		t.Error("User not removed")
	}
}
