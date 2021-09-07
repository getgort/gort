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
)

func TestCreateRole(t *testing.T) {
	router := createTestRouter()

	// Check role doesn't exist
	NewResponseTester("GET", "http://example.com/v2/roles/testCreateRole").WithStatus(http.StatusNotFound).Test(t, router)

	// Create role
	NewResponseTester("PUT", "http://example.com/v2/roles/testCreateRole").WithStatus(http.StatusOK).Test(t, router)

	// Expect role exists
	NewResponseTester("GET", "http://example.com/v2/roles/testCreateRole").WithStatus(http.StatusOK).Test(t, router)
}

func TestDeleteRole(t *testing.T) {
	router := createTestRouter()

	// Create role
	NewResponseTester("PUT", "http://example.com/v2/roles/testDeleteRole").WithStatus(http.StatusOK).Test(t, router)

	// Delete role
	NewResponseTester("DELETE", "http://example.com/v2/roles/testDeleteRole").WithStatus(http.StatusOK).Test(t, router)

	// Check role doesn't exist
	NewResponseTester("GET", "http://example.com/v2/roles/testDeleteRole").WithStatus(http.StatusNotFound).Test(t, router)
}

func TestGrantRolePermission(t *testing.T) {
	router := createTestRouter()

	// Create role
	NewResponseTester("PUT", "http://example.com/v2/roles/testGrantRolePermission").WithStatus(http.StatusOK).Test(t, router)

	// Grant permission
	NewResponseTester("PUT", "http://example.com/v2/roles/testGrantRolePermission/bundles/testbundle/permissions/testpermission").WithStatus(http.StatusOK).Test(t, router)
}

func TestGrantRolePermissionInvalidRole(t *testing.T) {
	router := createTestRouter()

	// Create role
	NewResponseTester("PUT", "http://example.com/v2/roles/testGrantRolePermissionInvalidRole").WithStatus(http.StatusOK).Test(t, router)

	// Grant permission to different role
	NewResponseTester("PUT", "http://example.com/v2/roles/testGrantRolePermissionInvalidRole2/bundles/testbundle/permissions/testpermission").WithStatus(http.StatusNotFound).Test(t, router)
}

func TestRevokeRolePermission(t *testing.T) {
	router := createTestRouter()

	// Create role
	NewResponseTester("PUT", "http://example.com/v2/roles/testRevokeRolePermission").WithStatus(http.StatusOK).Test(t, router)

	// Grant permission
	NewResponseTester("PUT", "http://example.com/v2/roles/testRevokeRolePermission/bundles/testbundle/permissions/testpermission").WithStatus(http.StatusOK).Test(t, router)

	// Revoke permission
	NewResponseTester("DELETE", "http://example.com/v2/roles/testRevokeRolePermission/bundles/testbundle/permissions/testpermission").WithStatus(http.StatusOK).Test(t, router)
}

func TestRevokeRolePermissionInvalidRole(t *testing.T) {
	router := createTestRouter()

	// Create role
	NewResponseTester("PUT", "http://example.com/v2/roles/testRevokeRolePermissionInvalidRole").WithStatus(http.StatusOK).Test(t, router)

	// Grant permission
	NewResponseTester("PUT", "http://example.com/v2/roles/testRevokeRolePermissionInvalidRole/bundles/testbundle/permissions/testpermission").WithStatus(http.StatusOK).Test(t, router)

	// Revoke permission from different role
	NewResponseTester("DELETE", "http://example.com/v2/roles/testRevokeRolePermissionInvalidRole2/bundles/testbundle/permissions/testpermission").WithStatus(http.StatusNotFound).Test(t, router)
}

func TestRevokeRolePermissionInvalidPermission(t *testing.T) {
	router := createTestRouter()

	// Create role
	NewResponseTester("PUT", "http://example.com/v2/roles/testRevokeRolePermissionInvalidPermission").WithStatus(http.StatusOK).Test(t, router)

	// Grant permission
	NewResponseTester("PUT", "http://example.com/v2/roles/testRevokeRolePermissionInvalidPermission/bundles/testbundle/permissions/testpermission").WithStatus(http.StatusOK).Test(t, router)

	// Revoke permission that doesn't exist, will be ignored and return 200
	// NOTE: Should we check for this? There may be implications with versioning.
	NewResponseTester("DELETE", "http://example.com/v2/roles/testRevokeRolePermissionInvalidPermission/bundles/testbundle/permissions/testpermission2").WithStatus(http.StatusOK).Test(t, router)
}
