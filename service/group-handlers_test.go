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
	NewResponseTester("PUT", "http://example.com/v2/groups/testgroup").WithBody(rest.Group{Name: "testgroup"}).WithStatus(http.StatusOK).Test(t, router)

	// Create role
	NewResponseTester("PUT", "http://example.com/v2/roles/testrole").WithStatus(http.StatusOK).Test(t, router)

	// Check no roles
	roles := []rest.Role{}
	NewResponseTester("GET", "http://example.com/v2/groups/testgroup/roles").WithOutput(&roles).WithStatus(http.StatusOK).Test(t, router)
	assert.Equal(t, len(roles), 0)

	// Grant role
	NewResponseTester("PUT", "http://example.com/v2/groups/testgroup/roles/testrole").WithStatus(http.StatusOK).Test(t, router)

	// Check new role added
	NewResponseTester("GET", "http://example.com/v2/groups/testgroup/roles").WithOutput(&roles).WithStatus(http.StatusOK).Test(t, router)
	assert.Equal(t, len(roles), 1)
}

func TestGrantGroupRoleInvalidGroup(t *testing.T) {
	router := createTestRouter()

	// Create group
	NewResponseTester("PUT", "http://example.com/v2/groups/testgroup").WithBody(rest.Group{Name: "testgroup"}).WithStatus(http.StatusOK).Test(t, router)

	// Create role
	NewResponseTester("PUT", "http://example.com/v2/roles/testrole").WithStatus(http.StatusOK).Test(t, router)

	// Grant role
	NewResponseTester("PUT", "http://example.com/v2/groups/testgroup2/roles/testrole").WithStatus(http.StatusNotFound).Test(t, router)

	// Check 0 roles
	roles := []rest.Role{}
	NewResponseTester("GET", "http://example.com/v2/groups/testgroup/roles").WithOutput(&roles).WithStatus(http.StatusOK).Test(t, router)
	assert.Equal(t, len(roles), 0)
}

func TestGrantGroupRoleInvalidRole(t *testing.T) {
	router := createTestRouter()

	// Create group
	NewResponseTester("PUT", "http://example.com/v2/groups/testgroup").WithBody(rest.Group{Name: "testgroup"}).WithStatus(http.StatusOK).Test(t, router)

	// Create role
	NewResponseTester("PUT", "http://example.com/v2/roles/testrole").WithStatus(http.StatusOK).Test(t, router)

	// Grant role
	NewResponseTester("PUT", "http://example.com/v2/groups/testgroup2/roles/testrole2").WithStatus(http.StatusNotFound).Test(t, router)

	// Check 0 roles
	roles := []rest.Role{}
	NewResponseTester("GET", "http://example.com/v2/groups/testgroup/roles").WithOutput(&roles).WithStatus(http.StatusOK).Test(t, router)
	assert.Equal(t, len(roles), 0)
}

func TestRevokeGroupRole(t *testing.T) {
	router := createTestRouter()

	// Create group
	NewResponseTester("PUT", "http://example.com/v2/groups/testgroup").WithBody(rest.Group{Name: "testgroup"}).WithStatus(http.StatusOK).Test(t, router)

	// Create role
	NewResponseTester("PUT", "http://example.com/v2/roles/testrole").WithStatus(http.StatusOK).Test(t, router)

	// Grant role
	NewResponseTester("PUT", "http://example.com/v2/groups/testgroup/roles/testrole").WithStatus(http.StatusOK).Test(t, router)

	// Revoke role
	NewResponseTester("DELETE", "http://example.com/v2/groups/testgroup/roles/testrole").WithStatus(http.StatusOK).Test(t, router)

	// Check 0 roles
	roles := []rest.Role{}
	NewResponseTester("GET", "http://example.com/v2/groups/testgroup/roles").WithOutput(&roles).WithStatus(http.StatusOK).Test(t, router)
	assert.Equal(t, len(roles), 0)
}

func TestRevokeGroupRoleInvalidGroup(t *testing.T) {
	router := createTestRouter()

	// Create group
	NewResponseTester("PUT", "http://example.com/v2/groups/testgroup").WithBody(rest.Group{Name: "testgroup"}).WithStatus(http.StatusOK).Test(t, router)

	// Create role
	NewResponseTester("PUT", "http://example.com/v2/roles/testrole").WithStatus(http.StatusOK).Test(t, router)

	// Grant role
	NewResponseTester("PUT", "http://example.com/v2/groups/testgroup/roles/testrole").WithStatus(http.StatusOK).Test(t, router)

	// Delete role for wrong group
	NewResponseTester("DELETE", "http://example.com/v2/groups/testgroup2/roles/testrole").WithStatus(http.StatusNotFound).Test(t, router)

	// Check 1 role
	roles := []rest.Role{}
	NewResponseTester("GET", "http://example.com/v2/groups/testgroup/roles").WithOutput(&roles).WithStatus(http.StatusOK).Test(t, router)
	assert.Equal(t, len(roles), 1)
}

func TestRevokeGroupRoleInvalidRole(t *testing.T) {
	router := createTestRouter()

	// Create group
	NewResponseTester("PUT", "http://example.com/v2/groups/testgroup").WithBody(rest.Group{Name: "testgroup"}).WithStatus(http.StatusOK).Test(t, router)

	// Create role
	NewResponseTester("PUT", "http://example.com/v2/roles/testrole").WithStatus(http.StatusOK).Test(t, router)

	// Grant role
	NewResponseTester("PUT", "http://example.com/v2/groups/testgroup/roles/testrole").WithStatus(http.StatusOK).Test(t, router)

	// Delete wrong role
	NewResponseTester("DELETE", "http://example.com/v2/groups/testgroup/roles/testrole2").WithStatus(http.StatusNotFound).Test(t, router)

	// Check 1 role
	roles := []rest.Role{}
	NewResponseTester("GET", "http://example.com/v2/groups/testgroup/roles").WithOutput(&roles).WithStatus(http.StatusOK).Test(t, router)
	assert.Equal(t, len(roles), 1)
}
