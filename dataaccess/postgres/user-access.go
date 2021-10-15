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

	"github.com/getgort/gort/data"
	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/dataaccess/errs"
	gerr "github.com/getgort/gort/errors"
	"github.com/getgort/gort/telemetry"
	"go.opentelemetry.io/otel"
)

// UserAuthenticate authenticates a username/password combination.
func (da PostgresDataAccess) UserAuthenticate(ctx context.Context, username string, password string) (bool, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.UserAuthenticate")
	defer sp.End()

	exists, err := da.UserExists(ctx, username)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, errs.ErrNoSuchUser
	}

	db, err := da.connect(ctx, DatabaseGort)
	if err != nil {
		return false, err
	}
	defer db.Close()

	query := `SELECT password_hash
		FROM users
		WHERE username=$1`

	var hash string
	err = db.QueryRowContext(ctx, query, username).Scan(&hash)
	if err != nil {
		err = gerr.Wrap(errs.ErrNoSuchUser, err)
	}

	return data.CompareHashAndPassword(hash, password), err
}

// UserCreate is used to create a new Gort user in the data store. An error is
// returned if the username is empty or if a user already exists.
func (da PostgresDataAccess) UserCreate(ctx context.Context, user rest.User) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.UserCreate")
	defer sp.End()

	if user.Username == "" {
		return errs.ErrEmptyUserName
	}

	exists, err := da.UserExists(ctx, user.Username)
	if err != nil {
		return err
	}
	if exists {
		return errs.ErrUserExists
	}

	db, err := da.connect(ctx, DatabaseGort)
	if err != nil {
		return err
	}
	defer db.Close()

	var hash string
	if user.Password != "" {
		hash, err = data.HashPassword(user.Password)
		if err != nil {
			return err
		}
	}

	userQuery := `INSERT INTO users (email, full_name, password_hash, username) VALUES ($1, $2, $3, $4);`
	if _, err := db.ExecContext(ctx, userQuery, user.Email, user.FullName, hash, user.Username); err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	adapterIDQuery := `INSERT INTO user_adapter_ids (username, adapter, id) VALUES ($1, $2, $3);`
	for adapter, id := range user.Mappings {
		if _, err := db.ExecContext(ctx, adapterIDQuery, user.Username, adapter, id); err != nil {
			return gerr.Wrap(errs.ErrDataAccess, err)
		}
	}

	return nil
}

// UserDelete deletes an existing user from the data store. An error is
// returned if the username parameter is empty or if the user doesn't
// exist.
func (da PostgresDataAccess) UserDelete(ctx context.Context, username string) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.UserDelete")
	defer sp.End()

	if username == "" {
		return errs.ErrEmptyUserName
	}

	// Thou Shalt Not Delete Admin
	if username == "admin" {
		return errs.ErrAdminUndeletable
	}

	exists, err := da.UserExists(ctx, username)
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

	query := `DELETE FROM groupusers WHERE username=$1;`
	_, err = db.ExecContext(ctx, query, username)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	query = "DELETE FROM tokens WHERE username=$1;"
	_, err = db.ExecContext(ctx, query, username)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	query = "DELETE FROM users WHERE username=$1;"
	_, err = db.ExecContext(ctx, query, username)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

// UserExists is used to determine whether a Gort user with the given username
// exists in the data store.
func (da PostgresDataAccess) UserExists(ctx context.Context, username string) (bool, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.UserExists")
	defer sp.End()

	db, err := da.connect(ctx, DatabaseGort)
	if err != nil {
		return false, err
	}
	defer db.Close()

	query := "SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)"
	exists := false

	err = db.QueryRowContext(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, gerr.Wrap(errs.ErrDataAccess, err)
	}

	return exists, nil
}

// UserGet returns a user from the data store. An error is returned if the
// username parameter is empty or if the user doesn't exist.
func (da PostgresDataAccess) UserGet(ctx context.Context, username string) (rest.User, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.UserGet")
	defer sp.End()

	if username == "" {
		return rest.User{}, errs.ErrEmptyUserName
	}

	db, err := da.connect(ctx, DatabaseGort)
	if err != nil {
		return rest.User{}, err
	}
	defer db.Close()

	query := `SELECT email, full_name, username
		FROM users
		WHERE username=$1`

	var user rest.User

	err = db.QueryRowContext(ctx, query, username).Scan(&user.Email, &user.FullName, &user.Username)
	switch {
	case err == sql.ErrNoRows:
		return rest.User{}, errs.ErrNoSuchUser
	case err != nil:
		return rest.User{}, gerr.Wrap(errs.ErrDataAccess, err)
	}

	if user.Mappings, err = da.doUserGetAdapterIDs(ctx, user.Username); err != nil {
		return rest.User{}, err
	}

	return user, nil
}

