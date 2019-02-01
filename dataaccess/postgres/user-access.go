package postgres

import (
	"github.com/clockworksoul/cog2/data"
	"github.com/clockworksoul/cog2/data/rest"
	"github.com/clockworksoul/cog2/dataaccess/errs"
	cogerr "github.com/clockworksoul/cog2/errors"
)

// UserAuthenticate authenticates a username/password combination.
func (da PostgresDataAccess) UserAuthenticate(username string, password string) (bool, error) {
	exists, err := da.UserExists(username)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, errs.ErrNoSuchUser
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
	if err != nil {
		err = cogerr.Wrap(errs.ErrNoSuchUser, err)
	}

	return data.CompareHashAndPassword(hash, password), err
}

// UserCreate is used to create a new Cog user in the data store. An error is
// returned if the username is empty or if a user already exists.
func (da PostgresDataAccess) UserCreate(user rest.User) error {
	if user.Username == "" {
		return errs.ErrEmptyUserName
	}

	exists, err := da.UserExists(user.Username)
	if err != nil {
		return err
	}
	if exists {
		return errs.ErrUserExists
	}

	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return err
	}

	hash := ""
	if user.Password != "" {
		hash, err = data.HashPassword(user.Password)
		if err != nil {
			return err
		}
	}

	query := `INSERT INTO users (email, full_name, password_hash, username)
		 VALUES ($1, $2, $3, $4);`
	_, err = db.Exec(query, user.Email, user.FullName, hash, user.Username)
	if err != nil {
		err = cogerr.Wrap(errs.ErrDataAccess, err)
	}

	return err
}

// UserDelete deletes an existing user from the data store. An error is
// returned if the username parameter is empty or if the user doesn't
// exist.
func (da PostgresDataAccess) UserDelete(username string) error {
	if username == "" {
		return errs.ErrEmptyUserName
	}

	// Thou Shalt Not Delete Admin
	if username == "admin" {
		return errs.ErrAdminUndeletable
	}

	exists, err := da.UserExists(username)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchUser
	}

	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return err
	}

	query := `DELETE FROM groupusers WHERE username=$1;`
	_, err = db.Exec(query, username)
	if err != nil {
		return cogerr.Wrap(errs.ErrDataAccess, err)
	}

	query = "DELETE FROM tokens WHERE username=$1;"
	_, err = db.Exec(query, username)
	if err != nil {
		return cogerr.Wrap(errs.ErrDataAccess, err)
	}

	query = "DELETE FROM users WHERE username=$1;"
	_, err = db.Exec(query, username)
	if err != nil {
		return cogerr.Wrap(errs.ErrDataAccess, err)
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
		return false, cogerr.Wrap(errs.ErrDataAccess, err)
	}

	return exists, nil
}

// UserGet returns a user from the data store. An error is returned if the
// username parameter is empty or if the user doesn't exist.
func (da PostgresDataAccess) UserGet(username string) (rest.User, error) {
	if username == "" {
		return rest.User{}, errs.ErrEmptyUserName
	}

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
	if err != nil {
		err = cogerr.Wrap(errs.ErrNoSuchUser, err)
	}

	return user, err
}

// UserGetByEmail returns a user from the data store. An error is returned if
// the email parameter is empty or if the user doesn't exist.
func (da PostgresDataAccess) UserGetByEmail(email string) (rest.User, error) {
	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return rest.User{}, err
	}

	query := `SELECT email, full_name, username
		FROM users
		WHERE email=$1`

	user := rest.User{}
	err = db.
		QueryRow(query, email).
		Scan(&user.Email, &user.FullName, &user.Username)
	if err != nil {
		err = cogerr.Wrap(errs.ErrNoSuchUser, err)
	}

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

		err = rows.Scan(&user.Email, &user.FullName, &user.Username)
		if err != nil {
			err = cogerr.Wrap(errs.ErrNoSuchUser, err)
		}

		users = append(users, user)
	}

	return users, nil
}

// UserUpdate is used to update an existing user. An error is returned if the
// username is empty or if the user doesn't exist.
func (da PostgresDataAccess) UserUpdate(user rest.User) error {
	if user.Username == "" {
		return errs.ErrEmptyUserName
	}

	exists, err := da.UserExists(user.Username)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchUser
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

	if err != nil {
		err = cogerr.Wrap(errs.ErrNoSuchUser, err)
	}

	if user.Email != "" {
		userOld.Email = user.Email
	}

	if user.FullName != "" {
		userOld.FullName = user.FullName
	}

	if user.Password != "" {
		userOld.Password, err = data.HashPassword(user.Password)
		if err != nil {
			return err
		}
	}

	query = `UPDATE users
	SET email=$1, full_name=$2, password_hash=$3
	WHERE username=$4;`

	_, err = db.Exec(query, userOld.Email, userOld.FullName, userOld.Password, userOld.Username)

	if err != nil {
		err = cogerr.Wrap(errs.ErrDataAccess, err)
	}

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
		return groups, cogerr.Wrap(errs.ErrDataAccess, err)
	}

	for rows.NextResultSet() && rows.Next() {
		group := rest.Group{}

		err = rows.Scan(&group.Name)
		if err != nil {
			err = cogerr.Wrap(errs.ErrDataAccess, err)
		}

		groups = append(groups, group)
	}

	return groups, nil
}

// UserGroupAdd comments TBD
func (da PostgresDataAccess) UserGroupAdd(username string, groupname string) error {
	if username == "" {
		return errs.ErrEmptyUserName
	}

	if groupname == "" {
		return errs.ErrEmptyGroupName
	}

	exists, err := da.UserExists(username)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchUser
	}

	exists, err = da.GroupExists(username)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchGroup
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
	if err != nil {
		err = cogerr.Wrap(errs.ErrDataAccess, err)
	}

	return err
}

// UserGroupDelete comments TBD
func (da PostgresDataAccess) UserGroupDelete(username string, groupname string) error {
	if username == "" {
		return errs.ErrEmptyUserName
	}

	if groupname == "" {
		return errs.ErrEmptyGroupName
	}

	exists, err := da.UserExists(username)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchUser
	}

	exists, err = da.GroupExists(username)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNoSuchGroup
	}

	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return err
	}

	query := `DELETE FROM groupusers WHERE groupname=$1 AND username=$2;`

	_, err = db.Exec(query, groupname, username)
	if err != nil {
		err = cogerr.Wrap(errs.ErrDataAccess, err)
	}

	return err
}
