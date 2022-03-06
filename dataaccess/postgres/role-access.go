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
	"log"
	"sort"

	"go.opentelemetry.io/otel"

	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/dataaccess/errs"
	gerr "github.com/getgort/gort/errors"
	"github.com/getgort/gort/telemetry"
)

// RoleCreate creates a new role.
func (da PostgresDataAccess) RoleCreate(ctx context.Context, name string) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.RoleCreate")
	defer sp.End()

	if name == "" {
		return errs.ErrEmptyRoleName
	}

	exists, err := da.RoleExists(ctx, name)
	if err != nil {
		return err
	}
	if exists {
		return errs.ErrRoleExists
	}

	conn, err := da.connect(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	query := `INSERT INTO roles (role_name) VALUES ($1);`
	_, err = conn.ExecContext(ctx, query, name)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return err
}

// RoleDelete
func (da PostgresDataAccess) RoleDelete(ctx context.Context, name string) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.RoleDelete")
	defer sp.End()

	if name == "" {
		return errs.ErrEmptyRoleName
	}

	// Thou Shalt Not Delete Admin
	if name == "admin" {
		return errs.ErrAdminUndeletable
	}

	exists, err := da.RoleExists(ctx, name)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchRole
	}

	conn, err := da.connect(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	query := `DELETE FROM group_roles WHERE group_name=$1;`
	_, err = conn.ExecContext(ctx, query, name)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	query = `DELETE FROM roles WHERE role_name=$1;`
	_, err = conn.ExecContext(ctx, query, name)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

// RoleExists is used to determine whether a group exists in the data store.
func (da PostgresDataAccess) RoleExists(ctx context.Context, rolename string) (bool, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.RoleExists")
	defer sp.End()

	conn, err := da.connect(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	query := "SELECT EXISTS(SELECT 1 FROM roles WHERE role_name=$1)"
	exists := false

	err = conn.QueryRowContext(ctx, query, rolename).Scan(&exists)
	if err != nil {
		return false, gerr.Wrap(errs.ErrNoSuchRole, err)
	}

	return exists, nil
}

// RoleGet gets a specific group.
func (da PostgresDataAccess) RoleGet(ctx context.Context, name string) (rest.Role, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.RoleGet")
	defer sp.End()

	if name == "" {
		return rest.Role{}, errs.ErrEmptyRoleName
	}

	conn, err := da.connect(ctx)
	if err != nil {
		return rest.Role{}, err
	}
	defer conn.Close()

	// There will be more fields here eventually
	query := `SELECT role_name
		FROM roles
		WHERE role_name=$1`

	role := rest.Role{}
	err = conn.QueryRowContext(ctx, query, name).Scan(&role.Name)
	if err != nil {
		return role, gerr.Wrap(errs.ErrNoSuchRole, err)
	}

	perms, err := da.doGetRolePermissions(ctx, name)
	if err != nil {
		return role, err
	}

	role.Permissions = perms

	return role, nil
}

func (da PostgresDataAccess) RoleGroupAdd(ctx context.Context, rolename, groupname string) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.RoleGroupAdd")
	defer sp.End()

	return da.GroupRoleAdd(ctx, groupname, rolename)
}

func (da PostgresDataAccess) RoleGroupDelete(ctx context.Context, rolename, groupname string) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.RoleGroupDelete")
	defer sp.End()

	return da.GroupRoleDelete(ctx, groupname, rolename)
}

func (da PostgresDataAccess) RoleGroupExists(ctx context.Context, rolename, groupname string) (bool, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.RoleGroupExists")
	defer sp.End()

	groups, err := da.RoleGroupList(ctx, rolename)
	if err != nil {
		return false, err
	}

	if exists, err := da.GroupExists(ctx, groupname); err != nil {
		return false, err
	} else if !exists {
		return false, errs.ErrNoSuchGroup
	}

	for _, g := range groups {
		if g.Name == groupname {
			return true, nil
		}
	}

	return false, nil
}

func (da PostgresDataAccess) RoleGroupList(ctx context.Context, rolename string) ([]rest.Group, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.RoleGroupList")
	defer sp.End()

	if rolename == "" {
		return nil, errs.ErrEmptyRoleName
	}

	exists, err := da.RoleExists(ctx, rolename)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errs.ErrNoSuchRole
	}

	conn, err := da.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	query := `SELECT group_name
		FROM group_roles
		WHERE role_name = $1
		ORDER BY role_name`

	rows, err := conn.QueryContext(ctx, query, rolename)
	if err != nil {
		return nil, gerr.Wrap(errs.ErrDataAccess, err)
	}

	groups := []rest.Group{}

	for rows.Next() {
		var name string

		err = rows.Scan(&name)
		if err != nil {
			return nil, gerr.Wrap(errs.ErrDataAccess, err)
		}

		group, err := da.GroupGet(ctx, name)
		if err != nil {
			return nil, err
		}

		groups = append(groups, group)
	}

	return groups, nil
}

