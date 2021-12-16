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
	"net/url"
	"strings"

	"go.opentelemetry.io/otel"

	"github.com/getgort/gort/data"
	"github.com/getgort/gort/dataaccess/errs"
	gerr "github.com/getgort/gort/errors"
	"github.com/getgort/gort/telemetry"
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
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.BundleCreate")
	defer sp.End()

	if bundle.Name == "" {
		return errs.ErrEmptyBundleName
	}

	if bundle.Version == "" {
		return errs.ErrEmptyBundleVersion
	}

	db, err := da.connect(ctx, "gort")
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	exists, err := da.doBundleExists(ctx, tx, bundle.Name, bundle.Version)
	if err != nil {
		return err
	} else if exists {
		return errs.ErrBundleExists
	}

	// Save bundle
	err = da.doBundleInsert(ctx, tx, bundle)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Save permissions
	err = da.doBundleInsertPermissions(ctx, tx, bundle)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Save commands
	err = da.doBundleInsertCommands(ctx, tx, bundle)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Save templates
	err = da.doBundleInsertTemplates(ctx, tx, bundle)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Save kubernetes config
	err = da.doBundleInsertKubernetes(ctx, tx, bundle)
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
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.BundleDelete")
	defer sp.End()

	if name == "" {
		return errs.ErrEmptyBundleName
	}

	if version == "" {
		return errs.ErrEmptyBundleVersion
	}

	db, err := da.connect(ctx, "gort")
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	exists, err := da.doBundleExists(ctx, tx, name, version)
	if err != nil {
		return err
	} else if !exists {
		return errs.ErrNoSuchBundle
	}

	err = da.doBundleDisable(ctx, tx, name, version)
	if err != nil {
		tx.Rollback()
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	err = da.doBundleDelete(ctx, tx, name, version)
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
func (da PostgresDataAccess) BundleDisable(ctx context.Context, name, version string) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.BundleDisable")
	defer sp.End()

	db, err := da.connect(ctx, "gort")
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	err = da.doBundleDisable(ctx, tx, name, version)
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
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.BundleEnable")
	defer sp.End()

	db, err := da.connect(ctx, "gort")
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	err = da.doBundleEnable(ctx, tx, name, version)
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

// BundleEnabledVersion returns the currently enabled version of the specified bundle.
// If no version is enabled an empty string will be returned.
func (da PostgresDataAccess) BundleEnabledVersion(ctx context.Context, bundlename string) (string, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.BundleEnabledVersion")
	defer sp.End()

	db, err := da.connect(ctx, "gort")
	if err != nil {
		return "", err
	}
	defer db.Close()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return "", gerr.Wrap(errs.ErrDataAccess, err)
	}

	enabled, err := da.doBundleEnabledVersion(ctx, tx, bundlename)
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
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.BundleExists")
	defer sp.End()

	db, err := da.connect(ctx, "gort")
	if err != nil {
		return false, err
	}
	defer db.Close()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return false, gerr.Wrap(errs.ErrDataAccess, err)
	}

	return da.doBundleExists(ctx, tx, name, version)
}

// BundleGet TBD
func (da PostgresDataAccess) BundleGet(ctx context.Context, name, version string) (data.Bundle, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.BundleGet")
	defer sp.End()

	if name == "" {
		return data.Bundle{}, errs.ErrEmptyBundleName
	}

	if version == "" {
		return data.Bundle{}, errs.ErrEmptyBundleVersion
	}

	db, err := da.connect(ctx, "gort")
	if err != nil {
		return data.Bundle{}, err
	}
	defer db.Close()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return data.Bundle{}, gerr.Wrap(errs.ErrDataAccess, err)
	}

	b, err := da.doBundleGet(ctx, tx, name, version)
	if err != nil {
		return data.Bundle{}, err
	}

	return b, err
}

