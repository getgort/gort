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
	status := jsonRequest(t, router, "PUT", "http://example.com/v2/groups/testgroup", rest.Group{Name: "testgroup"}, nil)
	assert.Equal(t, status, http.StatusOK)

	// Create role
	status = jsonRequest(t, router, "PUT", "http://example.com/v2/roles/testrole", nil, nil)
	assert.Equal(t, status, http.StatusOK)

	// Check no roles
	roles := []rest.Role{}
	status = jsonRequest(t, router, "GET", "http://example.com/v2/groups/testgroup/roles", nil, &roles)
	assert.Equal(t, status, http.StatusOK)
	assert.Equal(t, len(roles), 0)

	// Grant role
	status = jsonRequest(t, router, "PUT", "http://example.com/v2/groups/testgroup/roles/testrole", nil, nil)
	assert.Equal(t, status, http.StatusOK)

	// Check new role added
	status = jsonRequest(t, router, "GET", "http://example.com/v2/groups/testgroup/roles", nil, &roles)
	assert.Equal(t, status, http.StatusOK)
	assert.Equal(t, len(roles), 1)
}

func TestGrantGroupRoleInvalidGroup(t *testing.T) {
	router := createTestRouter()

	// Create group
	status := jsonRequest(t, router, "PUT", "http://example.com/v2/groups/testgroup", rest.Group{Name: "testgroup"}, nil)
	assert.Equal(t, status, http.StatusOK)

	// Create role
	status = jsonRequest(t, router, "PUT", "http://example.com/v2/roles/testrole", nil, nil)
	assert.Equal(t, status, http.StatusOK)

	// Grant role
	status = jsonRequest(t, router, "PUT", "http://example.com/v2/groups/testgroup2/roles/testrole", nil, nil)
	assert.Equal(t, status, http.StatusNotFound)

	// Check 0 roles
	roles := []rest.Role{}
	status = jsonRequest(t, router, "GET", "http://example.com/v2/groups/testgroup/roles", nil, &roles)
	assert.Equal(t, status, http.StatusOK)
	assert.Equal(t, len(roles), 0)
}

func TestGrantGroupRoleInvalidRole(t *testing.T) {
	router := createTestRouter()

	// Create group
	status := jsonRequest(t, router, "PUT", "http://example.com/v2/groups/testgroup", rest.Group{Name: "testgroup"}, nil)
	assert.Equal(t, status, http.StatusOK)

	// Create role
	status = jsonRequest(t, router, "PUT", "http://example.com/v2/roles/testrole", nil, nil)
	assert.Equal(t, status, http.StatusOK)

	// Grant role
	status = jsonRequest(t, router, "PUT", "http://example.com/v2/groups/testgroup/roles/testrole2", nil, nil)
	assert.Equal(t, status, http.StatusNotFound)

	// Check 0 roles
	roles := []rest.Role{}
	status = jsonRequest(t, router, "GET", "http://example.com/v2/groups/testgroup/roles", nil, &roles)
	assert.Equal(t, status, http.StatusOK)
	assert.Equal(t, len(roles), 0)
}

func TestRevokeGroupRole(t *testing.T) {
	router := createTestRouter()

	// Create group
	status := jsonRequest(t, router, "PUT", "http://example.com/v2/groups/testgroup", rest.Group{Name: "testgroup"}, nil)
	assert.Equal(t, status, http.StatusOK)

	// Create role
	status = jsonRequest(t, router, "PUT", "http://example.com/v2/roles/testrole", nil, nil)
	assert.Equal(t, status, http.StatusOK)

	// Grant role
	status = jsonRequest(t, router, "PUT", "http://example.com/v2/groups/testgroup/roles/testrole", nil, nil)
	assert.Equal(t, status, http.StatusOK)

	// Revoke role
	status = jsonRequest(t, router, "DELETE", "http://example.com/v2/groups/testgroup/roles/testrole", nil, nil)
	assert.Equal(t, status, http.StatusOK)

	// Check no roles
	roles := []rest.Role{}
	status = jsonRequest(t, router, "GET", "http://example.com/v2/groups/testgroup/roles", nil, &roles)
	assert.Equal(t, status, http.StatusOK)
	assert.Equal(t, len(roles), 0)

}

func TestRevokeGroupRoleInvalidGroup(t *testing.T) {
	router := createTestRouter()

	// Create group
	status := jsonRequest(t, router, "PUT", "http://example.com/v2/groups/testgroup", rest.Group{Name: "testgroup"}, nil)
	assert.Equal(t, status, http.StatusOK)

	// Create role
	status = jsonRequest(t, router, "PUT", "http://example.com/v2/roles/testrole", nil, nil)
	assert.Equal(t, status, http.StatusOK)

	// Grant role
	status = jsonRequest(t, router, "PUT", "http://example.com/v2/groups/testgroup/roles/testrole", nil, nil)
	assert.Equal(t, status, http.StatusOK)

	// Delete role for wrong group
	status = jsonRequest(t, router, "DELETE", "http://example.com/v2/groups/testgroup2/roles/testrole", nil, nil)
	assert.Equal(t, status, http.StatusNotFound)

	// Check 1 role
	roles := []rest.Role{}
	status = jsonRequest(t, router, "GET", "http://example.com/v2/groups/testgroup/roles", nil, &roles)
	assert.Equal(t, status, http.StatusOK)
	assert.Equal(t, len(roles), 1)
}

func TestRevokeGroupRoleInvalidRole(t *testing.T) {
	router := createTestRouter()

	// Create group
	status := jsonRequest(t, router, "PUT", "http://example.com/v2/groups/testgroup", rest.Group{Name: "testgroup"}, nil)
	assert.Equal(t, status, http.StatusOK)

	// Create role
	status = jsonRequest(t, router, "PUT", "http://example.com/v2/roles/testrole", nil, nil)
	assert.Equal(t, status, http.StatusOK)

	// Grant role
	status = jsonRequest(t, router, "PUT", "http://example.com/v2/groups/testgroup/roles/testrole", nil, nil)
	assert.Equal(t, status, http.StatusOK)

	// Delete wrong role
	status = jsonRequest(t, router, "DELETE", "http://example.com/v2/groups/testgroup/roles/testrole2", nil, nil)
	assert.Equal(t, status, http.StatusNotFound)

	// Check 1 role
	roles := []rest.Role{}
	status = jsonRequest(t, router, "GET", "http://example.com/v2/groups/testgroup/roles", nil, &roles)
	assert.Equal(t, status, http.StatusOK)
	assert.Equal(t, len(roles), 1)
}
