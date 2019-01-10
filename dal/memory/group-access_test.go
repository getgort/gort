package memory

import (
	"fmt"
	"testing"

	"github.com/clockworksoul/cog2/data/rest"
)

var (
	da = NewInMemoryDataAccess().(InMemoryDataAccess)
)

func TestGroupExists(t *testing.T) {
	var exists bool

	exists, _ = da.GroupExists("test-exists")
	if exists {
		t.Error("Group should not exist now")
	}

	// Now we add a group to find.
	da.groups["test-exists"] = &rest.Group{Name: "test-exists"}
	defer delete(da.groups, "test-exists")

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
	if err == nil {
		t.Error("Expected an error")
	}

	// Expect no error
	err = da.GroupCreate(rest.Group{Name: "test-create"})
	defer delete(da.groups, "test-create")
	if err != nil {
		t.Error("Expected no error")
	}

	// Expect an error
	err = da.GroupCreate(rest.Group{Name: "test-create"})
	if err == nil {
		t.Error("Expected no error")
	}
}

func TestGroupGet(t *testing.T) {
	var err error
	var group rest.Group

	// Expect an error
	_, err = da.GroupGet("")
	if err == nil {
		t.Error("Expected an error")
	}

	// Expect an error
	_, err = da.GroupGet("test-get")
	if err == nil {
		t.Error("Expected an error")
	}

	da.groups["test-get"] = &rest.Group{Name: "test-get"}
	defer delete(da.groups, "test-get")

	// da.Group should exist now
	exists, _ := da.GroupExists("test-get")
	if !exists {
		t.Error("Group should exist now")
	}

	// Expect no error
	group, err = da.GroupGet("test-get")
	if err != nil {
		t.Error("Expected no error")
	}
	if group.Name != "test-get" {
		t.Errorf("Group name mismatch: %q is not \"test-get\"", group.Name)
	}
}

func TestGroupList(t *testing.T) {
	da.GroupCreate(rest.Group{Name: "test-list-0"})
	da.GroupCreate(rest.Group{Name: "test-list-1"})
	da.GroupCreate(rest.Group{Name: "test-list-2"})
	da.GroupCreate(rest.Group{Name: "test-list-3"})

	defer func() { da.groups = make(map[string]*rest.Group) }()

	groups, err := da.GroupList()
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
	err := da.GroupDelete("")
	if err == nil {
		t.Error("Expected an error")
	}

	// Delete group that doesn't exist
	err = da.GroupDelete("no-such-group")
	if err == nil {
		t.Error("Expected an error")
	}

	group := rest.Group{Name: "test-delete"}
	da.GroupCreate(group) // This has its own test

	err = da.GroupDelete("test-delete")
	if err != nil {
		t.Error("Expected no error")
	}

	exists, _ := da.GroupExists("test-delete")
	if exists {
		t.Error("Shouldn't exist anymore!")
	}
}
func TestGroupAddUser(t *testing.T) {
	err := da.GroupAddUser("foo", "bar")
	if err == nil {
		t.Error("Expected an error")
	}

	da.GroupCreate(rest.Group{Name: "foo"})
	defer da.GroupDelete("foo")

	err = da.GroupAddUser("foo", "bar")
	if err == nil {
		t.Error("Expected an error")
	}

	da.UserCreate(rest.User{Username: "bar"})
	defer da.UserDelete("bar")

	err = da.GroupAddUser("foo", "bar")
	if err != nil {
		t.Error("Expected no error")
	}

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
	if err != nil {
		t.Error("Expected no error")
	}

	group, err := da.GroupGet("foo")
	if err != nil {
		t.Error("Expected no error")
	}

	if len(group.Users) != 1 {
		t.Error("Users list empty")
	}

	if len(group.Users) > 0 && group.Users[0].Username != "bat" {
		t.Error("Wrong user!")
	}

	err = da.GroupRemoveUser("foo", "bat")
	if err != nil {
		t.Error("Expected no error")
	}

	group, err = da.GroupGet("foo")
	if err != nil {
		t.Error("Expected no error")
	}

	if len(group.Users) != 0 {
		fmt.Println(group.Users)
		t.Error("User not removed")
	}
}
