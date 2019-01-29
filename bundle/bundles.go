package bundle

import (
	"io/ioutil"

	"github.com/clockworksoul/cog2/data"
	yaml "gopkg.in/yaml.v1"
)

// loadBundle is called by ...
func loadBundle(file string) (data.Bundle, error) {
	// Read file as a byte slice
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		return data.Bundle{}, err
	}

	var bundle data.Bundle

	err = yaml.Unmarshal(dat, &bundle)
	if err != nil {
		return data.Bundle{}, err
	}

	return bundle, nil
}
