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
	GortBundleVersion int                       `yaml:"gort_bundle_version,omitempty" json:",omitempty"`
	Name              string                    `yaml:",omitempty" json:",omitempty"`
	Version           string                    `yaml:",omitempty" json:",omitempty"`
	Enabled           bool                      `yaml:",omitempty" json:",omitempty"`
	Author            string                    `yaml:",omitempty" json:",omitempty"`
	Homepage          string                    `yaml:",omitempty" json:",omitempty"`
	Description       string                    `yaml:",omitempty" json:",omitempty"`
	Image             string                    `yaml:",omitempty" json:",omitempty"`
	InstalledOn       time.Time                 `yaml:"-" json:",omitempty"`
	InstalledBy       string                    `yaml:",omitempty" json:",omitempty"`
	LongDescription   string                    `yaml:"long_description,omitempty" json:",omitempty"`
	Kubernetes        BundleKubernetes          `yaml:",omitempty" json:",omitempty"`
	Permissions       []string                  `yaml:",omitempty" json:",omitempty"`
	Commands          map[string]*BundleCommand `yaml:",omitempty" json:",omitempty"`
	Default           bool                      `yaml:"-" json:",omitempty"`
	Templates         Templates                 `yaml:",omitempty" json:",omitempty"`
}

// ImageFull returns the full image name, consisting of a repository and tag.
func (b Bundle) ImageFull() string {
	if repo, tag := b.ImageFullParts(); repo != "" {
		return repo + ":" + tag
	}

	return ""
}

// ImageFullParts returns the image repository and tag. If the tag isn't
// specified in b.Image, the returned tag will be "latest".
func (b Bundle) ImageFullParts() (repository, tag string) {
	if b.Image == "" {
		return
	}

	ss := strings.SplitN(b.Image, ":", 2)

	repository = ss[0]

	if len(ss) > 1 {
		tag = ss[1]
	} else {
		tag = "latest"
	}

	return
}

// Semver returns b.Version as a semver.Version value, which makes it easier
// to compare and sort version numbers. If b.Version == "", a zero-value
// Version{} is returned. If b.Version isn't valid per Semantic Versioning
// 2.0.0 (https://semver.org), it will attempt to coerce it into a correct
// semantic version (since users be crazy). If it fails, a zero-value
// Version{} is returned.
func (b Bundle) Semver() semver.Version {
	if v, err := semver.NewVersion(CoerceVersionToSemver(b.Version)); err != nil {
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

// BundleKubernetes represents the "bundles/kubernetes" subsection of the config doc
type BundleKubernetes struct {
	ServiceAccountName string `yaml:"serviceAccountName,omitempty" json:"serviceAccountName,omitempty"`
	EnvSecret          string `yaml:"env_secret,omitempty" json:"env_secret,omitempty"`
}

// CoerceVersionToSemver takes a version number and attempts to coerce it
// into a semver-compliant dotted-tri format. It also understands semver
// pre-release and metadata decorations.
func CoerceVersionToSemver(version string) string {
	version = strings.TrimSpace(version)

	if version == "" {
		return "0.0.0"
	}

	if strings.ToLower(version)[0] == 'v' {
		version = version[1:]
	}

	v := version

	var metadata, preRelease string
	var ss []string
	var dotParts = make([]string, 3)

	ss = strings.SplitN(v, "+", 2)
	if len(ss) > 1 {
		v = ss[0]
		metadata = ss[1]
	}

	ss = strings.SplitN(v, "-", 2)
	if len(ss) > 1 {
		v = ss[0]
		preRelease = ss[1]
	}

	// If it turns out to be in dotted-tri format, return the original
	ss = strings.SplitN(v, ".", 4)
	for i := 0; i < len(ss) && i < 3; i++ {
		dotParts[i] = ss[i]
	}
	for i := 0; i < len(dotParts); i++ {
		if dotParts[i] == "" {
			dotParts[i] = "0"
		}
	}

	v = strings.Join(dotParts, ".")

	if preRelease != "" {
		v += "-" + preRelease
	}

	if metadata != "" {
		v += "+" + metadata
	}

	return v
}
