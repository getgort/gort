package memory

import (
	"github.com/clockworksoul/cog2/data/rest"
	"github.com/clockworksoul/cog2/dataaccess/errs"
)

// GroupAddUser adds a user to a group
func (da *InMemoryDataAccess) GroupAddUser(groupname string, username string) error {
	if groupname == "" {
		return errs.ErrEmptyGroupName
	}

	exists, err := da.GroupExists(groupname)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchGroup
	}

	if username == "" {
		return errs.ErrEmptyUserName
	}

	exists, err = da.UserExists(username)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchUser
	}

	group := da.groups[groupname]
	user := da.users[username]
	group.Users = append(group.Users, *user)

	return nil
}

// GroupCreate creates a new user group.
func (da *InMemoryDataAccess) GroupCreate(group rest.Group) error {
	if group.Name == "" {
		return errs.ErrEmptyGroupName
	}

	exists, err := da.GroupExists(group.Name)
	if err != nil {
		return err
	}
	if exists {
		return errs.ErrGroupExists
	}

	da.groups[group.Name] = &group

	return nil
}

// GroupDelete delete a group.
func (da *InMemoryDataAccess) GroupDelete(groupname string) error {
	if groupname == "" {
		return errs.ErrEmptyGroupName
	}

	// Thou Shalt Not Delete Admin
	if groupname == "admin" {
		return errs.ErrAdminUndeletable
	}

	exists, err := da.GroupExists(groupname)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchGroup
	}

	delete(da.groups, groupname)

	return nil
}

// GroupExists is used to determine whether a group exists in the data store.
func (da *InMemoryDataAccess) GroupExists(groupname string) (bool, error) {
	_, exists := da.groups[groupname]

	return exists, nil
}

// GroupGet gets a specific group.
func (da *InMemoryDataAccess) GroupGet(groupname string) (rest.Group, error) {
	if groupname == "" {
		return rest.Group{}, errs.ErrEmptyGroupName
	}

	exists, err := da.GroupExists(groupname)
	if err != nil {
		return rest.Group{}, err
	}
	if !exists {
		return rest.Group{}, errs.ErrNoSuchGroup
	}

	group := da.groups[groupname]

	return *group, nil
}

// GroupGrantRole grants one or more roles to a group.
func (da *InMemoryDataAccess) GroupGrantRole() error {
	return errs.ErrNotImplemented
}

// GroupList returns a list of all known groups in the datastore.
// Passwords are not included. Nice try.
func (da *InMemoryDataAccess) GroupList() ([]rest.Group, error) {
	list := make([]rest.Group, 0)

	for _, g := range da.groups {
		list = append(list, *g)
	}

	return list, nil
}

// GroupRemoveUser removes one or more users from a group.
func (da *InMemoryDataAccess) GroupRemoveUser(groupname string, username string) error {
	if groupname == "" {
		return errs.ErrEmptyGroupName
	}

	exists, err := da.GroupExists(groupname)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchGroup
	}

	group := da.groups[groupname]

	for i, u := range group.Users {
		if u.Username == username {
			group.Users = append(group.Users[:i], group.Users[i+1:]...)
			return nil
		}
	}

	return errs.ErrNoSuchUser
}

// GroupRevokeRole revokes one or more roles from a group.
func (da *InMemoryDataAccess) GroupRevokeRole() error {
	return errs.ErrNotImplemented
}

// GroupUpdate is used to update an existing group. An error is returned if the
// groupname is empty or if the group doesn't exist.
// TODO Should we let this create groups that don't exist?
func (da *InMemoryDataAccess) GroupUpdate(group rest.Group) error {
	if group.Name == "" {
		return errs.ErrEmptyGroupName
	}

	exists, err := da.GroupExists(group.Name)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchGroup
	}

	da.groups[group.Name] = &group

	return nil
}

// GroupUserList comments TBD
func (da *InMemoryDataAccess) GroupUserList(group string) ([]rest.User, error) {
	return []rest.User{}, errs.ErrNotImplemented
}

// GroupUserAdd comments TBD
func (da *InMemoryDataAccess) GroupUserAdd(group string, user string) error {
	return errs.ErrNotImplemented
}

// GroupUserDelete comments TBD
func (da *InMemoryDataAccess) GroupUserDelete(group string, user string) error {
	return errs.ErrNotImplemented
}
