package client

import (
	"errors"

	"github.com/clockworksoul/cog2/data/rest"
)

// UserList comments to be written...
func (c *CogClient) UserList() ([]rest.User, error) {
	return []rest.User{}, errors.New("not yet implemented")
}

// UserGet comments to be written...
func (c *CogClient) UserGet(username string) (rest.User, error) {
	return rest.User{}, errors.New("not yet implemented")
}

// UserSave comments to be written...
func (c *CogClient) UserSave(user rest.User) error {
	return errors.New("not yet implemented")
}

// UserDelete comments to be written...
func (c *CogClient) UserDelete(username string) error {
	return errors.New("not yet implemented")
}

// UserGroupList comments to be written...
func (c *CogClient) UserGroupList(username string) ([]rest.Group, error) {
	return []rest.Group{}, errors.New("not yet implemented")
}
