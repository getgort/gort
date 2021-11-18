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

package data

import (
	"strings"
	"time"

	"github.com/coreos/go-semver/semver"
)

// BundleInfo wraps a minimal amount of data about a bundle.
type BundleInfo struct {
	Name           string
	Versions       []string
	Enabled        bool
	EnabledVersion Bundle
}

// Bundle represents a bundle as defined in the "bundles" section of the
// config.
type Bundle struct {
	GortBundleVersion int                       `yaml:"gort_bundle_version,omitempty" json:"gort_bundle_version,omitempty"`
	Name              string                    `yaml:",omitempty" json:"name,omitempty"`
	Version           string                    `yaml:",omitempty" json:"version,omitempty"`
	Enabled           bool                      `yaml:",omitempty" json:"enabled"`
	Author            string                    `yaml:",omitempty" json:"author,omitempty"`
	Homepage          string                    `yaml:",omitempty" json:"homepage,omitempty"`
	Description       string                    `yaml:",omitempty" json:"description,omitempty"`
	InstalledOn       time.Time                 `yaml:"-" json:"installed_on,omitempty"`
	InstalledBy       string                    `yaml:",omitempty" json:"installed_by,omitempty"`
	LongDescription   string                    `yaml:"long_description,omitempty" json:"long_description,omitempty"`
	Docker            BundleDocker              `yaml:",omitempty" json:"docker,omitempty"`
	Permissions       []string                  `yaml:",omitempty" json:"permissions,omitempty"`
	Commands          map[string]*BundleCommand `yaml:",omitempty" json:"commands,omitempty"`
	Default           bool                      `yaml:"-" json:"default,omitempty"`
	Templates         Templates                 `yaml:",omitempty" json:"templates,omitempty"`
}

func (b Bundle) Semver() semver.Version {
	var version = b.Version

	if version == "" {
		return semver.Version{}
	}

	if strings.ToLower(version)[0] == 'v' {
		version = version[1:]
	}

	if v, err := semver.NewVersion(version); err != nil {
		return semver.Version{}
	} else {
		return *v
	}
}

// BundleCommand represents a bundle command, as defined in the "bundles/commands"
// section of the config.
type BundleCommand struct {
	Description     string    `yaml:",omitempty" json:"description,omitempty"`
	Executable      []string  `yaml:",omitempty,flow" json:"executable,omitempty"`
	LongDescription string    `yaml:"long_description,omitempty" json:"long_description,omitempty"`
	Name            string    `yaml:"-" json:"-"`
	Rules           []string  `yaml:",omitempty" json:"rules,omitempty"`
	Templates       Templates `yaml:",omitempty" json:"templates,omitempty"`
}

// BundleDocker represents the "bundles/docker" subsection of the config doc
type BundleDocker struct {
	Image string `yaml:",omitempty" json:"image,omitempty"`
	Tag   string `yaml:",omitempty" json:"tag,omitempty"`
}
