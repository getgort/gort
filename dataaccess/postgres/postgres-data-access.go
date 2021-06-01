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
	"database/sql"
	"fmt"

	"github.com/getgort/gort/data"
	"github.com/getgort/gort/dataaccess/errs"
	gerr "github.com/getgort/gort/errors"

	_ "github.com/lib/pq" // Load the Postgres drivers
)

// PostgresDataAccess is a data access implementation backed by a database.
type PostgresDataAccess struct {
	configs data.DatabaseConfigs
}

// NewPostgresDataAccess returns a new PostgresDataAccess based on the
// supplied config.
func NewPostgresDataAccess(configs data.DatabaseConfigs) PostgresDataAccess {
	return PostgresDataAccess{configs: configs}
}

// Initialize sets up the database.
func (da PostgresDataAccess) Initialize() error {
	// Does the database exist? If not, create it.
	err := da.ensureGortDatabaseExists()
	if err != nil {
		return err
	}

	// Establish a connection to the "gort" database
	db, err := da.connect("gort")
	if err != nil {
		return err
	}
	defer db.Close()

	// Check whether the users table exists
	exists, err := da.tableExists("users", db)
	if err != nil {
		return err
	}

	// If not, assume none of them do. Create them all.
	if !exists {
		err = da.createUsersTable(db)
		if err != nil {
			return err
		}
	}

	// Check whether the users table exists
	exists, err = da.tableExists("groups", db)
	if err != nil {
		return err
	}

	// If not, assume none of them do. Create them all.
	if !exists {
		err = da.createGroupsTable(db)
		if err != nil {
			return err
		}
	}

	// Check whether the users table exists
	exists, err = da.tableExists("groupusers", db)
	if err != nil {
		return err
	}

	// If not, assume none of them do. Create them all.
	if !exists {
		err = da.createGroupUsersTable(db)
		if err != nil {
			return err
		}
	}

	// Check whether the tokens table exists
	exists, err = da.tableExists("tokens", db)
	if err != nil {
		return err
	}

	// If not, assume none of them do. Create them all.
	if !exists {
		err = da.createTokensTable(db)
		if err != nil {
			return err
		}
	}

	// Check whether the bundles table exists
	exists, err = da.tableExists("bundles", db)
	if err != nil {
		return err
	}

	// If not, assume none of them do. Create them all.
	if !exists {
		err = da.createBundlesTables(db)
		if err != nil {
			return err
		}
	}

	return nil
}

func (da PostgresDataAccess) gortDatabaseExists(db *sql.DB) (bool, error) {
	rows, err := db.Query("SELECT datname " +
		"FROM pg_database " +
		"WHERE datistemplate = false AND datname = 'gort'")
	if err != nil {
		return false, gerr.Wrap(errs.ErrDataAccess, err)
	}
	defer rows.Close()

	datname := ""

	for rows.Next() {
		rows.Scan(&datname)

		if datname == "gort" {
			return true, nil
		}
	}

	if err := rows.Err(); err != nil {
		return false, gerr.Wrap(errs.ErrDataAccess, err)
	}

	return false, nil
}

func (da PostgresDataAccess) connect(dbname string) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		da.configs.Host, da.configs.Port, da.configs.User,
		da.configs.Password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, gerr.Wrap(errs.ErrDataAccess, err)
	}

	err = db.Ping()
	if err != nil {
		return nil, gerr.Wrap(errs.ErrDataAccess, err)
	}

	return db, nil
}