// UserGetByEmail returns a user from the data store. An error is returned if
// the email parameter is empty or if the user doesn't exist.
func (da PostgresDataAccess) UserGetByEmail(ctx context.Context, email string) (rest.User, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.UserGetByEmail")
	defer sp.End()

	if email == "" {
		return rest.User{}, errs.ErrEmptyUserEmail
	}

	db, err := da.connect(ctx, DatabaseGort)
	if err != nil {
		return rest.User{}, err
	}
	defer db.Close()

	query := `SELECT email, full_name, username
		FROM users
		WHERE email=$1`

	var user rest.User
	err = db.QueryRowContext(ctx, query, email).Scan(&user.Email, &user.FullName, &user.Username)
	switch {
	case err == sql.ErrNoRows:
		return rest.User{}, errs.ErrNoSuchUser
	case err != nil:
		return rest.User{}, gerr.Wrap(errs.ErrDataAccess, err)
	}

	if user.Mappings, err = da.doUserGetAdapterIDs(ctx, user.Username); err != nil {
		return rest.User{}, err
	}

	return user, nil
}

// UserGetByID returns a user from the data store. An error is returned if
// eitherparameter is empty or if the user doesn't exist.
func (da PostgresDataAccess) UserGetByID(ctx context.Context, adapter, id string) (rest.User, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.UserGetByID")
	defer sp.End()

	if adapter == "" {
		return rest.User{}, errs.ErrEmptyUserAdapter
	}

	if id == "" {
		return rest.User{}, errs.ErrEmptyUserID
	}

	db, err := da.connect(ctx, DatabaseGort)
	if err != nil {
		return rest.User{}, err
	}
	defer db.Close()

	query := `SELECT username
		FROM user_adapter_ids
		WHERE adapter=$1 AND id=$2`

	var username string
	err = db.QueryRowContext(ctx, query, adapter, id).Scan(&username)
	switch {
	case err == nil:
		return da.UserGet(ctx, username)
	case err == sql.ErrNoRows:
		return rest.User{}, errs.ErrNoSuchUser
	default:
		return rest.User{}, gerr.Wrap(errs.ErrDataAccess, err)
	}
}

// UserGroupList returns a slice of Group values representing the specified user's group memberships.
// The groups' Users slice is never populated, and is always nil.
func (da PostgresDataAccess) UserGroupList(ctx context.Context, username string) ([]rest.Group, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.UserGroupList")
	defer sp.End()

	groups := make([]rest.Group, 0)

	db, err := da.connect(ctx, DatabaseGort)
	if err != nil {
		return groups, err
	}
	defer db.Close()

	query := `SELECT groupname FROM groupusers WHERE username=$1`
	rows, err := db.QueryContext(ctx, query, username)
	if err != nil {
		return groups, gerr.Wrap(errs.ErrDataAccess, err)
	}

	for rows.Next() {
		group := rest.Group{}

		err = rows.Scan(&group.Name)
		if err != nil {
			err = gerr.Wrap(errs.ErrDataAccess, err)
		}

		groups = append(groups, group)
	}

	if rows.Err(); err != nil {
		return nil, gerr.Wrap(errs.ErrDataAccess, err)
	}

	return groups, err
}

// UserGroupAdd comments TBD
func (da PostgresDataAccess) UserGroupAdd(ctx context.Context, username string, groupname string) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.UserGroupAdd")
	defer sp.End()

	if username == "" {
		return errs.ErrEmptyUserName
	}

	if groupname == "" {
		return errs.ErrEmptyGroupName
	}

	exists, err := da.UserExists(ctx, username)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchUser
	}

	exists, err = da.GroupExists(ctx, username)
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

	query := `UPDATE groupusers
		SET groupname=$1, username=$2
		WHERE username=$2;`

	_, err = db.ExecContext(ctx, query, groupname, username)
	if err != nil {
		err = gerr.Wrap(errs.ErrDataAccess, err)
	}

	return err
}

// UserGroupDelete comments TBD
func (da PostgresDataAccess) UserGroupDelete(ctx context.Context, username string, groupname string) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.UserGroupDelete")
	defer sp.End()

	if username == "" {
		return errs.ErrEmptyUserName
	}

	if groupname == "" {
		return errs.ErrEmptyGroupName
	}

	exists, err := da.UserExists(ctx, username)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchUser
	}

	exists, err = da.GroupExists(ctx, username)
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

	query := `DELETE FROM groupusers WHERE groupname=$1 AND username=$2;`

	_, err = db.ExecContext(ctx, query, groupname, username)
	if err != nil {
		err = gerr.Wrap(errs.ErrDataAccess, err)
	}

	return err
}

