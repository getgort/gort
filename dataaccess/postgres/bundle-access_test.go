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

package postgres

import (
	"testing"

	"github.com/getgort/gort/bundles"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/dataaccess/errs"
	"github.com/stretchr/testify/assert"
)

func testBundleAccess(t *testing.T) {
	t.Run("testLoadTestData", testLoadTestData)
	t.Run("testBundleCreate", testBundleCreate)
	t.Run("testBundleCreateMissingRequired", testBundleCreateMissingRequired)
	t.Run("testBundleEnable", testBundleEnable)
	t.Run("testBundleEnableTwo", testBundleEnableTwo)
	t.Run("testBundleExists", testBundleExists)
	t.Run("testBundleDelete", testBundleDelete)
	t.Run("testBundleGet", testBundleGet)
	t.Run("testBundleList", testBundleList)
	t.Run("testBundleVersionList", testBundleVersionList)
	t.Run("testFindCommandEntry", testFindCommandEntry)
}

// Fail-fast: can the test bundle be loaded?
func testLoadTestData(t *testing.T) {
	_, err := getTestBundle()
	assert.NoError(t, err)
}

func testBundleCreate(t *testing.T) {
	// Expect an error
	err := da.BundleCreate(ctx, data.Bundle{})
	if !assert.Error(t, err, errs.ErrEmptyBundleName) {
		t.FailNow()
	}

	bundle, err := getTestBundle()
	assert.NoError(t, err)
	bundle.Name = "test-create"

	// Expect no error
	err = da.BundleCreate(ctx, bundle)
	defer da.BundleDelete(ctx, bundle.Name, bundle.Version)
	assert.NoError(t, err)

	// Expect an error
	err = da.BundleCreate(ctx, bundle)
	if !assert.Error(t, err, errs.ErrBundleExists) {
		t.FailNow()
	}
}

func testBundleCreateMissingRequired(t *testing.T) {
	bundle, err := getTestBundle()
	assert.NoError(t, err)
	bundle.Name = "test-missing-required"

	defer da.BundleDelete(ctx, bundle.Name, bundle.Version)

	// GortBundleVersion
	originalGortBundleVersion := bundle.GortBundleVersion
	bundle.GortBundleVersion = 0
	err = da.BundleCreate(ctx, bundle)
	if !assert.Error(t, err, errs.ErrFieldRequired) {
		t.FailNow()
	}
	bundle.GortBundleVersion = originalGortBundleVersion

	// Description
	originalDescription := bundle.Description
	bundle.Description = ""
	err = da.BundleCreate(ctx, bundle)
	if !assert.Error(t, err, errs.ErrFieldRequired) {
		t.FailNow()
	}
	bundle.Description = originalDescription
}

func testBundleEnable(t *testing.T) {
	bundle, err := getTestBundle()
	assert.NoError(t, err)
	bundle.Name = "test-enable"

	err = da.BundleCreate(ctx, bundle)
	assert.NoError(t, err)
	defer da.BundleDelete(ctx, bundle.Name, bundle.Version)

	// No version should be enabled
	enabled, err := da.BundleEnabledVersion(ctx, bundle.Name)
	assert.NoError(t, err)
	if enabled != "" {
		t.Error("Expected no version to be enabled")
		t.FailNow()
	}

	// Reload and verify enabled value is false
	bundle, err = da.BundleGet(ctx, bundle.Name, bundle.Version)
	assert.NoError(t, err)
	assert.False(t, bundle.Enabled)

	// Enable and verify
	err = da.BundleEnable(ctx, bundle.Name, bundle.Version)
	assert.NoError(t, err)

	enabled, err = da.BundleEnabledVersion(ctx, bundle.Name)
	assert.NoError(t, err)
	if enabled != bundle.Version {
		t.Errorf("Bundle should be enabled now. Expected=%q; Got=%q", bundle.Version, enabled)
		t.FailNow()
	}

	bundle, err = da.BundleGet(ctx, bundle.Name, bundle.Version)
	assert.NoError(t, err)
	assert.True(t, bundle.Enabled)

	// Should now delete cleanly
	err = da.BundleDelete(ctx, bundle.Name, bundle.Version)
	assert.NoError(t, err)
}

