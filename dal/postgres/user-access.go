package postgres

import (
	"fmt"

	"github.com/clockworksoul/cog2/dal"

	"github.com/clockworksoul/cog2/data/rest"
)

// UserAuthenticate authenticates a username/password combination.
func (da PostgresDataAccess) UserAuthenticate(username string, password string) (bool, error) {
	exists, err := da.UserExists(username)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, fmt.Errorf("no such user: %s", username)
	}

	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return false, err
	}

	query := `SELECT password_hash
		FROM users
		WHERE username=$1`

	var hash string
	err = db.QueryRow(query, username).Scan(&hash)

	return dal.CompareHashAndPassword(hash, password), err
}

// UserCreate is used to create a new Cog user in the data store. An error is
// returned if the username is empty or if a user already exists.
func (da PostgresDataAccess) UserCreate(user rest.User) error {
	if user.Username == "" {
		return fmt.Errorf("empty username")
	}

	exists, err := da.UserExists(user.Username)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("user %s already exists", user.Username)
	}

	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return err
	}

	hash := ""
	if user.Password != "" {
		hash, err = dal.HashPassword(user.Password)
		if err != nil {
			return err
		}
	}

	query := `INSERT INTO users (email, first_name, last_name, password_hash, username)
		 VALUES ($1, $2, $3, $4, $5);`
	_, err = db.Exec(query, user.Email, user.FirstName, user.LastName, hash, user.Username)

	return err
}

// UserDelete deletes an existing user from the data store. An error is
// returned if the username parameter is empty or if the user doesn't
// exist.
func (da PostgresDataAccess) UserDelete(username string) error {
	if username == "" {
		return fmt.Errorf("empty username")
	}

	exists, err := da.UserExists(username)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("no such user: %s", username)
	}

	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return err
	}

	query := `DELETE FROM groupusers WHERE username=$1;`
	_, err = db.Exec(query, username)
	if err != nil {
		return err
	}

	query = "DELETE FROM tokens WHERE username=$1;"
	_, err = db.Exec(query, username)
	if err != nil {
		return err
	}

	query = "DELETE FROM users WHERE username=$1;"
	_, err = db.Exec(query, username)
	if err != nil {
		return err
	}

	return nil
}

// UserExists is used to determine whether a Cog user with the given username
// exists in the data store.
func (da PostgresDataAccess) UserExists(username string) (bool, error) {
	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return false, err
	}

	query := "SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)"
	exists := false

	err = db.QueryRow(query, username).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// UserGet returns a user from the data store. An error is returned if the
// username parameter is empty or if the user doesn't exist.
func (da PostgresDataAccess) UserGet(username string) (rest.User, error) {
	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return rest.User{}, err
	}

	query := `SELECT email, first_name, last_name, username
		FROM users
		WHERE username=$1`

	user := rest.User{}
	err = db.
		QueryRow(query, username).
		Scan(&user.Email, &user.FirstName, &user.LastName, &user.Username)

	return user, err
}

// UserList returns a list of all known users in the datastore.
// Passwords are not included. Nice try.
func (da PostgresDataAccess) UserList() ([]rest.User, error) {
	users := make([]rest.User, 0)

	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return users, err
	}

	query := `SELECT email, first_name, last_name, username FROM users`
	rows, err := db.Query(query)
	if err != nil {
		return users, err
	}

	for rows.NextResultSet() && rows.Next() {
		user := rest.User{}
		rows.Scan(&user.Email, &user.FirstName, &user.LastName, &user.Username)
		users = append(users, user)
	}

	return users, nil
}

// UserUpdate is used to update an existing user. An error is returned if the
// username is empty or if the user doesn't exist.
// TODO Should we let this create users that don't exist?
func (da PostgresDataAccess) UserUpdate(user rest.User) error {
	if user.Username == "" {
		return fmt.Errorf("empty username")
	}

	exists, err := da.UserExists(user.Username)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("user %s doesn't exist", user.Username)
	}

	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return err
	}

	hash := ""
	if user.Password != "" {
		hash, err = dal.HashPassword(user.Password)
		if err != nil {
			return err
		}
	}

	query := `UPDATE users
	SET email=$1, first_name=$2, last_name=$3, password_hash=$4
	WHERE username=$5;`

	_, err = db.Exec(query, user.Email, user.FirstName, user.LastName, hash, user.Username)

	return err
}
