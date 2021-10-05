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

package service

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/getgort/gort/data/rest"
)

func TestGrantGroupRole(t *testing.T) {
	router := createTestRouter()

	// Create group
	NewResponseTester("PUT", "http://example.com/v2/groups/groupTestGrantGroupRole").WithBody(rest.Group{Name: "groupTestGrantGroupRole"}).WithStatus(http.StatusOK).Test(t, router)

	// Create role
	NewResponseTester("PUT", "http://example.com/v2/roles/roleTestGrantGroupRole").WithStatus(http.StatusOK).Test(t, router)

	// Check no roles
	roles := []rest.Role{}
	NewResponseTester("GET", "http://example.com/v2/groups/groupTestGrantGroupRole/roles").WithOutput(&roles).WithStatus(http.StatusOK).Test(t, router)
	assert.Equal(t, len(roles), 0)

	// Grant role
	NewResponseTester("PUT", "http://example.com/v2/groups/groupTestGrantGroupRole/roles/roleTestGrantGroupRole").WithStatus(http.StatusOK).Test(t, router)

	// Check new role added
	NewResponseTester("GET", "http://example.com/v2/groups/groupTestGrantGroupRole/roles").WithOutput(&roles).WithStatus(http.StatusOK).Test(t, router)
	assert.Equal(t, len(roles), 1)
}

func TestGrantGroupRoleInvalidGroup(t *testing.T) {
	router := createTestRouter()

	// Create group
	NewResponseTester("PUT", "http://example.com/v2/groups/groupTestGrantGroupRoleInvalidGroup").WithBody(rest.Group{Name: "groupTestGrantGroupRoleInvalidGroup"}).WithStatus(http.StatusOK).Test(t, router)

	// Create role
	NewResponseTester("PUT", "http://example.com/v2/roles/roleTestGrantGroupRoleInvalidGroup").WithStatus(http.StatusOK).Test(t, router)

	// Grant role
	NewResponseTester("PUT", "http://example.com/v2/groups/groupTestGrantGroupRoleInvalidGroup2/roles/testrole").WithStatus(http.StatusNotFound).Test(t, router)

	// Check 0 roles
	roles := []rest.Role{}
	NewResponseTester("GET", "http://example.com/v2/groups/groupTestGrantGroupRoleInvalidGroup/roles").WithOutput(&roles).WithStatus(http.StatusOK).Test(t, router)
	assert.Equal(t, len(roles), 0)
}

func TestGrantGroupRoleInvalidRole(t *testing.T) {
	router := createTestRouter()

	// Create group
	NewResponseTester("PUT", "http://example.com/v2/groups/groupTestGrantGroupRoleInvalidRole").WithBody(rest.Group{Name: "groupTestGrantGroupRoleInvalidRole"}).WithStatus(http.StatusOK).Test(t, router)

	// Create role
	NewResponseTester("PUT", "http://example.com/v2/roles/roleTestGrantGroupRoleInvalidRole").WithStatus(http.StatusOK).Test(t, router)

	// Grant role
	NewResponseTester("PUT", "http://example.com/v2/groups/groupTestGrantGroupRoleInvalidRole2/roles/roleTestGrantGroupRoleInvalidRole2").WithStatus(http.StatusNotFound).Test(t, router)

	// Check 0 roles
	roles := []rest.Role{}
	NewResponseTester("GET", "http://example.com/v2/groups/groupTestGrantGroupRoleInvalidRole/roles").WithOutput(&roles).WithStatus(http.StatusOK).Test(t, router)
	assert.Equal(t, len(roles), 0)
}

func TestRevokeGroupRole(t *testing.T) {
	router := createTestRouter()

	// Create group
	NewResponseTester("PUT", "http://example.com/v2/groups/groupTestRevokeGroupRole").WithBody(rest.Group{Name: "groupTestRevokeGroupRole"}).WithStatus(http.StatusOK).Test(t, router)

	// Create role
	NewResponseTester("PUT", "http://example.com/v2/roles/roleTestRevokeGroupRole").WithStatus(http.StatusOK).Test(t, router)

	// Grant role
	NewResponseTester("PUT", "http://example.com/v2/groups/groupTestRevokeGroupRole/roles/roleTestRevokeGroupRole").WithStatus(http.StatusOK).Test(t, router)

	// Revoke role
	NewResponseTester("DELETE", "http://example.com/v2/groups/groupTestRevokeGroupRole/roles/roleTestRevokeGroupRole").WithStatus(http.StatusOK).Test(t, router)

	// Check 0 roles
	roles := []rest.Role{}
	NewResponseTester("GET", "http://example.com/v2/groups/groupTestRevokeGroupRole/roles").WithOutput(&roles).WithStatus(http.StatusOK).Test(t, router)
	assert.Equal(t, len(roles), 0)
}

func TestRevokeGroupRoleInvalidGroup(t *testing.T) {
	router := createTestRouter()

	// Create group
	NewResponseTester("PUT", "http://example.com/v2/groups/groupTestRevokeGroupRoleInvalidGroup").WithBody(rest.Group{Name: "groupTestRevokeGroupRoleInvalidGroup"}).WithStatus(http.StatusOK).Test(t, router)

	// Create role
	NewResponseTester("PUT", "http://example.com/v2/roles/roleTestRevokeGroupRoleInvalidGroup").WithStatus(http.StatusOK).Test(t, router)

	// Grant role
	NewResponseTester("PUT", "http://example.com/v2/groups/groupTestRevokeGroupRoleInvalidGroup/roles/roleTestRevokeGroupRoleInvalidGroup").WithStatus(http.StatusOK).Test(t, router)

	// Delete role for wrong group
	NewResponseTester("DELETE", "http://example.com/v2/groups/groupTestRevokeGroupRoleInvalidGroup2/roles/roleTestRevokeGroupRoleInvalidGroup").WithStatus(http.StatusNotFound).Test(t, router)

	// Check 1 role
	roles := []rest.Role{}
	NewResponseTester("GET", "http://example.com/v2/groups/groupTestRevokeGroupRoleInvalidGroup/roles").WithOutput(&roles).WithStatus(http.StatusOK).Test(t, router)
	assert.Equal(t, len(roles), 1)
}

func TestRevokeGroupRoleInvalidRole(t *testing.T) {
	router := createTestRouter()

	// Create group
	NewResponseTester("PUT", "http://example.com/v2/groups/groupTestRevokeGroupRoleInvalidRole").WithBody(rest.Group{Name: "groupTestRevokeGroupRoleInvalidRole"}).WithStatus(http.StatusOK).Test(t, router)

	// Create role
	NewResponseTester("PUT", "http://example.com/v2/roles/roleTestRevokeGroupRoleInvalidRole").WithStatus(http.StatusOK).Test(t, router)

	// Grant role
	NewResponseTester("PUT", "http://example.com/v2/groups/groupTestRevokeGroupRoleInvalidRole/roles/roleTestRevokeGroupRoleInvalidRole").WithStatus(http.StatusOK).Test(t, router)

	// Delete wrong role
	NewResponseTester("DELETE", "http://example.com/v2/groups/groupTestRevokeGroupRoleInvalidRole/roles/roleTestRevokeGroupRoleInvalidRole2").WithStatus(http.StatusNotFound).Test(t, router)

	// Check 1 role
	roles := []rest.Role{}
	NewResponseTester("GET", "http://example.com/v2/groups/groupTestRevokeGroupRoleInvalidRole/roles").WithOutput(&roles).WithStatus(http.StatusOK).Test(t, router)
	assert.Equal(t, len(roles), 1)
}
