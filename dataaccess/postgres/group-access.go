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
	"context"

	"go.opentelemetry.io/otel"

	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/dataaccess/errs"
	gerr "github.com/getgort/gort/errors"
	"github.com/getgort/gort/telemetry"
)

// GroupAddUser adds a user to a group
func (da PostgresDataAccess) GroupAddUser(ctx context.Context, groupname string, username string) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.GroupAddUser")
	defer sp.End()

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

	db, err := da.connect(ctx, "gort")
	if err != nil {
		return err
	}
	defer db.Close()

	query := `INSERT INTO groupusers (groupname, username) VALUES ($1, $2);`
	_, err = db.ExecContext(ctx, query, groupname, username)
	if err != nil {
		err = gerr.Wrap(errs.ErrDataAccess, err)
	}

	return err
}

// GroupCreate creates a new user group.
func (da PostgresDataAccess) GroupCreate(ctx context.Context, group rest.Group) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.GroupCreate")
	defer sp.End()

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

	db, err := da.connect(ctx, "gort")
	if err != nil {
		return err
	}
	defer db.Close()

	query := `INSERT INTO groups (groupname) VALUES ($1);`
	_, err = db.ExecContext(ctx, query, group.Name)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return err
}

// GroupDelete deletes a group.
func (da PostgresDataAccess) GroupDelete(ctx context.Context, groupname string) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.GroupDelete")
	defer sp.End()

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

	db, err := da.connect(ctx, "gort")
	if err != nil {
		return err
	}
	defer db.Close()

	query := `DELETE FROM groupusers WHERE groupname=$1;`
	_, err = db.ExecContext(ctx, query, groupname)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	query = `DELETE FROM groups WHERE groupname=$1;`
	_, err = db.ExecContext(ctx, query, groupname)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

// GroupExists is used to determine whether a group exists in the data store.
func (da PostgresDataAccess) GroupExists(ctx context.Context, groupname string) (bool, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.GroupExists")
	defer sp.End()

	db, err := da.connect(ctx, "gort")
	if err != nil {
		return false, err
	}
	defer db.Close()

	query := "SELECT EXISTS(SELECT 1 FROM groups WHERE groupname=$1)"
	exists := false

	err = db.QueryRowContext(ctx, query, groupname).Scan(&exists)
	if err != nil {
		return false, gerr.Wrap(errs.ErrNoSuchGroup, err)
	}

	return exists, nil
}

// GroupGet gets a specific group.
func (da PostgresDataAccess) GroupGet(ctx context.Context, groupname string) (rest.Group, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.GroupGet")
	defer sp.End()

	if groupname == "" {
		return rest.Group{}, errs.ErrEmptyGroupName
	}

	db, err := da.connect(ctx, "gort")
	if err != nil {
		return rest.Group{}, err
	}
	defer db.Close()

	// There will be more fields here eventually
	query := `SELECT groupname
		FROM groups
		WHERE groupname=$1`

	group := rest.Group{}
	err = db.QueryRowContext(ctx, query, groupname).Scan(&group.Name)
	if err != nil {
		return group, gerr.Wrap(errs.ErrNoSuchGroup, err)
	}

	users, err := da.GroupListUsers(ctx, groupname)
	if err != nil {
		return group, err
	}

	group.Users = users

	return group, nil
}

// GroupGrantRole grants one or more roles to a group.
func (da PostgresDataAccess) GroupGrantRole(ctx context.Context) error {
	return errs.ErrNotImplemented
}

// GroupList returns a list of all known groups in the datastore.
// Passwords are not included. Nice try.
func (da PostgresDataAccess) GroupList(ctx context.Context) ([]rest.Group, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.GroupList")
	defer sp.End()

	groups := make([]rest.Group, 0)

	db, err := da.connect(ctx, "gort")
	if err != nil {
		return groups, err
	}
	defer db.Close()

	query := `SELECT groupname FROM groups`
	rows, err := db.QueryContext(ctx, query)
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
func (da PostgresDataAccess) GroupListUsers(ctx context.Context, groupname string) ([]rest.User, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.GroupListUsers")
	defer sp.End()

	users := make([]rest.User, 0)

	db, err := da.connect(ctx, "gort")
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

	rows, err := db.QueryContext(ctx, query, groupname)
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
func (da PostgresDataAccess) GroupRemoveUser(ctx context.Context, groupname string, username string) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.GroupRemoveUser")
	defer sp.End()

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

	db, err := da.connect(ctx, "gort")
	if err != nil {
		return err
	}
	defer db.Close()

	query := "DELETE FROM groupusers WHERE groupname=$1 AND username=$2;"
	_, err = db.ExecContext(ctx, query, groupname, username)
	if err != nil {
		err = gerr.Wrap(errs.ErrDataAccess, err)
	}

	return err
}

// GroupRevokeRole revokes a role from a group.
func (da PostgresDataAccess) GroupRevokeRole(ctx context.Context) error {
	return errs.ErrNotImplemented
}

// GroupUpdate is used to update an existing group. An error is returned if the
// groupname is empty or if the group doesn't exist.
// TODO Should we let this create groups that don't exist?
func (da PostgresDataAccess) GroupUpdate(ctx context.Context, group rest.Group) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.GroupUpdate")
	defer sp.End()

	if group.Name == "" {
		return errs.ErrEmptyGroupName
	}

	exists, err := da.UserExists(ctx, group.Name)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchGroup
	}

	db, err := da.connect(ctx, "gort")
	if err != nil {
		return err
	}
	defer db.Close()

	// There will be more eventually
	query := `UPDATE groupname
	SET groupname=$1
	WHERE groupname=$1;`

	_, err = db.ExecContext(ctx, query, group.Name)
	if err != nil {
		err = gerr.Wrap(errs.ErrDataAccess, err)
	}

	return err
}

// GroupUserList comments TBD
func (da PostgresDataAccess) GroupUserList(ctx context.Context, group string) ([]rest.User, error) {
	return nil, errs.ErrNotImplemented
}

// GroupUserAdd comments TBD
func (da PostgresDataAccess) GroupUserAdd(ctx context.Context, group string, user string) error {
	return errs.ErrNotImplemented
}

// GroupUserDelete comments TBD
func (da PostgresDataAccess) GroupUserDelete(ctx context.Context, group string, user string) error {
	return errs.ErrNotImplemented
}
