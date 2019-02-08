package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/clockworksoul/cog2/data"
)

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
