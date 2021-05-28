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

package memory

import (
	"testing"

	"github.com/clockworksoul/gort/bundle"
	"github.com/clockworksoul/gort/data"
	"github.com/clockworksoul/gort/dataaccess/errs"
	"github.com/stretchr/testify/assert"
)

func testBundleAccess(t *testing.T) {
	t.Run("testLoadTestData", testLoadTestData)
	t.Run("testBundleCreate", testBundleCreate)
	t.Run("testBundleCreateMissingRequired", testBundleCreateMissingRequired)
	t.Run("testBundleEnable", testBundleEnable)
	t.Run("testBundleExists", testBundleExists)
	t.Run("testBundleDelete", testBundleDelete)
	t.Run("testBundleGet", testBundleGet)
	t.Run("testBundleList", testBundleList)
	t.Run("testBundleListVersions", testBundleListVersions)
}

// Fail-fast: can the test bundle be loaded?
func testLoadTestData(t *testing.T) {
	_, err := getTestBundle()
	assert.NoError(t, err)
}

func testBundleCreate(t *testing.T) {
	// Expect an error
	err := da.BundleCreate(data.Bundle{})
	expectErr(t, err, errs.ErrEmptyBundleName)

	bundle, err := getTestBundle()
	assert.NoError(t, err)
	bundle.Name = "test-create"

	// Expect no error
	err = da.BundleCreate(bundle)
	defer da.BundleDelete(bundle.Name, bundle.Version)
	assert.NoError(t, err)

	// Expect an error
	err = da.BundleCreate(bundle)
	expectErr(t, err, errs.ErrBundleExists)
}

func testBundleCreateMissingRequired(t *testing.T) {
	bundle, err := getTestBundle()
	assert.NoError(t, err)
	bundle.Name = "test-missing-required"

	defer da.BundleDelete(bundle.Name, bundle.Version)

	// GortBundleVersion
	originalGortBundleVersion := bundle.GortBundleVersion
	bundle.GortBundleVersion = 0
	err = da.BundleCreate(bundle)
	expectErr(t, err, errs.ErrFieldRequired)
	bundle.GortBundleVersion = originalGortBundleVersion

	// Description
	originalDescription := bundle.Description
	bundle.Description = ""
	err = da.BundleCreate(bundle)
	expectErr(t, err, errs.ErrFieldRequired)
	bundle.Description = originalDescription
}

func testBundleEnable(t *testing.T) {
	bundle, err := getTestBundle()
	assert.NoError(t, err)
	bundle.Name = "test-enable"

	err = da.BundleCreate(bundle)
	assert.NoError(t, err)
	defer da.BundleDelete(bundle.Name, bundle.Version)

	// No version should be enabled
	enabled, err := da.BundleEnabledVersion(bundle.Name)
	assert.NoError(t, err)
	if enabled != "" {
		t.Error("Expected no version to be enabled")
	}

	// Enable and verify
	err = da.BundleEnable(bundle.Name, bundle.Version)
	assert.NoError(t, err)

	enabled, err = da.BundleEnabledVersion(bundle.Name)
	assert.NoError(t, err)
	if enabled != bundle.Version {
		t.Errorf("Bundle should be enabled now. Expected=%q; Got=%q",
			bundle.Version, enabled)
		t.FailNow()
	}

	// Should now delete cleanly
	err = da.BundleDelete(bundle.Name, bundle.Version)
	assert.NoError(t, err)
}

func testBundleExists(t *testing.T) {
	var exists bool

	bundle, err := getTestBundle()
	assert.NoError(t, err)
	bundle.Name = "test-exists"

	exists, _ = da.BundleExists(bundle.Name, bundle.Version)
	if exists {
		t.Error("Bundle should not exist now")
	}

	err = da.BundleCreate(bundle)
	defer da.BundleDelete(bundle.Name, bundle.Version)
	assert.NoError(t, err)

	exists, _ = da.BundleExists(bundle.Name, bundle.Version)
	if !exists {
		t.Error("Bundle should exist now")
	}
}

func testBundleDelete(t *testing.T) {
	// Delete blank bundle
	err := da.BundleDelete("", "0.0.1")
	expectErr(t, err, errs.ErrEmptyBundleName)

	// Delete blank bundle
	err = da.BundleDelete("foo", "")
	expectErr(t, err, errs.ErrEmptyBundleVersion)

	// Delete bundle that doesn't exist
	err = da.BundleDelete("no-such-bundle", "0.0.1")
	expectErr(t, err, errs.ErrNoSuchBundle)

	bundle, err := getTestBundle()
	assert.NoError(t, err)
	bundle.Name = "test-delete"

	err = da.BundleCreate(bundle) // This has its own test
	defer da.BundleDelete(bundle.Name, bundle.Version)
	assert.NoError(t, err)

	err = da.BundleDelete(bundle.Name, bundle.Version)
	assert.NoError(t, err)

	exists, _ := da.BundleExists(bundle.Name, bundle.Version)
	if exists {
		t.Error("Shouldn't exist anymore!")
	}
}