// BundleList TBD
func (da PostgresDataAccess) BundleList(ctx context.Context) ([]data.Bundle, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.BundleList")
	defer sp.End()

	// This is hacky as fuck. I know.
	// I'll optimize later.

	db, err := da.connect(ctx, "gort")
	if err != nil {
		return []data.Bundle{}, err
	}
	defer db.Close()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		tx.Rollback()
		return []data.Bundle{}, gerr.Wrap(errs.ErrDataAccess, err)
	}

	query := `SELECT name, version FROM bundles`
	rows, err := tx.QueryContext(ctx, query)
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
		bundle, err := da.doBundleGet(ctx, tx, bd.BundleName, bd.BundleVersion)
		if err != nil {
			tx.Rollback()
			return []data.Bundle{}, err
		}

		bundles = append(bundles, bundle)
	}

	tx.Commit()
	return bundles, nil
}

// BundleUpdate TBD
func (da PostgresDataAccess) BundleUpdate(ctx context.Context, bundle data.Bundle) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.BundleUpdate")
	defer sp.End()

	if bundle.Name == "" {
		return errs.ErrEmptyBundleName
	}

	if bundle.Version == "" {
		return errs.ErrEmptyBundleVersion
	}

	db, err := da.connect(ctx, "gort")
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	exists, err := da.doBundleExists(ctx, tx, bundle.Name, bundle.Version)
	if err != nil {
		return err
	} else if !exists {
		return errs.ErrNoSuchBundle
	}

	err = da.doBundleDelete(ctx, tx, bundle.Name, bundle.Version)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = da.doBundleInsert(ctx, tx, bundle)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

// BundleVersionList TBD
func (da PostgresDataAccess) BundleVersionList(ctx context.Context, name string) ([]data.Bundle, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.BundleVersionList")
	defer sp.End()

	// This is hacky as fuck. I know.
	// I'll optimize later.

	db, err := da.connect(ctx, "gort")
	if err != nil {
		return []data.Bundle{}, err
	}
	defer db.Close()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		tx.Rollback()
		return []data.Bundle{}, gerr.Wrap(errs.ErrDataAccess, err)
	}

	query := `SELECT name, version FROM bundles WHERE name=$1`
	rows, err := tx.QueryContext(ctx, query, name)
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
		bundle, err := da.doBundleGet(ctx, tx, bd.BundleName, bd.BundleVersion)
		if err != nil {
			tx.Rollback()
			return []data.Bundle{}, err
		}

		bundles = append(bundles, bundle)
	}

	return bundles, nil
}

// FindCommandEntry is used to find the enabled commands with the provided
// bundle and command names. If either is empty, it is treated as a wildcard.
// Importantly, this must only return ENABLED commands!
func (da PostgresDataAccess) FindCommandEntry(ctx context.Context, bundleName, commandName string) ([]data.CommandEntry, error) {
	db, err := da.connect(ctx, "gort")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, gerr.Wrap(errs.ErrDataAccess, err)
	}

	return da.doFindCommandEntry(ctx, tx, bundleName, commandName)
}

