package postgres

import (
	"testing"

	"github.com/clockworksoul/cog2/data/rest"
)

func TestUserNotExists(t *testing.T) {
	var exists bool

	err := da.Initialize()
	if err != nil {
		t.Error(err.Error())
	}

	exists, _ = da.UserExists("test-not-exists")
	if exists {
		t.Error("User should not exist now")
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
	err = da.UserCreate(rest.User{Username: "test-create", Email: "test-create@bar.com"})
	defer da.UserDelete("test-create")
	if err != nil {
		t.Error("Expected no error. Got:", err.Error())
	}

	// Expect an error
	err = da.UserCreate(rest.User{Username: "test-create", Email: "test-create@bar.com"})
	if err == nil {
		t.Error("Expected an error")
	}
}

func TestUserAuthenticate(t *testing.T) {
	var err error

	// Expect no error
	err = da.UserCreate(rest.User{
		Username: "test-auth",
		Email:    "test-auth@bar.com",
		Password: "password",
	})
	defer da.UserDelete("test-auth")
	if err != nil {
		t.Error("Expected no error. Got:", err.Error())
	}

	authenticated, err := da.UserAuthenticate("test-auth", "no-match")
	if err != nil {
		t.Error("Expected no error. Got:", err.Error())
	}
	if authenticated {
		t.Error("Expected false")
	}

	authenticated, err = da.UserAuthenticate("test-auth", "password")
	if err != nil {
		t.Error("Expected no error. Got:", err.Error())
	}
	if !authenticated {
		t.Error("Expected true")
	}
}
func TestUserExists(t *testing.T) {
	var exists bool

	exists, _ = da.UserExists("test-exists")
	if exists {
		t.Error("User should not exist now")
	}

	// Now we add a user to find.
	err := da.UserCreate(rest.User{Username: "test-exists", Email: "test-exists@bar.com"})
	defer da.UserDelete("test-exists")
	if err != nil {
		t.Error(err.Error())
	}

	exists, _ = da.UserExists("test-exists")
	if !exists {
		t.Error("User should exist now")
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
	defer da.UserDelete("test-delete")

	err = da.UserDelete("test-delete")
	if err != nil {
		t.Error("Expected no error")
	}

	exists, _ := da.UserExists("test-delete")
	if exists {
		t.Error("Shouldn't exist anymore!")
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

	err = da.UserCreate(rest.User{Username: "test-get", Email: "test-get@foo.com"})
	defer da.UserDelete("test-get")
	if err != nil {
		t.Error(err.Error())
	}

	// da.User should exist now
	exists, _ := da.UserExists("test-get")
	if !exists {
		t.Error("User should exist now")
	}

	// Expect no error
	user, err = da.UserGet("test-get")
	if err != nil {
		t.Error("Expected no error; got:", err.Error())
	}
	if user.Username != "test-get" {
		t.Errorf("User name mismatch: %q is not \"test-get\"", user.Username)
	}
}

func TestUserList(t *testing.T) {
	da.UserCreate(rest.User{Username: "test-list-0", Password: "password0!", Email: "test-list-0"})
	defer da.UserDelete("test-list-0")
	da.UserCreate(rest.User{Username: "test-list-1", Password: "password1!", Email: "test-list-1"})
	defer da.UserDelete("test-list-1")
	da.UserCreate(rest.User{Username: "test-list-2", Password: "password2!", Email: "test-list-2"})
	defer da.UserDelete("test-list-2")
	da.UserCreate(rest.User{Username: "test-list-3", Password: "password3!", Email: "test-list-3"})
	defer da.UserDelete("test-list-3")

	users, err := da.UserList()
	if err != nil {
		t.Error("Expected no error")
	}

	if len(users) != 4 {
		for i, u := range users {
			t.Logf("User %d: %v\n", i+1, u)
		}

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
	err = da.UserUpdate(rest.User{Username: "no-such-user"})
	if err == nil {
		t.Error("Expected an error")
	}

	userA := rest.User{Username: "test-update", Email: "foo1.example.com"}
	da.UserCreate(userA)
	defer da.UserDelete("test-update")

	// Get the user we just added. Emails should match.
	user1, _ := da.UserGet("test-update")
	if userA.Email != user1.Email {
		t.Errorf("Email mistatch: %q vs %q", userA.Email, user1.Email)
	}

	// Do the update
	userB := rest.User{Username: "test-update", Email: "foo2.example.com"}
	err = da.UserUpdate(userB)
	if err != nil {
		t.Error("Expected no error; got:", err.Error())
	}

	// Get the user we just updated. Emails should match.
	user2, _ := da.UserGet("test-update")
	if userB.Email != user2.Email {
		t.Errorf("Email mistatch: %q vs %q", userB.Email, user2.Email)
	}
}
