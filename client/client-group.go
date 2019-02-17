package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/clockworksoul/cog2/data/rest"
)

// GroupDelete comments to be written...
func (c *CogClient) GroupDelete(groupname string) error {
	url := fmt.Sprintf("%s/v2/groups/%s", c.profile.URL.String(), groupname)

	resp, err := c.doRequest("DELETE", url, []byte{})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return getResponseError(resp)
	}

	return nil
}

// GroupExists simply returns true if a group exists with the specified
// groupname; false otherwise.
func (c *CogClient) GroupExists(groupname string) (bool, error) {
	url := fmt.Sprintf("%s/v2/groups/%s", c.profile.URL.String(), groupname)
	resp, err := c.doRequest("GET", url, []byte{})
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return true, nil
	case 404:
		return false, nil
	default:
		return false, getResponseError(resp)
	}
}

// GroupGet comments to be written...
func (c *CogClient) GroupGet(groupname string) (rest.Group, error) {
	url := fmt.Sprintf("%s/v2/groups/%s", c.profile.URL.String(), groupname)
	resp, err := c.doRequest("GET", url, []byte{})
	if err != nil {
		return rest.Group{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return rest.Group{}, getResponseError(resp)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return rest.Group{}, err
	}

	group := rest.Group{}
	err = json.Unmarshal(body, &group)
	if err != nil {
		return rest.Group{}, err
	}

	return group, nil
}

// GroupList comments to be written...
func (c *CogClient) GroupList() ([]rest.Group, error) {
	url := fmt.Sprintf("%s/v2/groups", c.profile.URL.String())
	resp, err := c.doRequest("GET", url, []byte{})
	if err != nil {
		return []rest.Group{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return []rest.Group{}, getResponseError(resp)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []rest.Group{}, err
	}

	groups := []rest.Group{}
	err = json.Unmarshal(body, &groups)
	if err != nil {
		return []rest.Group{}, err
	}

	return groups, nil
}

// GroupMemberAdd comments to be written...
func (c *CogClient) GroupMemberAdd(groupname string, username string) error {
	url := fmt.Sprintf("%s/v2/groups/%s/members/%s", c.profile.URL.String(), groupname, username)
	resp, err := c.doRequest("PUT", url, []byte{})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return getResponseError(resp)
	}

	return nil
}

// GroupMemberDelete comments to be written...
func (c *CogClient) GroupMemberDelete(groupname string, username string) error {
	url := fmt.Sprintf("%s/v2/groups/%s/members/%s", c.profile.URL.String(), groupname, username)
	resp, err := c.doRequest("DELETE", url, []byte{})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return getResponseError(resp)
	}

	return nil
}

// GroupMemberList comments to be written...
func (c *CogClient) GroupMemberList(groupname string) ([]rest.User, error) {
	url := fmt.Sprintf("%s/v2/groups/%s/members", c.profile.URL.String(), groupname)
	resp, err := c.doRequest("GET", url, []byte{})
	if err != nil {
		return []rest.User{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return []rest.User{}, getResponseError(resp)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []rest.User{}, err
	}

	users := []rest.User{}
	err = json.Unmarshal(body, &users)
	if err != nil {
		return []rest.User{}, err
	}

	return users, nil
}

// GroupSave comments to be written...
func (c *CogClient) GroupSave(group rest.Group) error {
	url := fmt.Sprintf("%s/v2/groups/%s", c.profile.URL.String(), group.Name)

	bytes, err := json.Marshal(group)
	if err != nil {
		return err
	}

	resp, err := c.doRequest("PUT", url, bytes)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return getResponseError(resp)
	}

	return nil
}
