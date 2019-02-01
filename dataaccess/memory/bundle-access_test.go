package memory

import (
	"testing"

	"github.com/clockworksoul/cog2/data"
	"github.com/clockworksoul/cog2/dataaccess/errs"
)

func TestBundleExists(t *testing.T) {
	var exists bool

	exists, _ = da.BundleExists("test-exists")
	if exists {
		t.Error("Bundle should not exist now")
	}

	// Now we add a bundle to find.
	da.BundleCreate(data.Bundle{Name: "test-exists"})
	defer da.BundleDelete("test-exists")

	exists, _ = da.BundleExists("test-exists")
	if !exists {
		t.Error("Bundle should exist now")
	}
}

func TestBundleCreate(t *testing.T) {
	var err error
	var bundle data.Bundle

	// Expect an error
	err = da.BundleCreate(bundle)
	expectErr(t, err, errs.ErrEmptyBundleName)

	// Expect no error
	err = da.BundleCreate(data.Bundle{Name: "test-create"})
	defer da.BundleDelete("test-create")
	expectNoErr(t, err)

	// Expect an error
	err = da.BundleCreate(data.Bundle{Name: "test-create"})
	expectErr(t, err, errs.ErrBundleExists)
}

func TestBundleDelete(t *testing.T) {
	// Delete blank bundle
	err := da.BundleDelete("")
	expectErr(t, err, errs.ErrEmptyBundleName)

	// Delete bundle that doesn't exist
	err = da.BundleDelete("no-such-bundle")
	expectErr(t, err, errs.ErrNoSuchBundle)

	da.BundleCreate(data.Bundle{Name: "test-delete"}) // This has its own test
	defer da.BundleDelete("test-delete")

	err = da.BundleDelete("test-delete")
	expectNoErr(t, err)

	exists, _ := da.BundleExists("test-delete")
	if exists {
		t.Error("Shouldn't exist anymore!")
	}
}

func TestBundleGet(t *testing.T) {
	var err error
	var bundle data.Bundle

	// Expect an error
	_, err = da.BundleGet("")
	expectErr(t, err, errs.ErrEmptyBundleName)

	// Expect an error
	_, err = da.BundleGet("test-get")
	expectErr(t, err, errs.ErrNoSuchBundle)

	da.BundleCreate(data.Bundle{Name: "test-get"})
	defer da.BundleDelete("test-get")

	// da.Bundle should exist now
	exists, _ := da.BundleExists("test-get")
	if !exists {
		t.Error("Bundle should exist now")
	}

	// Expect no error
	bundle, err = da.BundleGet("test-get")
	expectNoErr(t, err)
	if bundle.Name != "test-get" {
		t.Errorf("Bundle name mismatch: %q is not \"test-get\"", bundle.Name)
	}
}

func TestBundleList(t *testing.T) {
	da.BundleCreate(data.Bundle{Name: "test-list-0"})
	defer da.BundleDelete("test-list-0")
	da.BundleCreate(data.Bundle{Name: "test-list-1"})
	defer da.BundleDelete("test-list-1")
	da.BundleCreate(data.Bundle{Name: "test-list-2"})
	defer da.BundleDelete("test-list-2")
	da.BundleCreate(data.Bundle{Name: "test-list-3"})
	defer da.BundleDelete("test-list-3")

	bundles, err := da.BundleList()
	expectNoErr(t, err)

	if len(bundles) != 4 {
		t.Errorf("Expected len(bundles) = 4; got %d", len(bundles))
	}

	for _, u := range bundles {
		if u.Name == "" {
			t.Error("Expected non-empty name")
		}
	}
}
