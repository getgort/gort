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
	"strings"

	"github.com/getgort/gort/data"
)

// DynamicConfigurationDelete
func (c *GortClient) DynamicConfigurationDelete(bundle string, layer data.ConfigurationLayer, owner, key string) error {
	switch {
	case bundle == "":
		return fmt.Errorf("dynamic configuration bundle is required")
	case layer == data.ConfigurationLayer(""):
		return fmt.Errorf("dynamic configuration layer is required")
	case layer.Validate() != nil:
		return layer.Validate()
	case owner == "" && layer != data.LayerBundle:
		return fmt.Errorf("dynamic configuration owner is required for layer %s", layer)
	case key == "":
		return fmt.Errorf("dynamic configuration key is required")
	}

	if owner == "" {
		owner = "-"
	}

	url := fmt.Sprintf("%s/v2/configs/%s/%s/%s/%s", c.profile.URL.String(), bundle, layer, owner, key)
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

// DynamicConfigurationExists
func (c *GortClient) DynamicConfigurationExists(bundle string, layer data.ConfigurationLayer, owner, key string) (bool, error) {
	url := fmt.Sprintf("%s/v2/configs/%s/%s/%s/%s", c.profile.URL.String(), bundle, layer, owner, key)
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

// DynamicConfigurationList
func (c *GortClient) DynamicConfigurationList(bundle string, layer data.ConfigurationLayer, owner, key string) ([]data.DynamicConfiguration, error) {
	p := func(s string) string {
		if s == "" {
			return "*"
		} else {
			return s
		}
	}

	url := fmt.Sprintf("%s/v2/configs/%s/%s/%s/%s", c.profile.URL.String(), p(bundle), p(string(layer)), p(owner), p(key))
	url = strings.TrimRight(url, "/")
	resp, err := c.doRequest("GET", url, []byte{})
	if err != nil {
		return []data.DynamicConfiguration{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []data.DynamicConfiguration{}, getResponseError(resp)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []data.DynamicConfiguration{}, err
	}

	configs := []data.DynamicConfiguration{}
	err = json.Unmarshal(body, &configs)
	if err != nil {
		return []data.DynamicConfiguration{}, err
	}

	return configs, nil
}

// DynamicConfigurationSave
func (c *GortClient) DynamicConfigurationSave(config data.DynamicConfiguration) error {
	switch {
	case config.Bundle == "":
		return fmt.Errorf("dynamic configuration bundle is required")
	case config.Layer == data.ConfigurationLayer(""):
		return fmt.Errorf("dynamic configuration layer is required")
	case config.Layer.Validate() != nil:
		return config.Layer.Validate()
	case config.Owner == "" && config.Layer != data.LayerBundle:
		return fmt.Errorf("dynamic configuration owner is required for layer %s", config.Layer)
	case config.Key == "":
		return fmt.Errorf("dynamic configuration key is required")
	}

	if config.Owner == "" {
		config.Owner = "-"
	}

	url := fmt.Sprintf("%s/v2/configs/%s/%s/%s/%s", c.profile.URL.String(), config.Bundle, config.Layer, config.Owner, config.Key)

	bytes, err := json.Marshal(config)
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
