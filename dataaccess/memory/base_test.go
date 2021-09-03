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

package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	ctx    context.Context
	cancel context.CancelFunc
	da     *InMemoryDataAccess
)

func testInitialize(t *testing.T) {
	// Reset for repeated runs
	Reset()

	ctx, cancel = context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	da = NewInMemoryDataAccess()

	err := da.Initialize(ctx)
	assert.NoError(t, err)
}

func TestMemoryDataAccessMain(t *testing.T) {
	t.Run("testInitialize", testInitialize)
	t.Run("testUserAccess", testUserAccess)
	t.Run("testGroupAccess", testGroupAccess)
	t.Run("testTokenAccess", testTokenAccess)
	t.Run("testBundleAccess", testBundleAccess)
	t.Run("testRoleAccess", testRoleAccess)
	t.Run("testRequestAccess", testRequestAccess)
}
