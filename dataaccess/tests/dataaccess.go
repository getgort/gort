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

package tests

import (
	"context"
	"time"

	"github.com/getgort/gort/bundles"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/data/rest"
)

// This is a copy of dataaccess.DataAccess, to break an import cycle.
// DO NOT USE THIS.
type TestDataAccess interface {
	bundles.CommandEntryFinder

	Initialize(context.Context) error

	RequestBegin(ctx context.Context, request *data.CommandRequest) error
	RequestUpdate(ctx context.Context, request data.CommandRequest) error
	RequestError(ctx context.Context, request data.CommandRequest, err error) error
	RequestClose(ctx context.Context, result data.CommandResponseEnvelope) error

	ScheduleCreate(ctx context.Context, command *data.ScheduledCommand) error
	ScheduleDelete(ctx context.Context, command data.ScheduledCommand) error
	SchedulesGet(ctx context.Context) ([]data.ScheduledCommand, error)

	BundleCreate(ctx context.Context, bundle data.Bundle) error
	BundleDelete(ctx context.Context, name string, version string) error
	BundleDisable(ctx context.Context, name string, version string) error
	BundleEnable(ctx context.Context, name string, version string) error
	BundleEnabledVersion(ctx context.Context, name string) (string, error)
	BundleExists(ctx context.Context, name string) (bool, error)
	BundleVersionExists(ctx context.Context, name string, version string) (bool, error)
	BundleGet(ctx context.Context, name string, version string) (data.Bundle, error)
	BundleList(ctx context.Context) ([]data.Bundle, error)
	BundleVersionList(ctx context.Context, name string) ([]data.Bundle, error)
	BundleUpdate(ctx context.Context, bundle data.Bundle) error

	DynamicConfigurationCreate(ctx context.Context, config data.DynamicConfiguration) error
	DynamicConfigurationDelete(ctx context.Context, layer data.ConfigurationLayer, bundle, owner, key string) error
	DynamicConfigurationExists(ctx context.Context, layer data.ConfigurationLayer, bundle, owner, key string) (bool, error)
	DynamicConfigurationGet(ctx context.Context, layer data.ConfigurationLayer, bundle, owner, key string) (data.DynamicConfiguration, error)
	DynamicConfigurationList(ctx context.Context, layer data.ConfigurationLayer, bundle, owner, key string) ([]data.DynamicConfiguration, error)

	GroupCreate(ctx context.Context, group rest.Group) error
	GroupDelete(ctx context.Context, groupname string) error
	GroupExists(ctx context.Context, groupname string) (bool, error)
	GroupGet(ctx context.Context, groupname string) (rest.Group, error)
	GroupList(ctx context.Context) ([]rest.Group, error)
	GroupPermissionList(ctx context.Context, groupname string) (rest.RolePermissionList, error)
	GroupRoleAdd(ctx context.Context, groupname, rolename string) error
	GroupRoleDelete(ctx context.Context, groupname, rolename string) error
	GroupRoleList(ctx context.Context, groupname string) ([]rest.Role, error)
	GroupUpdate(ctx context.Context, group rest.Group) error
	GroupUserAdd(ctx context.Context, groupname string, username string) error
	GroupUserDelete(ctx context.Context, groupname string, username string) error
	GroupUserList(ctx context.Context, groupname string) ([]rest.User, error)

	RoleCreate(ctx context.Context, rolename string) error
	RoleDelete(ctx context.Context, rolename string) error
	RoleGet(ctx context.Context, rolename string) (rest.Role, error)
	RoleGroupAdd(ctx context.Context, rolename, groupname string) error
	RoleGroupDelete(ctx context.Context, rolename, groupname string) error
	RoleGroupExists(ctx context.Context, rolename, groupname string) (bool, error)
	RoleGroupList(ctx context.Context, rolename string) ([]rest.Group, error)
	RoleList(ctx context.Context) ([]rest.Role, error)
	RoleExists(ctx context.Context, rolename string) (bool, error)
	RolePermissionAdd(ctx context.Context, rolename, bundlename, permission string) error
	RolePermissionDelete(ctx context.Context, rolename, bundlename, permission string) error
	RolePermissionExists(ctx context.Context, rolename, bundlename, permission string) (bool, error)
	RolePermissionList(ctx context.Context, rolename string) (rest.RolePermissionList, error)

	TokenEvaluate(ctx context.Context, token string) bool
	TokenGenerate(ctx context.Context, username string, duration time.Duration) (rest.Token, error)
	TokenInvalidate(ctx context.Context, token string) error
	TokenRetrieveByUser(ctx context.Context, username string) (rest.Token, error)
	TokenRetrieveByToken(ctx context.Context, token string) (rest.Token, error)

	UserAuthenticate(ctx context.Context, username string, password string) (bool, error)
	UserCreate(ctx context.Context, user rest.User) error
	UserDelete(ctx context.Context, username string) error
	UserExists(ctx context.Context, username string) (bool, error)
	UserGet(ctx context.Context, username string) (rest.User, error)
	UserGetByEmail(ctx context.Context, email string) (rest.User, error)
	UserGetByID(ctx context.Context, adapter, id string) (rest.User, error)
	UserGroupList(ctx context.Context, username string) ([]rest.Group, error)
	UserGroupAdd(ctx context.Context, username string, groupname string) error
	UserGroupDelete(ctx context.Context, username string, groupname string) error
	UserList(ctx context.Context) ([]rest.User, error)
	UserPermissionList(ctx context.Context, username string) (rest.RolePermissionList, error)
	UserRoleList(ctx context.Context, username string) ([]rest.Role, error)
	UserUpdate(ctx context.Context, user rest.User) error
}
