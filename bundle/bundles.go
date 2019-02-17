package bundle

import (
	"io/ioutil"

	"github.com/clockworksoul/cog2/data"
	cogerr "github.com/clockworksoul/cog2/errors"
	yaml "gopkg.in/yaml.v2"
)

// LoadBundle is called by ...
func LoadBundle(file string) (data.Bundle, error) {
	// Read file as a byte slice
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		return data.Bundle{}, cogerr.Wrap(cogerr.ErrIO, err)
	}

	var bundle data.Bundle

	err = yaml.Unmarshal(dat, &bundle)
	if err != nil {
		return data.Bundle{}, cogerr.Wrap(cogerr.ErrUnmarshal, err)
	}

	return bundle, nil
}