func testBundleGet(t *testing.T) {
	var err error

	// Empty bundle name. Expect a ErrEmptyBundleName.
	_, err = da.BundleGet("", "0.0.1")
	expectErr(t, err, errs.ErrEmptyBundleName)

	// Empty bundle name. Expect a ErrEmptyBundleVersion.
	_, err = da.BundleGet("test-get", "")
	expectErr(t, err, errs.ErrEmptyBundleVersion)

	// Bundle that doesn't exist. Expect a ErrNoSuchBundle.
	_, err = da.BundleGet("test-get", "0.0.1")
	expectErr(t, err, errs.ErrNoSuchBundle)

	// Get the test bundle. Expect no error.
	bundleCreate, err := getTestBundle()
	assert.NoError(t, err)

	// Set some values to non-defaults
	bundleCreate.Name = "test-get"
	// bundleCreate.Enabled = true

	// Save the test bundle. Expect no error.
	err = da.BundleCreate(bundleCreate)
	defer da.BundleDelete(bundleCreate.Name, bundleCreate.Version)
	assert.NoError(t, err)

	// Test bundle should now exist in the data store.
	exists, _ := da.BundleExists(bundleCreate.Name, bundleCreate.Version)
	if !exists {
		t.Error("Bundle should exist now, but it doesn't")
	}

	// Load the bundle from the data store. Expect no error
	bundleGet, err := da.BundleGet(bundleCreate.Name, bundleCreate.Version)
	assert.NoError(t, err)

	// This is set automatically on save, so we copy it here for the sake of the tests.
	bundleCreate.InstalledOn = bundleGet.InstalledOn

	assert.Equal(t, bundleCreate, bundleGet)
	assert.Equal(t, bundleCreate.Docker, bundleGet.Docker)
	assert.ElementsMatch(t, bundleCreate.Permissions, bundleGet.Permissions)
	assert.Equal(t, bundleCreate.Commands, bundleGet.Commands)
}

func testBundleList(t *testing.T) {
	da.BundleCreate(data.Bundle{GortBundleVersion: 5, Name: "test-list-0", Version: "0.0", Description: "foo"})
	defer da.BundleDelete("test-list-0", "0.0")
	da.BundleCreate(data.Bundle{GortBundleVersion: 5, Name: "test-list-0", Version: "0.1", Description: "foo"})
	defer da.BundleDelete("test-list-0", "0.1")
	da.BundleCreate(data.Bundle{GortBundleVersion: 5, Name: "test-list-1", Version: "0.0", Description: "foo"})
	defer da.BundleDelete("test-list-1", "0.0")
	da.BundleCreate(data.Bundle{GortBundleVersion: 5, Name: "test-list-1", Version: "0.1", Description: "foo"})
	defer da.BundleDelete("test-list-1", "0.1")

	bundles, err := da.BundleList()
	assert.NoError(t, err)

	if len(bundles) != 4 {
		for i, u := range bundles {
			t.Logf("Bundle %d: %v\n", i+1, u)
		}

		t.Errorf("Expected len(bundles) = 4; got %d", len(bundles))
	}
}

func testBundleListVersions(t *testing.T) {
	da.BundleCreate(data.Bundle{GortBundleVersion: 5, Name: "test-list-0", Version: "0.0", Description: "foo"})
	defer da.BundleDelete("test-list-0", "0.0")
	da.BundleCreate(data.Bundle{GortBundleVersion: 5, Name: "test-list-0", Version: "0.1", Description: "foo"})
	defer da.BundleDelete("test-list-0", "0.1")
	da.BundleCreate(data.Bundle{GortBundleVersion: 5, Name: "test-list-1", Version: "0.0", Description: "foo"})
	defer da.BundleDelete("test-list-1", "0.0")
	da.BundleCreate(data.Bundle{GortBundleVersion: 5, Name: "test-list-1", Version: "0.1", Description: "foo"})
	defer da.BundleDelete("test-list-1", "0.1")

	bundles, err := da.BundleListVersions("test-list-0")
	assert.NoError(t, err)

	if len(bundles) != 2 {
		for i, u := range bundles {
			t.Logf("Bundle %d: %v\n", i+1, u)
		}

		t.Errorf("Expected len(bundles) = 2; got %d", len(bundles))
	}
}

func getTestBundle() (data.Bundle, error) {
	return bundle.LoadBundle("../../testing/test-bundle.yml")
}
