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

package data

import (
	"crypto/rand"
	"encoding/base64"
	"errors"

	gorterr "github.com/clockworksoul/gort/errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrCryptoHash is returned by HashPassword and will wrap its
	// underlying error.
	ErrCryptoHash = errors.New("failed to generate password hash")

	// ErrCryptoIO is returned by GenerateRandomToken if it can't retrieve
	// random bytes from rand.Read()
	ErrCryptoIO = errors.New("failed to retrieve randomness")
)

// CompareHashAndPassword receives a plaintext password and its hash, and
// returns true if they match.
func CompareHashAndPassword(hashedPassword string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

// GenerateRandomToken generates a random character token.
func GenerateRandomToken(length int) (string, error) {
	byteCount := (length * 3) / 4
	bytes := make([]byte, byteCount)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", gorterr.Wrap(ErrCryptoIO, err)
	}

	sEnc := base64.StdEncoding.EncodeToString(bytes)

	return sEnc, nil
}

// HashPassword receives a plaintext password and returns its hashed equivalent.
func HashPassword(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.MinCost)
	if err != nil {
		return "", gorterr.Wrap(ErrCryptoHash, err)
	}

	return string(hash), nil
}
