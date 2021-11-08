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
	"context"
	"testing"

	"github.com/getgort/gort/dataaccess"
)

type DataAccessTest struct {
	dataaccess.DataAccess
	cancel context.CancelFunc
	ctx    context.Context
}

func NewDataAccessTest(ctx context.Context, cancel context.CancelFunc, da dataaccess.DataAccess) DataAccessTest {
	return DataAccessTest{
		DataAccess: da,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (da DataAccessTest) RunTests(t *testing.T) {
	t.Run("testUserAccess", da.testUserAccess)
	t.Run("testGroupAccess", da.testGroupAccess)
	t.Run("testTokenAccess", da.testTokenAccess)
	t.Run("testBundleAccess", da.testBundleAccess)
	t.Run("testRoleAccess", da.testRoleAccess)
	t.Run("testRequestAccess", da.testRequestAccess)
}
