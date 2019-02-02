package postgres

import (
	"database/sql"
	"strings"

	"github.com/clockworksoul/cog2/data"
	"github.com/clockworksoul/cog2/dataaccess/errs"
	cogerr "github.com/clockworksoul/cog2/errors"
)

// BundleCreate TBD
// TODO This func is a hot mess. Break it up and tidy it.
func (da PostgresDataAccess) BundleCreate(bundle data.Bundle) error {
	if bundle.Name == "" {
		return errs.ErrEmptyBundleName
	}

	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return err
	}

	exists, err := da.doBundleExists(db, bundle.Name, bundle.Version)
	if err != nil {
		return err
	}
	if exists {
		return errs.ErrBundleExists
	}

	tx, err := db.Begin()
	if err != nil {
		return cogerr.Wrap(errs.ErrDataAccess, err)
	}

	query := `INSERT INTO bundles (cog_bundle_version, name, version, active, author, 
								   homepage, description, long_description, docker_image,
								   docker_tag, install_user)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`

	_, err = tx.Exec(query, bundle.CogBundleVersion, bundle.Name, bundle.Version,
		bundle.Active, bundle.Author, bundle.Homepage, bundle.Description,
		bundle.LongDescription, bundle.Docker.Image, bundle.Docker.Tag, bundle.InstalledBy)
	if err != nil {
		tx.Rollback()

		if strings.Contains(err.Error(), "violates") {
			err = cogerr.Wrap(errs.ErrFieldRequired, err)
		} else {
			err = cogerr.Wrap(errs.ErrDataAccess, err)
		}

		return err
	}

	// Save permissions
	//
	query = `INSERT INTO bundle_permissions
		(bundle_name, bundle_version, index, permission)
		VALUES ($1, $2, $3, $4);`

	for i, perm := range bundle.Permissions {
		_, err = tx.Exec(query, bundle.Name, bundle.Version, i, perm)
		if err != nil {
			tx.Rollback()

			if strings.Contains(err.Error(), "violates") {
				err = cogerr.Wrap(errs.ErrFieldRequired, err)
			} else {
				err = cogerr.Wrap(errs.ErrDataAccess, err)
			}

			return err
		}
	}

	// Save commands
	//
	query = `INSERT INTO bundle_commands
		(bundle_name, bundle_version, name, description, executable)
		VALUES ($1, $2, $3, $4, $5);`

	for name, cmd := range bundle.Commands {
		_, err = tx.Exec(query, bundle.Name, bundle.Version, name,
			cmd.Description, cmd.Executable)

		if err != nil {
			tx.Rollback()

			if strings.Contains(err.Error(), "violates") {
				err = cogerr.Wrap(errs.ErrFieldRequired, err)
			} else {
				err = cogerr.Wrap(errs.ErrDataAccess, err)
			}

			return err
		}

		for _, rule := range cmd.Rules {
			ruleQuery := `INSERT INTO bundle_command_rules
				(bundle_name, bundle_version, command_name, rule)
				VALUES ($1, $2, $3, $4);`

			_, err = tx.Exec(ruleQuery, bundle.Name, bundle.Version, name, rule)

			if err != nil {
				tx.Rollback()

				if strings.Contains(err.Error(), "violates") {
					err = cogerr.Wrap(errs.ErrFieldRequired, err)
				} else {
					err = cogerr.Wrap(errs.ErrDataAccess, err)
				}

				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return cogerr.Wrap(errs.ErrDataAccess, err)
	}

	return err
}

// BundleDelete TBD
func (da PostgresDataAccess) BundleDelete(name, version string) error {
	if name == "" {
		return errs.ErrEmptyBundleName
	}

	if version == "" {
		return errs.ErrEmptyBundleVersion
	}

	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return err
	}

	exists, err := da.doBundleExists(db, name, version)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchBundle
	}

	tx, err := db.Begin()
	if err != nil {
		return cogerr.Wrap(errs.ErrDataAccess, err)
	}

	query := "DELETE FROM bundle_command_rules WHERE bundle_name=$1 AND bundle_version=$2;"
	_, err = tx.Exec(query, name, version)
	if err != nil {
		tx.Rollback()
		return cogerr.Wrap(errs.ErrDataAccess, err)
	}

	query = "DELETE FROM bundle_permissions WHERE bundle_name=$1 AND bundle_version=$2;"
	_, err = tx.Exec(query, name, version)
	if err != nil {
		tx.Rollback()
		return cogerr.Wrap(errs.ErrDataAccess, err)
	}

	query = "DELETE FROM bundle_commands WHERE bundle_name=$1 AND bundle_version=$2;"
	_, err = tx.Exec(query, name, version)
	if err != nil {
		tx.Rollback()
		return cogerr.Wrap(errs.ErrDataAccess, err)
	}

	query = "DELETE FROM bundles WHERE name=$1 AND version=$2;"
	_, err = tx.Exec(query, name, version)
	if err != nil {
		tx.Rollback()
		return cogerr.Wrap(errs.ErrDataAccess, err)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return cogerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

// BundleExists TBD
func (da PostgresDataAccess) BundleExists(name, version string) (bool, error) {
	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return false, err
	}

	return da.doBundleExists(db, name, version)
}

// BundleGet TBD
func (da PostgresDataAccess) BundleGet(name, version string) (data.Bundle, error) {
	if name == "" {
		return data.Bundle{}, errs.ErrEmptyBundleName
	}

	if version == "" {
		return data.Bundle{}, errs.ErrEmptyBundleVersion
	}

	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return data.Bundle{}, err
	}

	return da.doBundleGet(db, name, version)
}

