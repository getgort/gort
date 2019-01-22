package postgres

import (
	"errors"
	"time"

	"github.com/clockworksoul/cog2/dal"
)

// TokenEvaluate will test a token for validity. It returns true if the token
// exists and is still within its valid period; false otherwise.
func (da PostgresDataAccess) TokenEvaluate(tokenString string) bool {
	token, err := da.TokenRetrieveByToken(tokenString)
	if err != nil {
		return false
	}

	return !token.IsExpired()
}

// TokenGenerate generates a new token for the given user with a specified
// expiration duration. Any existing token for this user will be automatically
// invalidated. If the user doesn't exist an error is returned.
func (da PostgresDataAccess) TokenGenerate(username string, duration time.Duration) (dal.Token, error) {
	exists, err := da.UserExists(username)
	if err != nil {
		return dal.Token{}, err
	}
	if !exists {
		return dal.Token{}, errors.New("no such user")
	}

	// If a token already exists for this user, automatically invalidate it.
	token, err := da.TokenRetrieveByUser(username)
	if err == nil {
		da.TokenInvalidate(token.Token)
	}

	tokenString, err := dal.GenerateRandomToken(64)
	if err != nil {
		return dal.Token{}, err
	}

	validFrom := time.Now().UTC()
	validUntil := validFrom.Add(duration)

	token = dal.Token{
		Duration:   duration,
		Token:      tokenString,
		User:       username,
		ValidFrom:  validFrom,
		ValidUntil: validUntil,
	}

	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return dal.Token{}, err
	}

	query := `INSERT INTO tokens (token, username, valid_from, valid_until)
	VALUES ($1, $2, $3, $4);`
	_, err = db.Exec(query, token.Token, token.User, token.ValidFrom, token.ValidUntil)
	if err != nil {
		return dal.Token{}, err
	}

	return token, nil
}

// TokenInvalidate immediately invalidates the specified token. An error is
// returned if the token doesn't exist.
func (da PostgresDataAccess) TokenInvalidate(tokenString string) error {
	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return err
	}

	query := `DELETE FROM tokens WHERE token=$1;`
	_, err = db.Exec(query, tokenString)
	if err != nil {
		return err
	}

	return nil
}

// TokenRetrieveByUser retrieves the token associated with a username. An
// error is returned if no such token (or user) exists.
func (da PostgresDataAccess) TokenRetrieveByUser(username string) (dal.Token, error) {
	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return dal.Token{}, err
	}

	// There will be more here eventually
	query := `SELECT token, username, valid_from, valid_until
		FROM tokens
		WHERE username=$1`

	token := dal.Token{}
	err = db.
		QueryRow(query, username).
		Scan(&token.Token, &token.User, &token.ValidFrom, &token.ValidUntil)
	token.Duration = token.ValidUntil.Sub(token.ValidFrom)

	return token, err
}

// TokenRetrieveByToken retrieves the token by its value. An error is returned
// if no such token exists.
func (da PostgresDataAccess) TokenRetrieveByToken(tokenString string) (dal.Token, error) {
	db, err := da.connect("cog")
	defer db.Close()
	if err != nil {
		return dal.Token{}, err
	}

	// There will be more here eventually
	query := `SELECT token, username, valid_from, valid_until
		FROM tokens
		WHERE token=$1`

	token := dal.Token{}
	err = db.
		QueryRow(query, tokenString).
		Scan(&token.Token, &token.User, &token.ValidFrom, &token.ValidUntil)
	token.Duration = token.ValidUntil.Sub(token.ValidFrom)

	return token, err
}
