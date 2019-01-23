package client

import (
	"errors"

	"github.com/clockworksoul/cog2/data/rest"
)

// GroupList comments to be written...
func (c *CogClient) GroupList() ([]rest.Group, error) {
	return []rest.Group{}, errors.New("not yet implemented")
}

// GroupGet comments to be written...
func (c *CogClient) GroupGet(groupname string) (rest.Group, error) {
	return rest.Group{}, errors.New("not yet implemented")
}

// GroupSave comments to be written...
func (c *CogClient) GroupSave(group rest.Group) error {
	return errors.New("not yet implemented")
}

// GroupDelete comments to be written...
func (c *CogClient) GroupDelete(groupname string) error {
	return errors.New("not yet implemented")
}

// GroupMemberList comments to be written...
func (c *CogClient) GroupMemberList(groupname string) ([]rest.User, error) {
	return []rest.User{}, errors.New("not yet implemented")
}

// GroupMemberAdd comments to be written...
func (c *CogClient) GroupMemberAdd(groupname string, username string) error {
	return errors.New("not yet implemented")
}

// GroupMemberDelete comments to be written...
func (c *CogClient) GroupMemberDelete(groupname string, username string) error {
	return errors.New("not yet implemented")
}
