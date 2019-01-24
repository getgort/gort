package memory

import (
	"errors"
	"fmt"

	"github.com/clockworksoul/cog2/data/rest"
)

// GroupAddUser adds a user to a group
func (da InMemoryDataAccess) GroupAddUser(groupname string, username string) error {
	_, err := da.GroupGet(groupname)
	if err != nil {
		return err
	}

	_, err = da.UserGet(username)
	if err != nil {
		return err
	}

	group := da.groups[groupname]
	user := da.users[username]
	group.Users = append(group.Users, *user)

	return nil
}

// GroupCreate creates a new user group.
func (da InMemoryDataAccess) GroupCreate(group rest.Group) error {
	if group.Name == "" {
		return fmt.Errorf("empty group name")
	}

	exists, err := da.GroupExists(group.Name)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("group %s already exists", group.Name)
	}

	da.groups[group.Name] = &group

	return nil
}

// GroupDelete delete a group.
func (da InMemoryDataAccess) GroupDelete(name string) error {
	if name == "" {
		return fmt.Errorf("empty group name")
	}

	exists, err := da.GroupExists(name)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("no such group: %s", name)
	}

	delete(da.groups, name)

	return nil
}

// GroupExists is used to determine whether a group exists in the data store.
func (da InMemoryDataAccess) GroupExists(name string) (bool, error) {
	_, exists := da.groups[name]

	return exists, nil
}

// GroupGet gets a specific group.
func (da InMemoryDataAccess) GroupGet(name string) (rest.Group, error) {
	if name == "" {
		return rest.Group{}, fmt.Errorf("empty group name")
	}

	exists, err := da.GroupExists(name)
	if err != nil {
		return rest.Group{}, err
	}
	if !exists {
		return rest.Group{}, fmt.Errorf("no such group: %s", name)
	}

	group := da.groups[name]

	return *group, nil
}

// GroupGrantRole grants one or more roles to a group.
func (da InMemoryDataAccess) GroupGrantRole() error {
	return errors.New("Not yet supported")
}

// GroupList returns a list of all known groups in the datastore.
// Passwords are not included. Nice try.
func (da InMemoryDataAccess) GroupList() ([]rest.Group, error) {
	list := make([]rest.Group, 0)

	for _, g := range da.groups {
		list = append(list, *g)
	}

	return list, nil
}

// GroupRemoveUser removes one or more users from a group.
func (da InMemoryDataAccess) GroupRemoveUser(groupname string, username string) error {
	group := da.groups[groupname]

	for i, u := range group.Users {
		if u.Username == username {
			group.Users = append(group.Users[:i], group.Users[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("user %s is not in group %s", username, groupname)
}

// GroupRevokeRole revokes one or more roles from a group.
func (da InMemoryDataAccess) GroupRevokeRole() error {
	return errors.New("Not yet supported")
}

// GroupUpdate is used to update an existing group. An error is returned if the
// groupname is empty or if the group doesn't exist.
// TODO Should we let this create groups that don't exist?
func (da InMemoryDataAccess) GroupUpdate(group rest.Group) error {
	if group.Name == "" {
		return fmt.Errorf("empty group name")
	}

	exists, err := da.GroupExists(group.Name)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("group %s doesn't exist", group.Name)
	}

	da.groups[group.Name] = &group

	return nil
}

// GroupUserList comments TBD
func (da InMemoryDataAccess) GroupUserList(group string) ([]rest.User, error) {
	return []rest.User{}, fmt.Errorf("InMemoryDataAccess: not yet implemented")
}

// GroupUserAdd comments TBD
func (da InMemoryDataAccess) GroupUserAdd(group string, user string) error {
	return fmt.Errorf("InMemoryDataAccess: not yet implemented")
}

// GroupUserDelete comments TBD
func (da InMemoryDataAccess) GroupUserDelete(group string, user string) error {
	return fmt.Errorf("InMemoryDataAccess: not yet implemented")
}
