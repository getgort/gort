/*
 * Copyright 2021 The Gort Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package memory

import (
	"context"
	"sort"

	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/dataaccess/errs"
)

// GroupAddUser adds a user to a group
func (da *InMemoryDataAccess) GroupAddUser(ctx context.Context, groupname string, username string) error {
	if groupname == "" {
		return errs.ErrEmptyGroupName
	}

	exists, err := da.GroupExists(ctx, groupname)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchGroup
	}

	if username == "" {
		return errs.ErrEmptyUserName
	}

	exists, err = da.UserExists(ctx, username)
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
func (da *InMemoryDataAccess) GroupCreate(ctx context.Context, group rest.Group) error {
	if group.Name == "" {
		return errs.ErrEmptyGroupName
	}

	exists, err := da.GroupExists(ctx, group.Name)
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
func (da *InMemoryDataAccess) GroupDelete(ctx context.Context, groupname string) error {
	if groupname == "" {
		return errs.ErrEmptyGroupName
	}

	// Thou Shalt Not Delete Admin
	if groupname == "admin" {
		return errs.ErrAdminUndeletable
	}

	exists, err := da.GroupExists(ctx, groupname)
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
func (da *InMemoryDataAccess) GroupExists(ctx context.Context, groupname string) (bool, error) {
	_, exists := da.groups[groupname]

	return exists, nil
}

// GroupGet gets a specific group.
func (da *InMemoryDataAccess) GroupGet(ctx context.Context, groupname string) (rest.Group, error) {
	if groupname == "" {
		return rest.Group{}, errs.ErrEmptyGroupName
	}

	exists, err := da.GroupExists(ctx, groupname)
	if err != nil {
		return rest.Group{}, err
	}
	if !exists {
		return rest.Group{}, errs.ErrNoSuchGroup
	}

	group := da.groups[groupname]

	return *group, nil
}

// GroupRoleAdd grants one or more roles to a group.
func (da *InMemoryDataAccess) GroupRoleAdd(ctx context.Context, groupname, rolename string) error {
	b, err := da.GroupExists(ctx, groupname)
	if err != nil {
		return err
	} else if !b {
		return errs.ErrNoSuchGroup
	}

	b, err = da.RoleExists(ctx, rolename)
	if err != nil {
		return err
	} else if !b {
		return errs.ErrNoSuchRole
	}

	m := da.grouproles[groupname]
	if m == nil {
		m = make(map[string]*rest.Role)
		da.grouproles[groupname] = m
	}

	m[rolename] = da.roles[rolename]
	return nil
}

// GroupList returns a list of all known groups in the datastore.
// Passwords are not included. Nice try.
func (da *InMemoryDataAccess) GroupList(ctx context.Context) ([]rest.Group, error) {
	list := make([]rest.Group, 0)

	for _, g := range da.groups {
		list = append(list, *g)
	}

	return list, nil
}

func (da *InMemoryDataAccess) GroupListRoles(ctx context.Context, groupname string) ([]rest.Role, error) {
	roles := []rest.Role{}

	gr := da.grouproles[groupname]
	if gr == nil {
		return roles, nil
	}

	for _, r := range gr {
		roles = append(roles, *r)
	}

	sort.Slice(roles, func(i, j int) bool { return roles[i].Name < roles[j].Name })

	return roles, nil
}

func (da *InMemoryDataAccess) GroupListUsers(ctx context.Context, groupname string) ([]rest.User, error) {
	return nil, errs.ErrNotImplemented
}

// GroupRemoveUser removes one or more users from a group.
func (da *InMemoryDataAccess) GroupRemoveUser(ctx context.Context, groupname string, username string) error {
	if groupname == "" {
		return errs.ErrEmptyGroupName
	}

	exists, err := da.GroupExists(ctx, groupname)
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

// GroupRoleDelete revokes one or more roles from a group.
func (da *InMemoryDataAccess) GroupRoleDelete(ctx context.Context, groupname, rolename string) error {
	b, err := da.GroupExists(ctx, groupname)
	if err != nil {
		return err
	} else if !b {
		return errs.ErrNoSuchGroup
	}

	b, err = da.RoleExists(ctx, rolename)
	if err != nil {
		return err
	} else if !b {
		return errs.ErrNoSuchRole
	}

	m := da.grouproles[groupname]
	if m == nil {
		m = make(map[string]*rest.Role)
		da.grouproles[groupname] = m
	}

	delete(m, rolename)
	return nil
}

// GroupUpdate is used to update an existing group. An error is returned if the
// groupname is empty or if the group doesn't exist.
// TODO Should we let this create groups that don't exist?
func (da *InMemoryDataAccess) GroupUpdate(ctx context.Context, group rest.Group) error {
	if group.Name == "" {
		return errs.ErrEmptyGroupName
	}

	exists, err := da.GroupExists(ctx, group.Name)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchGroup
	}

	da.groups[group.Name] = &group

	return nil
}

// GroupUserAdd comments TBD
func (da *InMemoryDataAccess) GroupUserAdd(ctx context.Context, group string, user string) error {
	return errs.ErrNotImplemented
}

// GroupUserDelete comments TBD
func (da *InMemoryDataAccess) GroupUserDelete(ctx context.Context, group string, user string) error {
	return errs.ErrNotImplemented
}
