package memory

import (
	"fmt"

	"github.com/clockworksoul/cog2/data"
	"github.com/clockworksoul/cog2/dataaccess/errs"
)

// BundleCreate TBD
func (da InMemoryDataAccess) BundleCreate(bundle data.Bundle) error {
	if bundle.Name == "" {
		return errs.ErrEmptyBundleName
	}

	if bundle.CogBundleVersion == 0 || bundle.Version == "" || bundle.Description == "" {
		return errs.ErrFieldRequired
	}

	exists, err := da.BundleExists(bundle.Name, bundle.Version)
	if err != nil {
		return err
	}
	if exists {
		return errs.ErrBundleExists
	}

	da.bundles[bundleKey(bundle.Name, bundle.Version)] = &bundle

	return nil
}

// BundleDelete TBD
func (da InMemoryDataAccess) BundleDelete(name, version string) error {
	if name == "" {
		return errs.ErrEmptyBundleName
	}

	if version == "" {
		return errs.ErrEmptyBundleVersion
	}

	exists, err := da.BundleExists(name, version)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchBundle
	}

	delete(da.bundles, bundleKey(name, version))

	return nil
}

// BundleExists TBD
func (da InMemoryDataAccess) BundleExists(name, version string) (bool, error) {
	_, exists := da.bundles[bundleKey(name, version)]

	return exists, nil
}

// BundleGet TBD
func (da InMemoryDataAccess) BundleGet(name, version string) (data.Bundle, error) {
	if name == "" {
		return data.Bundle{}, errs.ErrEmptyBundleName
	}

	if version == "" {
		return data.Bundle{}, errs.ErrEmptyBundleVersion
	}

	exists, err := da.BundleExists(name, version)
	if err != nil {
		return data.Bundle{}, err
	}
	if !exists {
		return data.Bundle{}, errs.ErrNoSuchBundle
	}

	bundle := da.bundles[bundleKey(name, version)]

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

// BundleListVersions TBD
func (da InMemoryDataAccess) BundleListVersions(name string) ([]data.Bundle, error) {
	list := make([]data.Bundle, 0)

	for _, g := range da.bundles {
		if g.Name == name {
			list = append(list, *g)
		}
	}

	return list, nil
}

// BundleUpdate TBD
func (da InMemoryDataAccess) BundleUpdate(bundle data.Bundle) error {
	if bundle.Name == "" {
		return errs.ErrEmptyBundleName
	}

	if bundle.Version == "" {
		return errs.ErrEmptyBundleVersion
	}

	exists, err := da.BundleExists(bundle.Name, bundle.Version)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchBundle
	}

	da.bundles[bundleKey(bundle.Name, bundle.Version)] = &bundle

	return nil
}

func bundleKey(name, version string) string {
	return fmt.Sprintf("%q::%q", name, version)
}
