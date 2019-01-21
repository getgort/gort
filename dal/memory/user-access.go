package memory

import (
	"fmt"

	"github.com/clockworksoul/cog2/data/rest"
)

// UserAuthenticate authenticates a username/password combination.
func (da InMemoryDataAccess) UserAuthenticate(username string, password string) (bool, error) {
	exists, err := da.UserExists(username)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, fmt.Errorf("no such user: %s", username)
	}

	user, err := da.UserGet(username)
	if err != nil {
		return false, err
	}

	return password == user.Password, nil
}

// UserCreate is used to create a new Cog user in the data store. An error is
// returned if the username is empty or if a user already exists.
func (da InMemoryDataAccess) UserCreate(user rest.User) error {
	if user.Username == "" {
		return fmt.Errorf("empty username")
	}

	exists, err := da.UserExists(user.Username)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("user %s already exists", user.Username)
	}

	da.users[user.Username] = &user

	return nil
}

// UserDelete deletes an existing user from the data store. An error is
// returned if the username parameter is empty of if the user doesn't
// exist.
func (da InMemoryDataAccess) UserDelete(username string) error {
	if username == "" {
		return fmt.Errorf("empty username")
	}

	exists, err := da.UserExists(username)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("no such user: %s", username)
	}

	delete(da.users, username)

	return nil
}

// UserExists is used to determine whether a Cog user with the given username
// exists in the data store.
func (da InMemoryDataAccess) UserExists(username string) (bool, error) {
	_, exists := da.users[username]

	return exists, nil
}

// UserGet returns a user from the data store. An error is returned if the
// username parameter is empty or if the user doesn't exist.
func (da InMemoryDataAccess) UserGet(username string) (rest.User, error) {
	if username == "" {
		return rest.User{}, fmt.Errorf("empty username")
	}

	exists, err := da.UserExists(username)
	if err != nil {
		return rest.User{}, err
	}
	if !exists {
		return rest.User{}, fmt.Errorf("no such user: %s", username)
	}

	user := da.users[username]

	return *user, nil
}

// UserList returns a list of all known users in the datastore.
// Passwords are not included. Nice try.
func (da InMemoryDataAccess) UserList() ([]rest.User, error) {
	list := make([]rest.User, 0)

	for _, u := range da.users {
		u.Password = ""
		list = append(list, *u)
	}

	return list, nil
}

// UserUpdate is used to update an existing user. An error is returned if the
// username is empty or if the user doesn't exist.
// TODO Should we let this create users that don't exist?
func (da InMemoryDataAccess) UserUpdate(user rest.User) error {
	if user.Username == "" {
		return fmt.Errorf("empty username")
	}

	exists, err := da.UserExists(user.Username)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("user %s doesn't exist", user.Username)
	}

	da.users[user.Username] = &user

	return nil
}
