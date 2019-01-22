package postgres

import (
	"errors"
	"fmt"

	"github.com/clockworksoul/cog2/data/rest"
)

// GroupAddUser adds a user to a group
func (da PostgresDataAccess) GroupAddUser(groupname string, username string) error {
	if groupname == "" {
		return fmt.Errorf("empty group name")
	}

	if username == "" {
		return fmt.Errorf("empty user name")
	}

	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return err
	}

	query := `INSERT INTO groupusers (groupname, username) VALUES ($1, $2);`
	_, err = db.Exec(query, groupname, username)

	return err
}

// GroupCreate creates a new user group.
func (da PostgresDataAccess) GroupCreate(group rest.Group) error {
	if group.Name == "" {
		return fmt.Errorf("empty group name")
	}

	exists, err := da.GroupExists(group.Name)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("group %s already exists", group.Name)
	}

	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return err
	}

	query := `INSERT INTO groups (groupname) VALUES ($1);`
	_, err = db.Exec(query, group.Name)

	return err
}

// GroupDelete deletes a group.
func (da PostgresDataAccess) GroupDelete(groupname string) error {
	if groupname == "" {
		return fmt.Errorf("empty group name")
	}

	exists, err := da.GroupExists(groupname)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("no such group: %s", groupname)
	}

	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return err
	}

	query := `DELETE FROM groupusers WHERE groupname=$1;`
	_, err = db.Exec(query, groupname)
	if err != nil {
		return err
	}

	query = `DELETE FROM groups WHERE groupname=$1;`
	_, err = db.Exec(query, groupname)
	if err != nil {
		return err
	}

	return nil
}

// GroupExists is used to determine whether a group exists in the data store.
func (da PostgresDataAccess) GroupExists(groupname string) (bool, error) {
	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return false, err
	}

	query := "SELECT EXISTS(SELECT 1 FROM groups WHERE groupname=$1)"
	exists := false

	err = db.QueryRow(query, groupname).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// GroupGet gets a specific group.
func (da PostgresDataAccess) GroupGet(groupname string) (rest.Group, error) {
	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return rest.Group{}, err
	}

	// There will be more here eventually
	query := `SELECT groupname
		FROM groups
		WHERE groupname=$1`

	group := rest.Group{}
	err = db.QueryRow(query, groupname).Scan(&group.Name)
	if err != nil {
		return group, err
	}

	users, err := da.GroupListUsers(groupname)
	if err != nil {
		return group, err
	}

	group.Users = users

	return group, nil
}

// GroupGrantRole grants one or more roles to a group.
func (da PostgresDataAccess) GroupGrantRole() error {
	return errors.New("Not yet supported")
}

// GroupList returns a list of all known groups in the datastore.
// Passwords are not included. Nice try.
func (da PostgresDataAccess) GroupList() ([]rest.Group, error) {
	groups := make([]rest.Group, 0)

	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return groups, err
	}

	query := `SELECT groupname FROM groups`
	rows, err := db.Query(query)
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

// GroupListUsers returns a list of all known users in a group.
func (da PostgresDataAccess) GroupListUsers(groupname string) ([]rest.User, error) {
	users := make([]rest.User, 0)

	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return users, err
	}

	query := `SELECT email, full_name, username
	FROM users
	WHERE username IN (
		SELECT username 
		FROM groupusers
		WHERE groupname = $1
	)`

	rows, err := db.Query(query, groupname)
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

// GroupRemoveUser removes a user from a group.
func (da PostgresDataAccess) GroupRemoveUser(groupname string, username string) error {
	if groupname == "" {
		return fmt.Errorf("empty group name")
	}

	exists, err := da.GroupExists(groupname)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("no such group: %s", groupname)
	}

	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return err
	}

	query := "DELETE FROM groupusers WHERE groupname=$1 AND username=$2;"
	_, err = db.Exec(query, groupname, username)

	return err
}

// GroupRevokeRole revokes a role from a group.
func (da PostgresDataAccess) GroupRevokeRole() error {
	return errors.New("Not yet supported")
}

// GroupUpdate is used to update an existing group. An error is returned if the
// groupname is empty or if the group doesn't exist.
// TODO Should we let this create groups that don't exist?
func (da PostgresDataAccess) GroupUpdate(group rest.Group) error {
	if group.Name == "" {
		return fmt.Errorf("empty group name")
	}

	exists, err := da.UserExists(group.Name)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("group %s doesn't exist", group.Name)
	}

	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return err
	}

	// There will be more eventually
	query := `UPDATE groupname
	SET groupname=$1
	WHERE groupname=$1;`

	_, err = db.Exec(query, group.Name)

	return err
}