// UserList returns a list of all known users in the datastore.
// Passwords are not included. Nice try.
func (da PostgresDataAccess) UserList(ctx context.Context) ([]rest.User, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.UserList")
	defer sp.End()

	db, err := da.connect(ctx, DatabaseGort)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := `SELECT email, full_name, username FROM users`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]rest.User, 0)

	for rows.Next() {
		user := rest.User{}
		err = rows.Scan(&user.Email, &user.FullName, &user.Username)
		if err != nil {
			err = gerr.Wrap(errs.ErrNoSuchUser, err)
		}
		users = append(users, user)
	}

	if rows.Err(); err != nil {
		return nil, gerr.Wrap(errs.ErrDataAccess, err)
	}

	return users, err
}

// UserPermissionList returns an alphabetically-sorted list of permissions
// available to the specified user.
func (da PostgresDataAccess) UserPermissionList(ctx context.Context, username string) (rest.RolePermissionList, error) {
	// TODO This is HORRIBLY inefficient -- use a real SQL query instead!

	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.UserPermissionList")
	defer sp.End()

	pp := []rest.RolePermission{}

	groups, err := da.UserGroupList(ctx, username)
	if err != nil {
		return nil, err
	}

	for _, group := range groups {
		roles, err := da.GroupRoleList(ctx, group.Name)
		if err != nil {
			return nil, err
		}

		for _, role := range roles {
			rp, err := da.RolePermissionList(ctx, role.Name)
			if err != nil {
				return nil, err
			}

			for _, p := range rp {
				pp = append(pp, p)
			}
		}
	}

	sort.Slice(pp, func(i, j int) bool { return pp[i].String() < pp[j].String() })

	return pp, nil
}

// UserRoleList returns a slice of Role values representing the specified
// user's indirect roles (indirect because users are members of groups,
// and groups have roles).
func (da PostgresDataAccess) UserRoleList(ctx context.Context, username string) ([]rest.Role, error) {
	rm := map[string]rest.Role{}

	groups, err := da.UserGroupList(ctx, username)
	if err != nil {
		return []rest.Role{}, err
	}

	for _, gr := range groups {
		rl, err := da.GroupRoleList(ctx, gr.Name)
		if err != nil {
			return []rest.Role{}, err
		}

		for _, r := range rl {
			rm[r.Name] = r
		}
	}

	roles := []rest.Role{}

	for _, r := range rm {
		roles = append(roles, r)
	}

	sort.Slice(roles, func(i, j int) bool { return roles[i].Name < roles[j].Name })

	return roles, nil
}

// UserUpdate is used to update an existing user. An error is returned if the
// username is empty or if the user doesn't exist.
func (da PostgresDataAccess) UserUpdate(ctx context.Context, user rest.User) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.UserUpdate")
	defer sp.End()

	if user.Username == "" {
		return errs.ErrEmptyUserName
	}

	exists, err := da.UserExists(ctx, user.Username)
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

	query := `SELECT email, full_name, username, password_hash
		FROM users
		WHERE username=$1`

	userOld := rest.User{}
	err = db.
		QueryRowContext(ctx, query, user.Username).
		Scan(&userOld.Email, &userOld.FullName, &userOld.Username, &userOld.Password)

	if err != nil {
		return gerr.Wrap(errs.ErrNoSuchUser, err)
	}

	if user.Email != "" {
		userOld.Email = user.Email
	}

	if user.FullName != "" {
		userOld.FullName = user.FullName
	}

	if user.Password != "" {
		userOld.Password, err = data.HashPassword(user.Password)
		if err != nil {
			return err
		}
	}

	query = `UPDATE users
	SET email=$1, full_name=$2, password_hash=$3
	WHERE username=$4;`

	_, err = db.ExecContext(ctx, query, userOld.Email, userOld.FullName, userOld.Password, userOld.Username)

	if err != nil {
		err = gerr.Wrap(errs.ErrDataAccess, err)
	}

	return err
}

// doUserGetAdapterIDs retrieves any adapter ID mappings associated with the
// user. If none are found, an emptry map is returned.
func (da PostgresDataAccess) doUserGetAdapterIDs(ctx context.Context, username string) (map[string]string, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.doUserGetAdapterIDs")
	defer sp.End()

	db, err := da.connect(ctx, DatabaseGort)
	if err != nil {
		return nil, gerr.Wrap(errs.ErrDataAccess, err)
	}
	defer db.Close()

	query := `SELECT adapter, id
		FROM user_adapter_ids
		WHERE username=$1`

	rows, err := db.QueryContext(ctx, query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var m = map[string]string{}

	for rows.Next() {
		var adapter, id string

		if err := rows.Scan(&adapter, &id); err != nil {
			return m, gerr.Wrap(errs.ErrDataAccess, err)
		}

		m[adapter] = id
	}

	if rows.Err(); err != nil {
		return nil, gerr.Wrap(errs.ErrDataAccess, err)
	}

	return m, nil
}
