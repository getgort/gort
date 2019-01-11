package postgres

import (
	"database/sql"
	"fmt"

	"github.com/clockworksoul/cog2/dal"
	"github.com/clockworksoul/cog2/data"

	_ "github.com/lib/pq" // Load the Postgres drivers
)

// PostgresDataAccess is a data access implementation backed by a database.
type PostgresDataAccess struct {
	dal.DataAccess

	configs data.DatabaseConfigs
	db      *sql.DB
}

// NewPostgresDataAccess will create and return a new PostgresDataAccess instance.
func NewPostgresDataAccess(configs data.DatabaseConfigs) dal.DataAccess {
	return PostgresDataAccess{configs: configs}
}

// Initialize sets up the database.
func (da PostgresDataAccess) Initialize() error {
	// Does the database exist? If not, create it.
	err := da.ensureCogDatabaseExists()
	if err != nil {
		return err
	}

	// Establish a connection to the "cog" database
	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return err
	}

	// Check whether the users table exists
	exists, err := da.tableExists("users", db)
	if err != nil {
		return err
	}

	// If not, assume none of them do. Create them all.
	if !exists {
		err = da.initializeCogDatabase(db)
		if err != nil {
			return err
		}
	}

	return nil
}

func (da PostgresDataAccess) cogDatabaseExists(db *sql.DB) (bool, error) {
	rows, err := db.Query("SELECT datname FROM pg_database WHERE datistemplate = false;")
	defer rows.Close()
	if err != nil {
		return false, err
	}

	datname := ""
	for rows.NextResultSet() {
		rows.Next()

		err = rows.Err()
		if err != nil {
			return false, err
		}

		rows.Scan(&datname)

		if datname == "cog" {
			return true, nil
		}
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
		return nil, err
	}

	return db, db.Ping()
}

// ensureCogDatabaseExists simply checks whether the "cog" database exists,
// and creates the empty database if it doesn't.
func (da PostgresDataAccess) ensureCogDatabaseExists() error {
	db, err := da.connect("postgres")
	defer db.Close()
	if err != nil {
		return err
	}

	exists, err := da.cogDatabaseExists(db)
	if err != nil {
		return err
	}

	if !exists {
		_, err := db.Exec("CREATE DATABASE Cog")
		if err != nil {
			return err
		}
	}

	return nil
}

func (da PostgresDataAccess) initializeCogDatabase(db *sql.DB) error {
	var err error

	createUserQuery := `CREATE TABLE users (
		email TEXT PRIMARY KEY,
		first_name TEXT,
		last_name TEXT,
		password_hash TEXT,
		username TEXT UNIQUE NOT NULL
	  );`

	_, err = db.Exec(createUserQuery)
	if err != nil {
		return err
	}

	createGroupQuery := `CREATE TABLE groups (
		name TEXT PRIMARY KEY
	  );`

	_, err = db.Exec(createGroupQuery)
	if err != nil {
		return err
	}

	return nil
}

func (da PostgresDataAccess) tableExists(table string, db *sql.DB) (bool, error) {
	var result string

	rows, err := db.Query(fmt.Sprintf("SELECT to_regclass('public.%s');", table))
	defer rows.Close()

	for rows.NextResultSet() {
		rows.Next()
		err = rows.Err()

		if err != nil {
			return false, err
		}

		rows.Scan(&result)

		if result == table {
			return true, nil
		}
	}

	return false, err
}
