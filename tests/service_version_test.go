package tests

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thilak009/kong-assignment/models"
)

// TestCreateServiceVersion tests POST /v1/orgs/{orgId}/services/{serviceId}/versions endpoint
func TestCreateServiceVersion(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	t.Run("Success", func(t *testing.T) {
		// Setup test user, organization and service
		_, token := helpers.CreateTestUser("test@example.com", "Test User", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")
		service := helpers.CreateTestService(token, org.ID, "Test Service", "Service for version testing")

		payload := map[string]interface{}{
			"version":          "1.0.0",
			"description":      "Initial version of the service",
			"releaseTimestamp": time.Now().Format(time.RFC3339),
		}

		resp, err := helpers.MakeAuthenticatedRequest("POST", fmt.Sprintf("/v1/orgs/%s/services/%s/versions", org.ID, service.ID), payload, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusOK)

		var serviceVersion models.ServiceVersion
		helpers.AssertJSONResponse(resp, &serviceVersion)

		// Validate response structure with testify
		helpers.AssertServiceVersionFields(serviceVersion, service.ID, "1.0.0", "Initial version of the service")
	})

	t.Run("ValidationErrors", func(t *testing.T) {
		// Setup test user, organization and service
		_, token := helpers.CreateTestUser("test2@example.com", "Test User 2", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")
		service := helpers.CreateTestService(token, org.ID, "Test Service", "Service for version testing")

		testCases := []struct {
			name         string
			payload      map[string]interface{}
			expectedCode int
		}{
			{
				name:         "Missing version",
				payload:      map[string]interface{}{"description": "Valid description with enough length"},
				expectedCode: http.StatusBadRequest,
			},
			{
				name:         "Missing description",
				payload:      map[string]interface{}{"version": "1.0.0"},
				expectedCode: http.StatusBadRequest,
			},
			{
				name:         "Invalid semantic version",
				payload:      map[string]interface{}{"version": "1.0", "description": "Valid description with enough length"},
				expectedCode: http.StatusBadRequest,
			},
			{
				name:         "Invalid semantic version format",
				payload:      map[string]interface{}{"version": "v1.0.0", "description": "Valid description with enough length"},
				expectedCode: http.StatusBadRequest,
			},
			{
				name:         "Description too short",
				payload:      map[string]interface{}{"version": "1.0.0", "description": "Short"},
				expectedCode: http.StatusBadRequest,
			},
			{
				name:         "Description too long",
				payload:      map[string]interface{}{"version": "1.0.0", "description": string(make([]byte, 1001))},
				expectedCode: http.StatusBadRequest,
			},
			{
				name:         "Empty request body",
				payload:      map[string]interface{}{},
				expectedCode: http.StatusBadRequest,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := helpers.MakeAuthenticatedRequest("POST", fmt.Sprintf("/v1/orgs/%s/services/%s/versions", org.ID, service.ID), tc.payload, token)
				if err != nil {
					t.Fatalf("Failed to make request: %v", err)
				}

				helpers.AssertStatusCode(resp, tc.expectedCode)
				helpers.AssertErrorResponseNotEmpty(resp)
			})
		}
	})

	t.Run("Unauthorized", func(t *testing.T) {
		// Setup test user, organization and service
		_, token := helpers.CreateTestUser("test3@example.com", "Test User 3", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")
		service := helpers.CreateTestService(token, org.ID, "Test Service", "Service for version testing")

		payload := map[string]interface{}{
			"version":          "1.0.0",
			"description":      "Initial version of the service",
			"releaseTimestamp": time.Now().Format(time.RFC3339),
		}

		// Test without token
		resp, err := helpers.MakeRequest("POST", fmt.Sprintf("/v1/orgs/%s/services/%s/versions", org.ID, service.ID), payload)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusUnauthorized)
	})

	t.Run("Forbidden", func(t *testing.T) {
		// Setup first user and organization
		_, token1 := helpers.CreateTestUser("test4@example.com", "Test User 4", TestPassword)
		org1 := helpers.CreateTestOrganization(token1, "Test Organization 1", "Test org description")
		service1 := helpers.CreateTestService(token1, org1.ID, "Test Service", "Service for version testing")

		// Setup second user (different user)
		_, token2 := helpers.CreateTestUser("test5@example.com", "Test User 5", TestPassword)

		payload := map[string]interface{}{
			"version":          "1.0.0",
			"description":      "Initial version of the service",
			"releaseTimestamp": time.Now().Format(time.RFC3339),
		}

		// Try to create version in org1 service using token2 (user5 is not member of org1)
		resp, err := helpers.MakeAuthenticatedRequest("POST", fmt.Sprintf("/v1/orgs/%s/services/%s/versions", org1.ID, service1.ID), payload, token2)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusForbidden)
	})

	t.Run("DuplicateVersion", func(t *testing.T) {
		// Setup test user, organization and service
		_, token := helpers.CreateTestUser("test6@example.com", "Test User 6", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")
		service := helpers.CreateTestService(token, org.ID, "Test Service", "Service for version testing")

		// Create first version
		helpers.CreateTestServiceVersion(token, org.ID, service.ID, "1.0.0", "First version")

		payload := map[string]interface{}{
			"version":          "1.0.0", // Same version
			"description":      "Duplicate version attempt",
			"releaseTimestamp": time.Now().Format(time.RFC3339),
		}

		resp, err := helpers.MakeAuthenticatedRequest("POST", fmt.Sprintf("/v1/orgs/%s/services/%s/versions", org.ID, service.ID), payload, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusInternalServerError) // Database constraint violation
	})
}

// TestGetServiceVersions tests GET /v1/orgs/{orgId}/services/{serviceId}/versions endpoint
func TestGetServiceVersions(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	t.Run("EmptyList", func(t *testing.T) {
		// Setup test user, organization and service
		_, token := helpers.CreateTestUser("test@example.com", "Test User", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")
		service := helpers.CreateTestService(token, org.ID, "Test Service", "Service for version testing")

		resp, err := helpers.MakeAuthenticatedRequest("GET", fmt.Sprintf("/v1/orgs/%s/services/%s/versions", org.ID, service.ID), nil, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusOK)

		var result models.PaginatedResult[models.ServiceVersion]
		helpers.AssertJSONResponse(resp, &result)

		assert.Len(t, result.Data, 0, "Expected empty versions data")
		assert.Equal(t, 0, result.Meta.TotalCount, "Expected total count to be 0")
	})

	t.Run("WithData", func(t *testing.T) {
		// Setup test user, organization and service
		_, token := helpers.CreateTestUser("test2@example.com", "Test User 2", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")
		service := helpers.CreateTestService(token, org.ID, "Test Service", "Service for version testing")

		// Create test version
		helpers.CreateTestServiceVersion(token, org.ID, service.ID, "1.0.0", "Initial version")

		resp, err := helpers.MakeAuthenticatedRequest("GET", fmt.Sprintf("/v1/orgs/%s/services/%s/versions", org.ID, service.ID), nil, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusOK)

		var result models.PaginatedResult[models.ServiceVersion]
		helpers.AssertJSONResponse(resp, &result)

		assert.Len(t, result.Data, 1, "Expected 1 version in data")
		assert.Equal(t, 1, result.Meta.TotalCount, "Expected total count to be 1")
		assert.Equal(t, 0, result.Meta.CurrentPage, "Expected current page to be 0")
	})

	t.Run("WithQueryParameters", func(t *testing.T) {
		// Setup test user, organization and service
		_, token := helpers.CreateTestUser("test3@example.com", "Test User 3", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")
		service := helpers.CreateTestService(token, org.ID, "Test Service", "Service for version testing")

		// Create test versions
		helpers.CreateTestServiceVersion(token, org.ID, service.ID, "1.0.0", "Initial version")
		helpers.CreateTestServiceVersion(token, org.ID, service.ID, "2.0.0", "Major update")

		testCases := []struct {
			name       string
			query      string
			shouldFind bool
			expectedCount int
		}{
			{
				name:       "Search by version - exact match",
				query:      "?q=1.0.0",
				shouldFind: true,
				expectedCount: 1,
			},
			{
				name:       "Search by version - partial match",
				query:      "?q=1.",
				shouldFind: true,
				expectedCount: 1,
			},
			{
				name:       "Search by version - no match",
				query:      "?q=3.0.0",
				shouldFind: false,
				expectedCount: 0,
			},
			{
				name:       "Sort by version ascending",
				query:      "?sort_by=version&sort=asc",
				shouldFind: true,
				expectedCount: 2,
			},
			{
				name:       "Sort by created_at descending",
				query:      "?sort_by=created_at&sort=desc",
				shouldFind: true,
				expectedCount: 2,
			},
			{
				name:       "Pagination - page 0",
				query:      "?page=0&per_page=1",
				shouldFind: true,
				expectedCount: 1,
			},
			{
				name:       "Pagination - page 1",
				query:      "?page=1&per_page=1",
				shouldFind: true,
				expectedCount: 1,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := helpers.MakeAuthenticatedRequest("GET", fmt.Sprintf("/v1/orgs/%s/services/%s/versions%s", org.ID, service.ID, tc.query), nil, token)
				if err != nil {
					t.Fatalf("Failed to make request: %v", err)
				}

				helpers.AssertStatusCode(resp, http.StatusOK)

				var result models.PaginatedResult[models.ServiceVersion]
				helpers.AssertJSONResponse(resp, &result)

				if tc.shouldFind {
					assert.Len(t, result.Data, tc.expectedCount, "Expected count mismatch")
				} else {
					assert.Empty(t, result.Data, "Expected empty result")
				}
			})
		}
	})
}

// TestGetServiceVersion tests GET /v1/orgs/{orgId}/services/{serviceId}/versions/{versionId} endpoint
func TestGetServiceVersion(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	t.Run("Success", func(t *testing.T) {
		// Setup test user, organization and service
		_, token := helpers.CreateTestUser("test@example.com", "Test User", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")
		service := helpers.CreateTestService(token, org.ID, "Test Service", "Service for version testing")

		// Create test version
		version := helpers.CreateTestServiceVersion(token, org.ID, service.ID, "1.0.0", "Initial version")

		resp, err := helpers.MakeAuthenticatedRequest("GET", fmt.Sprintf("/v1/orgs/%s/services/%s/versions/%s", org.ID, service.ID, version.ID), nil, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusOK)

		var retrievedVersion models.ServiceVersion
		helpers.AssertJSONResponse(resp, &retrievedVersion)

		assert.Equal(t, version.ID, retrievedVersion.ID, "Version ID should match")
		assert.Equal(t, "1.0.0", retrievedVersion.Version, "Version should match")
		assert.Equal(t, service.ID, retrievedVersion.ServiceID, "Service ID should match")
	})

	t.Run("NotFound", func(t *testing.T) {
		// Setup test user, organization and service
		_, token := helpers.CreateTestUser("test2@example.com", "Test User 2", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")
		service := helpers.CreateTestService(token, org.ID, "Test Service", "Service for version testing")

		nonExistentID := "non-existent-id"
		resp, err := helpers.MakeAuthenticatedRequest("GET", fmt.Sprintf("/v1/orgs/%s/services/%s/versions/%s", org.ID, service.ID, nonExistentID), nil, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusNotFound)
		helpers.AssertErrorResponseNotEmpty(resp)
	})
}

// TestUpdateServiceVersion tests PATCH /v1/orgs/{orgId}/services/{serviceId}/versions/{versionId} endpoint
func TestUpdateServiceVersion(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	t.Run("Success", func(t *testing.T) {
		// Setup test user, organization and service
		_, token := helpers.CreateTestUser("test@example.com", "Test User", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")
		service := helpers.CreateTestService(token, org.ID, "Test Service", "Service for version testing")

		// Create test version
		version := helpers.CreateTestServiceVersion(token, org.ID, service.ID, "1.0.0", "Initial version")

		payload := map[string]interface{}{
			"description": "Updated version description",
		}

		resp, err := helpers.MakeAuthenticatedRequest("PATCH", fmt.Sprintf("/v1/orgs/%s/services/%s/versions/%s", org.ID, service.ID, version.ID), payload, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusOK)

		var updatedVersion models.ServiceVersion
		helpers.AssertJSONResponse(resp, &updatedVersion)

		assert.Equal(t, "Updated version description", updatedVersion.Description, "Description should be updated")
		assert.Equal(t, "1.0.0", updatedVersion.Version, "Version should remain unchanged")
		assert.Equal(t, service.ID, updatedVersion.ServiceID, "Service ID should remain unchanged")
	})

	t.Run("NotFound", func(t *testing.T) {
		// Setup test user, organization and service
		_, token := helpers.CreateTestUser("test2@example.com", "Test User 2", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")
		service := helpers.CreateTestService(token, org.ID, "Test Service", "Service for version testing")

		payload := map[string]interface{}{
			"description": "Updated description",
		}

		nonExistentID := "non-existent-id"
		resp, err := helpers.MakeAuthenticatedRequest("PATCH", fmt.Sprintf("/v1/orgs/%s/services/%s/versions/%s", org.ID, service.ID, nonExistentID), payload, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusNotFound)
	})
}

// TestDeleteServiceVersion tests DELETE /v1/orgs/{orgId}/services/{serviceId}/versions/{versionId} endpoint
func TestDeleteServiceVersion(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	t.Run("Success", func(t *testing.T) {
		// Setup test user, organization and service
		_, token := helpers.CreateTestUser("test@example.com", "Test User", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")
		service := helpers.CreateTestService(token, org.ID, "Test Service", "Service for version testing")

		// Create test version
		version := helpers.CreateTestServiceVersion(token, org.ID, service.ID, "1.0.0", "Initial version")

		resp, err := helpers.MakeAuthenticatedRequest("DELETE", fmt.Sprintf("/v1/orgs/%s/services/%s/versions/%s", org.ID, service.ID, version.ID), nil, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusNoContent)

		// Verify version is deleted by trying to get it
		getResp, err := helpers.MakeAuthenticatedRequest("GET", fmt.Sprintf("/v1/orgs/%s/services/%s/versions/%s", org.ID, service.ID, version.ID), nil, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(getResp, http.StatusNotFound)
	})

	t.Run("NotFound", func(t *testing.T) {
		// Setup test user, organization and service
		_, token := helpers.CreateTestUser("test2@example.com", "Test User 2", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")
		service := helpers.CreateTestService(token, org.ID, "Test Service", "Service for version testing")

		nonExistentID := "non-existent-id"
		resp, err := helpers.MakeAuthenticatedRequest("DELETE", fmt.Sprintf("/v1/orgs/%s/services/%s/versions/%s", org.ID, service.ID, nonExistentID), nil, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusNotFound)
	})
}