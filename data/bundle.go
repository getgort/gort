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
	"time"
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
	GortBundleVersion int                       `yaml:"gort_bundle_version,omitempty" json:",omitempty"`
	Name              string                    `yaml:",omitempty" json:",omitempty"`
	Version           string                    `yaml:",omitempty" json:",omitempty"`
	Enabled           bool                      `yaml:",omitempty" json:""`
	Author            string                    `yaml:",omitempty" json:",omitempty"`
	Homepage          string                    `yaml:",omitempty" json:",omitempty"`
	Description       string                    `yaml:",omitempty" json:",omitempty"`
	InstalledOn       time.Time                 `yaml:"-" json:",omitempty"`
	InstalledBy       string                    `yaml:",omitempty" json:",omitempty"`
	LongDescription   string                    `yaml:"long_description,omitempty" json:",omitempty"`
	Docker            BundleDocker              `yaml:",omitempty" json:",omitempty"`
	Permissions       []string                  `yaml:",omitempty" json:",omitempty"`
	Commands          map[string]*BundleCommand `yaml:",omitempty" json:",omitempty"`
	Default           bool                      `yaml:"-" json:",omitempty"`
	Templates         Templates                 `yaml:",omitempty" json:",omitempty"`
}

// BundleCommand represents a bundle command, as defined in the "bundles/commands"
// section of the config.
type BundleCommand struct {
	Description     string    `yaml:",omitempty" json:",omitempty"`
	Executable      []string  `yaml:",omitempty,flow" json:",omitempty"`
	LongDescription string    `yaml:"long_description,omitempty" json:",omitempty"`
	Name            string    `yaml:"-" json:"-"`
	Rules           []string  `yaml:",omitempty" json:",omitempty"`
	Templates       Templates `yaml:",omitempty" json:",omitempty"`
}

// BundleDocker represents the "bundles/docker" subsection of the config doc
type BundleDocker struct {
	Image string `yaml:",omitempty" json:",omitempty"`
	Tag   string `yaml:",omitempty" json:",omitempty"`
}
