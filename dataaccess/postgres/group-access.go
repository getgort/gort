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

package postgres

import (
	"github.com/clockworksoul/gort/data/rest"
	"github.com/clockworksoul/gort/dataaccess/errs"
	gerr "github.com/clockworksoul/gort/errors"
)

// GroupAddUser adds a user to a group
func (da PostgresDataAccess) GroupAddUser(groupname string, username string) error {
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

	db, err := da.connect("gort")
	if err != nil {
		return err
	}
	defer db.Close()

	query := `INSERT INTO groupusers (groupname, username) VALUES ($1, $2);`
	_, err = db.Exec(query, groupname, username)
	if err != nil {
		err = gerr.Wrap(errs.ErrDataAccess, err)
	}

	return err
}

// GroupCreate creates a new user group.
func (da PostgresDataAccess) GroupCreate(group rest.Group) error {
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

	db, err := da.connect("gort")
	if err != nil {
		return err
	}
	defer db.Close()

	query := `INSERT INTO groups (groupname) VALUES ($1);`
	_, err = db.Exec(query, group.Name)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return err
}

// GroupDelete deletes a group.
func (da PostgresDataAccess) GroupDelete(groupname string) error {
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

	db, err := da.connect("gort")
	if err != nil {
		return err
	}
	defer db.Close()

	query := `DELETE FROM groupusers WHERE groupname=$1;`
	_, err = db.Exec(query, groupname)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	query = `DELETE FROM groups WHERE groupname=$1;`
	_, err = db.Exec(query, groupname)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

// GroupExists is used to determine whether a group exists in the data store.
func (da PostgresDataAccess) GroupExists(groupname string) (bool, error) {
	db, err := da.connect("gort")
	if err != nil {
		return false, err
	}
	defer db.Close()

	query := "SELECT EXISTS(SELECT 1 FROM groups WHERE groupname=$1)"
	exists := false

	err = db.QueryRow(query, groupname).Scan(&exists)
	if err != nil {
		return false, gerr.Wrap(errs.ErrNoSuchGroup, err)
	}

	return exists, nil
}

// GroupGet gets a specific group.
func (da PostgresDataAccess) GroupGet(groupname string) (rest.Group, error) {
	if groupname == "" {
		return rest.Group{}, errs.ErrEmptyGroupName
	}

	db, err := da.connect("gort")
	if err != nil {
		return rest.Group{}, err
	}
	defer db.Close()

	// There will be more fields here eventually
	query := `SELECT groupname
		FROM groups
		WHERE groupname=$1`

	group := rest.Group{}
	err = db.QueryRow(query, groupname).Scan(&group.Name)
	if err != nil {
		return group, gerr.Wrap(errs.ErrNoSuchGroup, err)
	}

	users, err := da.GroupListUsers(groupname)
	if err != nil {
		return group, err
	}

	group.Users = users

	return group, nil
}

// GroupGrantRole grants one or more roles to a group.
func (da PostgresDataAccess) GroupGrantRole() error {
	return errs.ErrNotImplemented
}

// GroupList returns a list of all known groups in the datastore.
// Passwords are not included. Nice try.
func (da PostgresDataAccess) GroupList() ([]rest.Group, error) {
	groups := make([]rest.Group, 0)

	db, err := da.connect("gort")
	if err != nil {
		return groups, err
	}
	defer db.Close()

	query := `SELECT groupname FROM groups`
	rows, err := db.Query(query)
	if err != nil {
		return groups, gerr.Wrap(errs.ErrDataAccess, err)
	}

	for rows.Next() {
		group := rest.Group{}

		err = rows.Scan(&group.Name)
		if err != nil {
			return groups, gerr.Wrap(errs.ErrNoSuchGroup, err)
		}

		groups = append(groups, group)
	}

	return groups, nil
}

// GroupListUsers returns a list of all known users in a group.
func (da PostgresDataAccess) GroupListUsers(groupname string) ([]rest.User, error) {
	users := make([]rest.User, 0)

	db, err := da.connect("gort")
	if err != nil {
		return users, err
	}
	defer db.Close()

	query := `SELECT email, full_name, username
	FROM users
	WHERE username IN (
		SELECT username
		FROM groupusers
		WHERE groupname = $1
	)`

	rows, err := db.Query(query, groupname)
	if err != nil {
		return users, gerr.Wrap(errs.ErrDataAccess, err)
	}

	for rows.Next() {
		user := rest.User{}

		err = rows.Scan(&user.Email, &user.FullName, &user.Username)
		if err != nil {
			return users, gerr.Wrap(errs.ErrNoSuchUser, err)
		}

		users = append(users, user)
	}

	return users, nil
}

// GroupRemoveUser removes a user from a group.
func (da PostgresDataAccess) GroupRemoveUser(groupname string, username string) error {
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

	db, err := da.connect("gort")
	if err != nil {
		return err
	}
	defer db.Close()

	query := "DELETE FROM groupusers WHERE groupname=$1 AND username=$2;"
	_, err = db.Exec(query, groupname, username)
	if err != nil {
		err = gerr.Wrap(errs.ErrDataAccess, err)
	}

	return err
}

// GroupRevokeRole revokes a role from a group.
func (da PostgresDataAccess) GroupRevokeRole() error {
	return errs.ErrNotImplemented
}

// GroupUpdate is used to update an existing group. An error is returned if the
// groupname is empty or if the group doesn't exist.
// TODO Should we let this create groups that don't exist?
func (da PostgresDataAccess) GroupUpdate(group rest.Group) error {
	if group.Name == "" {
		return errs.ErrEmptyGroupName
	}

	exists, err := da.UserExists(group.Name)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchGroup
	}

	db, err := da.connect("gort")
	if err != nil {
		return err
	}
	defer db.Close()

	// There will be more eventually
	query := `UPDATE groupname
	SET groupname=$1
	WHERE groupname=$1;`

	_, err = db.Exec(query, group.Name)
	if err != nil {
		err = gerr.Wrap(errs.ErrDataAccess, err)
	}

	return err
}

// GroupUserList comments TBD
func (da PostgresDataAccess) GroupUserList(group string) ([]rest.User, error) {
	return []rest.User{}, errs.ErrNotImplemented
}

// GroupUserAdd comments TBD
func (da PostgresDataAccess) GroupUserAdd(group string, user string) error {
	return errs.ErrNotImplemented
}

// GroupUserDelete comments TBD
func (da PostgresDataAccess) GroupUserDelete(group string, user string) error {
	return errs.ErrNotImplemented
}
