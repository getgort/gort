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

package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/getgort/gort/data/rest"
)

// UserDelete comments to be written...
func (c *GortClient) UserDelete(username string) error {
	url := fmt.Sprintf("%s/v2/users/%s", c.profile.URL.String(), username)

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

// UserExists simply returns true if a user exists with the specified
// username; false otherwise.
func (c *GortClient) UserExists(username string) (bool, error) {
	url := fmt.Sprintf("%s/v2/users/%s", c.profile.URL.String(), username)
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
func (c *GortClient) UserGet(username string) (rest.User, error) {
	url := fmt.Sprintf("%s/v2/users/%s", c.profile.URL.String(), username)
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

// UserGroupList comments to be written...
func (c *GortClient) UserGroupList(username string) ([]rest.Group, error) {
	url := fmt.Sprintf("%s/v2/users/%s/groups", c.profile.URL.String(), username)
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

// UserList comments to be written...
func (c *GortClient) UserList() ([]rest.User, error) {
	url := fmt.Sprintf("%s/v2/users", c.profile.URL.String())
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

// UserSave will create or update a user. Note the the key is the username: if
// this is called with a user whose username exists that user is updated
// (empty fields will not be overwritten); otherwise a new user is created.
func (c *GortClient) UserSave(user rest.User) error {
	url := fmt.Sprintf("%s/v2/users/%s", c.profile.URL.String(), user.Username)

	bytes, err := json.Marshal(user)
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
