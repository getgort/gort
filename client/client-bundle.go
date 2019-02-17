package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/clockworksoul/cog2/data"
)

// BundleExists simply returns true if a bundle exists with the specified
// bundlename; false otherwise.
func (c *CogClient) BundleExists(bundlename string, version string) (bool, error) {
	url := fmt.Sprintf("%s/v2/bundles/%s/version/%s",
		c.profile.URL.String(), bundlename, version)
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

// BundleList comments to be written...
func (c *CogClient) BundleList() ([]data.Bundle, error) {
	url := fmt.Sprintf("%s/v2/bundles", c.profile.URL.String())
	resp, err := c.doRequest("GET", url, []byte{})
	if err != nil {
		return []data.Bundle{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
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

// BundleSave comments to be written...
func (c *CogClient) BundleSave(bundle data.Bundle) error {
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

	if resp.StatusCode != 200 {
		return getResponseError(resp)
	}

	return nil
}
