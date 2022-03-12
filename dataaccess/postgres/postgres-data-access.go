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
	"sync"

	"github.com/getgort/gort/data"
	"github.com/getgort/gort/dataaccess/errs"
	gerr "github.com/getgort/gort/errors"
	"github.com/getgort/gort/telemetry"

	_ "github.com/jackc/pgx/v4/stdlib"
	"go.opentelemetry.io/otel"
)

const (
	DatabaseGort = "gort"
	DriverName   = "pgx"
)

// PostgresDataAccess is a data access implementation backed by a database.
type PostgresDataAccess struct {
	configs data.DatabaseConfigs
	dbs     map[string]*sql.DB
	mutex   *sync.Mutex
}

// NewPostgresDataAccess returns a new PostgresDataAccess based on the
// supplied config.
func NewPostgresDataAccess(configs data.DatabaseConfigs) PostgresDataAccess {
	return PostgresDataAccess{
		configs: configs,
		dbs:     map[string]*sql.DB{},
		mutex:   &sync.Mutex{},
	}
}

// Initialize sets up the database.
func (da PostgresDataAccess) Initialize(ctx context.Context) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.Initialize")
	defer sp.End()

	if err := da.initializeGortData(ctx); err != nil {
		return gerr.Wrap(fmt.Errorf("failed to initialize gort data"), err)
	}
	if err := da.initializeAuditData(ctx); err != nil {
		return gerr.Wrap(fmt.Errorf("failed to initialize audit data"), err)
	}

	return nil
}

func (da PostgresDataAccess) columnExists(ctx context.Context, table, column string, conn *sql.Conn) (bool, error) {
	if table == "" || column == "" {
		return false, nil
	}

	query := "SELECT EXISTS (SELECT column_name FROM information_schema.columns WHERE table_name=$1 AND column_name=$2)"
	exists := false

	if err := conn.QueryRowContext(ctx, query, table, column).Scan(&exists); err != nil {
		return false, gerr.Wrap(errs.ErrDataAccess, err)
	}

	return exists, nil
}

