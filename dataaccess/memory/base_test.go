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
	"testing"

	gerrs "github.com/getgort/gort/errors"
	"github.com/stretchr/testify/assert"
)

var (
	da *InMemoryDataAccess
)

func expectErr(t *testing.T, err error, expected error) {
	if err == nil {
		t.Error("Expected an error")
	} else if !gerrs.Is(err, expected) {
		t.Errorf("Wrong error: Expected: %q Got: %q\n", expected.Error(), err.Error())
	}
}

func testInitialize(t *testing.T) {
	da = NewInMemoryDataAccess()

	err := da.Initialize()
	assert.NoError(t, err)
}

func TestMain(t *testing.T) {
	t.Run("testInitialize", testInitialize)

	t.Run("testUserAccess", testUserAccess)
	t.Run("testGroupAccess", testGroupAccess)
	t.Run("testTokenAccess", testTokenAccess)
	t.Run("testBundleAccess", testBundleAccess)
}
