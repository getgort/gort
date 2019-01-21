package memory

import (
	"errors"
	"time"

	"github.com/clockworksoul/cog2/dal"
)

var (
	tokensByUser  map[string]dal.Token // key=username
	tokensByValue map[string]dal.Token // key=token
)

func init() {
	tokensByUser = make(map[string]dal.Token)
	tokensByValue = make(map[string]dal.Token)
}

// TokenEvaluate will test a token for validity. It returns true if the token
// exists and is still within its valid period; false otherwise.
func (da InMemoryDataAccess) TokenEvaluate(tokenString string) bool {
	token, err := da.TokenRetrieveByToken(tokenString)
	if err != nil {
		return false
	}

	return !token.IsExpired()
}

// TokenGenerate generates a new token for the given user with a specified
// expiration duration. Any existing token for this user will be automatically
// invalidated. If the user doesn't exist an error is returned.
func (da InMemoryDataAccess) TokenGenerate(username string, duration time.Duration) (dal.Token, error) {
	exists, err := da.UserExists(username)
	if err != nil {
		return dal.Token{}, err
	}
	if !exists {
		return dal.Token{}, errors.New("no such user")
	}

	tokenString, err := dal.GenerateRandomToken()
	if err != nil {
		return dal.Token{}, err
	}

	validFrom := time.Now()
	validUntil := validFrom.Add(duration)

	token := dal.Token{
		Duration:   duration,
		Token:      tokenString,
		User:       username,
		ValidFrom:  validFrom,
		ValidUntil: validUntil,
	}

	tokensByUser[username] = token
	tokensByValue[tokenString] = token

	return token, nil
}

// TokenInvalidate immediately invalidates the specified token. An error is
// returned if the token doesn't exist.
func (da InMemoryDataAccess) TokenInvalidate(tokenString string) error {
	token, err := da.TokenRetrieveByToken(tokenString)
	if err != nil {
		return err
	}

	delete(tokensByUser, token.User)
	delete(tokensByValue, token.Token)

	return nil
}

// TokenRetrieveByUser retrieves the token associated with a username. An
// error is returned if no such token (or user) exists.
func (da InMemoryDataAccess) TokenRetrieveByUser(username string) (dal.Token, error) {
	if token, ok := tokensByUser[username]; ok {
		return token, nil
	}

	return dal.Token{}, errors.New("no token for given user")
}

// TokenRetrieveByToken retrieves the token by its value. An error is returned
// if no such token exists.
func (da InMemoryDataAccess) TokenRetrieveByToken(tokenString string) (dal.Token, error) {
	if token, ok := tokensByValue[tokenString]; ok {
		return token, nil
	}

	return dal.Token{}, errors.New("no such token")
}
