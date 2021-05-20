package bundle

import (
	"io/ioutil"

	"github.com/clockworksoul/gort/data"
	gorterr "github.com/clockworksoul/gort/errors"
	yaml "gopkg.in/yaml.v3"
)

// LoadBundle is called by ...
func LoadBundle(file string) (data.Bundle, error) {
	// Read file as a byte slice
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		return data.Bundle{}, gorterr.Wrap(gorterr.ErrIO, err)
	}

	var bundle data.Bundle

	err = yaml.Unmarshal(dat, &bundle)
	if err != nil {
		return data.Bundle{}, gorterr.Wrap(gorterr.ErrUnmarshal, err)
	}

	return bundle, nil
}
