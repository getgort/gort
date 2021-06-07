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
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/getgort/gort/data"
	"github.com/getgort/gort/dataaccess/errs"
	gerr "github.com/getgort/gort/errors"
)

type bundleData struct {
	BundleName    string
	BundleVersion string
}

type bundleCommandData struct {
	data.BundleCommand
	bundleData
}

// BundleCreate TBD
func (da PostgresDataAccess) BundleCreate(ctx context.Context, bundle data.Bundle) error {
	if bundle.Name == "" {
		return errs.ErrEmptyBundleName
	}

	if bundle.Version == "" {
		return errs.ErrEmptyBundleVersion
	}

	db, err := da.connect("gort")
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	exists, err := da.doBundleExists(tx, bundle.Name, bundle.Version)
	if err != nil {
		return err
	} else if exists {
		return errs.ErrBundleExists
	}

	// Save bundle
	err = da.doBundleInsert(tx, bundle)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Save permissions
	err = da.doBundleInsertPermissions(tx, bundle)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Save commands
	err = da.doBundleInsertCommands(tx, bundle)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return err
}

// BundleDelete TBD
func (da PostgresDataAccess) BundleDelete(ctx context.Context, name, version string) error {
	if name == "" {
		return errs.ErrEmptyBundleName
	}

	if version == "" {
		return errs.ErrEmptyBundleVersion
	}

	db, err := da.connect("gort")
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	exists, err := da.doBundleExists(tx, name, version)
	if err != nil {
		return err
	} else if !exists {
		return errs.ErrNoSuchBundle
	}

	err = da.doBundleDisable(tx, name)
	if err != nil {
		tx.Rollback()
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	err = da.doBundleDelete(tx, name, version)
	if err != nil {
		tx.Rollback()
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

// BundleDisable TBD
func (da PostgresDataAccess) BundleDisable(ctx context.Context, name string) error {
	db, err := da.connect("gort")
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	err = da.doBundleDisable(tx, name)
	if err != nil {
		tx.Rollback()
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

// BundleEnable TBD
func (da PostgresDataAccess) BundleEnable(ctx context.Context, name, version string) error {
	db, err := da.connect("gort")
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	err = da.doBundleEnable(tx, name, version)
	if err != nil {
		tx.Rollback()
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

// BundleEnabledVersion TBD
func (da PostgresDataAccess) BundleEnabledVersion(ctx context.Context, name string) (string, error) {
	db, err := da.connect("gort")
	if err != nil {
		return "", err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return "", gerr.Wrap(errs.ErrDataAccess, err)
	}

	enabled, err := da.doBundleEnabledVersion(tx, name)
	if err != nil {
		tx.Rollback()
		return "", gerr.Wrap(errs.ErrDataAccess, err)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return enabled, gerr.Wrap(errs.ErrDataAccess, err)
	}

	return enabled, nil
}

// BundleExists TBD
func (da PostgresDataAccess) BundleExists(ctx context.Context, name, version string) (bool, error) {
	db, err := da.connect("gort")
	if err != nil {
		return false, err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return false, gerr.Wrap(errs.ErrDataAccess, err)
	}

	return da.doBundleExists(tx, name, version)
}

// BundleGet TBD
func (da PostgresDataAccess) BundleGet(ctx context.Context, name, version string) (data.Bundle, error) {
	if name == "" {
		return data.Bundle{}, errs.ErrEmptyBundleName
	}

	if version == "" {
		return data.Bundle{}, errs.ErrEmptyBundleVersion
	}

	db, err := da.connect("gort")
	if err != nil {
		return data.Bundle{}, err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return data.Bundle{}, gerr.Wrap(errs.ErrDataAccess, err)
	}

	b, err := da.doBundleGet(tx, name, version)
	if err != nil {
		return data.Bundle{}, err
	}

	return b, err
}

// BundleList TBD
func (da PostgresDataAccess) BundleList(ctx context.Context) ([]data.Bundle, error) {
	// This is hacky as fuck. I know.
	// I'll optimize later.

	db, err := da.connect("gort")
	if err != nil {
		return []data.Bundle{}, err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		return []data.Bundle{}, gerr.Wrap(errs.ErrDataAccess, err)
	}

	query := `SELECT name, version FROM bundles`
	rows, err := tx.Query(query)
	if err != nil {
		tx.Rollback()
		return []data.Bundle{}, gerr.Wrap(errs.ErrDataAccess, err)
	}

	bds := make([]bundleData, 0)
	for rows.Next() {
		var bd bundleData

		err = rows.Scan(&bd.BundleName, &bd.BundleVersion)
		if err != nil {
			rows.Close()
			tx.Rollback()
			return []data.Bundle{}, gerr.Wrap(errs.ErrDataAccess, err)
		}

		bds = append(bds, bd)
	}
	rows.Close()

	bundles := make([]data.Bundle, 0)
	for _, bd := range bds {
		bundle, err := da.doBundleGet(tx, bd.BundleName, bd.BundleVersion)
		if err != nil {
			tx.Rollback()
			return []data.Bundle{}, err
		}

		bundles = append(bundles, bundle)
	}

	tx.Commit()
	return bundles, nil
}

// BundleListVersions TBD
func (da PostgresDataAccess) BundleListVersions(ctx context.Context, name string) ([]data.Bundle, error) {
	// This is hacky as fuck. I know.
	// I'll optimize later.

	db, err := da.connect("gort")
	if err != nil {
		return []data.Bundle{}, err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		return []data.Bundle{}, gerr.Wrap(errs.ErrDataAccess, err)
	}

	query := `SELECT name, version FROM bundles WHERE name=$1`
	rows, err := tx.Query(query, name)
	if err != nil {
		tx.Rollback()
		return []data.Bundle{}, gerr.Wrap(errs.ErrDataAccess, err)
	}

	bds := make([]bundleData, 0)
	for rows.Next() {
		var bd bundleData

		err = rows.Scan(&bd.BundleName, &bd.BundleVersion)
		if err != nil {
			rows.Close()
			tx.Rollback()
			return []data.Bundle{}, gerr.Wrap(errs.ErrDataAccess, err)
		}

		bds = append(bds, bd)
	}
	rows.Close()

	bundles := make([]data.Bundle, 0)
	for _, bd := range bds {
		bundle, err := da.doBundleGet(tx, bd.BundleName, bd.BundleVersion)
		if err != nil {
			tx.Rollback()
			return []data.Bundle{}, err
		}

		bundles = append(bundles, bundle)
	}

	return bundles, nil
}

// BundleUpdate TBD
func (da PostgresDataAccess) BundleUpdate(ctx context.Context, bundle data.Bundle) error {
	if bundle.Name == "" {
		return errs.ErrEmptyBundleName
	}

	if bundle.Version == "" {
		return errs.ErrEmptyBundleVersion
	}

	db, err := da.connect("gort")
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	exists, err := da.doBundleExists(tx, bundle.Name, bundle.Version)
	if err != nil {
		return err
	} else if !exists {
		return errs.ErrNoSuchBundle
	}

	err = da.doBundleDelete(tx, bundle.Name, bundle.Version)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = da.doBundleInsert(tx, bundle)
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

// FindCommandEntry is used to find the enabled commands with the provided
// bundle and command names. If either is empty, it is treated as a wildcard.
// Importantly, this must only return ENABLED commands!
func (da PostgresDataAccess) FindCommandEntry(ctx context.Context, bundleName, commandName string) ([]data.CommandEntry, error) {
	db, err := da.connect("gort")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, gerr.Wrap(errs.ErrDataAccess, err)
	}

	return da.doFindCommandEntry(tx, bundleName, commandName)
}

func (da PostgresDataAccess) doBundleDelete(tx *sql.Tx, name string, version string) error {
	query := "DELETE FROM bundle_command_rules WHERE bundle_name=$1 AND bundle_version=$2;"
	_, err := tx.Exec(query, name, version)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	query = "DELETE FROM bundle_permissions WHERE bundle_name=$1 AND bundle_version=$2;"
	_, err = tx.Exec(query, name, version)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	query = "DELETE FROM bundle_commands WHERE bundle_name=$1 AND bundle_version=$2;"
	_, err = tx.Exec(query, name, version)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	query = "DELETE FROM bundles WHERE name=$1 AND version=$2;"
	_, err = tx.Exec(query, name, version)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

// doBundleDisable TBD
func (da PostgresDataAccess) doBundleDisable(tx *sql.Tx, name string) error {
	query := `DELETE FROM bundle_enabled WHERE bundle_name=$1`

	_, err := tx.Exec(query, name)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

// BundleEnable TBD
func (da PostgresDataAccess) doBundleEnable(tx *sql.Tx, name string, version string) error {
	enabled, err := da.doBundleEnabledVersion(tx, name)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	query := ""

	if enabled == "" {
		query = `INSERT INTO bundle_enabled (bundle_name, bundle_version)
			VALUES ($1, $2);`
	} else {
		query = `UPDATE bundle_enabled
			SET bundle_name=$1, bundle_version=$2
			WHERE bundle_name=$1;`
	}

	_, err = tx.Exec(query, name, version)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

// BundleExists TBD
func (da PostgresDataAccess) doBundleEnabledVersion(tx *sql.Tx, name string) (string, error) {
	query := `SELECT
		COALESCE(
		(SELECT bundle_version FROM bundle_enabled WHERE bundle_name=$1),
		''
		) AS bundle_version;`

	enabled := ""

	err := tx.QueryRow(query, name).Scan(&enabled)
	if err != nil {
		return "", gerr.Wrap(errs.ErrDataAccess, err)
	}

	return enabled, nil
}

// BundleExists TBD
func (da PostgresDataAccess) doBundleExists(tx *sql.Tx, name string, version string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM bundles WHERE name=$1 AND version=$2)"
	exists := false

	err := tx.QueryRow(query, name, version).Scan(&exists)
	if err != nil {
		return false, gerr.Wrap(errs.ErrDataAccess, err)
	}

	return exists, nil
}

func (da PostgresDataAccess) doBundleGet(tx *sql.Tx, name string, version string) (data.Bundle, error) {
	query := `SELECT gort_bundle_version, name, version, author, homepage,
			description, long_description, docker_image, docker_tag,
			install_timestamp, install_user
		FROM bundles
		WHERE name=$1 AND version=$2`

	bundle := data.Bundle{}
	row := tx.QueryRow(query, name, version)
	err := row.Scan(&bundle.GortBundleVersion, &bundle.Name, &bundle.Version,
		&bundle.Author, &bundle.Homepage, &bundle.Description,
		&bundle.LongDescription, &bundle.Docker.Image, &bundle.Docker.Tag,
		&bundle.InstalledOn, &bundle.InstalledBy)
	if err != nil {
		return bundle, gerr.Wrap(errs.ErrNoSuchBundle, err)
	}

	enabledVersion, err := da.doBundleEnabledVersion(tx, name)
	if err != nil {
		return bundle, gerr.Wrap(fmt.Errorf("failed to get bundle enabled version"), err)
	}
	bundle.Enabled = (bundle.Version == enabledVersion)

	// Load bundle permissions
	bundle.Permissions, err = da.doBundlePermissionsGet(tx, name, version)
	if err != nil {
		return bundle, gerr.Wrap(fmt.Errorf("failed to get bundle permissions"), err)
	}

	// Load all commands (and their rules) for this bundle
	commandSlice, err := da.doBundleCommandsGet(tx, name, version, "")
	if err != nil {
		return bundle, gerr.Wrap(fmt.Errorf("failed to get bundle commands"), err)
	}

	bundle.Commands = make(map[string]*data.BundleCommand)

	for _, command := range commandSlice {
		bundle.Commands[command.Name] = command
	}

	return bundle, nil
}

// doBundleCommandsDataGet is a helper method that retrieves zero or more
// commands for the specified bundle name+version, along with the owning
// bundle's name and version. Empty string parameters are treated as wildcards.
func (da PostgresDataAccess) doBundleCommandsDataGet(tx *sql.Tx, bundleName, bundleVersion, commandName string, enabledOnly bool) ([]bundleCommandData, error) {
	var query string

	if bundleName == "" {
		bundleName = "%"
	}
	if bundleVersion == "" {
		bundleVersion = "%"
	}
	if commandName == "" {
		commandName = "%"
	}

	if enabledOnly {
		query = `SELECT bundle_commands.bundle_name, bundle_commands.bundle_version, name, description, executable
			FROM bundle_commands
			INNER JOIN bundle_enabled ON bundle_commands.bundle_name=bundle_enabled.bundle_name
			WHERE bundle_commands.bundle_name LIKE $1 AND bundle_commands.bundle_version LIKE $2 AND name LIKE $3`
	} else {
		query = `SELECT bundle_commands.bundle_name, bundle_commands.bundle_version, name, description, executable
			FROM bundle_commands
			WHERE bundle_commands.bundle_name LIKE $1 AND bundle_commands.bundle_version LIKE $2 AND name LIKE $3`
	}

	rows, err := tx.Query(query, bundleName, bundleVersion, commandName)
	if err != nil {
		return nil, gerr.Wrap(errs.ErrDataAccess, err)
	}
	defer rows.Close()

	commands := make([]bundleCommandData, 0)

	for rows.Next() {
		cd := bundleCommandData{}

		err = rows.Scan(&cd.BundleName, &cd.BundleVersion, &cd.Name, &cd.Description, &cd.Executable)
		if err != nil {
			return nil, gerr.Wrap(errs.ErrDataAccess, err)
		}

		commands = append(commands, cd)
	}

	return commands, nil
}

// doBundleCommandGet empty strings become wildcards
func (da PostgresDataAccess) doBundleCommandsGet(tx *sql.Tx, bundleName, bundleVersion, commandName string) ([]*data.BundleCommand, error) {
	bcd, err := da.doBundleCommandsDataGet(tx, bundleName, bundleVersion, commandName, false)
	if err != nil {
		return nil, gerr.Wrap(errs.ErrDataAccess, err)
	}

	commands := make([]*data.BundleCommand, 0)

	for _, bc := range bcd {
		bc.BundleCommand.Rules, err = da.doBundleCommandRulesGet(tx, bundleName, bundleVersion, bc.Name)
		if err != nil {
			return nil, gerr.Wrap(fmt.Errorf("failed to get bundle command rules"), err)
		}

		commands = append(commands, &bc.BundleCommand)
	}

	return commands, nil
}

// doFindCommandEntry returns all command entries for any enabled bundle
// matching the specified bundle and command names. The bundle parameter may be
// empty, in which case it will match all bundles.
func (da PostgresDataAccess) doFindCommandEntry(tx *sql.Tx, bundle, command string) ([]data.CommandEntry, error) {
	bcd, err := da.doBundleCommandsDataGet(tx, bundle, "", command, true)
	if err != nil {
		return nil, gerr.Wrap(errs.ErrDataAccess, err)
	}

	entries := make([]data.CommandEntry, 0)

	for _, cd := range bcd {
		entry := data.CommandEntry{}

		// Load the appropriate bundle
		entry.Bundle, err = da.doBundleGet(tx, cd.BundleName, cd.BundleVersion)
		if err != nil {
			return nil, gerr.Wrap(errs.ErrDataAccess, err)
		}

		// Load the relevant bundle command (there should be exactly one)
		commands, err := da.doBundleCommandsGet(tx, cd.BundleName, cd.BundleVersion, cd.Name)
		if err != nil {
			return nil, gerr.Wrap(errs.ErrDataAccess, err)
		}
		if len(commands) != 1 {
			return nil, gerr.Wrap(errs.ErrDataAccess, fmt.Errorf("unexpected commands count: %d", len(commands)))
		}
		entry.Command = *commands[0]

		entries = append(entries, entry)
	}

	return entries, nil
}

func (da PostgresDataAccess) doBundleCommandRulesGet(tx *sql.Tx, bundleName, bundleVersion, commandName string) ([]string, error) {
	cmdQuery := `SELECT rule
		FROM bundle_command_rules
		WHERE bundle_name=$1 AND bundle_version=$2 AND command_name=$3`

	rows, err := tx.Query(cmdQuery, bundleName, bundleVersion, commandName)
	if err != nil {
		return nil, gerr.Wrap(errs.ErrDataAccess, err)
	}
	defer rows.Close()

	rules := make([]string, 0)
	for rows.Next() {
		var rule string

		err = rows.Scan(&rule)
		if err != nil {
			return nil, gerr.Wrap(errs.ErrDataAccess, err)
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

func (da PostgresDataAccess) doBundlePermissionsGet(tx *sql.Tx, bundleName, bundleVersion string) ([]string, error) {
	// Load permissions
	query := `SELECT permission
		FROM bundle_permissions
		WHERE bundle_name=$1 AND bundle_version=$2
		ORDER BY index`

	rows, err := tx.Query(query, bundleName, bundleVersion)
	if err != nil {
		return nil, gerr.Wrap(errs.ErrDataAccess, err)
	}
	defer rows.Close()

	permissions := make([]string, 0)

	for rows.Next() {
		var perm string

		err = rows.Scan(&perm)
		if err != nil {
			return nil, gerr.Wrap(errs.ErrDataAccess, err)
		}

		permissions = append(permissions, perm)
	}
	rows.Close()

	return permissions, nil
}

func (da PostgresDataAccess) doBundleInsert(tx *sql.Tx, bundle data.Bundle) error {
	query := `INSERT INTO bundles (gort_bundle_version, name, version, author,
		homepage, description, long_description, docker_image,
		docker_tag, install_user)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);`

	_, err := tx.Exec(query, bundle.GortBundleVersion, bundle.Name, bundle.Version,
		bundle.Author, bundle.Homepage, bundle.Description, bundle.LongDescription,
		bundle.Docker.Image, bundle.Docker.Tag, bundle.InstalledBy)

	if err != nil {
		if strings.Contains(err.Error(), "violates") {
			err = gerr.Wrap(errs.ErrFieldRequired, err)
		} else {
			err = gerr.Wrap(errs.ErrDataAccess, err)
		}

		return err
	}

	return nil
}

func (da PostgresDataAccess) doBundleInsertCommandRules(
	tx *sql.Tx, bundle data.Bundle, command *data.BundleCommand) error {

	query := `INSERT INTO bundle_command_rules
		(bundle_name, bundle_version, command_name, rule)
		VALUES ($1, $2, $3, $4);`

	for _, rule := range command.Rules {
		_, err := tx.Exec(query, bundle.Name, bundle.Version, command.Name, rule)
		if err != nil {
			if strings.Contains(err.Error(), "violates") {
				err = gerr.Wrap(errs.ErrFieldRequired, err)
			} else {
				err = gerr.Wrap(errs.ErrDataAccess, err)
			}

			return err
		}
	}

	return nil
}

func (da PostgresDataAccess) doBundleInsertCommands(tx *sql.Tx, bundle data.Bundle) error {
	query := `INSERT INTO bundle_commands
		(bundle_name, bundle_version, name, description, executable)
		VALUES ($1, $2, $3, $4, $5);`

	for name, cmd := range bundle.Commands {
		cmd.Name = name

		_, err := tx.Exec(query, bundle.Name, bundle.Version,
			cmd.Name, cmd.Description, cmd.Executable)

		if err != nil {
			if strings.Contains(err.Error(), "violates") {
				err = gerr.Wrap(errs.ErrFieldRequired, err)
			} else {
				err = gerr.Wrap(errs.ErrDataAccess, err)
			}

			return err
		}

		err = da.doBundleInsertCommandRules(tx, bundle, cmd)
		if err != nil {
			return err
		}
	}

	return nil
}

func (da PostgresDataAccess) doBundleInsertPermissions(tx *sql.Tx, bundle data.Bundle) error {
	query := `INSERT INTO bundle_permissions
		(bundle_name, bundle_version, index, permission)
		VALUES ($1, $2, $3, $4);`

	for i, perm := range bundle.Permissions {
		_, err := tx.Exec(query, bundle.Name, bundle.Version, i, perm)
		if err != nil {
			if strings.Contains(err.Error(), "violates") {
				err = gerr.Wrap(errs.ErrFieldRequired, err)
			} else {
				err = gerr.Wrap(errs.ErrDataAccess, err)
			}

			return err
		}
	}

	return nil
}