// BundleList TBD
func (da PostgresDataAccess) BundleList() ([]data.Bundle, error) {
	// This is hacky as fuck. I know.
	// I'll optimize later.

	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return []data.Bundle{}, err
	}

	query := `SELECT name, version FROM bundles`
	rows, err := db.Query(query)
	if err != nil {
		return []data.Bundle{}, cogerr.Wrap(errs.ErrDataAccess, err)
	}

	bundles := make([]data.Bundle, 0)
	for rows.NextResultSet() && rows.Next() {
		var name, version string

		err = rows.Scan(&name, &version)
		if err != nil {
			return []data.Bundle{}, cogerr.Wrap(errs.ErrDataAccess, err)
		}

		bundle, err := da.doBundleGet(db, name, version)
		if err != nil {
			return []data.Bundle{}, err
		}

		bundles = append(bundles, bundle)
	}

	return bundles, nil
}

// BundleListVersions TBD
func (da PostgresDataAccess) BundleListVersions(name string) ([]data.Bundle, error) {
	// This is hacky as fuck. I know.
	// I'll optimize later.

	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return []data.Bundle{}, err
	}

	query := `SELECT name, version FROM bundles WHERE name=$1`
	rows, err := db.Query(query, name)
	if err != nil {
		return []data.Bundle{}, cogerr.Wrap(errs.ErrDataAccess, err)
	}

	bundles := make([]data.Bundle, 0)
	for rows.NextResultSet() && rows.Next() {
		var name, version string

		err = rows.Scan(&name, &version)
		if err != nil {
			return []data.Bundle{}, cogerr.Wrap(errs.ErrDataAccess, err)
		}

		bundle, err := da.doBundleGet(db, name, version)
		if err != nil {
			return []data.Bundle{}, err
		}

		bundles = append(bundles, bundle)
	}

	return bundles, nil
}

// BundleUpdate TBD
func (da PostgresDataAccess) BundleUpdate(bundle data.Bundle) error {
	return errs.ErrNotImplemented
}

// BundleExists TBD
func (da PostgresDataAccess) doBundleExists(db *sql.DB, name string, version string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM bundles WHERE name=$1 AND version=$2)"
	exists := false

	err := db.QueryRow(query, name, version).Scan(&exists)
	if err != nil {
		return false, cogerr.Wrap(errs.ErrDataAccess, err)
	}

	return exists, nil
}

func (da PostgresDataAccess) doBundleGet(db *sql.DB, name string, version string) (data.Bundle, error) {
	query := `SELECT cog_bundle_version, name, version, active, author, homepage,
			description, long_description, docker_image, docker_tag, 
			install_timestamp, install_user
		FROM bundles
		WHERE name=$1 AND version=$2`

	bundle := data.Bundle{}
	err := db.
		QueryRow(query, name, version).
		Scan(&bundle.CogBundleVersion, &bundle.Name, &bundle.Version,
			&bundle.Active, &bundle.Author, &bundle.Homepage, &bundle.Description,
			&bundle.LongDescription, &bundle.Docker.Image, &bundle.Docker.Tag,
			&bundle.InstalledOn, &bundle.InstalledBy)
	if err != nil {
		return bundle, cogerr.Wrap(errs.ErrNoSuchBundle, err)
	}

	// Load permissions
	query = `SELECT permission
		FROM bundle_permissions
		WHERE bundle_name=$1 AND bundle_version=$2
		ORDER BY index`
	rows, err := db.Query(query, name, version)
	if err != nil {
		return bundle, cogerr.Wrap(errs.ErrDataAccess, err)
	}

	permissions := make([]string, 0)
	for rows.NextResultSet() && rows.Next() {
		var perm string

		err = rows.Scan(&perm)
		if err != nil {
			return bundle, cogerr.Wrap(errs.ErrDataAccess, err)
		}

		permissions = append(permissions, perm)
	}
	bundle.Permissions = permissions

	// Load commands
	query = `SELECT name, description, executable
		FROM bundle_commands
		WHERE bundle_name=$1 AND bundle_version=$2`
	rows, err = db.Query(query, name, version)
	if err != nil {
		return bundle, cogerr.Wrap(errs.ErrDataAccess, err)
	}

	commands := make(map[string]data.BundleCommand, 0)
	for rows.NextResultSet() && rows.Next() {
		command := data.BundleCommand{}

		err = rows.Scan(&command.Name, &command.Description, &command.Executable)
		if err != nil {
			return bundle, cogerr.Wrap(errs.ErrDataAccess, err)
		}

		cmdQuery := `SELECT rule
			FROM bundle_command_rules
			WHERE bundle_name=$1 AND bundle_version=$2 AND command_name=$3`
		cmdRows, err := db.Query(cmdQuery, name, version, command.Name)
		if err != nil {
			return bundle, cogerr.Wrap(errs.ErrDataAccess, err)
		}

		rules := make([]string, 0)
		for cmdRows.NextResultSet() && cmdRows.Next() {
			var rule string

			err = cmdRows.Scan(&rule)
			if err != nil {
				return bundle, cogerr.Wrap(errs.ErrDataAccess, err)
			}

			rules = append(rules, rule)
		}

		command.Rules = rules
		commands[command.Name] = command
	}

	bundle.Commands = commands

	return bundle, nil
}
