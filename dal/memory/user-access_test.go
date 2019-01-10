package memory

import (
	"testing"

	"github.com/clockworksoul/cog2/data/rest"
)

func TestUserExists(t *testing.T) {
	var exists bool

	exists, _ = da.UserExists("test-exists")
	if exists {
		t.Error("User should not exist now")
	}

	// Now we add a user to find.
	da.users["test-exists"] = &rest.User{Username: "test-exists"}
	defer delete(da.users, "test-exists")

	exists, _ = da.UserExists("test-exists")
	if !exists {
		t.Error("User should exist now")
	}
}
func TestUserCreate(t *testing.T) {
	var err error
	var user rest.User

	// Expect an error
	err = da.UserCreate(user)
	if err == nil {
		t.Error("Expected an error")
	}

	// Expect no error
	err = da.UserCreate(rest.User{Username: "test-create"})
	defer delete(da.users, "test-create")
	if err != nil {
		t.Error("Expected no error")
	}

	// Expect an error
	err = da.UserCreate(rest.User{Username: "test-create"})
	if err == nil {
		t.Error("Expected no error")
	}
}

func TestUserGet(t *testing.T) {
	var err error
	var user rest.User

	// Expect an error
	_, err = da.UserGet("")
	if err == nil {
		t.Error("Expected an error")
	}

	// Expect an error
	_, err = da.UserGet("test-get")
	if err == nil {
		t.Error("Expected an error")
	}

	da.users["test-get"] = &rest.User{Username: "test-get"}
	defer delete(da.users, "test-get")

	// da.User should exist now
	exists, _ := da.UserExists("test-get")
	if !exists {
		t.Error("User should exist now")
	}

	// Expect no error
	user, err = da.UserGet("test-get")
	if err != nil {
		t.Error("Expected no error")
	}
	if user.Username != "test-get" {
		t.Errorf("User name mismatch: %q is not \"test-get\"", user.Username)
	}
}

func TestUserList(t *testing.T) {
	da.UserCreate(rest.User{Username: "test-list-0", Password: "password0!"})
	da.UserCreate(rest.User{Username: "test-list-1", Password: "password1!"})
	da.UserCreate(rest.User{Username: "test-list-2", Password: "password2!"})
	da.UserCreate(rest.User{Username: "test-list-3", Password: "password3!"})
	defer func() { da.users = make(map[string]*rest.User) }()

	users, err := da.UserList()
	if err != nil {
		t.Error("Expected no error")
	}

	if len(users) != 4 {
		t.Errorf("Expected len(users) = 4; got %d", len(users))
	}

	for _, u := range users {
		if u.Password != "" {
			t.Error("Expected empty password")
		}

		if u.Username == "" {
			t.Error("Expected non-empty username")
		}
	}
}

func TestUserUpdate(t *testing.T) {
	// Update blank user
	err := da.UserUpdate(rest.User{})
	if err == nil {
		t.Error("Expected an error")
	}

	// Update user that doesn't exist
	err = da.UserUpdate(rest.User{Username: "no-such-usere"})
	if err == nil {
		t.Error("Expected an error")
	}

	userA := rest.User{Username: "test-update", Email: "foo1.example.com"}
	da.UserCreate(userA)
	defer delete(da.users, "test-update")

	// Get the user we just added. Emails should match.
	user1, _ := da.UserGet("test-update")
	if userA.Email != user1.Email {
		t.Errorf("Email mistatch: %q vs %q", userA.Email, user1.Email)
	}

	// Do the update
	userB := rest.User{Username: "test-update", Email: "foo2.example.com"}
	err = da.UserUpdate(userB)
	if err != nil {
		t.Error("Expected no error")
	}

	// Get the user we just updated. Emails should match.
	user2, _ := da.UserGet("test-update")
	if userB.Email != user2.Email {
		t.Errorf("Email mistatch: %q vs %q", userB.Email, user2.Email)
	}
}
func TestUserDelete(t *testing.T) {
	// Delete blank user
	err := da.UserDelete("")
	if err == nil {
		t.Error("Expected an error")
	}

	// Delete user that doesn't exist
	err = da.UserDelete("no-such-user")
	if err == nil {
		t.Error("Expected an error")
	}

	user := rest.User{Username: "test-delete", Email: "foo1.example.com"}
	da.UserCreate(user) // This has its own test

	err = da.UserDelete("test-delete")
	if err != nil {
		t.Error("Expected no error")
	}

	exists, _ := da.UserExists("test-delete")
	if exists {
		t.Error("Shouldn't exist anymore!")
	}
}
