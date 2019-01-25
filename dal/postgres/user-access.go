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

	query := `INSERT INTO users (email, full_name, password_hash, username)
		 VALUES ($1, $2, $3, $4);`
	_, err = db.Exec(query, user.Email, user.FullName, hash, user.Username)

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

	query := `SELECT email, full_name, username
		FROM users
		WHERE username=$1`

	user := rest.User{}
	err = db.
		QueryRow(query, username).
		Scan(&user.Email, &user.FullName, &user.Username)

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

	query := `SELECT email, full_name, username FROM users`
	rows, err := db.Query(query)
	if err != nil {
		return users, err
	}

	for rows.NextResultSet() && rows.Next() {
		user := rest.User{}
		rows.Scan(&user.Email, &user.FullName, &user.Username)
		users = append(users, user)
	}

	return users, nil
}

// UserUpdate is used to update an existing user. An error is returned if the
// username is empty or if the user doesn't exist.
func (da PostgresDataAccess) UserUpdate(user rest.User) error {
	if user.Username == "" {
		return fmt.Errorf("empty username")
	}

	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return err
	}

	query := `SELECT email, full_name, username, password_hash
		FROM users
		WHERE username=$1`

	userOld := rest.User{}
	err = db.
		QueryRow(query, user.Username).
		Scan(&userOld.Email, &userOld.FullName, &userOld.Username, &userOld.Password)

	if user.Email != "" {
		userOld.Email = user.Email
	}

	if user.FullName != "" {
		userOld.FullName = user.FullName
	}

	if user.Password != "" {
		userOld.Password, err = dal.HashPassword(user.Password)
		if err != nil {
			return err
		}
	}

	query = `UPDATE users
	SET email=$1, full_name=$2, password_hash=$3
	WHERE username=$4;`

	_, err = db.Exec(query, userOld.Email, userOld.FullName, userOld.Password, userOld.Username)

	return err
}

// UserGroupList comments TBD
func (da PostgresDataAccess) UserGroupList(username string) ([]rest.Group, error) {
	groups := make([]rest.Group, 0)

	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return groups, err
	}

	query := `SELECT groupname FROM groupusers WHERE username=$1`
	rows, err := db.Query(query, username)
	if err != nil {
		return groups, err
	}

	for rows.NextResultSet() && rows.Next() {
		group := rest.Group{}
		rows.Scan(&group.Name)
		groups = append(groups, group)
	}

	return groups, nil
}

// UserGroupAdd comments TBD
func (da PostgresDataAccess) UserGroupAdd(username string, groupname string) error {
	if username == "" {
		return fmt.Errorf("empty username")
	}

	if groupname == "" {
		return fmt.Errorf("empty groupname")
	}

	exists, err := da.UserExists(username)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("user %s doesn't exist", username)
	}

	exists, err = da.GroupExists(username)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("group %s doesn't exist", groupname)
	}

	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return err
	}

	query := `UPDATE groupusers
		SET groupname=$1, username=$2
		WHERE username=$2;`

	_, err = db.Exec(query, groupname, username)

	return err
}

// UserGroupDelete comments TBD
func (da PostgresDataAccess) UserGroupDelete(username string, groupname string) error {
	if username == "" {
		return fmt.Errorf("empty username")
	}

	if groupname == "" {
		return fmt.Errorf("empty groupname")
	}

	exists, err := da.UserExists(username)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("user %s doesn't exist", username)
	}

	exists, err = da.GroupExists(username)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("group %s doesn't exist", groupname)
	}

	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return err
	}

	query := `DELETE FROM groupusers WHERE groupname=$1 AND username=$2;`

	_, err = db.Exec(query, groupname, username)

	return err
}
