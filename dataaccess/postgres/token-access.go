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
	"time"

	"github.com/getgort/gort/data"
	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/dataaccess/errs"
	gerr "github.com/getgort/gort/errors"
)

// TokenEvaluate will test a token for validity. It returns true if the token
// exists and is still within its valid period; false otherwise.
func (da PostgresDataAccess) TokenEvaluate(ctx context.Context, tokenString string) bool {
	token, err := da.TokenRetrieveByToken(ctx, tokenString)
	if err != nil {
		return false
	}

	return !token.IsExpired()
}

// TokenGenerate generates a new token for the given user with a specified
// expiration duration. Any existing token for this user will be automatically
// invalidated. If the user doesn't exist an error is returned.
func (da PostgresDataAccess) TokenGenerate(ctx context.Context, username string, duration time.Duration) (rest.Token, error) {
	exists, err := da.UserExists(ctx, username)
	if err != nil {
		return rest.Token{}, err
	}
	if !exists {
		return rest.Token{}, errs.ErrNoSuchUser
	}

	// If a token already exists for this user, automatically invalidate it.
	token, err := da.TokenRetrieveByUser(ctx, username)
	if err == nil {
		da.TokenInvalidate(ctx, token.Token)
	}

	tokenString, err := data.GenerateRandomToken(64)
	if err != nil {
		return rest.Token{}, err
	}

	validFrom := time.Now().UTC()
	validUntil := validFrom.Add(duration)

	token = rest.Token{
		Duration:   duration,
		Token:      tokenString,
		User:       username,
		ValidFrom:  validFrom,
		ValidUntil: validUntil,
	}

	db, err := da.connect("gort")
	if err != nil {
		return rest.Token{}, err
	}
	defer db.Close()

	query := `INSERT INTO tokens (token, username, valid_from, valid_until)
	VALUES ($1, $2, $3, $4);`
	_, err = db.Exec(query, token.Token, token.User, token.ValidFrom, token.ValidUntil)
	if err != nil {
		return rest.Token{}, gerr.Wrap(errs.ErrDataAccess, err)
	}

	return token, nil
}

// TokenInvalidate immediately invalidates the specified token. An error is
// returned if the token doesn't exist.
func (da PostgresDataAccess) TokenInvalidate(ctx context.Context, tokenString string) error {
	db, err := da.connect("gort")
	if err != nil {
		return err
	}
	defer db.Close()

	query := `DELETE FROM tokens WHERE token=$1;`
	_, err = db.Exec(query, tokenString)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

// TokenRetrieveByUser retrieves the token associated with a username. An
// error is returned if no such token (or user) exists.
func (da PostgresDataAccess) TokenRetrieveByUser(ctx context.Context, username string) (rest.Token, error) {
	db, err := da.connect("gort")
	if err != nil {
		return rest.Token{}, err
	}
	defer db.Close()

	// There will be more here eventually
	query := `SELECT token, username, valid_from, valid_until
		FROM tokens
		WHERE username=$1`

	token := rest.Token{}

	err = db.
		QueryRow(query, username).
		Scan(&token.Token, &token.User, &token.ValidFrom, &token.ValidUntil)

	if err != nil {
		err = gerr.Wrap(errs.ErrNoSuchToken, err)
	}

	token.Duration = token.ValidUntil.Sub(token.ValidFrom)

	return token, err
}

// TokenRetrieveByToken retrieves the token by its value. An error is returned
// if no such token exists.
func (da PostgresDataAccess) TokenRetrieveByToken(ctx context.Context, tokenString string) (rest.Token, error) {
	db, err := da.connect("gort")
	if err != nil {
		return rest.Token{}, err
	}
	defer db.Close()

	// There will be more here eventually
	query := `SELECT token, username, valid_from, valid_until
		FROM tokens
		WHERE token=$1`

	token := rest.Token{}
	err = db.
		QueryRow(query, tokenString).
		Scan(&token.Token, &token.User, &token.ValidFrom, &token.ValidUntil)

	if err != nil {
		err = gerr.Wrap(errs.ErrNoSuchToken, err)
	}

	token.Duration = token.ValidUntil.Sub(token.ValidFrom)

	return token, err
}