// RoleList gets all roles.
func (da PostgresDataAccess) RoleList(ctx context.Context) ([]rest.Role, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.RoleList")
	defer sp.End()

	conn, err := da.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var rolesByName = make(map[string]*rest.Role)
	// Load all role names and add to the roles map
	query := `SELECT role_name
		FROM roles`

	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return nil, gerr.Wrap(errs.ErrNoSuchRole, err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatal(err)
		}
		rolesByName[name] = &rest.Role{Name: name}
	}

	// Load all permissions and add to role objects
	query = `SELECT role_name, bundle_name, permission
		FROM role_permissions`
	rows, err = conn.QueryContext(ctx, query)
	if err != nil {
		return nil, gerr.Wrap(errs.ErrNoSuchRole, err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			rolename   string
			permission rest.RolePermission
		)
		if err := rows.Scan(&rolename, &permission.BundleName, &permission.Permission); err != nil {
			log.Fatal(err)
		}
		rolesByName[rolename].Permissions = append(rolesByName[rolename].Permissions, permission)
	}

	var roles []rest.Role
	for _, role := range rolesByName {
		roles = append(roles, *role)
	}

	return roles, nil
}

func (da PostgresDataAccess) RolePermissionAdd(ctx context.Context, rolename, bundle, permission string) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.RolePermissionAdd")
	defer sp.End()

	if rolename == "" {
		return errs.ErrEmptyRoleName
	}

	if bundle == "" {
		return errs.ErrEmptyBundleName
	}

	if permission == "" {
		return errs.ErrEmptyPermission
	}

	exists, err := da.RoleExists(ctx, rolename)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchRole
	}

	conn, err := da.connect(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	query := `INSERT INTO role_permissions (role_name, bundle_name, permission)
		VALUES ($1, $2, $3);`
	_, err = conn.ExecContext(ctx, query, rolename, bundle, permission)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return err
}

func (da PostgresDataAccess) RolePermissionDelete(ctx context.Context, rolename, bundle, permission string) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.RolePermissionDelete")
	defer sp.End()

	if rolename == "" {
		return errs.ErrEmptyRoleName
	}

	if bundle == "" {
		return errs.ErrEmptyBundleName
	}

	if permission == "" {
		return errs.ErrEmptyPermission
	}

	conn, err := da.connect(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	query := `DELETE FROM role_permissions
		WHERE role_name=$1 AND bundle_name=$2 AND permission=$3;`
	_, err = conn.ExecContext(ctx, query, rolename, bundle, permission)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return err
}

// RolePermissionExists returns true if the given role has been granted the
// specified permission. It returns an error if rolename is empty or if no
// such role exists.
func (da PostgresDataAccess) RolePermissionExists(ctx context.Context, rolename, bundlename, permission string) (bool, error) {
	// TODO Make this more efficient.

	perms, err := da.RolePermissionList(ctx, rolename)
	if err != nil {
		return false, err
	}

	for _, p := range perms {
		if p.BundleName == bundlename && p.Permission == permission {
			return true, nil
		}
	}

	return false, nil
}

// RolePermissionList returns returns an alphabetically-sorted list of
// fully-qualified (i.e., "bundle:permission") permissions granted to
// the role.
func (da PostgresDataAccess) RolePermissionList(ctx context.Context, rolename string) (rest.RolePermissionList, error) {
	// TODO Make this more efficient.

	role, err := da.RoleGet(ctx, rolename)
	if err != nil {
		return nil, err
	}

	perms := role.Permissions

	sort.Slice(perms, func(i, j int) bool { return perms[i].String() < perms[j].String() })

	return perms, nil
}

func (da PostgresDataAccess) doGetRolePermissions(ctx context.Context, name string) (rest.RolePermissionList, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.doGetRolePermissions")
	defer sp.End()

	perms := make([]rest.RolePermission, 0)

	conn, err := da.connect(ctx)
	if err != nil {
		return perms, err
	}
	defer conn.Close()

	query := `SELECT bundle_name, permission
		FROM role_permissions
		WHERE role_name = $1`

	rows, err := conn.QueryContext(ctx, query, name)
	if err != nil {
		return perms, gerr.Wrap(errs.ErrDataAccess, err)
	}

	for rows.Next() {
		perm := rest.RolePermission{}

		err = rows.Scan(&perm.BundleName, &perm.Permission)
		if err != nil {
			return perms, gerr.Wrap(errs.ErrNoSuchUser, err)
		}

		perms = append(perms, perm)
	}

	return perms, nil
}
