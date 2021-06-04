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

	"github.com/getgort/gort/data"
)

// BundleDisable comments to be written...
func (c *GortClient) BundleDisable(bundlename string) error {
	return c.doBundleEnable(bundlename, "-", false)
}

// BundleEnable comments to be written...
func (c *GortClient) BundleEnable(bundlename string, version string) error {
	return c.doBundleEnable(bundlename, version, true)
}

// BundleExists simply returns true if a bundle exists with the specified
// bundlename; false otherwise.
func (c *GortClient) BundleExists(bundlename string, version string) (bool, error) {
	url := fmt.Sprintf("%s/v2/bundles/%s/version/%s",
		c.profile.URL.String(), bundlename, version)

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

// BundleGet comments to be written...
func (c *GortClient) BundleGet(bundlename string, version string) (data.Bundle, error) {
	url := fmt.Sprintf("%s/v2/bundles/%s/versions/%s",
		c.profile.URL.String(), bundlename, version)

	resp, err := c.doRequest("GET", url, []byte{})
	if err != nil {
		return data.Bundle{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return data.Bundle{}, getResponseError(resp)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return data.Bundle{}, err
	}

	bundle := data.Bundle{}
	err = json.Unmarshal(body, &bundle)
	if err != nil {
		return data.Bundle{}, err
	}

	return bundle, nil
}

// BundleList comments to be written...
func (c *GortClient) BundleList() ([]data.Bundle, error) {
	url := fmt.Sprintf("%s/v2/bundles", c.profile.URL.String())

	resp, err := c.doRequest("GET", url, []byte{})
	if err != nil {
		return []data.Bundle{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []data.Bundle{}, getResponseError(resp)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []data.Bundle{}, err
	}

	bundles := []data.Bundle{}
	err = json.Unmarshal(body, &bundles)
	if err != nil {
		return []data.Bundle{}, err
	}

	return bundles, nil
}

// BundleListVersions comments to be written...
func (c *GortClient) BundleListVersions(bundlename string) ([]data.Bundle, error) {
	url := fmt.Sprintf("%s/v2/bundles/%s/versions", c.profile.URL.String(), bundlename)

	resp, err := c.doRequest("GET", url, []byte{})
	if err != nil {
		return []data.Bundle{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []data.Bundle{}, getResponseError(resp)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []data.Bundle{}, err
	}

	bundles := []data.Bundle{}
	err = json.Unmarshal(body, &bundles)
	if err != nil {
		return []data.Bundle{}, err
	}

	return bundles, nil
}

// BundleInstall comments to be written...
func (c *GortClient) BundleInstall(bundle data.Bundle) error {
	url := fmt.Sprintf("%s/v2/bundles/%s/versions/%s",
		c.profile.URL.String(), bundle.Name, bundle.Version)

	bytes, err := json.Marshal(bundle)
	if err != nil {
		return err
	}

	resp, err := c.doRequest("PUT", url, bytes)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return getResponseError(resp)
	}

	return nil
}

// BundleUninstall comments to be written...
func (c *GortClient) BundleUninstall(bundlename string, version string) error {
	url := fmt.Sprintf("%s/v2/bundles/%s/versions/%s",
		c.profile.URL.String(), bundlename, version)

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

// doBundleEnable allows a bundle to be enabled or disabled. The value of
// version is ignored when disabling a bundle.
func (c *GortClient) doBundleEnable(bundlename string, version string, enabled bool) error {
	url := fmt.Sprintf("%s/v2/bundles/%s/versions/%s?enabled=%v",
		c.profile.URL.String(), bundlename, version, enabled)

	// TODO Get latest if version == 'latest'

	resp, err := c.doRequest("PATCH", url, []byte{})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	default:
		return getResponseError(resp)
	}
}
