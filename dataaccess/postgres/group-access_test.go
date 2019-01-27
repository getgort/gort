package postgres

import (
	"fmt"
	"testing"

	"github.com/clockworksoul/cog2/data"
	"github.com/clockworksoul/cog2/data/rest"
)

var (
	configs = data.DatabaseConfigs{
		Host:       "localhost",
		Password:   "password",
		Port:       5432,
		SSLEnabled: false,
		User:       "cog",
	}

	da = PostgresDataAccess{configs: configs}
)

func TestGroupExists(t *testing.T) {
	var exists bool

	err := da.Initialize()
	if err != nil {
		t.Error(err.Error())
	}

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
	if err == nil {
		t.Error("Expected an error")
	}

	// Expect no error
	err = da.GroupCreate(rest.Group{Name: "test-create"})
	defer da.GroupDelete("test-create")
	if err != nil {
		t.Error("Expected no error; got:", err.Error())
	}

	// Expect an error
	err = da.GroupCreate(rest.Group{Name: "test-create"})
	if err == nil {
		t.Error("Expected no error; got:", err.Error())
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

	da.GroupCreate(rest.Group{Name: "test-delete"}) // This has its own test
	defer da.GroupDelete("test-delete")

	err = da.GroupDelete("test-delete")
	if err != nil {
		t.Error("Expected no error; got:", err.Error())
	}

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
	if err == nil {
		t.Error("Expected an error")
	}

	// Expect an error
	_, err = da.GroupGet("test-get")
	if err == nil {
		t.Error("Expected an error")
	}

	da.GroupCreate(rest.Group{Name: "test-get"})
	defer da.GroupDelete("test-get")

	// da.Group should exist now
	exists, _ := da.GroupExists("test-get")
	if !exists {
		t.Error("Group should exist now")
	}

	// Expect no error
	group, err = da.GroupGet("test-get")
	if err != nil {
		t.Error("Expected no error; got:", err.Error())
	}
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
	if err != nil {
		t.Error("Expected no error; got:", err.Error())
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

	da.UserCreate(rest.User{Username: "bar", Email: "bar"})
	defer da.UserDelete("bar")

	err = da.GroupAddUser("foo", "bar")
	if err != nil {
		t.Error("Expected no error; got:", err.Error())
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
		t.Error("Expected no error; got:", err.Error())
	}

	group, err := da.GroupGet("foo")
	if err != nil {
		t.Error("Expected no error; got:", err.Error())
	}

	if len(group.Users) != 1 {
		t.Error("Users list empty")
	}

	if len(group.Users) > 0 && group.Users[0].Username != "bat" {
		t.Error("Wrong user!")
	}

	err = da.GroupRemoveUser("foo", "bat")
	if err != nil {
		t.Error("Expected no error; got:", err.Error())
	}

	group, err = da.GroupGet("foo")
	if err != nil {
		t.Error("Expected no error; got:", err.Error())
	}

	if len(group.Users) != 0 {
		fmt.Println(group.Users)
		t.Error("User not removed")
	}
}
