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

// GroupList returns a list of all known groups in the datastore.
// Passwords are not included. Nice try.
func (da *InMemoryDataAccess) GroupList(ctx context.Context) ([]rest.Group, error) {
	list := make([]rest.Group, 0)

	for _, g := range da.groups {
		list = append(list, *g)
	}

	return list, nil
}

func (da *InMemoryDataAccess) GroupPermissionList(ctx context.Context, groupname string) (rest.RolePermissionList, error) {
	roles, err := da.GroupRoleList(ctx, groupname)
	if err != nil {
		return rest.RolePermissionList{}, err
	}

	mp := map[string]rest.RolePermission{}

	for _, r := range roles {
		rpl, err := da.RolePermissionList(ctx, r.Name)
		if err != nil {
			return rest.RolePermissionList{}, err
		}

		for _, rp := range rpl {
			mp[rp.String()] = rp
		}
	}

	pp := []rest.RolePermission{}

	for _, p := range mp {
		pp = append(pp, p)
	}

	sort.Slice(pp, func(i, j int) bool { return pp[i].String() < pp[j].String() })

	return pp, nil
}

func (da *InMemoryDataAccess) GroupRoleList(ctx context.Context, groupname string) ([]rest.Role, error) {
	gr := da.groups[groupname]
	if gr == nil {
		return []rest.Role{}, nil
	}

	sort.Slice(gr.Roles, func(i, j int) bool { return gr.Roles[i].Name < gr.Roles[j].Name })

	return gr.Roles, nil
}

// GroupRoleAdd grants one or more roles to a group.
func (da *InMemoryDataAccess) GroupRoleAdd(ctx context.Context, groupname, rolename string) error {
	group, exists := da.groups[groupname]
	if !exists {
		return errs.ErrNoSuchGroup
	}

	role, exists := da.roles[rolename]
	if !exists {
		return errs.ErrNoSuchRole
	}

	group.Roles = append(group.Roles, *role)
	role.Groups = append(role.Groups, *group)

	return nil
}

// GroupRoleDelete revokes one or more roles from a group.
func (da *InMemoryDataAccess) GroupRoleDelete(ctx context.Context, groupname, rolename string) error {
	group, exists := da.groups[groupname]
	if !exists {
		return errs.ErrNoSuchGroup
	}

	role, exists := da.roles[rolename]
	if !exists {
		return errs.ErrNoSuchRole
	}

	for i, r := range group.Roles {
		if r.Name == rolename {
			group.Roles = append(group.Roles[:i], group.Roles[i+1:]...)
			break
		}
	}

	for i, g := range role.Groups {
		if g.Name == groupname {
			role.Groups = append(role.Groups[:i], role.Groups[i+1:]...)
			break
		}
	}

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

// GroupUserAdd adds a user to a group
func (da *InMemoryDataAccess) GroupUserAdd(ctx context.Context, groupname string, username string) error {
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

// GroupUserDelete removes one or more users from a group.
func (da *InMemoryDataAccess) GroupUserDelete(ctx context.Context, groupname string, username string) error {
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

func (da *InMemoryDataAccess) GroupUserList(ctx context.Context, groupname string) ([]rest.User, error) {
	group, exists := da.groups[groupname]
	if !exists {
		return []rest.User{}, errs.ErrNoSuchGroup
	}

	return group.Users, nil
}
