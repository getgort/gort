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
	"testing"

	"github.com/coreos/go-semver/semver"
	"github.com/stretchr/testify/assert"
)

func TestBundleImageFullParts(t *testing.T) {
	tests := []struct {
		Image        string
		ExpectedRepo string
		ExpectedTag  string
	}{
		{"", "", ""},
		{"ubuntu", "ubuntu", "latest"},
		{"ubuntu:20.04", "ubuntu", "20.04"},
		{"linux:ubuntu:20.04", "linux", "ubuntu:20.04"},
	}

	for _, test := range tests {
		b := Bundle{Image: test.Image}
		repo, tag := b.ImageFullParts()

		assert.Equal(t, test.ExpectedRepo, repo)
		assert.Equal(t, test.ExpectedTag, tag)
	}
}

func TestCoerceVersionToSemver(t *testing.T) {
	tests := []struct {
		Version  string
		Expected string
	}{
		// Zero value
		{"", "0.0.0"},

		// Malicious? edge cases
		{"v", "0.0.0"},
		{"  v1  ", "1.0.0"},

		// Examples from semver.org
		{"1.0.0-alpha", "1.0.0-alpha"},
		{"v1.0.0-alpha", "1.0.0-alpha"},
		{"1.0.0-alpha.1", "1.0.0-alpha.1"},
		{"v1.0.0-alpha.1", "1.0.0-alpha.1"},
		{"1.0.0-0.3.7", "1.0.0-0.3.7"},
		{"v1.0.0-0.3.7", "1.0.0-0.3.7"},
		{"1.0.0-x.7.z.92", "1.0.0-x.7.z.92"},
		{"v1.0.0-x.7.z.92", "1.0.0-x.7.z.92"},
		{"1.0.0-alpha+001", "1.0.0-alpha+001"},
		{"v1.0.0-alpha+001", "1.0.0-alpha+001"},
		{"1.0.0+20130313144700", "1.0.0+20130313144700"},
		{"v1.0.0+20130313144700", "1.0.0+20130313144700"},
		{"1.0.0-beta+exp.sha.5114f85", "1.0.0-beta+exp.sha.5114f85"},
		{"v1.0.0-beta+exp.sha.5114f85", "1.0.0-beta+exp.sha.5114f85"},

		// Version coercion
		{"1", "1.0.0"},
		{"v1", "1.0.0"},
		{"1.2", "1.2.0"},
		{"v1.2", "1.2.0"},
		{"1.2.3", "1.2.3"},
		{"v1.2.3", "1.2.3"},
		{"1.2.3.4", "1.2.3"}, // Truncate to 3 parts.
		{"v1.2.3.4", "1.2.3"},
	}

	for _, test := range tests {
		result := CoerceVersionToSemver(test.Version)
		assert.Equal(t, test.Expected, result, "Test case: %q", test.Version)

		_, err := semver.NewVersion(result)
		assert.NoError(t, err, "Test case: %q", test.Version)
	}
}

func TestSemver(t *testing.T) {
	tests := []struct {
		Version  string
		Expected semver.Version
	}{
		// Zero value
		{"", semver.Version{}},

		// Malicious? edge cases
		{"v", semver.Version{}},
		{"  v1  ", *semver.New("1.0.0")},

		// Examples from *semver.org
		{"1.0.0-alpha", *semver.New("1.0.0-alpha")},
		{"v1.0.0-alpha", *semver.New("1.0.0-alpha")},
		{"1.0.0-alpha.1", *semver.New("1.0.0-alpha.1")},
		{"v1.0.0-alpha.1", *semver.New("1.0.0-alpha.1")},
		{"1.0.0-0.3.7", *semver.New("1.0.0-0.3.7")},
		{"v1.0.0-0.3.7", *semver.New("1.0.0-0.3.7")},
		{"1.0.0-x.7.z.92", *semver.New("1.0.0-x.7.z.92")},
		{"v1.0.0-x.7.z.92", *semver.New("1.0.0-x.7.z.92")},
		{"1.0.0-alpha+001", *semver.New("1.0.0-alpha+001")},
		{"v1.0.0-alpha+001", *semver.New("1.0.0-alpha+001")},
		{"1.0.0+20130313144700", *semver.New("1.0.0+20130313144700")},
		{"v1.0.0+20130313144700", *semver.New("1.0.0+20130313144700")},
		{"1.0.0-beta+exp.sha.5114f85", *semver.New("1.0.0-beta+exp.sha.5114f85")},
		{"v1.0.0-beta+exp.sha.5114f85", *semver.New("1.0.0-beta+exp.sha.5114f85")},

		// Version coercion
		{"1", *semver.New("1.0.0")},
		{"v1", *semver.New("1.0.0")},
		{"1.2", *semver.New("1.2.0")},
		{"v1.2", *semver.New("1.2.0")},
		{"1.2.3", *semver.New("1.2.3")},
		{"v1.2.3", *semver.New("1.2.3")},
		{"1.2.3.4", *semver.New("1.2.3")},
		{"v1.2.3.4", *semver.New("1.2.3")},
	}

	for _, test := range tests {
		b := Bundle{Version: test.Version}
		result := b.Semver()
		assert.Equal(t, test.Expected, result, "Test case: %q", test.Version)
	}
}
