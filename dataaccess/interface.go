package dataaccess

import (
	"time"

	"github.com/clockworksoul/gort/data"
	"github.com/clockworksoul/gort/data/rest"
)

// DataAccess represents a common DataAccessObject, backed either by a
// database or an in-memory datastore.
type DataAccess interface {
	Initialize() error

	BundleCreate(bundle data.Bundle) error
	BundleDelete(name string, version string) error
	BundleDisable(name string, version string) error
	BundleEnable(name string, version string) error
	BundleEnabledVersion(name string) (string, error)
	BundleExists(name string, version string) (bool, error)
	BundleGet(name string, version string) (data.Bundle, error)
	BundleList() ([]data.Bundle, error)
	BundleListVersions(name string) ([]data.Bundle, error)
	BundleUpdate(bundle data.Bundle) error

	GroupAddUser(groupname string, username string) error
	GroupCreate(group rest.Group) error
	GroupDelete(groupname string) error
	GroupExists(groupname string) (bool, error)
	GroupGet(groupname string) (rest.Group, error)
	GroupGrantRole() error
	GroupList() ([]rest.Group, error)
	GroupRemoveUser(groupname string, username string) error
	GroupRevokeRole() error
	GroupUpdate(rest.Group) error
	GroupUserList(groupname string) ([]rest.User, error)
	GroupUserAdd(groupname string, user string) error
	GroupUserDelete(groupname string, username string) error

	TokenEvaluate(token string) bool
	TokenGenerate(username string, duration time.Duration) (rest.Token, error)
	TokenInvalidate(token string) error
	TokenRetrieveByUser(username string) (rest.Token, error)
	TokenRetrieveByToken(token string) (rest.Token, error)

	UserAuthenticate(username string, password string) (bool, error)
	UserCreate(user rest.User) error
	UserDelete(username string) error
	UserExists(username string) (bool, error)
	UserGet(username string) (rest.User, error)
	UserGetByEmail(email string) (rest.User, error)
	UserGroupList(username string) ([]rest.Group, error)
	UserGroupAdd(username string, groupname string) error
	UserGroupDelete(username string, groupname string) error
	UserList() ([]rest.User, error)
	UserUpdate(user rest.User) error
}
