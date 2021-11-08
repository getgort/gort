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

package tests

import (
	"testing"
	"time"

	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/dataaccess/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (da DataAccessTester) testTokenAccess(t *testing.T) {
	t.Run("testTokenGenerate", da.testTokenGenerate)
	t.Run("testTokenRetrieveByUser", da.testTokenRetrieveByUser)
	t.Run("testTokenRetrieveByToken", da.testTokenRetrieveByToken)
	t.Run("testTokenExpiry", da.testTokenExpiry)
	t.Run("testTokenInvalidate", da.testTokenInvalidate)
}

func (da DataAccessTester) testTokenGenerate(t *testing.T) {
	err := da.UserCreate(da.ctx, rest.User{Username: "test_generate"})
	defer da.UserDelete(da.ctx, "test_generate")
	assert.NoError(t, err)

	token, err := da.TokenGenerate(da.ctx, "test_generate", 10*time.Minute)
	defer da.TokenInvalidate(da.ctx, token.Token)
	assert.NoError(t, err)

	require.Equal(t, token.Duration, 10*time.Minute)
	require.Equal(t, token.User, "test_generate")
	require.Equal(t, token.ValidFrom.Add(10*time.Minute), token.ValidUntil)
}

func (da DataAccessTester) testTokenRetrieveByUser(t *testing.T) {
	_, err := da.TokenRetrieveByUser(da.ctx, "no-such-user")
	assert.Error(t, err, errs.ErrNoSuchToken)

	err = da.UserCreate(da.ctx, rest.User{Username: "test_uretrieve", Email: "test_uretrieve"})
	defer da.UserDelete(da.ctx, "test_uretrieve")
	assert.NoError(t, err)

	token, err := da.TokenGenerate(da.ctx, "test_uretrieve", 10*time.Minute)
	defer da.TokenInvalidate(da.ctx, token.Token)
	assert.NoError(t, err)

	rtoken, err := da.TokenRetrieveByUser(da.ctx, "test_uretrieve")
	assert.NoError(t, err)
	require.Equal(t, token.Token, rtoken.Token)
}

func (da DataAccessTester) testTokenRetrieveByToken(t *testing.T) {
	_, err := da.TokenRetrieveByToken(da.ctx, "no-such-token")
	assert.Error(t, err, errs.ErrNoSuchToken)

	err = da.UserCreate(da.ctx, rest.User{Username: "test_tretrieve", Email: "test_tretrieve"})
	defer da.UserDelete(da.ctx, "test_tretrieve")
	assert.NoError(t, err)

	token, err := da.TokenGenerate(da.ctx, "test_tretrieve", 10*time.Minute)
	defer da.TokenInvalidate(da.ctx, token.Token)
	assert.NoError(t, err)

	rtoken, err := da.TokenRetrieveByToken(da.ctx, token.Token)
	assert.NoError(t, err)
	require.Equal(t, token.Token, rtoken.Token)
}

func (da DataAccessTester) testTokenExpiry(t *testing.T) {
	err := da.UserCreate(da.ctx, rest.User{Username: "test_expires", Email: "test_expires"})
	defer da.UserDelete(da.ctx, "test_expires")
	assert.NoError(t, err)

	token, err := da.TokenGenerate(da.ctx, "test_expires", 1*time.Second)
	defer da.TokenInvalidate(da.ctx, token.Token)
	assert.NoError(t, err)
	require.False(t, token.IsExpired())

	time.Sleep(time.Second)

	require.True(t, token.IsExpired())
}

func (da DataAccessTester) testTokenInvalidate(t *testing.T) {
	err := da.UserCreate(da.ctx, rest.User{Username: "test_invalidate", Email: "test_invalidate"})
	defer da.UserDelete(da.ctx, "test_invalidate")
	assert.NoError(t, err)

	token, err := da.TokenGenerate(da.ctx, "test_invalidate", 10*time.Minute)
	defer da.TokenInvalidate(da.ctx, token.Token)
	assert.NoError(t, err)
	require.True(t, da.TokenEvaluate(da.ctx, token.Token))

	err = da.TokenInvalidate(da.ctx, token.Token)
	assert.NoError(t, err)
	require.False(t, da.TokenEvaluate(da.ctx, token.Token))
}
