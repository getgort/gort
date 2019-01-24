package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/clockworksoul/cog2/data/rest"
)

// UserList comments to be written...
func (c *CogClient) UserList() ([]rest.User, error) {
	url := fmt.Sprintf("%s/v2/user", c.profile.URL.String())
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

// UserExists simply returns true if a user exists with the specified
// username; false otherwise.
func (c *CogClient) UserExists(username string) (bool, error) {
	url := fmt.Sprintf("%s/v2/user/%s", c.profile.URL.String(), username)
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

// UserGet comments to be written...
func (c *CogClient) UserGet(username string) (rest.User, error) {
	url := fmt.Sprintf("%s/v2/user/%s", c.profile.URL.String(), username)
	resp, err := c.doRequest("GET", url, []byte{})
	if err != nil {
		return rest.User{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return rest.User{}, getResponseError(resp)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return rest.User{}, err
	}

	user := rest.User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		return rest.User{}, err
	}

	return user, nil
}

// UserSave will create or update a user. Note the the key is the username: if
// this is called with a user whose username exists that user is updated;
// otherwise a new user is created.
func (c *CogClient) UserSave(user rest.User) error {
	url := fmt.Sprintf("%s/v2/user/%s", c.profile.URL.String(), user.Username)

	bytes, err := json.Marshal(user)
	if err != nil {
		return err
	}

	resp, err := c.doRequest("POST", url, bytes)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return getResponseError(resp)
	}

	return nil
}

// UserDelete comments to be written...
func (c *CogClient) UserDelete(username string) error {
	url := fmt.Sprintf("%s/v2/user/%s", c.profile.URL.String(), username)

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

// UserGroupList comments to be written...
func (c *CogClient) UserGroupList(username string) ([]rest.Group, error) {
	url := fmt.Sprintf("%s/v2/user/%s/group", c.profile.URL.String(), username)
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

	user := []rest.Group{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		return []rest.Group{}, err
	}

	return user, nil
}
