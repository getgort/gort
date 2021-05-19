package memory

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/clockworksoul/cog2/data"
	"github.com/clockworksoul/cog2/dataaccess/errs"
	yaml "gopkg.in/yaml.v3"

	cogerr "github.com/clockworksoul/cog2/errors"
)

func getTestBundle() (data.Bundle, error) {
	bundle := data.Bundle{}

	dat, err := ioutil.ReadFile("../../testing/test-bundle.yml")
	if err != nil {
		return bundle, cogerr.Wrap(cogerr.ErrIO, err)
	}

	err = yaml.Unmarshal(dat, &bundle)
	if err != nil {
		return bundle, cogerr.Wrap(cogerr.ErrUnmarshal, err)
	}

	return bundle, nil
}

func TestLoadTestData(t *testing.T) {
	_, err := getTestBundle()
	expectNoErr(t, err)
}

func TestBundleCreate(t *testing.T) {
	// Expect an error
	err := da.BundleCreate(data.Bundle{})
	expectErr(t, err, errs.ErrEmptyBundleName)

	bundle, err := getTestBundle()
	expectNoErr(t, err)
	bundle.Name = "test-create"

	// Expect no error
	err = da.BundleCreate(bundle)
	defer da.BundleDelete(bundle.Name, bundle.Version)
	expectNoErr(t, err)

	// Expect an error
	err = da.BundleCreate(bundle)
	expectErr(t, err, errs.ErrBundleExists)
}

func TestBundleCreateMissingRequired(t *testing.T) {
	bundle, err := getTestBundle()
	expectNoErr(t, err)
	bundle.Name = "test-missing-required"

	defer da.BundleDelete(bundle.Name, bundle.Version)

	// CogBundleVersion
	originalCogBundleVersion := bundle.CogBundleVersion
	bundle.CogBundleVersion = 0
	err = da.BundleCreate(bundle)
	expectErr(t, err, errs.ErrFieldRequired)
	bundle.CogBundleVersion = originalCogBundleVersion

	// Description
	originalDescription := bundle.Description
	bundle.Description = ""
	err = da.BundleCreate(bundle)
	expectErr(t, err, errs.ErrFieldRequired)
	bundle.Description = originalDescription
}

func TestBundleExists(t *testing.T) {
	var exists bool

	bundle, err := getTestBundle()
	expectNoErr(t, err)
	bundle.Name = "test-exists"

	exists, _ = da.BundleExists(bundle.Name, bundle.Version)
	if exists {
		t.Error("Bundle should not exist now")
	}

	err = da.BundleCreate(bundle)
	defer da.BundleDelete(bundle.Name, bundle.Version)
	expectNoErr(t, err)

	exists, _ = da.BundleExists(bundle.Name, bundle.Version)
	if !exists {
		t.Error("Bundle should exist now")
	}
}

func TestBundleDelete(t *testing.T) {
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
	expectNoErr(t, err)
	bundle.Name = "test-delete"

	err = da.BundleCreate(bundle) // This has its own test
	defer da.BundleDelete(bundle.Name, bundle.Version)
	expectNoErr(t, err)

	err = da.BundleDelete(bundle.Name, bundle.Version)
	expectNoErr(t, err)

	exists, _ := da.BundleExists(bundle.Name, bundle.Version)
	if exists {
		t.Error("Shouldn't exist anymore!")
	}
}

func TestBundleGet(t *testing.T) {
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
	expectNoErr(t, err)
	bundleCreate.Name = "test-get"

	// Set some values to non-defaults
	bundleCreate.Enabled = true

	// Save the test bundle. Expect no error.
	err = da.BundleCreate(bundleCreate)
	defer da.BundleDelete(bundleCreate.Name, bundleCreate.Version)
	expectNoErr(t, err)

	// Test bundle should now exist in the data store.
	exists, _ := da.BundleExists(bundleCreate.Name, bundleCreate.Version)
	if !exists {
		t.Error("Bundle should exist now, but it doesn't")
	}

	// Load the bundle from the data store. Expect no error
	bundleGet, err := da.BundleGet(bundleCreate.Name, bundleCreate.Version)
	expectNoErr(t, err)

	matches, mismatch, expected, got, err := compareFields(
		bundleCreate, bundleGet,
		"CogBundleVersion", "Name", "Version", "Author", "Homepage",
		"Description", "LongDescription", "InstalledBy")
	expectNoErr(t, err)
	if err == nil && !matches {
		t.Errorf("Create/Got mismatch on field %q (expected=%s; got=%s)", mismatch, expected, got)
	}

	matches, mismatch, expected, got, err = compareFields(
		bundleCreate.Docker, bundleGet.Docker,
		"Image", "Tag")
	expectNoErr(t, err)
	if err == nil && !matches {
		t.Errorf("Create/Got mismatch on field %q (expected=%s; got=%s)", mismatch, expected, got)
	}

	err = compareStringSlices(bundleCreate.Permissions, bundleGet.Permissions)
	if err != nil {
		t.Errorf("Create/Got mismatch on permissions: %s", err)
	}

	if len(bundleCreate.Commands) != len(bundleGet.Commands) {
		t.Errorf("Create/Got mismatch on commands")
	}

	for k := range bundleCreate.Commands {
		matches, mismatch, expected, got, err = compareFields(
			bundleCreate.Commands[k], bundleGet.Commands[k],
			"Description", "Executable")
		expectNoErr(t, err)
		if err == nil && !matches {
			t.Errorf("Create/Got mismatch on Command field %q (expected=%s; got=%s)", mismatch, expected, got)
		}

		err = compareStringSlices(bundleCreate.Commands[k].Rules, bundleGet.Commands[k].Rules)
		if err != nil {
			t.Errorf("Create/Got mismatch on command rules: %s", err)
		}
	}
}

