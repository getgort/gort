package client

import (
	"io/ioutil"
	"net/url"
	"os"

	"github.com/clockworksoul/cog2/data/rest"
	cogerr "github.com/clockworksoul/cog2/errors"
	yaml "gopkg.in/yaml.v3"
)

// Profile represents a set of user profiles from a $HOME/.cog/profiles file
type Profile struct {
	Defaults ProfileDefaults
	Profiles map[string]ProfileEntry `yaml:",inline"`
}

// Default returns this Profile's default entry. If there's no default, or if
// the default doesn't exist, an empty ProfileEntry is returned.
func (p Profile) Default() ProfileEntry {
	entry, ok := p.Profiles[p.Defaults.Profile]

	if ok {
		entry.Name = p.Defaults.Profile
	}

	return entry
}

// ProfileDefaults is used to store default values for a cog client profile.
type ProfileDefaults struct {
	Profile string
}

// ProfileEntry represents a single profile entry.
type ProfileEntry struct {
	Name      string   `yaml:"-"`
	URLString string   `yaml:"url"`
	Password  string   `yaml:"password"`
	URL       *url.URL `yaml:"-"`
	Username  string   `yaml:"user"`
}

// User is a convenience method that returns a rest.User pre-set with the
// entry's username and password.
func (pe ProfileEntry) User() rest.User {
	return rest.User{Password: pe.Password, Username: pe.Username}
}

// loadClientProfile loads and returns the complete client profile. If there's
// no profile file, an empty Profile is returned. An error is returned if
// there's an underlying IO error.
func loadClientProfile() (Profile, error) {
	profile := Profile{Profiles: make(map[string]ProfileEntry)}

	configDir, err := getCogConfigDir()
	if err != nil {
		return Profile{}, err
	}

	profileFile := configDir + "/profile"

	_, err = os.Stat(profileFile)

	// File doesn't exist. Treat as not an "error"
	if err != nil && os.IsNotExist(err) {
		return profile, nil
	}

	// An actual error
	if err != nil {
		return profile, cogerr.Wrap(cogerr.ErrIO, err)
	}

	// The file exists!
	bytes, err := ioutil.ReadFile(profileFile)
	if err != nil {
		return profile, cogerr.Wrap(cogerr.ErrIO, err)
	}

	err = yaml.Unmarshal(bytes, &profile)
	if err != nil {
		return profile, cogerr.Wrap(cogerr.ErrUnmarshal, err)
	}

	// Ensure that the URL field gets set.
	for k, entry := range profile.Profiles {
		if entry.URLString != "" {
			url, err := parseHostURL(entry.URLString)
			if err != nil {
				return profile, err
			}
			entry.URL = url
			profile.Profiles[k] = entry
		}
	}

	return profile, err
}

func saveClientProfile(profile Profile) error {
	configDir, err := getCogConfigDir()
	if err != nil {
		return err
	}

	profileFile := configDir + "/profile"

	bytes, err := yaml.Marshal(profile)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(profileFile, bytes, 0600)
	if err != nil {
		return err
	}

	return nil
}
