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
)

func (da DataAccessTest) testTokenAccess(t *testing.T) {
	t.Run("testTokenGenerate", da.testTokenGenerate)
	t.Run("testTokenRetrieveByUser", da.testTokenRetrieveByUser)
	t.Run("testTokenRetrieveByToken", da.testTokenRetrieveByToken)
	t.Run("testTokenExpiry", da.testTokenExpiry)
	t.Run("testTokenInvalidate", da.testTokenInvalidate)
}

func (da DataAccessTest) testTokenGenerate(t *testing.T) {
	err := da.UserCreate(da.ctx, rest.User{Username: "test_generate"})
	defer da.UserDelete(da.ctx, "test_generate")
	assert.NoError(t, err)

	token, err := da.TokenGenerate(da.ctx, "test_generate", 10*time.Minute)
	defer da.TokenInvalidate(da.ctx, token.Token)
	assert.NoError(t, err)

	if token.Duration != 10*time.Minute {
		t.Errorf("Duration mismatch: %v vs %v\n", token.Duration, 10*time.Minute)
		t.FailNow()
	}

	if token.User != "test_generate" {
		t.Error("User mismatch")
		t.FailNow()
	}

	if token.ValidFrom.Add(10*time.Minute) != token.ValidUntil {
		t.Error("Validity duration mismatch")
		t.FailNow()
	}
}

func (da DataAccessTest) testTokenRetrieveByUser(t *testing.T) {
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

	if token.Token != rtoken.Token {
		t.Error("token mismatch")
		t.FailNow()
	}
}

func (da DataAccessTest) testTokenRetrieveByToken(t *testing.T) {
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

	if token.Token != rtoken.Token {
		t.Error("token mismatch")
		t.FailNow()
	}
}

func (da DataAccessTest) testTokenExpiry(t *testing.T) {
	err := da.UserCreate(da.ctx, rest.User{Username: "test_expires", Email: "test_expires"})
	defer da.UserDelete(da.ctx, "test_expires")
	assert.NoError(t, err)

	token, err := da.TokenGenerate(da.ctx, "test_expires", 1*time.Second)
	defer da.TokenInvalidate(da.ctx, token.Token)
	assert.NoError(t, err)

	if token.IsExpired() {
		t.Error("Expected token to be unexpired")
		t.FailNow()
	}

	time.Sleep(time.Second)

	if !token.IsExpired() {
		t.Error("Expected token to be expired")
		t.FailNow()
	}
}

func (da DataAccessTest) testTokenInvalidate(t *testing.T) {
	err := da.UserCreate(da.ctx, rest.User{Username: "test_invalidate", Email: "test_invalidate"})
	defer da.UserDelete(da.ctx, "test_invalidate")
	assert.NoError(t, err)

	token, err := da.TokenGenerate(da.ctx, "test_invalidate", 10*time.Minute)
	defer da.TokenInvalidate(da.ctx, token.Token)
	assert.NoError(t, err)

	if !da.TokenEvaluate(da.ctx, token.Token) {
		t.Error("Expected token to be valid")
		t.FailNow()
	}

	err = da.TokenInvalidate(da.ctx, token.Token)
	assert.NoError(t, err)

	if da.TokenEvaluate(da.ctx, token.Token) {
		t.Error("Expected token to be invalid")
		t.FailNow()
	}
}