func TestBundleList(t *testing.T) {
	da.BundleCreate(data.Bundle{CogBundleVersion: 5, Name: "test-list-0", Version: "0.0", Description: "foo"})
	defer da.BundleDelete("test-list-0", "0.0")
	da.BundleCreate(data.Bundle{CogBundleVersion: 5, Name: "test-list-0", Version: "0.1", Description: "foo"})
	defer da.BundleDelete("test-list-0", "0.1")
	da.BundleCreate(data.Bundle{CogBundleVersion: 5, Name: "test-list-1", Version: "0.0", Description: "foo"})
	defer da.BundleDelete("test-list-1", "0.0")
	da.BundleCreate(data.Bundle{CogBundleVersion: 5, Name: "test-list-1", Version: "0.1", Description: "foo"})
	defer da.BundleDelete("test-list-1", "0.1")

	bundles, err := da.BundleList()
	expectNoErr(t, err)

	if len(bundles) != 4 {
		for i, u := range bundles {
			t.Logf("Bundle %d: %v\n", i+1, u)
		}

		t.Errorf("Expected len(bundles) = 4; got %d", len(bundles))
	}
}

func TestBundleListVersions(t *testing.T) {
	da.BundleCreate(data.Bundle{CogBundleVersion: 5, Name: "test-list-0", Version: "0.0", Description: "foo"})
	defer da.BundleDelete("test-list-0", "0.0")
	da.BundleCreate(data.Bundle{CogBundleVersion: 5, Name: "test-list-0", Version: "0.1", Description: "foo"})
	defer da.BundleDelete("test-list-0", "0.1")
	da.BundleCreate(data.Bundle{CogBundleVersion: 5, Name: "test-list-1", Version: "0.0", Description: "foo"})
	defer da.BundleDelete("test-list-1", "0.0")
	da.BundleCreate(data.Bundle{CogBundleVersion: 5, Name: "test-list-1", Version: "0.1", Description: "foo"})
	defer da.BundleDelete("test-list-1", "0.1")

	bundles, err := da.BundleListVersions("test-list-0")
	expectNoErr(t, err)

	if len(bundles) != 2 {
		for i, u := range bundles {
			t.Logf("Bundle %d: %v\n", i+1, u)
		}

		t.Errorf("Expected len(bundles) = 2; got %d", len(bundles))
	}
}

// Returns: matches?, mismatching field name, expected field value, got field value, error
func compareFields(ob1 interface{}, ob2 interface{}, fields ...string) (bool, string, string, string, error) {
	v1 := reflect.ValueOf(ob1)
	v2 := reflect.ValueOf(ob2)

	for _, fname := range fields {
		f1 := v1.FieldByName(fname)
		if !f1.IsValid() {
			return false, fname, "", "", fmt.Errorf("Type %T has no field %q", ob1, fname)
		}

		f2 := v2.FieldByName(fname)
		if !f2.IsValid() {
			return false, fname, "", "", fmt.Errorf("Type %T has no field %q", ob1, fname)
		}

		if f1.Interface() != f2.Interface() {
			s1 := fmt.Sprintf("%v", f1.Interface())
			s2 := fmt.Sprintf("%v", f2.Interface())
			return false, fname, s1, s2, nil
		}
	}

	return true, "", "", "", nil
}

func compareStringSlices(s1, s2 []string) error {
	if len(s1) != len(s2) {
		return fmt.Errorf("different length slices: %d vs %d", len(s1), len(s2))
	}

	for i := 0; i < len(s1); i++ {
		if s1[i] != s2[i] {
			return fmt.Errorf("value mismatch: %q vs %q", s1[i], s2[i])
		}
	}

	return nil
}