func testBundleEnableTwo(t *testing.T) {
	bundleA, err := getTestBundle()
	assert.NoError(t, err)
	bundleA.Name = "test-enable-2"
	bundleA.Version = "0.0.1"

	err = da.BundleCreate(ctx, bundleA)
	assert.NoError(t, err)
	defer da.BundleDelete(ctx, bundleA.Name, bundleA.Version)

	// Enable and verify
	err = da.BundleEnable(ctx, bundleA.Name, bundleA.Version)
	assert.NoError(t, err)

	enabled, err := da.BundleEnabledVersion(ctx, bundleA.Name)
	assert.NoError(t, err)

	if enabled != bundleA.Version {
		t.Errorf("Bundle should be enabled now. Expected=%q; Got=%q", bundleA.Version, enabled)
		t.FailNow()
	}

	// Create a new version of the same bundle

	bundleB, err := getTestBundle()
	assert.NoError(t, err)
	bundleB.Name = bundleA.Name
	bundleB.Version = "0.0.2"

	err = da.BundleCreate(ctx, bundleB)
	assert.NoError(t, err)
	defer da.BundleDelete(ctx, bundleB.Name, bundleB.Version)

	// BundleA should still be enabled

	enabled, err = da.BundleEnabledVersion(ctx, bundleA.Name)
	assert.NoError(t, err)

	if enabled != bundleA.Version {
		t.Errorf("Bundle should be enabled now. Expected=%q; Got=%q", bundleA.Version, enabled)
		t.FailNow()
	}

	enabled, err = da.BundleEnabledVersion(ctx, bundleA.Name)
	assert.NoError(t, err)

	if enabled != bundleA.Version {
		t.Errorf("Bundle should be enabled now. Expected=%q; Got=%q", bundleA.Version, enabled)
		t.FailNow()
	}

	// Enable and verify
	err = da.BundleEnable(ctx, bundleB.Name, bundleB.Version)
	assert.NoError(t, err)

	enabled, err = da.BundleEnabledVersion(ctx, bundleB.Name)
	assert.NoError(t, err)

	if enabled != bundleB.Version {
		t.Errorf("Bundle should be enabled now. Expected=%q; Got=%q", bundleB.Version, enabled)
		t.FailNow()
	}
}

func testBundleExists(t *testing.T) {
	var exists bool

	bundle, err := getTestBundle()
	assert.NoError(t, err)
	bundle.Name = "test-exists"

	exists, _ = da.BundleExists(ctx, bundle.Name, bundle.Version)
	if exists {
		t.Error("Bundle should not exist now")
		t.FailNow()
	}

	err = da.BundleCreate(ctx, bundle)
	defer da.BundleDelete(ctx, bundle.Name, bundle.Version)
	assert.NoError(t, err)

	exists, _ = da.BundleExists(ctx, bundle.Name, bundle.Version)
	if !exists {
		t.Error("Bundle should exist now")
		t.FailNow()
	}
}

func testBundleDelete(t *testing.T) {
	// Delete blank bundle
	err := da.BundleDelete(ctx, "", "0.0.1")
	if !assert.Error(t, err, errs.ErrEmptyBundleName) {
		t.FailNow()
	}

	// Delete blank bundle
	err = da.BundleDelete(ctx, "foo", "")
	if !assert.Error(t, err, errs.ErrEmptyBundleVersion) {
		t.FailNow()
	}

	// Delete bundle that doesn't exist
	err = da.BundleDelete(ctx, "no-such-bundle", "0.0.1")
	if !assert.Error(t, err, errs.ErrNoSuchBundle) {
		t.FailNow()
	}

	bundle, err := getTestBundle()
	assert.NoError(t, err)
	bundle.Name = "test-delete"

	err = da.BundleCreate(ctx, bundle) // This has its own test
	defer da.BundleDelete(ctx, bundle.Name, bundle.Version)
	assert.NoError(t, err)

	err = da.BundleDelete(ctx, bundle.Name, bundle.Version)
	assert.NoError(t, err)

	exists, _ := da.BundleExists(ctx, bundle.Name, bundle.Version)
	if exists {
		t.Error("Shouldn't exist anymore!")
		t.FailNow()
	}
}

func testBundleGet(t *testing.T) {
	var err error

	// Empty bundle name. Expect a ErrEmptyBundleName.
	_, err = da.BundleGet(ctx, "", "0.0.1")
	if !assert.Error(t, err, errs.ErrEmptyBundleName) {
		t.FailNow()
	}

	// Empty bundle name. Expect a ErrEmptyBundleVersion.
	_, err = da.BundleGet(ctx, "test-get", "")
	if !assert.Error(t, err, errs.ErrEmptyBundleVersion) {
		t.FailNow()
	}

	// Bundle that doesn't exist. Expect a ErrNoSuchBundle.
	_, err = da.BundleGet(ctx, "test-get", "0.0.1")
	if !assert.Error(t, err, errs.ErrNoSuchBundle) {
		t.FailNow()
	}

	// Get the test bundle. Expect no error.
	bundleCreate, err := getTestBundle()
	assert.NoError(t, err)

	// Set some values to non-defaults
	bundleCreate.Name = "test-get"
	// bundleCreate.Enabled = true

	// Save the test bundle. Expect no error.
	err = da.BundleCreate(ctx, bundleCreate)
	defer da.BundleDelete(ctx, bundleCreate.Name, bundleCreate.Version)
	assert.NoError(t, err)

	// Test bundle should now exist in the data store.
	exists, _ := da.BundleExists(ctx, bundleCreate.Name, bundleCreate.Version)
	if !exists {
		t.Error("Bundle should exist now, but it doesn't")
		t.FailNow()
	}

	// Load the bundle from the data store. Expect no error
	bundleGet, err := da.BundleGet(ctx, bundleCreate.Name, bundleCreate.Version)
	assert.NoError(t, err)

	// This is set automatically on save, so we copy it here for the sake of the tests.
	bundleCreate.InstalledOn = bundleGet.InstalledOn

	assert.Equal(t, bundleCreate.Docker, bundleGet.Docker)
	assert.ElementsMatch(t, bundleCreate.Permissions, bundleGet.Permissions)
	assert.Equal(t, bundleCreate.Commands, bundleGet.Commands)

	// Compare everything for good measure
	assert.Equal(t, bundleCreate, bundleGet)
}

