package dataaccess

import (
	"errors"
	"fmt"
	"sync"

	"github.com/clockworksoul/cog2/data/rest"
)

// A temporary place to store data until we can get a database into place.
var _groups = struct {
	sync.RWMutex
	m map[string]*rest.Group
}{m: make(map[string]*rest.Group)}

// GroupAddUser adds a user to a group
func GroupAddUser(groupname string, username string) error {
	_, err := GroupGet(groupname)
	if err != nil {
		return err
	}

	_, err = UserGet(username)
	if err != nil {
		return err
	}

	_groups.RLock()
	_users.RLock()

	group := _groups.m[groupname]
	user := _users.m[username]
	group.Users = append(group.Users, *user)

	_users.RUnlock()
	_groups.RUnlock()

	return nil
}

// GroupCreate creates a new user group.
func GroupCreate(group rest.Group) error {
	if group.Name == "" {
		return fmt.Errorf("empty group name")
	}

	exists, err := GroupExists(group.Name)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("group %s already exists", group.Name)
	}

	_groups.Lock()
	_groups.m[group.Name] = &group
	_groups.Unlock()

	return nil
}

// GroupDelete delete a group.
func GroupDelete(name string) error {
	if name == "" {
		return fmt.Errorf("empty group name")
	}

	exists, err := GroupExists(name)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("no such group: %s", name)
	}

	_groups.Lock()
	delete(_groups.m, name)
	_groups.Unlock()

	return nil
}

// GroupExists is used to determine whether a group exists in the data store.
func GroupExists(name string) (bool, error) {
	_groups.RLock()
	_, exists := _groups.m[name]
	_groups.RUnlock()

	return exists, nil
}

// GroupGet gets a specific group.
func GroupGet(name string) (rest.Group, error) {
	if name == "" {
		return rest.Group{}, fmt.Errorf("empty group name")
	}

	exists, err := GroupExists(name)
	if err != nil {
		return rest.Group{}, err
	}
	if !exists {
		return rest.Group{}, fmt.Errorf("no such group: %s", name)
	}

	_groups.RLock()
	group := _groups.m[name]
	_groups.RUnlock()

	return *group, nil
}

// GroupGrantRole grants one or more roles to a group.
func GroupGrantRole() error {
	return errors.New("Not yet supported")
}

// GroupList returns a list of all known groups in the datastore.
// Passwords are not included. Nice try.
func GroupList() ([]rest.Group, error) {
	list := make([]rest.Group, 0)

	for _, g := range _groups.m {
		list = append(list, *g)
	}

	return list, nil
}

// GroupRemoveUser removes one or more users from a group.
func GroupRemoveUser(groupname string, username string) error {
	_groups.RLock()
	group := _groups.m[groupname]
	defer _groups.RUnlock()

	for i, u := range group.Users {
		if u.Username == username {
			group.Users = append(group.Users[:i], group.Users[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("user %s is not in group %s", username, groupname)
}

// GroupRevokeRole revokes one or more roles from a group.
func GroupRevokeRole() error {
	return errors.New("Not yet supported")
}

// GroupUpdate is used to update an existing group. An error is returned if the
// groupname is empty or if the group doesn't exist.
// TODO Should we let this create groups that don't exist?
func GroupUpdate(group rest.Group) error {
	if group.Name == "" {
		return fmt.Errorf("empty group name")
	}

	exists, err := GroupExists(group.Name)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("group %s doesn't exist", group.Name)
	}

	_groups.Lock()
	_groups.m[group.Name] = &group
	_groups.Unlock()

	return nil
}
