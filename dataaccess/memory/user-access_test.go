package memory

import (
	"testing"

	"github.com/clockworksoul/gort/data/rest"
	"github.com/clockworksoul/gort/dataaccess/errs"
	"github.com/stretchr/testify/assert"
)

func testUserAccess(t *testing.T) {
	t.Run("testUserNotExists", testUserNotExists)
	t.Run("testUserCreate", testUserCreate)
	t.Run("testUserAuthenticate", testUserAuthenticate)
	t.Run("testUserExists", testUserExists)
	t.Run("testUserDelete", testUserDelete)
	t.Run("testUserGet", testUserGet)
	t.Run("testUserList", testUserList)
	t.Run("testUserUpdate", testUserUpdate)
}

func testUserNotExists(t *testing.T) {
	var exists bool

	err := da.Initialize()
	assert.NoError(t, err)

	exists, _ = da.UserExists("test-not-exists")
	if exists {
		t.Error("User should not exist now")
	}
}

func testUserCreate(t *testing.T) {
	var err error
	var user rest.User

	// Expect an error
	err = da.UserCreate(user)
	expectErr(t, err, errs.ErrEmptyUserName)

	// Expect no error
	err = da.UserCreate(rest.User{Username: "test-create", Email: "test-create@bar.com"})
	defer da.UserDelete("test-create")
	assert.NoError(t, err)

	// Expect an error
	err = da.UserCreate(rest.User{Username: "test-create", Email: "test-create@bar.com"})
	expectErr(t, err, errs.ErrUserExists)
}

func testUserAuthenticate(t *testing.T) {
	var err error

	authenticated, err := da.UserAuthenticate("test-auth", "no-match")
	expectErr(t, err, errs.ErrNoSuchUser)
	if authenticated {
		t.Error("Expected false")
	}

	// Expect no error
	err = da.UserCreate(rest.User{
		Username: "test-auth",
		Email:    "test-auth@bar.com",
		Password: "password",
	})
	defer da.UserDelete("test-auth")
	assert.NoError(t, err)

	authenticated, err = da.UserAuthenticate("test-auth", "no-match")
	assert.NoError(t, err)
	if authenticated {
		t.Error("Expected false")
	}

	authenticated, err = da.UserAuthenticate("test-auth", "password")
	assert.NoError(t, err)
	if !authenticated {
		t.Error("Expected true")
	}
}
func testUserExists(t *testing.T) {
	var exists bool

	exists, _ = da.UserExists("test-exists")
	if exists {
		t.Error("User should not exist now")
	}

	// Now we add a user to find.
	err := da.UserCreate(rest.User{Username: "test-exists", Email: "test-exists@bar.com"})
	defer da.UserDelete("test-exists")
	assert.NoError(t, err)

	exists, _ = da.UserExists("test-exists")
	if !exists {
		t.Error("User should exist now")
	}
}

func testUserDelete(t *testing.T) {
	// Delete blank user
	err := da.UserDelete("")
	expectErr(t, err, errs.ErrEmptyUserName)

	// Delete admin user
	err = da.UserDelete("admin")
	expectErr(t, err, errs.ErrAdminUndeletable)

	// Delete user that doesn't exist
	err = da.UserDelete("no-such-user")
	expectErr(t, err, errs.ErrNoSuchUser)

	user := rest.User{Username: "test-delete", Email: "foo1.example.com"}
	da.UserCreate(user) // This has its own test
	defer da.UserDelete("test-delete")

	err = da.UserDelete("test-delete")
	assert.NoError(t, err)

	exists, _ := da.UserExists("test-delete")
	if exists {
		t.Error("Shouldn't exist anymore!")
	}
}

func testUserGet(t *testing.T) {
	var err error
	var user rest.User

	// Expect an error
	_, err = da.UserGet("")
	expectErr(t, err, errs.ErrEmptyUserName)

	// Expect an error
	_, err = da.UserGet("test-get")
	expectErr(t, err, errs.ErrNoSuchUser)

	err = da.UserCreate(rest.User{Username: "test-get", Email: "test-get@foo.com"})
	defer da.UserDelete("test-get")
	assert.NoError(t, err)

	// da.User should exist now
	exists, _ := da.UserExists("test-get")
	if !exists {
		t.Error("User should exist now")
	}

	// Expect no error
	user, err = da.UserGet("test-get")
	assert.NoError(t, err)
	if user.Username != "test-get" {
		t.Errorf("User name mismatch: %q is not \"test-get\"", user.Username)
	}
}

func testUserList(t *testing.T) {
	da.UserCreate(rest.User{Username: "test-list-0", Password: "password0!", Email: "test-list-0"})
	defer da.UserDelete("test-list-0")
	da.UserCreate(rest.User{Username: "test-list-1", Password: "password1!", Email: "test-list-1"})
	defer da.UserDelete("test-list-1")
	da.UserCreate(rest.User{Username: "test-list-2", Password: "password2!", Email: "test-list-2"})
	defer da.UserDelete("test-list-2")
	da.UserCreate(rest.User{Username: "test-list-3", Password: "password3!", Email: "test-list-3"})
	defer da.UserDelete("test-list-3")

	users, err := da.UserList()
	assert.NoError(t, err)

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

func testUserUpdate(t *testing.T) {
	// Update blank user
	err := da.UserUpdate(rest.User{})
	expectErr(t, err, errs.ErrEmptyUserName)

	// Update user that doesn't exist
	err = da.UserUpdate(rest.User{Username: "no-such-user"})
	expectErr(t, err, errs.ErrNoSuchUser)

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
	assert.NoError(t, err)

	// Get the user we just updated. Emails should match.
	user2, _ := da.UserGet("test-update")
	if userB.Email != user2.Email {
		t.Errorf("Email mistatch: %q vs %q", userB.Email, user2.Email)
	}
}