func (da PostgresDataAccess) doBundleDelete(ctx context.Context, tx *sql.Tx, name string, version string) error {
	query := "DELETE FROM bundle_kubernetes WHERE bundle_name=$1 AND bundle_version=$2;"
	_, err := tx.ExecContext(ctx, query, name, version)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	query = "DELETE FROM bundle_command_rules WHERE bundle_name=$1 AND bundle_version=$2;"
	_, err = tx.ExecContext(ctx, query, name, version)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	query = "DELETE FROM bundle_permissions WHERE bundle_name=$1 AND bundle_version=$2;"
	_, err = tx.ExecContext(ctx, query, name, version)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	query = "DELETE FROM bundle_commands WHERE bundle_name=$1 AND bundle_version=$2;"
	_, err = tx.ExecContext(ctx, query, name, version)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	query = "DELETE FROM bundles WHERE name=$1 AND version=$2;"
	_, err = tx.ExecContext(ctx, query, name, version)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

// doBundleDisable TBD
func (da PostgresDataAccess) doBundleDisable(ctx context.Context, tx *sql.Tx, name string, version string) error {
	query := `DELETE FROM bundle_enabled WHERE bundle_name=$1 AND bundle_version=$2`

	_, err := tx.ExecContext(ctx, query, name, version)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

// BundleEnable TBD
func (da PostgresDataAccess) doBundleEnable(ctx context.Context, tx *sql.Tx, name string, version string) error {
	enabled, err := da.doBundleEnabledVersion(ctx, tx, name)
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

	_, err = tx.ExecContext(ctx, query, name, version)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

// BundleExists TBD
func (da PostgresDataAccess) doBundleEnabledVersion(ctx context.Context, tx *sql.Tx, name string) (string, error) {
	query := `SELECT
		COALESCE(
		(SELECT bundle_version FROM bundle_enabled WHERE bundle_name=$1),
		''
		) AS bundle_version;`

	enabled := ""

	err := tx.QueryRowContext(ctx, query, name).Scan(&enabled)
	if err != nil {
		return "", gerr.Wrap(errs.ErrDataAccess, err)
	}

	return enabled, nil
}

// BundleExists TBD
func (da PostgresDataAccess) doBundleExists(ctx context.Context, tx *sql.Tx, name string, version string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM bundles WHERE name=$1 AND version=$2)"
	exists := false

	err := tx.QueryRowContext(ctx, query, name, version).Scan(&exists)
	if err != nil {
		return false, gerr.Wrap(errs.ErrDataAccess, err)
	}

	return exists, nil
}

// doFindCommandEntry returns all command entries for any enabled bundle
// matching the specified bundle and command names. The bundle parameter may be
// empty, in which case it will match all bundles.
func (da PostgresDataAccess) doFindCommandEntry(ctx context.Context, tx *sql.Tx, bundle, command string) ([]data.CommandEntry, error) {
	bcd, err := da.doBundleGetCommandsData(ctx, tx, bundle, "", command, true)
	if err != nil {
		return nil, gerr.Wrap(errs.ErrDataAccess, err)
	}

	entries := make([]data.CommandEntry, 0)

	for _, cd := range bcd {
		entry := data.CommandEntry{}

		// Load the appropriate bundle
		entry.Bundle, err = da.doBundleGet(ctx, tx, cd.BundleName, cd.BundleVersion)
		if err != nil {
			return nil, gerr.Wrap(errs.ErrDataAccess, err)
		}

		// Load the relevant bundle command (there should be exactly one)
		commands, err := da.doBundleGetCommands(ctx, tx, cd.BundleName, cd.BundleVersion, cd.Name)
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

func (da PostgresDataAccess) doBundleGet(ctx context.Context, tx *sql.Tx, name string, version string) (data.Bundle, error) {
	query := `SELECT gort_bundle_version, name, version, author, homepage,
			description, long_description, image_repository, image_tag,
			install_timestamp, install_user
		FROM bundles
		WHERE name=$1 AND version=$2`

	var repository, tag string

	bundle := data.Bundle{}
	row := tx.QueryRowContext(ctx, query, name, version)
	err := row.Scan(&bundle.GortBundleVersion, &bundle.Name, &bundle.Version,
		&bundle.Author, &bundle.Homepage, &bundle.Description,
		&bundle.LongDescription, &repository, &tag,
		&bundle.InstalledOn, &bundle.InstalledBy)
	if err != nil {
		return bundle, gerr.Wrap(errs.ErrNoSuchBundle, err)
	}

	if repository != "" {
		if tag == "" {
			tag = "latest"
		}

		bundle.Image = repository + ":" + tag
	}

	enabledVersion, err := da.doBundleEnabledVersion(ctx, tx, name)
	if err != nil {
		return bundle, gerr.Wrap(fmt.Errorf("failed to get bundle enabled version"), err)
	}
	bundle.Enabled = (bundle.Version == enabledVersion)

	// Load bundle permissions
	bundle.Permissions, err = da.doBundleGetPermissions(ctx, tx, name, version)
	if err != nil {
		return bundle, gerr.Wrap(fmt.Errorf("failed to get bundle permissions"), err)
	}

	// Load all commands (and their rules) for this bundle
	commandSlice, err := da.doBundleGetCommands(ctx, tx, name, version, "")
	if err != nil {
		return bundle, gerr.Wrap(fmt.Errorf("failed to get bundle commands"), err)
	}

	bundle.Commands = make(map[string]*data.BundleCommand)
	for _, command := range commandSlice {
		bundle.Commands[command.Name] = command
	}

	bundle.Templates, err = da.doBundleGetTemplates(ctx, tx, name, version)
	if err != nil {
		return bundle, gerr.Wrap(fmt.Errorf("failed to get bundle templates"), err)
	}

	bundle.Kubernetes, err = da.doBundleGetKubernetes(ctx, tx, name, version)
	if err != nil {
		return bundle, gerr.Wrap(fmt.Errorf("failed to get bundle kubernetes config"), err)
	}

	return bundle, nil
}

// doBundleGetCommandsData is a helper method that retrieves zero or more
// commands for the specified bundle name+version, along with the owning
// bundle's name and version. Empty string parameters are treated as wildcards.
func (da PostgresDataAccess) doBundleGetCommandsData(ctx context.Context, tx *sql.Tx, bundleName, bundleVersion, commandName string, enabledOnly bool) ([]bundleCommandData, error) {
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
		query = `SELECT bundle_commands.bundle_name, bundle_commands.bundle_version, name, description, executable, long_description
			FROM bundle_commands
			INNER JOIN bundle_enabled ON bundle_commands.bundle_name=bundle_enabled.bundle_name
			WHERE bundle_commands.bundle_name LIKE $1 AND bundle_commands.bundle_version LIKE $2 AND name LIKE $3`
	} else {
		query = `SELECT bundle_commands.bundle_name, bundle_commands.bundle_version, name, description, executable, long_description
			FROM bundle_commands
			WHERE bundle_commands.bundle_name LIKE $1 AND bundle_commands.bundle_version LIKE $2 AND name LIKE $3`
	}

	rows, err := tx.QueryContext(ctx, query, bundleName, bundleVersion, commandName)
	if err != nil {
		return nil, gerr.Wrap(errs.ErrDataAccess, err)
	}
	defer rows.Close()

	commands := make([]bundleCommandData, 0)

	for rows.Next() {
		var enc string
		cd := bundleCommandData{}

		err = rows.Scan(&cd.BundleName, &cd.BundleVersion, &cd.Name, &cd.Description, &enc, &cd.LongDescription)
		if err != nil {
			return nil, gerr.Wrap(errs.ErrDataAccess, err)
		}

		cd.Executable = decodeStringSlice(enc)
		commands = append(commands, cd)
	}

	return commands, nil
}

// doBundleGetCommands empty strings become wildcards
func (da PostgresDataAccess) doBundleGetCommands(ctx context.Context, tx *sql.Tx, bundleName, bundleVersion, commandName string) ([]*data.BundleCommand, error) {
	bcd, err := da.doBundleGetCommandsData(ctx, tx, bundleName, bundleVersion, commandName, false)
	if err != nil {
		return nil, gerr.Wrap(errs.ErrDataAccess, err)
	}

	commands := make([]*data.BundleCommand, 0)

	for _, bc := range bcd {
		bc.BundleCommand.Rules, err = da.doBundleGetCommandRules(ctx, tx, bundleName, bundleVersion, bc.Name)
		if err != nil {
			return nil, gerr.Wrap(fmt.Errorf("failed to get bundle command rules"), err)
		}

		bc.BundleCommand.Templates, err = da.doBundleGetCommandTemplates(ctx, tx, bundleName, bundleVersion, bc.Name)
		if err != nil {
			return nil, gerr.Wrap(fmt.Errorf("failed to get bundle command templates"), err)
		}

		command := bc.BundleCommand
		commands = append(commands, &command)
	}

	return commands, nil
}

func (da PostgresDataAccess) doBundleGetCommandRules(ctx context.Context, tx *sql.Tx, bundleName, bundleVersion, commandName string) ([]string, error) {
	cmdQuery := `SELECT rule
		FROM bundle_command_rules
		WHERE bundle_name=$1 AND bundle_version=$2 AND command_name=$3`

	rows, err := tx.QueryContext(ctx, cmdQuery, bundleName, bundleVersion, commandName)
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

func (da PostgresDataAccess) doBundleGetCommandTemplates(ctx context.Context, tx *sql.Tx, bundleName, bundleVersion, commandName string) (data.Templates, error) {
	query := `SELECT command, command_error, message, message_error
		FROM bundle_command_templates
		WHERE bundle_name=$1 AND bundle_version=$2 AND command_name=$3`

	var templates data.Templates

	err := tx.QueryRowContext(ctx, query, bundleName, bundleVersion, commandName).
		Scan(&templates.Command, &templates.CommandError, &templates.Message, &templates.MessageError)

	switch {
	case err == sql.ErrNoRows:
		return data.Templates{}, errs.ErrNoSuchUser
	case err != nil:
		return data.Templates{}, gerr.Wrap(errs.ErrDataAccess, err)
	}

	return templates, nil
}

func (da PostgresDataAccess) doBundleGetKubernetes(ctx context.Context, tx *sql.Tx, bundleName, bundleVersion string) (data.BundleKubernetes, error) {
	query := `SELECT service_account_name, env_secret
		FROM bundle_kubernetes
		WHERE bundle_name=$1 AND bundle_version=$2`

	var kubernetes data.BundleKubernetes

	err := tx.QueryRowContext(ctx, query, bundleName, bundleVersion).
		Scan(&kubernetes.ServiceAccountName, &kubernetes.EnvSecret)

	switch {
	case err == sql.ErrNoRows:
		return data.BundleKubernetes{}, errs.ErrNoSuchUser
	case err != nil:
		return data.BundleKubernetes{}, gerr.Wrap(errs.ErrDataAccess, err)
	}

	return kubernetes, nil
}

func (da PostgresDataAccess) doBundleGetPermissions(ctx context.Context, tx *sql.Tx, bundleName, bundleVersion string) ([]string, error) {
	// Load permissions
	query := `SELECT permission
		FROM bundle_permissions
		WHERE bundle_name=$1 AND bundle_version=$2
		ORDER BY index`

	rows, err := tx.QueryContext(ctx, query, bundleName, bundleVersion)
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

func (da PostgresDataAccess) doBundleGetTemplates(ctx context.Context, tx *sql.Tx, bundleName, bundleVersion string) (data.Templates, error) {
	query := `SELECT command, command_error, message, message_error FROM bundle_templates
		WHERE bundle_name=$1 AND bundle_version=$2`

	var templates data.Templates

	err := tx.QueryRowContext(ctx, query, bundleName, bundleVersion).
		Scan(&templates.Command, &templates.CommandError, &templates.Message, &templates.MessageError)

	switch {
	case err == sql.ErrNoRows:
		return data.Templates{}, errs.ErrNoSuchUser
	case err != nil:
		return data.Templates{}, gerr.Wrap(errs.ErrDataAccess, err)
	}

	return templates, nil
}

func (da PostgresDataAccess) doBundleInsert(ctx context.Context, tx *sql.Tx, bundle data.Bundle) error {
	query := `INSERT INTO bundles (gort_bundle_version, name, version, author,
		homepage, description, long_description, image_repository, image_tag,
		install_user)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);`

	repository, tag := bundle.ImageFullParts()

	_, err := tx.ExecContext(ctx, query, bundle.GortBundleVersion, bundle.Name, bundle.Version,
		bundle.Author, bundle.Homepage, bundle.Description, bundle.LongDescription,
		repository, tag, bundle.InstalledBy)

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

func (da PostgresDataAccess) doBundleInsertCommandRules(ctx context.Context,
	tx *sql.Tx, bundle data.Bundle, command *data.BundleCommand) error {

	query := `INSERT INTO bundle_command_rules
		(bundle_name, bundle_version, command_name, rule)
		VALUES ($1, $2, $3, $4);`

	for _, rule := range command.Rules {
		_, err := tx.ExecContext(ctx, query, bundle.Name, bundle.Version, command.Name, rule)
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

func (da PostgresDataAccess) doBundleInsertCommandTemplates(ctx context.Context,
	tx *sql.Tx, bundle data.Bundle, command *data.BundleCommand) error {

	query := `INSERT INTO bundle_command_templates
		(bundle_name, bundle_version, command_name, command, command_error, message, message_error)
		VALUES ($1, $2, $3, $4, $5, $6, $7);`

	_, err := tx.ExecContext(ctx, query, bundle.Name, bundle.Version, command.Name,
		command.Templates.Command, command.Templates.CommandError,
		command.Templates.Message, command.Templates.MessageError)

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

func (da PostgresDataAccess) doBundleInsertCommands(ctx context.Context, tx *sql.Tx, bundle data.Bundle) error {
	query := `INSERT INTO bundle_commands
		(bundle_name, bundle_version, name, description, executable, long_description)
		VALUES ($1, $2, $3, $4, $5, $6);`

	for name, cmd := range bundle.Commands {
		cmd.Name = name

		enc := encodeStringSlice(cmd.Executable)

		_, err := tx.ExecContext(ctx, query, bundle.Name, bundle.Version,
			cmd.Name, cmd.Description, enc, cmd.LongDescription)

		if err != nil {
			if strings.Contains(err.Error(), "violates") {
				err = gerr.Wrap(errs.ErrFieldRequired, err)
			} else {
				err = gerr.Wrap(errs.ErrDataAccess, err)
			}

			return err
		}

		err = da.doBundleInsertCommandRules(ctx, tx, bundle, cmd)
		if err != nil {
			return err
		}

		err = da.doBundleInsertCommandTemplates(ctx, tx, bundle, cmd)
		if err != nil {
			return err
		}
	}

	return nil
}

func (da PostgresDataAccess) doBundleInsertPermissions(ctx context.Context, tx *sql.Tx, bundle data.Bundle) error {
	query := `INSERT INTO bundle_permissions
		(bundle_name, bundle_version, index, permission)
		VALUES ($1, $2, $3, $4);`

	for i, perm := range bundle.Permissions {
		_, err := tx.ExecContext(ctx, query, bundle.Name, bundle.Version, i, perm)
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

func (da PostgresDataAccess) doBundleInsertTemplates(ctx context.Context, tx *sql.Tx, bundle data.Bundle) error {
	query := `INSERT INTO bundle_templates
		(bundle_name, bundle_version, command, command_error, message, message_error)
		VALUES ($1, $2, $3, $4, $5, $6);`

	_, err := tx.ExecContext(ctx, query, bundle.Name, bundle.Version,
		bundle.Templates.Command, bundle.Templates.CommandError,
		bundle.Templates.Message, bundle.Templates.MessageError)

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

func (da PostgresDataAccess) doBundleInsertKubernetes(ctx context.Context, tx *sql.Tx, bundle data.Bundle) error {
	query := `INSERT INTO bundle_kubernetes
		(bundle_name, bundle_version, service_account_name, env_secret)
		VALUES ($1, $2, $3, $4);`

	_, err := tx.ExecContext(ctx, query, bundle.Name, bundle.Version,
		bundle.Kubernetes.ServiceAccountName, bundle.Kubernetes.EnvSecret)

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

func decodeStringSlice(str string) []string {
	if str == "" {
		return []string{}
	}

	enc := strings.Split(str, ",")

	for i, s := range enc {
		enc[i], _ = url.QueryUnescape(s)
	}

	return enc
}

func encodeStringSlice(strs []string) string {
	if len(strs) == 0 {
		return ""
	}

	enc := make([]string, len(strs))

	for i, s := range strs {
		enc[i] = url.QueryEscape(s)
	}

	return strings.Join(enc, ",")
}
