package service

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateRole(t *testing.T) {
	router := createTestRouter()

	// Check role doesn't exist
	status := jsonRequest(t, router, "GET", "http://example.com/v2/roles/testrole", nil, nil)
	assert.Equal(t, status, http.StatusNotFound)

	// Create role
	status = jsonRequest(t, router, "PUT", "http://example.com/v2/roles/testrole", nil, nil)
	assert.Equal(t, status, http.StatusOK)

	// Expect role exists
	status = jsonRequest(t, router, "GET", "http://example.com/v2/roles/testrole", nil, nil)
	assert.Equal(t, status, http.StatusOK)
}

func TestDeleteRole(t *testing.T) {
	router := createTestRouter()

	// Create role
	status := jsonRequest(t, router, "PUT", "http://example.com/v2/roles/testrole", nil, nil)
	assert.Equal(t, status, http.StatusOK)

	// Delete role
	status = jsonRequest(t, router, "DELETE", "http://example.com/v2/roles/testrole", nil, nil)
	assert.Equal(t, status, http.StatusOK)

	// Check role doesn't exist
	status = jsonRequest(t, router, "GET", "http://example.com/v2/roles/testrole", nil, nil)
	assert.Equal(t, status, http.StatusNotFound)

}

func TestGrantRolePermission(t *testing.T) {
	router := createTestRouter()

	// Create role
	status := jsonRequest(t, router, "PUT", "http://example.com/v2/roles/testrole", nil, nil)
	assert.Equal(t, status, http.StatusOK)

	// Grant permission
	status = jsonRequest(t, router, "PUT", "http://example.com/v2/roles/testrole/bundles/testbundle/permissions/testpermission", nil, nil)
	assert.Equal(t, status, http.StatusOK)
}

func TestGrantRolePermissionInvalidRole(t *testing.T) {
	router := createTestRouter()

	// Create role
	status := jsonRequest(t, router, "PUT", "http://example.com/v2/roles/testrole", nil, nil)
	assert.Equal(t, status, http.StatusOK)

	// Grant permission to different role
	status = jsonRequest(t, router, "PUT", "http://example.com/v2/roles/testrole2/bundles/testbundle/permissions/testpermission", nil, nil)
	assert.Equal(t, status, http.StatusNotFound)
}

func TestRevokeRolePermission(t *testing.T) {
	router := createTestRouter()

	// Create role
	status := jsonRequest(t, router, "PUT", "http://example.com/v2/roles/testrole", nil, nil)
	assert.Equal(t, status, http.StatusOK)

	// Grant permission
	status = jsonRequest(t, router, "PUT", "http://example.com/v2/roles/testrole/bundles/testbundle/permissions/testpermission", nil, nil)
	assert.Equal(t, status, http.StatusOK)

	// Revoke permission
	status = jsonRequest(t, router, "DELETE", "http://example.com/v2/roles/testrole/bundles/testbundle/permissions/testpermission", nil, nil)
	assert.Equal(t, status, http.StatusOK)
}

func TestRevokeRolePermissionInvalidRole(t *testing.T) {
	router := createTestRouter()

	// Create role
	status := jsonRequest(t, router, "PUT", "http://example.com/v2/roles/testrole", nil, nil)
	assert.Equal(t, status, http.StatusOK)

	// Grant permission
	status = jsonRequest(t, router, "PUT", "http://example.com/v2/roles/testrole/bundles/testbundle/permissions/testpermission", nil, nil)
	assert.Equal(t, status, http.StatusOK)

	// Revoke permission from different role
	status = jsonRequest(t, router, "DELETE", "http://example.com/v2/roles/testrole2/bundles/testbundle/permissions/testpermission", nil, nil)
	assert.Equal(t, status, http.StatusNotFound)
}

func TestRevokeRolePermissionInvalidPermission(t *testing.T) {
	router := createTestRouter()

	// Create role
	status := jsonRequest(t, router, "PUT", "http://example.com/v2/roles/testrole", nil, nil)
	assert.Equal(t, status, http.StatusOK)

	// Grant permission
	status = jsonRequest(t, router, "PUT", "http://example.com/v2/roles/testrole/bundles/testbundle/permissions/testpermission", nil, nil)
	assert.Equal(t, status, http.StatusOK)

	// Revoke permission that doesn't exist, will be ignored and return 200
	// NOTE: Should we check for this? There may be implications with versioning.
	status = jsonRequest(t, router, "DELETE", "http://example.com/v2/roles/testrole/bundles/testbundle/permissions/testpermission2", nil, nil)
	assert.Equal(t, status, http.StatusOK)
}
