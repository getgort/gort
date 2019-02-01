package memory

import (
	"github.com/clockworksoul/cog2/data"
	"github.com/clockworksoul/cog2/dataaccess/errs"
)

// BundleCreate TBD
func (da InMemoryDataAccess) BundleCreate(bundle data.Bundle) error {
	if bundle.Name == "" {
		return errs.ErrEmptyBundleName
	}

	exists, err := da.BundleExists(bundle.Name)
	if err != nil {
		return err
	}
	if exists {
		return errs.ErrBundleExists
	}

	da.bundles[bundle.Name] = &bundle

	return nil
}

// BundleDelete TBD
func (da InMemoryDataAccess) BundleDelete(bundlename string) error {
	if bundlename == "" {
		return errs.ErrEmptyBundleName
	}

	exists, err := da.BundleExists(bundlename)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchBundle
	}

	delete(da.bundles, bundlename)

	return nil
}

// BundleExists TBD
func (da InMemoryDataAccess) BundleExists(bundlename string) (bool, error) {
	_, exists := da.bundles[bundlename]

	return exists, nil
}

// BundleGet TBD
func (da InMemoryDataAccess) BundleGet(bundlename string) (data.Bundle, error) {
	if bundlename == "" {
		return data.Bundle{}, errs.ErrEmptyBundleName
	}

	exists, err := da.BundleExists(bundlename)
	if err != nil {
		return data.Bundle{}, err
	}
	if !exists {
		return data.Bundle{}, errs.ErrNoSuchBundle
	}

	bundle := da.bundles[bundlename]

	return *bundle, nil
}

// BundleList TBD
func (da InMemoryDataAccess) BundleList() ([]data.Bundle, error) {
	list := make([]data.Bundle, 0)

	for _, g := range da.bundles {
		list = append(list, *g)
	}

	return list, nil
}

// BundleUpdate TBD
func (da InMemoryDataAccess) BundleUpdate(bundle data.Bundle) error {
	if bundle.Name == "" {
		return errs.ErrEmptyBundleName
	}

	exists, err := da.BundleExists(bundle.Name)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchBundle
	}

	da.bundles[bundle.Name] = &bundle

	return nil
}
