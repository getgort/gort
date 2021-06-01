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
	"fmt"

	"github.com/getgort/gort/data"
	"github.com/getgort/gort/dataaccess/errs"
)

// BundleCreate TBD
func (da *InMemoryDataAccess) BundleCreate(bundle data.Bundle) error {
	if bundle.Name == "" {
		return errs.ErrEmptyBundleName
	}

	if bundle.GortBundleVersion == 0 || bundle.Version == "" || bundle.Description == "" {
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
func (da *InMemoryDataAccess) BundleDelete(name, version string) error {
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

// BundleDisable TBD
func (da *InMemoryDataAccess) BundleDisable(name, version string) error {
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

	bundle := da.bundles[bundleKey(name, version)]
	bundle.Enabled = false

	return nil
}

// BundleEnable TBD
func (da *InMemoryDataAccess) BundleEnable(name, version string) error {
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

	bundle := da.bundles[bundleKey(name, version)]
	bundle.Enabled = true

	return nil
}

// BundleEnabledVersion TBD
func (da *InMemoryDataAccess) BundleEnabledVersion(name string) (string, error) {
	if name == "" {
		return "", errs.ErrEmptyBundleName
	}

	exists := false

	for _, v := range da.bundles {
		if v.Name != name {
			continue
		}

		exists = true

		if v.Enabled {
			return v.Version, nil
		}
	}

	if !exists {
		return "", errs.ErrNoSuchBundle
	}

	return "", nil
}

// BundleExists TBD
func (da *InMemoryDataAccess) BundleExists(name, version string) (bool, error) {
	_, exists := da.bundles[bundleKey(name, version)]

	return exists, nil
}

// BundleGet TBD
func (da *InMemoryDataAccess) BundleGet(name, version string) (data.Bundle, error) {
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
func (da *InMemoryDataAccess) BundleList() ([]data.Bundle, error) {
	list := make([]data.Bundle, 0)

	for _, g := range da.bundles {
		list = append(list, *g)
	}

	return list, nil
}

// BundleListVersions TBD
func (da *InMemoryDataAccess) BundleListVersions(name string) ([]data.Bundle, error) {
	list := make([]data.Bundle, 0)

	for _, g := range da.bundles {
		if g.Name == name {
			list = append(list, *g)
		}
	}

	return list, nil
}

// BundleUpdate TBD
func (da *InMemoryDataAccess) BundleUpdate(bundle data.Bundle) error {
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
