package client

import (
	"encoding/json"
	"errors"
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

// UserSave comments to be written...
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
	return []rest.Group{}, errors.New("not yet implemented")
}
