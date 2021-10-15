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
	"database/sql"
	"sort"

	"go.opentelemetry.io/otel"

	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/dataaccess/errs"
	gerr "github.com/getgort/gort/errors"
	"github.com/getgort/gort/telemetry"
)

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

	db, err := da.connect(ctx, DatabaseGort)
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

	db, err := da.connect(ctx, DatabaseGort)
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

	db, err := da.connect(ctx, DatabaseGort)
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

	db, err := da.connect(ctx, DatabaseGort)
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
	if err == sql.ErrNoRows {
		return group, errs.ErrNoSuchGroup
	} else if err != nil {
		return group, gerr.Wrap(errs.ErrDataAccess, err)
	}

	users, err := da.GroupUserList(ctx, groupname)
	if err != nil {
		return group, err
	}

	group.Users = users

	return group, nil
}

// GroupList returns a list of all known groups in the datastore.
// Passwords are not included. Nice try.
func (da PostgresDataAccess) GroupList(ctx context.Context) ([]rest.Group, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.GroupList")
	defer sp.End()

	groups := make([]rest.Group, 0)

	db, err := da.connect(ctx, DatabaseGort)
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

	if rows.Err(); err != nil {
		return nil, gerr.Wrap(errs.ErrDataAccess, err)
	}

	return groups, nil
}

func (da PostgresDataAccess) GroupPermissionList(ctx context.Context, groupname string) (rest.RolePermissionList, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.GroupPermissionList")
	defer sp.End()

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

// GroupRoleAdd grants one or more roles to a group.
func (da PostgresDataAccess) GroupRoleAdd(ctx context.Context, groupname, rolename string) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.GroupRoleAdd")
	defer sp.End()

	if rolename == "" {
		return errs.ErrEmptyRoleName
	}

	exists, err := da.GroupExists(ctx, groupname)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchGroup
	}

	exists, err = da.RoleExists(ctx, rolename)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchRole
	}

	db, err := da.connect(ctx, DatabaseGort)
	if err != nil {
		return err
	}
	defer db.Close()

	query := `INSERT INTO group_roles (group_name, role_name)
		VALUES ($1, $2);`
	_, err = db.ExecContext(ctx, query, groupname, rolename)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return err
}

// GroupRoleDelete revokes a role from a group.
func (da PostgresDataAccess) GroupRoleDelete(ctx context.Context, groupname, rolename string) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.GroupRoleDelete")
	defer sp.End()

	if groupname == "" {
		return errs.ErrEmptyGroupName
	}

	if rolename == "" {
		return errs.ErrEmptyRoleName
	}

	db, err := da.connect(ctx, DatabaseGort)
	if err != nil {
		return err
	}
	defer db.Close()

	query := `DELETE FROM group_roles
		WHERE group_name=$1 AND role_name=$2;`
	_, err = db.ExecContext(ctx, query, groupname, rolename)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return err
}

func (da PostgresDataAccess) GroupRoleList(ctx context.Context, groupname string) ([]rest.Role, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.GroupRoleList")
	defer sp.End()

	if groupname == "" {
		return nil, errs.ErrEmptyGroupName
	}

	exists, err := da.GroupExists(ctx, groupname)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errs.ErrNoSuchGroup
	}

	db, err := da.connect(ctx, DatabaseGort)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := `SELECT role_name
		FROM group_roles
		WHERE group_name = $1
		ORDER BY role_name`

	rows, err := db.QueryContext(ctx, query, groupname)
	if err != nil {
		return nil, gerr.Wrap(errs.ErrDataAccess, err)
	}

	roles := []rest.Role{}

	for rows.Next() {
		var name string

		err = rows.Scan(&name)
		if err != nil {
			return nil, gerr.Wrap(errs.ErrDataAccess, err)
		}

		role, err := da.RoleGet(ctx, name)
		if err != nil {
			return nil, err
		}

		roles = append(roles, role)
	}

	return roles, nil
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

	db, err := da.connect(ctx, DatabaseGort)
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

// GroupUserAdd adds a user to a group
func (da PostgresDataAccess) GroupUserAdd(ctx context.Context, groupname string, username string) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.GroupUserAdd")
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

	db, err := da.connect(ctx, DatabaseGort)
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

// GroupUserDelete removes a user from a group.
func (da PostgresDataAccess) GroupUserDelete(ctx context.Context, groupname string, username string) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.GroupUserDelete")
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

	db, err := da.connect(ctx, DatabaseGort)
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

// GroupUserList returns a list of all known users in a group.
func (da PostgresDataAccess) GroupUserList(ctx context.Context, groupname string) ([]rest.User, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.GroupUserList")
	defer sp.End()

	users := make([]rest.User, 0)

	if groupname == "" {
		return users, errs.ErrEmptyGroupName
	}

	exists, err := da.GroupExists(ctx, groupname)
	if err != nil {
		return users, err
	}
	if !exists {
		return users, errs.ErrNoSuchGroup
	}

	db, err := da.connect(ctx, DatabaseGort)
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

		user.Mappings, err = da.doUserGetAdapterIDs(ctx, user.Username)
		if err != nil {
			return users, err
		}

		users = append(users, user)
	}

	if rows.Err(); err != nil {
		return nil, gerr.Wrap(errs.ErrDataAccess, err)
	}

	return users, nil
}
