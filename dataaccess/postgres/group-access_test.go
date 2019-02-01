package postgres

import (
	"fmt"
	"testing"

	"github.com/clockworksoul/cog2/data/rest"
	"github.com/clockworksoul/cog2/dataaccess/errs"
)

func TestGroupExists(t *testing.T) {
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

func TestGroupCreate(t *testing.T) {
	var err error
	var group rest.Group

	// Expect an error
	err = da.GroupCreate(group)
	expectErr(t, err, errs.ErrEmptyGroupName)

	// Expect no error
	err = da.GroupCreate(rest.Group{Name: "test-create"})
	defer da.GroupDelete("test-create")
	expectNoErr(t, err)

	// Expect an error
	err = da.GroupCreate(rest.Group{Name: "test-create"})
	expectErr(t, err, errs.ErrGroupExists)
}

func TestGroupDelete(t *testing.T) {
	// Delete blank group
	err := da.GroupDelete("")
	expectErr(t, err, errs.ErrEmptyGroupName)

	// Delete group that doesn't exist
	err = da.GroupDelete("no-such-group")
	expectErr(t, err, errs.ErrNoSuchGroup)

	da.GroupCreate(rest.Group{Name: "test-delete"}) // This has its own test
	defer da.GroupDelete("test-delete")

	err = da.GroupDelete("test-delete")
	expectNoErr(t, err)

	exists, _ := da.GroupExists("test-delete")
	if exists {
		t.Error("Shouldn't exist anymore!")
	}
}

func TestGroupGet(t *testing.T) {
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
	expectNoErr(t, err)
	if group.Name != "test-get" {
		t.Errorf("Group name mismatch: %q is not \"test-get\"", group.Name)
	}
}

func TestGroupList(t *testing.T) {
	da.GroupCreate(rest.Group{Name: "test-list-0"})
	defer da.GroupDelete("test-list-0")
	da.GroupCreate(rest.Group{Name: "test-list-1"})
	defer da.GroupDelete("test-list-1")
	da.GroupCreate(rest.Group{Name: "test-list-2"})
	defer da.GroupDelete("test-list-2")
	da.GroupCreate(rest.Group{Name: "test-list-3"})
	defer da.GroupDelete("test-list-3")

	groups, err := da.GroupList()
	expectNoErr(t, err)

	if len(groups) != 4 {
		t.Errorf("Expected len(groups) = 4; got %d", len(groups))
	}

	for _, u := range groups {
		if u.Name == "" {
			t.Error("Expected non-empty name")
		}
	}
}

func TestGroupAddUser(t *testing.T) {
	err := da.GroupAddUser("foo", "bar")
	expectErr(t, err, errs.ErrNoSuchGroup)

	da.GroupCreate(rest.Group{Name: "foo"})
	defer da.GroupDelete("foo")

	err = da.GroupAddUser("foo", "bar")
	expectErr(t, err, errs.ErrNoSuchUser)

	da.UserCreate(rest.User{Username: "bar", Email: "bar"})
	defer da.UserDelete("bar")

	err = da.GroupAddUser("foo", "bar")
	expectNoErr(t, err)

	group, _ := da.GroupGet("foo")

	if len(group.Users) != 1 {
		t.Error("Users list empty")
	}

	if len(group.Users) > 0 && group.Users[0].Username != "bar" {
		t.Error("Wrong user!")
	}
}

func TestGroupRemoveUser(t *testing.T) {
	da.GroupCreate(rest.Group{Name: "foo"})
	defer da.GroupDelete("foo")

	da.UserCreate(rest.User{Username: "bat"})
	defer da.UserDelete("bat")

	err := da.GroupAddUser("foo", "bat")
	expectNoErr(t, err)

	group, err := da.GroupGet("foo")
	expectNoErr(t, err)

	if len(group.Users) != 1 {
		t.Error("Users list empty")
	}

	if len(group.Users) > 0 && group.Users[0].Username != "bat" {
		t.Error("Wrong user!")
	}

	err = da.GroupRemoveUser("foo", "bat")
	expectNoErr(t, err)

	group, err = da.GroupGet("foo")
	expectNoErr(t, err)

	if len(group.Users) != 0 {
		fmt.Println(group.Users)
		t.Error("User not removed")
	}
}