func (da PostgresDataAccess) createBundlesTables(db *sql.DB) error {
	var err error

	createBundlesQuery := `CREATE TABLE bundles (
		gort_bundle_version  INT NOT NULL CHECK(gort_bundle_version > 0),
		name				TEXT NOT NULL CHECK(name <> ''),
		version				TEXT NOT NULL CHECK(version <> ''),
		author				TEXT,
		homepage			TEXT,
		description			TEXT NOT NULL CHECK(description <> ''),
		long_description	TEXT,
		docker_image		TEXT,
		docker_tag			TEXT,
		install_timestamp	TIMESTAMP WITH TIME ZONE,
		install_user		TEXT,
		CONSTRAINT 			unq_bundle UNIQUE(name, version),
		PRIMARY KEY 		(name, version)
	);

	ALTER TABLE bundles ALTER COLUMN install_timestamp SET DEFAULT now();

	CREATE TABLE bundle_enabled (
		bundle_name			TEXT NOT NULL,
		bundle_version		TEXT NOT NULL,
		CONSTRAINT			unq_bundle_enabled UNIQUE(bundle_name),
		PRIMARY KEY			(bundle_name),
		FOREIGN KEY 		(bundle_name, bundle_version) REFERENCES bundles(name, version)
	);

	CREATE TABLE bundle_permissions (
		bundle_name			TEXT NOT NULL,
		bundle_version		TEXT NOT NULL,
		index				INT NOT NULL CHECK(index >= 0),
		permission			TEXT,
		CONSTRAINT			unq_bundle_permission UNIQUE(bundle_name, bundle_version, index),
		PRIMARY KEY			(bundle_name, bundle_version, index),
		FOREIGN KEY 		(bundle_name, bundle_version) REFERENCES bundles(name, version)
	);

	CREATE TABLE bundle_commands (
		bundle_name			TEXT NOT NULL,
		bundle_version		TEXT NOT NULL,
		name				TEXT NOT NULL CHECK(name <> ''),
		description			TEXT NOT NULL CHECK(description <> ''),
		executable			TEXT NOT NULL CHECK(executable <> ''),
		CONSTRAINT			unq_bundle_command UNIQUE(bundle_name, bundle_version, name),
		PRIMARY KEY			(bundle_name, bundle_version, name),
		FOREIGN KEY 		(bundle_name, bundle_version) REFERENCES bundles(name, version)
	);

	CREATE TABLE bundle_command_rules (
		bundle_name			TEXT NOT NULL,
		bundle_version		TEXT NOT NULL,
		command_name		TEXT NOT NULL,
		rule				TEXT NOT NULL CHECK(rule <> ''),
		PRIMARY KEY			(bundle_name, bundle_version, command_name),
		FOREIGN KEY 		(bundle_name, bundle_version, command_name)
			REFERENCES bundle_commands(bundle_name, bundle_version, name)
	);
	`

	_, err = db.Exec(createBundlesQuery)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

func (da PostgresDataAccess) createGroupsTable(db *sql.DB) error {
	var err error

	createGroupQuery := `CREATE TABLE groups (
		groupname TEXT PRIMARY KEY
	  );`

	_, err = db.Exec(createGroupQuery)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

func (da PostgresDataAccess) createGroupUsersTable(db *sql.DB) error {
	var err error

	createGroupUsersQuery := `CREATE TABLE groupusers (
		groupname TEXT REFERENCES groups,
		username  TEXT REFERENCES users,
		PRIMARY KEY (groupname, username)
	);`

	_, err = db.Exec(createGroupUsersQuery)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

func (da PostgresDataAccess) createTokensTable(db *sql.DB) error {
	var err error

	createTokensQuery := `CREATE TABLE tokens (
		token       TEXT,
		username    TEXT REFERENCES users,
		valid_from  TIMESTAMP WITH TIME ZONE,
		valid_until TIMESTAMP WITH TIME ZONE,
		PRIMARY KEY (username)
	);

	CREATE UNIQUE INDEX tokens_token ON tokens (token);
	`

	_, err = db.Exec(createTokensQuery)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

func (da PostgresDataAccess) createUsersTable(db *sql.DB) error {
	var err error

	createUserQuery := `CREATE TABLE users (
		email         TEXT UNIQUE NOT NULL,
		full_name     TEXT,
		password_hash TEXT,
		username TEXT PRIMARY KEY
	  );`

	_, err = db.Exec(createUserQuery)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

// ensureGortDatabaseExists simply checks whether the "gort" database exists,
// and creates the empty database if it doesn't.
func (da PostgresDataAccess) ensureGortDatabaseExists() error {
	db, err := da.connect("postgres")
	if err != nil {
		return err
	}
	defer db.Close()

	exists, err := da.gortDatabaseExists(db)
	if err != nil {
		return err
	}

	if !exists {
		_, err := db.Exec("CREATE DATABASE Gort")
		if err != nil {
			return gerr.Wrap(errs.ErrDataAccess, err)
		}
	}

	return nil
}

func (da PostgresDataAccess) tableExists(table string, db *sql.DB) (bool, error) {
	var result string

	rows, err := db.Query(fmt.Sprintf("SELECT to_regclass('public.%s');", table))
	if err != nil {
		return false, gerr.Wrap(errs.ErrDataAccess, err)
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&result)

		if result == table {
			return true, nil
		}
	}

	if err := rows.Err(); err != nil {
		return false, gerr.Wrap(errs.ErrDataAccess, err)
	}

	return false, err
}
