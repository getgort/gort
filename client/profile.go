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
	"io/ioutil"
	"net/url"
	"os"

	"github.com/getgort/gort/data/rest"
	gerrs "github.com/getgort/gort/errors"

	"gopkg.in/yaml.v3"
)

// Profile represents a set of user profiles from a $HOME/.gort/profiles file
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

// ProfileDefaults is used to store default values for a gort client profile.
type ProfileDefaults struct {
	Profile string
}

// ProfileEntry represents a single profile entry.
type ProfileEntry struct {
	Name          string   `json:"-" yaml:"-"`
	URLString     string   `json:"URL,omitempty" yaml:"url,omitempty"`
	Password      string   `json:",omitempty" yaml:"password,omitempty"`
	URL           *url.URL `json:"-" yaml:"-"`
	Username      string   `json:",omitempty" yaml:"user,omitempty"`
	AllowInsecure bool     `json:",omitempty" yaml:"allow_insecure,omitempty"`
	TLSCertFile   string   `json:",omitempty" yaml:"tls_cert_file,omitempty"`
}

// User is a convenience method that returns a rest.User pre-set with the
// entry's username and password.
func (pe ProfileEntry) User() rest.User {
	return rest.User{Password: pe.Password, Username: pe.Username}
}

// loadClientProfile loads and returns the complete client profile. If there's
// no profile file, an empty Profile is returned. An error is returned if
// there's an underlying IO error.
func LoadClientProfile() (Profile, error) {
	profile := Profile{Profiles: make(map[string]ProfileEntry)}

	configDir, err := getGortConfigDir()
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
		return profile, gerrs.Wrap(gerrs.ErrIO, err)
	}

	// The file exists!
	bytes, err := ioutil.ReadFile(profileFile)
	if err != nil {
		return profile, gerrs.Wrap(gerrs.ErrIO, err)
	}

	err = yaml.Unmarshal(bytes, &profile)
	if err != nil {
		return profile, gerrs.Wrap(gerrs.ErrUnmarshal, err)
	}

	if profile.Profiles == nil {
		profile.Profiles = make(map[string]ProfileEntry)
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

func SaveClientProfile(profile Profile) error {
	configDir, err := getGortConfigDir()
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