func testBundleList(t *testing.T) {
	da.BundleCreate(ctx, data.Bundle{GortBundleVersion: 5, Name: "test-list-0", Version: "0.0", Description: "foo"})
	defer da.BundleDelete(ctx, "test-list-0", "0.0")
	da.BundleCreate(ctx, data.Bundle{GortBundleVersion: 5, Name: "test-list-0", Version: "0.1", Description: "foo"})
	defer da.BundleDelete(ctx, "test-list-0", "0.1")
	da.BundleCreate(ctx, data.Bundle{GortBundleVersion: 5, Name: "test-list-1", Version: "0.0", Description: "foo"})
	defer da.BundleDelete(ctx, "test-list-1", "0.0")
	da.BundleCreate(ctx, data.Bundle{GortBundleVersion: 5, Name: "test-list-1", Version: "0.1", Description: "foo"})
	defer da.BundleDelete(ctx, "test-list-1", "0.1")

	bundles, err := da.BundleList(ctx)
	assert.NoError(t, err)

	if len(bundles) != 4 {
		for i, u := range bundles {
			t.Logf("Bundle %d: %v\n", i+1, u)
		}

		t.Errorf("Expected len(bundles) = 4; got %d", len(bundles))
		t.FailNow()
	}
}

func testBundleVersionList(t *testing.T) {
	da.BundleCreate(ctx, data.Bundle{GortBundleVersion: 5, Name: "test-list-0", Version: "0.0", Description: "foo"})
	defer da.BundleDelete(ctx, "test-list-0", "0.0")
	da.BundleCreate(ctx, data.Bundle{GortBundleVersion: 5, Name: "test-list-0", Version: "0.1", Description: "foo"})
	defer da.BundleDelete(ctx, "test-list-0", "0.1")
	da.BundleCreate(ctx, data.Bundle{GortBundleVersion: 5, Name: "test-list-1", Version: "0.0", Description: "foo"})
	defer da.BundleDelete(ctx, "test-list-1", "0.0")
	da.BundleCreate(ctx, data.Bundle{GortBundleVersion: 5, Name: "test-list-1", Version: "0.1", Description: "foo"})
	defer da.BundleDelete(ctx, "test-list-1", "0.1")

	bundles, err := da.BundleVersionList(ctx, "test-list-0")
	assert.NoError(t, err)

	if len(bundles) != 2 {
		for i, u := range bundles {
			t.Logf("Bundle %d: %v\n", i+1, u)
		}

		t.Errorf("Expected len(bundles) = 2; got %d", len(bundles))
		t.FailNow()
	}
}

func testFindCommandEntry(t *testing.T) {
	const BundleName = "test"
	const BundleVersion = "0.0.1"
	const CommandName = "echox"

	tb, err := getTestBundle()
	assert.NoError(t, err)

	// Save to data store
	err = da.BundleCreate(ctx, tb)
	assert.NoError(t, err)

	// Load back from the data store
	tb, err = da.BundleGet(ctx, tb.Name, tb.Version)
	assert.NoError(t, err)

	// Sanity testing. Has the test case changed?
	assert.Equal(t, BundleName, tb.Name)
	assert.Equal(t, BundleVersion, tb.Version)
	assert.NotNil(t, tb.Commands[CommandName])

	// Not yet enabled. Should find nothing.
	ce, err := da.FindCommandEntry(ctx, BundleName, CommandName)
	assert.NoError(t, err)
	assert.Len(t, ce, 0)

	err = da.BundleEnable(ctx, BundleName, BundleVersion)
	assert.NoError(t, err)

	// Reload to capture enabled status
	tb, err = da.BundleGet(ctx, tb.Name, tb.Version)
	assert.NoError(t, err)

	// Enabled. Should find commands.
	ce, err = da.FindCommandEntry(ctx, BundleName, CommandName)
	assert.NoError(t, err)
	assert.Len(t, ce, 1)

	// Is the loaded bundle correct?
	assert.Equal(t, tb, ce[0].Bundle)

	tc := tb.Commands[CommandName]
	cmd := ce[0].Command
	assert.Equal(t, tc.Description, cmd.Description)
	assert.Equal(t, tc.Executable, cmd.Executable)
	assert.Equal(t, tc.Name, cmd.Name)
	assert.Equal(t, tc.Rules, cmd.Rules)
}

func getTestBundle() (data.Bundle, error) {
	return bundles.LoadBundle("../../testing/test-bundle.yml")
}
