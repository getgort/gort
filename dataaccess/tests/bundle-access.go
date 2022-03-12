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

package tests

import (
	"fmt"
	"testing"

	"github.com/getgort/gort/bundles"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/dataaccess/errs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (da DataAccessTester) testBundleAccess(t *testing.T) {
	t.Run("testLoadTestData", da.testLoadTestData)
	t.Run("testBundleCreate", da.testBundleCreate)
	t.Run("testBundleCreateMissingRequired", da.testBundleCreateMissingRequired)
	t.Run("testBundleEnable", da.testBundleEnable)
	t.Run("testBundleEnableTwo", da.testBundleEnableTwo)
	t.Run("testBundleExists", da.testBundleExists)
	t.Run("testBundleVersionExists", da.testBundleVersionExists)
	t.Run("testBundleDelete", da.testBundleDelete)
	t.Run("testBundleDeleteDoesntDisable", da.testBundleDeleteDoesntDisable)
	t.Run("testBundleGet", da.testBundleGet)
	t.Run("testBundleImageConsistency", da.testBundleImageConsistency)
	t.Run("testBundleList", da.testBundleList)
	t.Run("testBundleVersionList", da.testBundleVersionList)
	t.Run("testFindCommandEntry", da.testFindCommandEntry)
}

// Fail-fast: can the test bundle be loaded?
func (da DataAccessTester) testLoadTestData(t *testing.T) {
	b, err := getTestBundle()
	assert.NoError(t, err)

	assert.NotZero(t, b.Templates)

	assert.NotEmpty(t, b.Commands)
	assert.NotEmpty(t, b.Commands["echox"].Description)
	assert.NotEmpty(t, b.Commands["echox"].Executable)
	assert.NotEmpty(t, b.Commands["echox"].LongDescription)
	assert.NotEmpty(t, b.Commands["echox"].Name)
	assert.NotEmpty(t, b.Commands["echox"].Rules)
}

func (da DataAccessTester) testBundleCreate(t *testing.T) {
	// Expect an error
	err := da.BundleCreate(da.ctx, data.Bundle{})
	require.Error(t, err, errs.ErrEmptyBundleName)

	bundle, err := getTestBundle()
	assert.NoError(t, err)
	bundle.Name = "test-create"

	// Expect no error
	err = da.BundleCreate(da.ctx, bundle)
	defer da.BundleDelete(da.ctx, bundle.Name, bundle.Version)
	assert.NoError(t, err)

	// Expect an error
	err = da.BundleCreate(da.ctx, bundle)
	require.Error(t, err, errs.ErrBundleExists)

}

func (da DataAccessTester) testBundleCreateMissingRequired(t *testing.T) {
	bundle, err := getTestBundle()
	assert.NoError(t, err)
	bundle.Name = "test-missing-required"

	defer da.BundleDelete(da.ctx, bundle.Name, bundle.Version)

	// GortBundleVersion
	originalGortBundleVersion := bundle.GortBundleVersion
	bundle.GortBundleVersion = 0
	err = da.BundleCreate(da.ctx, bundle)
	require.Error(t, err, errs.ErrFieldRequired)
	bundle.GortBundleVersion = originalGortBundleVersion

	// Description
	originalDescription := bundle.Description
	bundle.Description = ""
	err = da.BundleCreate(da.ctx, bundle)
	require.Error(t, err, errs.ErrFieldRequired)
	bundle.Description = originalDescription
}

func (da DataAccessTester) testBundleEnable(t *testing.T) {
	bundle, err := getTestBundle()
	assert.NoError(t, err)
	bundle.Name = "test-enable"

	err = da.BundleCreate(da.ctx, bundle)
	assert.NoError(t, err)
	defer da.BundleDelete(da.ctx, bundle.Name, bundle.Version)

	// No version should be enabled
	enabled, err := da.BundleEnabledVersion(da.ctx, bundle.Name)
	assert.NoError(t, err)
	require.Empty(t, enabled)

	// Reload and verify enabled value is false
	bundle, err = da.BundleGet(da.ctx, bundle.Name, bundle.Version)
	assert.NoError(t, err)
	assert.False(t, bundle.Enabled)

	// Enable and verify
	err = da.BundleEnable(da.ctx, bundle.Name, bundle.Version)
	assert.NoError(t, err)

	enabled, err = da.BundleEnabledVersion(da.ctx, bundle.Name)
	assert.NoError(t, err)
	require.Equal(t, enabled, bundle.Version)

	bundle, err = da.BundleGet(da.ctx, bundle.Name, bundle.Version)
	assert.NoError(t, err)
	assert.True(t, bundle.Enabled)

	// Should now delete cleanly
	err = da.BundleDelete(da.ctx, bundle.Name, bundle.Version)
	assert.NoError(t, err)
}

func (da DataAccessTester) testBundleEnableTwo(t *testing.T) {
	bundleA, err := getTestBundle()
	assert.NoError(t, err)
	bundleA.Name = "test-enable-2"
	bundleA.Version = "0.0.1"

	err = da.BundleCreate(da.ctx, bundleA)
	assert.NoError(t, err)
	defer da.BundleDelete(da.ctx, bundleA.Name, bundleA.Version)

	// Enable and verify
	err = da.BundleEnable(da.ctx, bundleA.Name, bundleA.Version)
	assert.NoError(t, err)

	enabled, err := da.BundleEnabledVersion(da.ctx, bundleA.Name)
	assert.NoError(t, err)
	require.Equal(t, enabled, bundleA.Version)

	// Create a new version of the same bundle

	bundleB, err := getTestBundle()
	assert.NoError(t, err)
	bundleB.Name = bundleA.Name
	bundleB.Version = "0.0.2"

	err = da.BundleCreate(da.ctx, bundleB)
	assert.NoError(t, err)
	defer da.BundleDelete(da.ctx, bundleB.Name, bundleB.Version)

	// BundleA should still be enabled

	enabled, err = da.BundleEnabledVersion(da.ctx, bundleA.Name)
	assert.NoError(t, err)
	require.Equal(t, enabled, bundleA.Version)

	enabled, err = da.BundleEnabledVersion(da.ctx, bundleA.Name)
	assert.NoError(t, err)
	require.Equal(t, enabled, bundleA.Version)

	// Enable and verify
	err = da.BundleEnable(da.ctx, bundleB.Name, bundleB.Version)
	assert.NoError(t, err)

	enabled, err = da.BundleEnabledVersion(da.ctx, bundleB.Name)
	assert.NoError(t, err)

	if enabled != bundleB.Version {
		t.Errorf("Bundle should be enabled now. Expected=%q; Got=%q", bundleB.Version, enabled)
		t.FailNow()
	}
}

func (da DataAccessTester) testBundleExists(t *testing.T) {
	var exists bool

	bundle, err := getTestBundle()
	assert.NoError(t, err)
	bundle.Name = "test-exists-version"

	exists, _ = da.BundleExists(da.ctx, bundle.Name)
	if exists {
		t.Error("Bundle should not exist now")
		t.FailNow()
	}

	err = da.BundleCreate(da.ctx, bundle)
	defer da.BundleDelete(da.ctx, bundle.Name, bundle.Version)
	assert.NoError(t, err)

	exists, _ = da.BundleExists(da.ctx, bundle.Name)
	if !exists {
		t.Error("Bundle should exist now")
		t.FailNow()
	}
}

func (da DataAccessTester) testBundleVersionExists(t *testing.T) {
	var exists bool

	bundle, err := getTestBundle()
	assert.NoError(t, err)
	bundle.Name = "test-exists"

	exists, _ = da.BundleVersionExists(da.ctx, bundle.Name, bundle.Version)
	if exists {
		t.Error("Bundle should not exist now")
		t.FailNow()
	}

	err = da.BundleCreate(da.ctx, bundle)
	defer da.BundleDelete(da.ctx, bundle.Name, bundle.Version)
	assert.NoError(t, err)

	exists, _ = da.BundleVersionExists(da.ctx, bundle.Name, bundle.Version)
	if !exists {
		t.Error("Bundle should exist now")
		t.FailNow()
	}
}

func (da DataAccessTester) testBundleDelete(t *testing.T) {
	// Delete blank bundle
	err := da.BundleDelete(da.ctx, "", "0.0.1")
	require.Error(t, err, errs.ErrEmptyBundleName)

	// Delete blank bundle
	err = da.BundleDelete(da.ctx, "foo", "")
	require.Error(t, err, errs.ErrEmptyBundleVersion)

	// Delete bundle that doesn't exist
	err = da.BundleDelete(da.ctx, "no-such-bundle", "0.0.1")
	require.Error(t, err, errs.ErrNoSuchBundle)

	bundle, err := getTestBundle()
	assert.NoError(t, err)
	bundle.Name = "test-delete"

	err = da.BundleCreate(da.ctx, bundle) // This has its own test
	defer da.BundleDelete(da.ctx, bundle.Name, bundle.Version)
	assert.NoError(t, err)

	err = da.BundleDelete(da.ctx, bundle.Name, bundle.Version)
	assert.NoError(t, err)

	exists, _ := da.BundleVersionExists(da.ctx, bundle.Name, bundle.Version)
	if exists {
		t.Error("Shouldn't exist anymore!")
		t.FailNow()
	}
}

func (da DataAccessTester) testBundleDeleteDoesntDisable(t *testing.T) {
	var err error

	bundle, _ := getTestBundle()
	bundle.Name = "test-delete2"
	bundle.Version = "0.0.1"
	err = da.BundleCreate(da.ctx, bundle)
	require.NoError(t, err)
	defer da.BundleDelete(da.ctx, bundle.Name, bundle.Version)

	bundle2, _ := getTestBundle()
	bundle2.Name = "test-delete2"
	bundle2.Version = "0.0.2"
	err = da.BundleCreate(da.ctx, bundle2)
	require.NoError(t, err)
	defer da.BundleDelete(da.ctx, bundle2.Name, bundle2.Version)

	err = da.BundleEnable(da.ctx, bundle2.Name, bundle2.Version)
	require.NoError(t, err)

	err = da.BundleDelete(da.ctx, bundle.Name, bundle.Version)
	require.NoError(t, err)

	bundle2, err = da.BundleGet(da.ctx, bundle2.Name, bundle2.Version)
	require.NoError(t, err)
	assert.True(t, bundle2.Enabled)
}

func (da DataAccessTester) testBundleGet(t *testing.T) {
	var err error

	// Empty bundle name. Expect a ErrEmptyBundleName.
	_, err = da.BundleGet(da.ctx, "", "0.0.1")
	require.Error(t, err, errs.ErrEmptyBundleName)

	// Empty bundle name. Expect a ErrEmptyBundleVersion.
	_, err = da.BundleGet(da.ctx, "test-get", "")
	require.Error(t, err, errs.ErrEmptyBundleVersion)

	// Bundle that doesn't exist. Expect a ErrNoSuchBundle.
	_, err = da.BundleGet(da.ctx, "test-get", "0.0.1")
	require.Error(t, err, errs.ErrNoSuchBundle)

	// Get the test bundle. Expect no error.
	bundleCreate, err := getTestBundle()
	assert.NoError(t, err)

	// Set some values to non-defaults
	bundleCreate.Name = "test-get"
	// bundleCreate.Enabled = true

	// Save the test bundle. Expect no error.
	err = da.BundleCreate(da.ctx, bundleCreate)
	defer da.BundleDelete(da.ctx, bundleCreate.Name, bundleCreate.Version)
	assert.NoError(t, err)

	// Test bundle should now exist in the data store.
	exists, _ := da.BundleVersionExists(da.ctx, bundleCreate.Name, bundleCreate.Version)
	if !exists {
		t.Error("Bundle should exist now, but it doesn't")
		t.FailNow()
	}

	// Load the bundle from the data store. Expect no error
	bundleGet, err := da.BundleGet(da.ctx, bundleCreate.Name, bundleCreate.Version)
	assert.NoError(t, err)

	// This is set automatically on save, so we copy it here for the sake of the tests.
	bundleCreate.InstalledOn = bundleGet.InstalledOn

	assert.Equal(t, bundleCreate.Image, bundleGet.Image)
	assert.ElementsMatch(t, bundleCreate.Permissions, bundleGet.Permissions)
	assert.Equal(t, bundleCreate.Commands, bundleGet.Commands)
	assert.Equal(t, bundleCreate.Kubernetes, bundleGet.Kubernetes)

	// Compare everything for good measure
	assert.Equal(t, bundleCreate, bundleGet)
}

func (da DataAccessTester) testBundleImageConsistency(t *testing.T) {
	tests := []struct {
		B             data.Bundle
		ExpectedImage string
	}{
		{data.Bundle{GortBundleVersion: 1, Name: "test-image-0", Version: "0.0.0", Description: "Foo"}, ""},
		{data.Bundle{GortBundleVersion: 1, Name: "test-image-1", Version: "0.0.1", Description: "Foo", Image: "ubuntu:20.04"}, "ubuntu:20.04"},
		{data.Bundle{GortBundleVersion: 1, Name: "test-image-2", Version: "0.0.2", Description: "Foo", Image: "ubuntu:latest"}, "ubuntu:latest"},
		{data.Bundle{GortBundleVersion: 1, Name: "test-image-3", Version: "0.0.3", Description: "Foo", Image: "ubuntu"}, "ubuntu:latest"},
	}

	const msg = "Test case %d: Name:%q Image:%q"

	for i, test := range tests {
		err := da.BundleCreate(da.ctx, test.B)
		require.NoError(t, err, msg, i, test.B.Name, test.B.Image)
		defer da.BundleDelete(da.ctx, test.B.Name, test.B.Version)

		b, err := da.BundleGet(da.ctx, test.B.Name, test.B.Version)
		require.NoError(t, err, msg, i, test.B.Name, test.B.Image)

		assert.Equal(t, test.ExpectedImage, b.Image)
	}
}

func (da DataAccessTester) testBundleList(t *testing.T) {
	da.BundleCreate(da.ctx, data.Bundle{GortBundleVersion: 5, Name: "test-list-0", Version: "0.0", Description: "foo"})
	defer da.BundleDelete(da.ctx, "test-list-0", "0.0")
	da.BundleCreate(da.ctx, data.Bundle{GortBundleVersion: 5, Name: "test-list-0", Version: "0.1", Description: "foo"})
	defer da.BundleDelete(da.ctx, "test-list-0", "0.1")
	da.BundleCreate(da.ctx, data.Bundle{GortBundleVersion: 5, Name: "test-list-1", Version: "0.0", Description: "foo"})
	defer da.BundleDelete(da.ctx, "test-list-1", "0.0")
	da.BundleCreate(da.ctx, data.Bundle{GortBundleVersion: 5, Name: "test-list-1", Version: "0.1", Description: "foo"})
	defer da.BundleDelete(da.ctx, "test-list-1", "0.1")

	bundles, err := da.BundleList(da.ctx)
	assert.NoError(t, err)
	require.Len(t, bundles, 4)
}

func (da DataAccessTester) testBundleVersionList(t *testing.T) {
	da.BundleCreate(da.ctx, data.Bundle{GortBundleVersion: 5, Name: "test-list-0", Version: "0.0", Description: "foo"})
	defer da.BundleDelete(da.ctx, "test-list-0", "0.0")
	da.BundleCreate(da.ctx, data.Bundle{GortBundleVersion: 5, Name: "test-list-0", Version: "0.1", Description: "foo"})
	defer da.BundleDelete(da.ctx, "test-list-0", "0.1")
	da.BundleCreate(da.ctx, data.Bundle{GortBundleVersion: 5, Name: "test-list-1", Version: "0.0", Description: "foo"})
	defer da.BundleDelete(da.ctx, "test-list-1", "0.0")
	da.BundleCreate(da.ctx, data.Bundle{GortBundleVersion: 5, Name: "test-list-1", Version: "0.1", Description: "foo"})
	defer da.BundleDelete(da.ctx, "test-list-1", "0.1")

	bundles, err := da.BundleVersionList(da.ctx, "test-list-0")
	assert.NoError(t, err)
	require.Len(t, bundles, 2)
}

func (da DataAccessTester) testFindCommandEntry(t *testing.T) {
	const BundleName = "test"
	const BundleVersion = "0.0.1"
	const CommandName = "echox"

	tb, err := getTestBundle()
	assert.NoError(t, err)

	// Save to data store
	err = da.BundleCreate(da.ctx, tb)
	assert.NoError(t, err)

	// Load back from the data store
	tb, err = da.BundleGet(da.ctx, tb.Name, tb.Version)
	assert.NoError(t, err)

	// Sanity testing. Has the test case changed?
	assert.Equal(t, BundleName, tb.Name)
	assert.Equal(t, BundleVersion, tb.Version)
	assert.NotNil(t, tb.Commands[CommandName])

	// Not yet enabled. Should find nothing.
	ce, err := da.FindCommandEntry(da.ctx, BundleName, CommandName)
	assert.NoError(t, err)
	assert.Len(t, ce, 0)

	err = da.BundleEnable(da.ctx, BundleName, BundleVersion)
	assert.NoError(t, err)

	// Reload to capture enabled status
	tb, err = da.BundleGet(da.ctx, tb.Name, tb.Version)
	assert.NoError(t, err)

	// Enabled. Should find commands.
	ce, err = da.FindCommandEntry(da.ctx, BundleName, CommandName)
	assert.NoError(t, err)
	assert.Len(t, ce, 1)

	// Is the loaded bundle correct?
	fmt.Println(ce)
	assert.Equal(t, tb, ce[0].Bundle)

	tc := tb.Commands[CommandName]
	cmd := ce[0].Command
	assert.Equal(t, tc.Description, cmd.Description)
	assert.Equal(t, tc.LongDescription, cmd.LongDescription)
	assert.Equal(t, tc.Executable, cmd.Executable)
	assert.Equal(t, tc.Name, cmd.Name)
	assert.Equal(t, tc.Rules, cmd.Rules)
	assert.Equal(t, tc.Triggers, cmd.Triggers)
}

func getTestBundle() (data.Bundle, error) {
	return bundles.LoadBundleFromFile("../../testing/test-bundle.yml")
}
