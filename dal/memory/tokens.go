package memory

import (
	"errors"
	"time"

	"github.com/clockworksoul/cog2/data"
	"github.com/clockworksoul/cog2/data/rest"
)

var (
	tokensByUser  map[string]rest.Token // key=username
	tokensByValue map[string]rest.Token // key=token
)

func init() {
	tokensByUser = make(map[string]rest.Token)
	tokensByValue = make(map[string]rest.Token)
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
func (da InMemoryDataAccess) TokenGenerate(username string, duration time.Duration) (rest.Token, error) {
	exists, err := da.UserExists(username)
	if err != nil {
		return rest.Token{}, err
	}
	if !exists {
		return rest.Token{}, errors.New("no such user")
	}

	tokenString, err := data.GenerateRandomToken(64)
	if err != nil {
		return rest.Token{}, err
	}

	validFrom := time.Now()
	validUntil := validFrom.Add(duration)

	token := rest.Token{
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
func (da InMemoryDataAccess) TokenRetrieveByUser(username string) (rest.Token, error) {
	if token, ok := tokensByUser[username]; ok {
		return token, nil
	}

	return rest.Token{}, errors.New("no token for given user")
}

// TokenRetrieveByToken retrieves the token by its value. An error is returned
// if no such token exists.
func (da InMemoryDataAccess) TokenRetrieveByToken(tokenString string) (rest.Token, error) {
	if token, ok := tokensByValue[tokenString]; ok {
		return token, nil
	}

	return rest.Token{}, errors.New("no such token")
}