func (da PostgresDataAccess) connect(ctx context.Context) (*sql.Conn, error) {
	db, err := da.open(ctx, DatabaseGort)
	if err != nil {
		return nil, gerr.WrapStr("failed to open Gort database", err)
	}

	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (da PostgresDataAccess) createBundlesTables(ctx context.Context, conn *sql.Conn) error {
	var err error

	createBundlesQuery := `CREATE TABLE IF NOT EXISTS bundles (
		gort_bundle_version INT NOT NULL CHECK(gort_bundle_version > 0),
		name				TEXT NOT NULL CHECK(name <> ''),
		version				TEXT NOT NULL CHECK(version <> ''),
		author				TEXT,
		homepage			TEXT,
		description			TEXT NOT NULL CHECK(description <> ''),
		long_description	TEXT,
		image_repository	TEXT,
		image_tag			TEXT,
		install_timestamp	TIMESTAMP WITH TIME ZONE,
		install_user		TEXT,
		CONSTRAINT 			unq_bundle UNIQUE(name, version),
		PRIMARY KEY 		(name, version)
	);

	ALTER TABLE bundles ALTER COLUMN install_timestamp SET DEFAULT now();

	CREATE TABLE IF NOT EXISTS bundle_enabled (
		bundle_name			TEXT NOT NULL,
		bundle_version		TEXT NOT NULL,
		CONSTRAINT			unq_bundle_enabled UNIQUE(bundle_name),
		PRIMARY KEY			(bundle_name),
		FOREIGN KEY 		(bundle_name, bundle_version) REFERENCES bundles(name, version)
		ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS bundle_permissions (
		bundle_name			TEXT NOT NULL,
		bundle_version		TEXT NOT NULL,
		index				INT NOT NULL CHECK(index >= 0),
		permission			TEXT,
		CONSTRAINT			unq_bundle_permission UNIQUE(bundle_name, bundle_version, index),
		PRIMARY KEY			(bundle_name, bundle_version, index),
		FOREIGN KEY 		(bundle_name, bundle_version) REFERENCES bundles(name, version)
		ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS bundle_templates (
		bundle_name			TEXT NOT NULL,
		bundle_version		TEXT NOT NULL,
		command				TEXT,
		command_error		TEXT,
		message				TEXT,
		message_error		TEXT,
		CONSTRAINT			unq_bundle_template UNIQUE(bundle_name, bundle_version),
		PRIMARY KEY			(bundle_name, bundle_version),
		FOREIGN KEY 		(bundle_name, bundle_version) REFERENCES bundles(name, version)
		ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS bundle_commands (
		bundle_name			TEXT NOT NULL,
		bundle_version		TEXT NOT NULL,
		name				TEXT NOT NULL CHECK(name <> ''),
		description			TEXT NOT NULL,
		executable			TEXT NOT NULL,
		long_description	TEXT,
		CONSTRAINT			unq_bundle_command UNIQUE(bundle_name, bundle_version, name),
		PRIMARY KEY			(bundle_name, bundle_version, name),
		FOREIGN KEY 		(bundle_name, bundle_version) REFERENCES bundles(name, version)
		ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS bundle_command_io (
		bundle_name			TEXT NOT NULL,
		bundle_version		TEXT NOT NULL,
		command_name		TEXT NOT NULL,
		io_type 			TEXT NOT NULL,
		advanced			BOOLEAN NOT NULL,
		PRIMARY KEY			(bundle_name, bundle_version, command_name, io_type),
		FOREIGN KEY 		(bundle_name, bundle_version, command_name)
		REFERENCES 			bundle_commands(bundle_name, bundle_version, name)
		ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS bundle_command_triggers (
		bundle_name			TEXT NOT NULL,
		bundle_version		TEXT NOT NULL,
		command_name		TEXT NOT NULL,
		match 				TEXT NOT NULL,
		PRIMARY KEY			(bundle_name, bundle_version, command_name, match),
		FOREIGN KEY 		(bundle_name, bundle_version, command_name)
		REFERENCES 			bundle_commands(bundle_name, bundle_version, name)
		ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS bundle_command_rules (
		bundle_name			TEXT NOT NULL,
		bundle_version		TEXT NOT NULL,
		command_name		TEXT NOT NULL,
		rule				TEXT NOT NULL CHECK(rule <> ''),
		PRIMARY KEY			(bundle_name, bundle_version, command_name, rule),
		FOREIGN KEY 		(bundle_name, bundle_version, command_name)
		REFERENCES 			bundle_commands(bundle_name, bundle_version, name)
		ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS bundle_command_templates (
		bundle_name			TEXT NOT NULL,
		bundle_version		TEXT NOT NULL,
		command_name		TEXT NOT NULL,
		command				TEXT,
		command_error		TEXT,
		message				TEXT,
		message_error		TEXT,
		CONSTRAINT			unq_bundle_command_templates UNIQUE(bundle_name, bundle_version, command_name),
		PRIMARY KEY			(bundle_name, bundle_version, command_name),
		FOREIGN KEY 		(bundle_name, bundle_version, command_name)
		REFERENCES 			bundle_commands(bundle_name, bundle_version, name)
		ON DELETE CASCADE
	);
	`

	_, err = conn.ExecContext(ctx, createBundlesQuery)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

func (da PostgresDataAccess) createBundleKubernetesTables(ctx context.Context, conn *sql.Conn) error {
	var err error

	createBundlesQuery := `CREATE TABLE bundle_kubernetes (
		bundle_version 			TEXT NOT NULL,
		bundle_name				TEXT NOT NULL,
		service_account_name 	TEXT NOT NULL,
		env_secret            TEXT NOT NULL
	);
	`

	_, err = conn.ExecContext(ctx, createBundlesQuery)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

func (da PostgresDataAccess) createConfigsTable(ctx context.Context, conn *sql.Conn) error {
	var err error

	createConfigsQuery := `CREATE TABLE configs (
		bundle_name		TEXT NOT NULL CHECK(bundle_name <> ''),
		layer			TEXT NOT NULL,
		owner			TEXT NOT NULL,
		key				TEXT NOT NULL,
		value			TEXT NOT NULL,
		secret			BOOLEAN NOT NULL
	);`

	_, err = conn.ExecContext(ctx, createConfigsQuery)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

func (da PostgresDataAccess) createGroupsTable(ctx context.Context, conn *sql.Conn) error {
	var err error

	createGroupQuery := `CREATE TABLE groups (
		groupname TEXT PRIMARY KEY
	  );`

	_, err = conn.ExecContext(ctx, createGroupQuery)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

func (da PostgresDataAccess) createGroupUsersTable(ctx context.Context, conn *sql.Conn) error {
	var err error

	createGroupUsersQuery := `CREATE TABLE groupusers (
		groupname TEXT REFERENCES groups,
		username  TEXT REFERENCES users,
		PRIMARY KEY (groupname, username)
	);`

	_, err = conn.ExecContext(ctx, createGroupUsersQuery)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

func (da PostgresDataAccess) createRolesTables(ctx context.Context, conn *sql.Conn) error {
	var err error

	createRolesQuery := `CREATE TABLE roles (
		role_name		TEXT NOT NULL,
		PRIMARY KEY 	(role_name)
	);

	CREATE TABLE group_roles (
		group_name		TEXT NOT NULL,
		role_name		TEXT NOT NULL,
		CONSTRAINT		unq_group_role UNIQUE(group_name, role_name),
		PRIMARY KEY		(group_name, role_name),
		FOREIGN KEY 	(group_name) REFERENCES groups(groupname)
		ON DELETE CASCADE
	);

	CREATE TABLE role_permissions (
		role_name			TEXT NOT NULL,
		bundle_name			TEXT NOT NULL,
		permission			TEXT NOT NULL,
		CONSTRAINT			unq_role_permission UNIQUE(role_name, bundle_name, permission),
		PRIMARY KEY			(role_name, bundle_name, permission),
		FOREIGN KEY 		(role_name) REFERENCES roles(role_name)
		ON DELETE CASCADE
	);
	`

	_, err = conn.ExecContext(ctx, createRolesQuery)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

func (da PostgresDataAccess) createTokensTable(ctx context.Context, conn *sql.Conn) error {
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

	_, err = conn.ExecContext(ctx, createTokensQuery)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

func (da PostgresDataAccess) createUsersTable(ctx context.Context, conn *sql.Conn) error {
	var err error

	createUserQuery := `CREATE TABLE users (
		email         	TEXT,
		full_name     	TEXT,
		password_hash 	TEXT,
		username 		TEXT PRIMARY KEY
	  );`

	_, err = conn.ExecContext(ctx, createUserQuery)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

func (da PostgresDataAccess) createUsersAdapterIDsTable(ctx context.Context, conn *sql.Conn) error {
	var err error

	createTableQuery := `CREATE TABLE user_adapter_ids (
		username            TEXT NOT NULL,
		adapter             TEXT NOT NULL,
		id                  TEXT NOT NULL,
		CONSTRAINT          unq_adapter_id UNIQUE(username, adapter),
		PRIMARY KEY         (adapter, id),
		FOREIGN KEY         (username) REFERENCES users(username)
		ON DELETE CASCADE
	);
	`

	_, err = conn.ExecContext(ctx, createTableQuery)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

func (da PostgresDataAccess) databaseExists(ctx context.Context, conn *sql.Conn, dbName string) (bool, error) {
	const query = `SELECT datname
		FROM pg_database
		WHERE datistemplate = false AND datname = $1`

	rows, err := conn.QueryContext(ctx, query, dbName)
	if err != nil {
		return false, gerr.Wrap(errs.ErrDataAccess, err)
	}
	defer rows.Close()

	datname := ""

	for rows.Next() {
		rows.Scan(&datname)

		if datname == dbName {
			return true, nil
		}
	}

	if err := rows.Err(); err != nil {
		return false, gerr.Wrap(errs.ErrDataAccess, err)
	}

	return false, nil
}

// ensureGortDatabaseExists simply checks whether the dbname database exists,
// and creates the empty database if it doesn't.
func (da PostgresDataAccess) ensureDatabaseExists(ctx context.Context, dbName string) error {
	db, err := da.open(ctx, "postgres")
	if err != nil {
		return gerr.WrapStr("failed to open postgres database", err)
	}

	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	exists, err := da.databaseExists(ctx, conn, dbName)
	if err != nil {
		return err
	}

	if !exists {
		_, err := conn.ExecContext(ctx, "CREATE DATABASE ?", dbName)
		if err != nil {
			return gerr.Wrap(errs.ErrDataAccess,
				gerr.Wrap(fmt.Errorf("failed to create database"),
					err))
		}
	}

	return nil
}

func (da PostgresDataAccess) initializeAuditData(ctx context.Context) error {
	// Does the database exist? If not, create it.
	err := da.ensureDatabaseExists(ctx, DatabaseGort)
	if err != nil {
		return gerr.Wrap(fmt.Errorf("cannot ensure gort database exists"), err)
	}

	// Establish a connection to the "gort" database
	conn, err := da.connect(ctx)
	if err != nil {
		return gerr.Wrap(fmt.Errorf("cannot connect to gort database"), err)
	}
	defer conn.Close()

	// Check whether the users table exists
	exists, err := da.tableExists(ctx, "commands", conn)
	if err != nil {
		return err
	}

	// If not, assume none of them do. Create them all.
	if !exists {
		err = da.createCommandsTable(ctx, conn)
		if err != nil {
			return gerr.Wrap(fmt.Errorf("failed to create commands table"), err)
		}
	}

	return nil
}

func (da PostgresDataAccess) initializeGortData(ctx context.Context) error {
	// Does the database exist? If not, create it.
	err := da.ensureDatabaseExists(ctx, DatabaseGort)
	if err != nil {
		return err
	}

	// Establish a connection to the "gort" database
	conn, err := da.connect(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Check whether the users table exists
	exists, err := da.tableExists(ctx, "users", conn)
	if err != nil {
		return err
	}
	if !exists {
		err = da.createUsersTable(ctx, conn)
		if err != nil {
			return err
		}
	}

	// Check whether the user adapter ids table exists
	if exists, err = da.tableExists(ctx, "user_adapter_ids", conn); err != nil {
		return err
	} else if !exists {
		err = da.createUsersAdapterIDsTable(ctx, conn)
		if err != nil {
			return err
		}
	}

	// Check whether the groups table exists
	exists, err = da.tableExists(ctx, "groups", conn)
	if err != nil {
		return err
	}
	if !exists {
		err = da.createGroupsTable(ctx, conn)
		if err != nil {
			return err
		}
	}

	// Check whether the groupusers table exists
	exists, err = da.tableExists(ctx, "groupusers", conn)
	if err != nil {
		return err
	}
	if !exists {
		err = da.createGroupUsersTable(ctx, conn)
		if err != nil {
			return err
		}
	}

	// Check whether the tokens table exists
	exists, err = da.tableExists(ctx, "tokens", conn)
	if err != nil {
		return err
	}
	if !exists {
		err = da.createTokensTable(ctx, conn)
		if err != nil {
			return err
		}
	}

	// Upsert bundles tables to make sure it and related tables exist with appropriate columns
	err = da.createBundlesTables(ctx, conn)
	if err != nil {
		return err
	}

	// Check whether the bundles_kubernetes table exists
	exists, err = da.tableExists(ctx, "bundle_kubernetes", conn)
	if err != nil {
		return err
	}
	if !exists {
		err = da.createBundleKubernetesTables(ctx, conn)
		if err != nil {
			return err
		}
	}

	// Check whether the roles table exists
	exists, err = da.tableExists(ctx, "roles", conn)
	if err != nil {
		return err
	}
	if !exists {
		err = da.createRolesTables(ctx, conn)
		if err != nil {
			return err
		}
	}

	// Check whether the configs table exists
	exists, err = da.tableExists(ctx, "configs", conn)
	if err != nil {
		return err
	}
	if !exists {
		err = da.createConfigsTable(ctx, conn)
		if err != nil {
			return err
		}
	}

	return nil
}

func (da PostgresDataAccess) open(ctx context.Context, databaseName string) (*sql.DB, error) {
	da.mutex.Lock()
	defer da.mutex.Unlock()

	if db, exists := da.dbs[databaseName]; exists {
		return db, nil
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s database=%s",
		da.configs.Host, da.configs.Port, da.configs.User,
		da.configs.Password, databaseName)

	if !da.configs.SSLEnabled {
		psqlInfo = psqlInfo + " sslmode=disable"
	}

	db, err := sql.Open(DriverName, psqlInfo)
	if err != nil {
		return nil, gerr.WrapStr(fmt.Sprintf("failed to open database %q", databaseName), err)
	}

	if da.configs.MaxIdleConnections != 0 {
		db.SetMaxIdleConns(da.configs.MaxIdleConnections)
	}
	if da.configs.MaxOpenConnections != 0 {
		db.SetMaxOpenConns(da.configs.MaxOpenConnections)
	}
	if da.configs.ConnectionMaxIdleTime != 0 {
		db.SetConnMaxIdleTime(da.configs.ConnectionMaxIdleTime)
	}
	if da.configs.ConnectionMaxLifetime != 0 {
		db.SetConnMaxLifetime(da.configs.ConnectionMaxLifetime)
	}

	err = db.PingContext(ctx)
	if err != nil {
		return nil, gerr.Wrap(errs.ErrDataAccess, err)
	}

	da.dbs[databaseName] = db

	return db, nil
}

func (da PostgresDataAccess) tableExists(ctx context.Context, table string, conn *sql.Conn) (bool, error) {
	var result string

	if table == "" {
		return false, nil
	}

	rows, err := conn.QueryContext(ctx, fmt.Sprintf("SELECT to_regclass('public.%s');", table))
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
