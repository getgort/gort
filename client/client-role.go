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
	"net/http"

	"github.com/getgort/gort/data/rest"
)

// RoleDelete deletes an existing role.
func (c *GortClient) RoleDelete(rolename string) error {
	url := fmt.Sprintf("%s/v2/roles/%s", c.profile.URL.String(), rolename)

	resp, err := c.doRequest("DELETE", url, []byte{})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return getResponseError(resp)
	}

	return nil
}

// RoleCreate creates a new role.
func (c *GortClient) RoleCreate(rolename string) error {
	url := fmt.Sprintf("%s/v2/roles/%s", c.profile.URL.String(), rolename)

	resp, err := c.doRequest("PUT", url, []byte{})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return getResponseError(resp)
	}

	return nil
}

// RoleList comments to be written...
func (c *GortClient) RoleList() ([]rest.Group, error) {
	url := fmt.Sprintf("%s/v2/roles", c.profile.URL.String())
	resp, err := c.doRequest("GET", url, []byte{})
	if err != nil {
		return []rest.Group{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []rest.Group{}, getResponseError(resp)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []rest.Group{}, err
	}

	roles := []rest.Group{}
	err = json.Unmarshal(body, &roles)
	if err != nil {
		return []rest.Group{}, err
	}

	return roles, nil
}

// RoleExists simply returns true if a role exists with the specified
// rolename; false otherwise.
func (c *GortClient) RoleExists(rolename string) (bool, error) {
	url := fmt.Sprintf("%s/v2/roles/%s", c.profile.URL.String(), rolename)
	resp, err := c.doRequest("GET", url, []byte{})
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, getResponseError(resp)
	}
}

// RoleGet gets an existing role.
func (c *GortClient) RoleGet(rolename string) (rest.Role, error) {
	url := fmt.Sprintf("%s/v2/roles/%s", c.profile.URL.String(), rolename)
	resp, err := c.doRequest("GET", url, []byte{})
	if err != nil {
		return rest.Role{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return rest.Role{}, getResponseError(resp)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return rest.Role{}, err
	}

	role := rest.Role{}
	err = json.Unmarshal(body, &role)
	if err != nil {
		return rest.Role{}, err
	}

	return role, nil
}

// RolePermissionList comments to be written...
func (c *GortClient) RolePermissionList(username string) (rest.RolePermissionList, error) {
	url := fmt.Sprintf("%s/v2/roles/%s/permissions", c.profile.URL.String(), username)
	resp, err := c.doRequest("GET", url, []byte{})
	if err != nil {
		return rest.RolePermissionList{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return rest.RolePermissionList{}, getResponseError(resp)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return rest.RolePermissionList{}, err
	}

	rpl := rest.RolePermissionList{}
	err = json.Unmarshal(body, &rpl)
	if err != nil {
		return rest.RolePermissionList{}, err
	}

	return rpl, nil
}

// RolePermissionRevoke revokes an existing permission from a role
func (c *GortClient) RolePermissionRevoke(rolename string, bundlename string, permissionname string) error {
	url := fmt.Sprintf("%s/v2/roles/%s/bundles/%s/permissions/%s", c.profile.URL.String(), rolename, bundlename, permissionname)

	resp, err := c.doRequest("DELETE", url, []byte{})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return getResponseError(resp)
	}

	return nil
}

// RolePermissionGrant grants a permission to an existing role
func (c *GortClient) RolePermissionGrant(rolename string, bundlename string, permissionname string) error {
	url := fmt.Sprintf("%s/v2/roles/%s/bundles/%s/permissions/%s", c.profile.URL.String(), rolename, bundlename, permissionname)

	resp, err := c.doRequest("PUT", url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return getResponseError(resp)
	}

	return nil
}
