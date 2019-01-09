package dataaccess

import (
	"testing"

	"github.com/clockworksoul/cog2/data/rest"
)

func TestGroupExists(t *testing.T) {
	var exists bool

	exists, _ = GroupExists("test-exists")
	if exists {
		t.Error("Group should not exist now")
	}

	// Now we add a group to find.
	_groups.m["test-exists"] = &rest.Group{Name: "test-exists"}
	defer delete(_groups.m, "test-exists")

	exists, _ = GroupExists("test-exists")
	if !exists {
		t.Error("Group should exist now")
	}
}
func TestGroupCreate(t *testing.T) {
	var err error
	var group rest.Group

	// Expect an error
	err = GroupCreate(group)
	if err == nil {
		t.Error("Expected an error")
	}

	// Expect no error
	err = GroupCreate(rest.Group{Name: "test-create"})
	defer delete(_groups.m, "test-create")
	if err != nil {
		t.Error("Expected no error")
	}

	// Expect an error
	err = GroupCreate(rest.Group{Name: "test-create"})
	if err == nil {
		t.Error("Expected no error")
	}
}

func TestGroupGet(t *testing.T) {
	var err error
	var group rest.Group

	// Expect an error
	_, err = GroupGet("")
	if err == nil {
		t.Error("Expected an error")
	}

	// Expect an error
	_, err = GroupGet("test-get")
	if err == nil {
		t.Error("Expected an error")
	}

	_groups.m["test-get"] = &rest.Group{Name: "test-get"}
	defer delete(_groups.m, "test-get")

	// Group should exist now
	exists, _ := GroupExists("test-get")
	if !exists {
		t.Error("Group should exist now")
	}

	// Expect no error
	group, err = GroupGet("test-get")
	if err != nil {
		t.Error("Expected no error")
	}
	if group.Name != "test-get" {
		t.Errorf("Group name mismatch: %q is not \"test-get\"", group.Name)
	}
}

func TestGroupList(t *testing.T) {
	GroupCreate(rest.Group{Name: "test-list-0"})
	GroupCreate(rest.Group{Name: "test-list-1"})
	GroupCreate(rest.Group{Name: "test-list-2"})
	GroupCreate(rest.Group{Name: "test-list-3"})

	defer func() { _groups.m = make(map[string]*rest.Group) }()

	groups, err := GroupList()
	if err != nil {
		t.Error("Expected no error")
	}

	if len(groups) != 4 {
		t.Errorf("Expected len(groups) = 4; got %d", len(groups))
	}

	for _, u := range groups {
		if u.Name == "" {
			t.Error("Expected non-empty name")
		}
	}
}

func TestGroupDelete(t *testing.T) {
	// Delete blank group
	err := GroupDelete("")
	if err == nil {
		t.Error("Expected an error")
	}

	// Delete group that doesn't exist
	err = GroupDelete("no-such-group")
	if err == nil {
		t.Error("Expected an error")
	}

	group := rest.Group{Name: "test-delete"}
	GroupCreate(group) // This has its own test

	err = GroupDelete("test-delete")
	if err != nil {
		t.Error("Expected no error")
	}

	exists, _ := GroupExists("test-delete")
	if exists {
		t.Error("Shouldn't exist anymore!")
	}
}
func TestGroupAddUser(t *testing.T) {
	err := GroupAddUser("foo", "bar")
	if err == nil {
		t.Error("Expected an error")
	}

	GroupCreate(rest.Group{Name: "foo"})
	defer GroupDelete("foo")

	err = GroupAddUser("foo", "bar")
	if err == nil {
		t.Error("Expected an error")
	}

	UserCreate(rest.User{Username: "bar"})
	defer UserDelete("bar")

	err = GroupAddUser("foo", "bar")
	if err != nil {
		t.Error("Expected no error")
	}

	group, _ := GroupGet("foo")

	if len(group.Users) != 1 {
		t.Error("Users list empty")
	}

	if len(group.Users) > 0 && group.Users[0].Username != "bar" {
		t.Error("Wrong user!")
	}
}
