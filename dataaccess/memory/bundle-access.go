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
	"context"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/getgort/gort/data"
	"github.com/getgort/gort/dataaccess/errs"
)

// BundleCreate TBD
func (da *InMemoryDataAccess) BundleCreate(ctx context.Context, bundle data.Bundle) error {
	if bundle.Name == "" {
		return errs.ErrEmptyBundleName
	}

	if bundle.GortBundleVersion == 0 || bundle.Version == "" || bundle.Description == "" {
		return errs.ErrFieldRequired
	}

	exists, err := da.BundleVersionExists(ctx, bundle.Name, bundle.Version)
	if err != nil {
		return err
	}
	if exists {
		return errs.ErrBundleExists
	}

	for n := range bundle.Commands {
		(bundle.Commands[n]).Name = n
	}

	log.
		WithField("bundle.name", bundle.Name).
		WithField("bundle.commands.length", len(bundle.Commands)).
		WithField("command.names", strings.Join(func() []string {
			var names []string
			for _, c := range bundle.Commands {
				names = append(names, c.Name)
			}
			return names
		}(), ",")).
		Info("Creating bundle")

	bundle.Image = bundle.ImageFull()

	da.bundles[bundleKey(bundle.Name, bundle.Version)] = &bundle

	return nil
}

// BundleDelete TBD
func (da *InMemoryDataAccess) BundleDelete(ctx context.Context, name, version string) error {
	if name == "" {
		return errs.ErrEmptyBundleName
	}

	if version == "" {
		return errs.ErrEmptyBundleVersion
	}

	exists, err := da.BundleVersionExists(ctx, name, version)
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
func (da *InMemoryDataAccess) BundleDisable(ctx context.Context, name, version string) error {
	if name == "" {
		return errs.ErrEmptyBundleName
	}

	foundMatch := false

	for _, b := range da.bundles {
		if b.Name != name || b.Version != version {
			continue
		}

		foundMatch = true
		b.Enabled = false
	}

	if !foundMatch {
		return errs.ErrNoSuchBundle
	}

	return nil
}

// BundleEnable TBD
func (da *InMemoryDataAccess) BundleEnable(ctx context.Context, name, version string) error {
	if name == "" {
		return errs.ErrEmptyBundleName
	}

	if version == "" {
		return errs.ErrEmptyBundleVersion
	}

	exists, err := da.BundleVersionExists(ctx, name, version)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchBundle
	}

	for _, v := range da.bundles {
		if v.Name != name {
			continue
		}

		v.Enabled = (version == v.Version)
	}

	return nil
}

// BundleEnabledVersion TBD
func (da *InMemoryDataAccess) BundleEnabledVersion(ctx context.Context, name string) (string, error) {
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
func (da *InMemoryDataAccess) BundleExists(ctx context.Context, name string) (bool, error) {
	if name == "" {
		return false, errs.ErrEmptyBundleName
	}

	for _, v := range da.bundles {
		if v.Name == name {
			return true, nil
		}
	}

	return false, nil
}

// BundleVersionExists TBD
func (da *InMemoryDataAccess) BundleVersionExists(ctx context.Context, name, version string) (bool, error) {
	if name == "" {
		return false, errs.ErrEmptyBundleName
	}

	if version == "" {
		return false, errs.ErrEmptyBundleVersion
	}

	_, exists := da.bundles[bundleKey(name, version)]

	return exists, nil
}

// BundleGet TBD
func (da *InMemoryDataAccess) BundleGet(ctx context.Context, name, version string) (data.Bundle, error) {
	if name == "" {
		return data.Bundle{}, errs.ErrEmptyBundleName
	}

	if version == "" {
		return data.Bundle{}, errs.ErrEmptyBundleVersion
	}

	exists, err := da.BundleVersionExists(ctx, name, version)
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
func (da *InMemoryDataAccess) BundleList(ctx context.Context) ([]data.Bundle, error) {
	list := make([]data.Bundle, 0)

	for _, g := range da.bundles {
		list = append(list, *g)
	}

	return list, nil
}

// BundleListVersions TBD
func (da *InMemoryDataAccess) BundleVersionList(ctx context.Context, name string) ([]data.Bundle, error) {
	list := make([]data.Bundle, 0)

	for _, g := range da.bundles {
		if g.Name == name {
			list = append(list, *g)
		}
	}

	return list, nil
}

// BundleUpdate TBD
func (da *InMemoryDataAccess) BundleUpdate(ctx context.Context, bundle data.Bundle) error {
	if bundle.Name == "" {
		return errs.ErrEmptyBundleName
	}

	if bundle.Version == "" {
		return errs.ErrEmptyBundleVersion
	}

	exists, err := da.BundleVersionExists(ctx, bundle.Name, bundle.Version)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchBundle
	}

	da.bundles[bundleKey(bundle.Name, bundle.Version)] = &bundle

	return nil
}

// FindCommandEntry is used to find the enabled commands with the provided
// bundle and command names. If either is empty, it is treated as a wildcard.
// Importantly, this must only return ENABLED commands!
func (da *InMemoryDataAccess) FindCommandEntry(ctx context.Context, bundleName, commandName string) ([]data.CommandEntry, error) {
	entries := make([]data.CommandEntry, 0)

	base := log.WithField("search.bundleName", bundleName).WithField("search.commandName", commandName)

	for _, bundle := range da.bundles {
		e := base.WithField("bundle.name", bundle.Name)
		if bundleName != "" && bundleName != bundle.Name {
			e.Debug("bundle name mismatch")
			continue
		}

		if !bundle.Enabled {
			e.Debug("bundle disabled")
			continue
		}

		for _, cmd := range bundle.Commands {
			l := e.WithField("cmd.name", cmd.Name)
			if cmd.Name == commandName {
				l.Debug("match")
				e := data.CommandEntry{Bundle: *bundle, Command: *cmd}
				entries = append(entries, e)
			} else {
				l.Debug("mismatch")
			}
		}
	}

	return entries, nil
}

func (da *InMemoryDataAccess) FindCommandEntryByTrigger(ctx context.Context, tokens []string) ([]data.CommandEntry, error) {
	entries := make([]data.CommandEntry, 0)

	for _, bundle := range da.bundles {
		if !bundle.Enabled {
			continue
		}

		for _, cmd := range bundle.Commands {
			matched, err := cmd.MatchTrigger(ctx, strings.Join(tokens, " "))
			if err != nil {
				return nil, err
			}
			if matched {
				e := data.CommandEntry{Bundle: *bundle, Command: *cmd}
				entries = append(entries, e)
			}
		}
	}

	return entries, nil
}

func bundleKey(name, version string) string {
	return fmt.Sprintf("%q::%q", name, version)
}
