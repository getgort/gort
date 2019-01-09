package dataaccess

import (
	"fmt"
	"sync"

	"github.com/clockworksoul/cog2/data/rest"
)

// A temporary place to store data until we can get a database into place.
var _users = struct {
	sync.RWMutex
	m map[string]rest.User
}{m: make(map[string]rest.User)}

// UserCreate is used to create a new Cog user in the data store. An error is
// returned if the username is empty or if a user already exists.
func UserCreate(user rest.User) error {
	if user.Username == "" {
		return fmt.Errorf("empty username")
	}

	exists, err := UserExists(user.Username)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("user %s already exists", user.Username)
	}

	_users.Lock()
	_users.m[user.Username] = user
	_users.Unlock()

	return nil
}

// UserDelete deletes an existing user from the data store. An error is
// returned if the username parameter is empty of if the user doesn't
// exist.
func UserDelete(username string) error {
	if username == "" {
		return fmt.Errorf("empty username")
	}

	exists, err := UserExists(username)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("no such user: %s", username)
	}

	_users.Lock()
	delete(_users.m, username)
	_users.Unlock()

	return nil
}

// UserExists is used to determine whether a Cog user with the given username
// exists in the data store.
func UserExists(username string) (bool, error) {
	_users.RLock()
	_, exists := _users.m[username]
	_users.RUnlock()

	return exists, nil
}

// UserGet returns a user from the data store. An error is returned if the
// username parameter is empty or if the user doesn't exist.
func UserGet(username string) (rest.User, error) {
	if username == "" {
		return rest.User{}, fmt.Errorf("empty username")
	}

	exists, err := UserExists(username)
	if err != nil {
		return rest.User{}, err
	}
	if !exists {
		return rest.User{}, fmt.Errorf("no such user: %s", username)
	}

	_users.RLock()
	user := _users.m[username]
	_users.RUnlock()

	return user, nil
}

// UserList returns a list of all known users in the datastore.
// Passwords are not included. Nice try.
func UserList() ([]rest.User, error) {
	list := make([]rest.User, 0)

	for _, u := range _users.m {
		u.Password = ""
		list = append(list, u)
	}

	return list, nil
}

// UserUpdate is used to update an existing user. An error is returned if the
// username is empty or if the user doesn't exist.
// TODO Should we let this create users that don't exist?
func UserUpdate(user rest.User) error {
	if user.Username == "" {
		return fmt.Errorf("empty username")
	}

	exists, err := UserExists(user.Username)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("user %s doesn't exist", user.Username)
	}

	_users.Lock()
	_users.m[user.Username] = user
	_users.Unlock()

	return nil
}
