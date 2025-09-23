package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thilak009/kong-assignment/models"
)

// TestCreateService tests POST /v1/orgs/{orgId}/services endpoint
func TestCreateService(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	t.Run("Success", func(t *testing.T) {
		// Setup test user and organization
		_, token := helpers.CreateTestUser("test@example.com", "Test User", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")

		payload := map[string]interface{}{
			"name":        "Test Service",
			"description": "This is a test service for integration testing",
		}

		resp, err := helpers.MakeAuthenticatedRequest("POST", fmt.Sprintf("/v1/orgs/%s/services", org.ID), payload, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusOK)

		var service models.Service
		helpers.AssertJSONResponse(resp, &service)

		// Validate response structure with helper
		helpers.AssertServiceFields(service, "Test Service", "This is a test service for integration testing")
		assert.Equal(t, org.ID, service.OrganizationID, "Service should belong to the organization")
	})

	t.Run("ValidationErrors", func(t *testing.T) {
		// Setup test user and organization
		_, token := helpers.CreateTestUser("test2@example.com", "Test User 2", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")

		testCases := []struct {
			name         string
			payload      map[string]interface{}
			expectedCode int
		}{
			{
				name:         "Missing name",
				payload:      map[string]interface{}{"description": "Valid description with enough length"},
				expectedCode: http.StatusBadRequest,
			},
			{
				name:         "Name too short",
				payload:      map[string]interface{}{"name": "AB", "description": "Valid description with enough length"},
				expectedCode: http.StatusBadRequest,
			},
			{
				name:         "Name too long",
				payload:      map[string]interface{}{"name": string(make([]byte, 101)), "description": "Valid description with enough length"},
				expectedCode: http.StatusBadRequest,
			},
			{
				name:         "Description too short",
				payload:      map[string]interface{}{"name": "Valid Name", "description": "Short"},
				expectedCode: http.StatusBadRequest,
			},
			{
				name:         "Description too long",
				payload:      map[string]interface{}{"name": "Valid Name", "description": string(make([]byte, 1001))},
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
				resp, err := helpers.MakeAuthenticatedRequest("POST", fmt.Sprintf("/v1/orgs/%s/services", org.ID), tc.payload, token)
				if err != nil {
					t.Fatalf("Failed to make request: %v", err)
				}

				helpers.AssertStatusCode(resp, tc.expectedCode)
				helpers.AssertErrorResponseNotEmpty(resp)
			})
		}
	})

	t.Run("Unauthorized", func(t *testing.T) {
		// Setup test user and organization
		_, token := helpers.CreateTestUser("test3@example.com", "Test User 3", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")

		payload := map[string]interface{}{
			"name":        "Test Service",
			"description": "This is a test service for integration testing",
		}

		// Test without token
		resp, err := helpers.MakeRequest("POST", fmt.Sprintf("/v1/orgs/%s/services", org.ID), payload)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusUnauthorized)
	})

	t.Run("Forbidden", func(t *testing.T) {
		// Setup first user and organization
		_, token1 := helpers.CreateTestUser("test4@example.com", "Test User 4", TestPassword)
		org1 := helpers.CreateTestOrganization(token1, "Test Organization 1", "Test org description")

		// Setup second user (different user)
		_, token2 := helpers.CreateTestUser("test5@example.com", "Test User 5", TestPassword)

		payload := map[string]interface{}{
			"name":        "Test Service",
			"description": "This is a test service for integration testing",
		}

		// Try to create service in org1 using token2 (user5 is not member of org1)
		resp, err := helpers.MakeAuthenticatedRequest("POST", fmt.Sprintf("/v1/orgs/%s/services", org1.ID), payload, token2)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusForbidden)
	})
}

// TestGetAllServices tests GET /v1/orgs/{orgId}/services endpoint
func TestGetAllServices(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	t.Run("EmptyList", func(t *testing.T) {
		// Setup test user and organization
		_, token := helpers.CreateTestUser("test@example.com", "Test User", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")

		resp, err := helpers.MakeAuthenticatedRequest("GET", fmt.Sprintf("/v1/orgs/%s/services", org.ID), nil, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusOK)

		var result models.PaginatedResult[models.Service]
		helpers.AssertJSONResponse(resp, &result)

		assert.Len(t, result.Data, 0, "Expected empty services data")
		assert.Equal(t, 0, result.Meta.TotalCount, "Expected total count to be 0")
	})

	t.Run("WithData", func(t *testing.T) {
		// Setup test user and organization
		_, token := helpers.CreateTestUser("test2@example.com", "Test User 2", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")

		// Create test service
		helpers.CreateTestService(token, org.ID, "Integration Test Service", "Service for integration testing")

		resp, err := helpers.MakeAuthenticatedRequest("GET", fmt.Sprintf("/v1/orgs/%s/services", org.ID), nil, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusOK)

		var result models.PaginatedResult[models.Service]
		helpers.AssertJSONResponse(resp, &result)

		assert.Len(t, result.Data, 1, "Expected 1 service in data")
		assert.Equal(t, 1, result.Meta.TotalCount, "Expected total count to be 1")
		assert.Equal(t, 0, result.Meta.CurrentPage, "Expected current page to be 0")
	})

	t.Run("WithQueryParameters", func(t *testing.T) {
		// Setup test user and organization
		_, token := helpers.CreateTestUser("test3@example.com", "Test User 3", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")

		// Create test service
		helpers.CreateTestService(token, org.ID, "Integration Test Service", "Service for integration testing")

		testCases := []struct {
			name       string
			query      string
			shouldFind bool
		}{
			{
				name:       "Search by name - exact match",
				query:      "?q=Integration Test Service",
				shouldFind: true,
			},
			{
				name:       "Search by name - partial match",
				query:      "?q=Integration",
				shouldFind: true,
			},
			{
				name:       "Search by name - no match",
				query:      "?q=NonExistent",
				shouldFind: false,
			},
			{
				name:       "Sort by name ascending",
				query:      "?sort_by=name&sort=asc",
				shouldFind: true,
			},
			{
				name:       "Sort by created_at descending",
				query:      "?sort_by=created_at&sort=desc",
				shouldFind: true,
			},
			{
				name:       "Pagination - page 0",
				query:      "?page=0&per_page=5",
				shouldFind: true,
			},
			{
				name:       "Pagination - page 1 (should be empty)",
				query:      "?page=1&per_page=5",
				shouldFind: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := helpers.MakeAuthenticatedRequest("GET", fmt.Sprintf("/v1/orgs/%s/services%s", org.ID, tc.query), nil, token)
				if err != nil {
					t.Fatalf("Failed to make request: %v", err)
				}

				helpers.AssertStatusCode(resp, http.StatusOK)

				var result models.PaginatedResult[models.Service]
				helpers.AssertJSONResponse(resp, &result)

				if tc.shouldFind {
					assert.NotEmpty(t, result.Data, "Expected to find services but got empty result")
				} else {
					assert.Empty(t, result.Data, "Expected empty result")
				}
			})
		}
	})

	t.Run("WithIncludeVersionCount", func(t *testing.T) {
		// Setup test user and organization
		_, token := helpers.CreateTestUser("test4@example.com", "Test User 4", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")

		// Create test service
		helpers.CreateTestService(token, org.ID, "Integration Test Service", "Service for integration testing")

		resp, err := helpers.MakeAuthenticatedRequest("GET", fmt.Sprintf("/v1/orgs/%s/services?include=versionCount", org.ID), nil, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusOK)

		var result models.PaginatedResult[models.Service]
		helpers.AssertJSONResponse(resp, &result)

		if len(result.Data) > 0 {
			service := *result.Data[0]
			helpers.AssertVersionCountIncluded(service, 0)
		}
	})
}

// TestGetService tests GET /v1/orgs/{orgId}/services/{serviceId} endpoint
func TestGetService(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	t.Run("Success", func(t *testing.T) {
		// Setup test user and organization
		_, token := helpers.CreateTestUser("test@example.com", "Test User", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")

		// Create test service
		service := helpers.CreateTestService(token, org.ID, "Integration Test Service", "Service for integration testing")

		resp, err := helpers.MakeAuthenticatedRequest("GET", fmt.Sprintf("/v1/orgs/%s/services/%s", org.ID, service.ID), nil, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusOK)

		var retrievedService models.Service
		helpers.AssertJSONResponse(resp, &retrievedService)

		assert.Equal(t, service.ID, retrievedService.ID, "Service ID should match")
		assert.Equal(t, "Integration Test Service", retrievedService.Name, "Service name should match")
		assert.Equal(t, org.ID, retrievedService.OrganizationID, "Service should belong to the organization")
	})

	t.Run("NotFound", func(t *testing.T) {
		// Setup test user and organization
		_, token := helpers.CreateTestUser("test2@example.com", "Test User 2", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")

		nonExistentID := "non-existent-id"
		resp, err := helpers.MakeAuthenticatedRequest("GET", fmt.Sprintf("/v1/orgs/%s/services/%s", org.ID, nonExistentID), nil, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusNotFound)
		helpers.AssertErrorResponseNotEmpty(resp)
	})

	t.Run("WithIncludeVersionCount", func(t *testing.T) {
		// Setup test user and organization
		_, token := helpers.CreateTestUser("test3@example.com", "Test User 3", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")

		// Create test service
		service := helpers.CreateTestService(token, org.ID, "Integration Test Service", "Service for integration testing")

		resp, err := helpers.MakeAuthenticatedRequest("GET", fmt.Sprintf("/v1/orgs/%s/services/%s?include=versionCount", org.ID, service.ID), nil, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusOK)

		var retrievedService models.Service
		helpers.AssertJSONResponse(resp, &retrievedService)

		helpers.AssertVersionCountIncluded(retrievedService, 0)
	})
}

// TestUpdateService tests PATCH /v1/orgs/{orgId}/services/{serviceId} endpoint
func TestUpdateService(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	t.Run("Success", func(t *testing.T) {
		// Setup test user and organization
		_, token := helpers.CreateTestUser("test@example.com", "Test User", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")

		// Create test service
		service := helpers.CreateTestService(token, org.ID, "Integration Test Service", "Service for integration testing")

		payload := map[string]interface{}{
			"name":        "Updated Test Service",
			"description": "This is an updated test service description",
		}

		resp, err := helpers.MakeAuthenticatedRequest("PATCH", fmt.Sprintf("/v1/orgs/%s/services/%s", org.ID, service.ID), payload, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusOK)

		var updatedService models.Service
		helpers.AssertJSONResponse(resp, &updatedService)

		assert.Equal(t, "Updated Test Service", updatedService.Name, "Service name should be updated")
		assert.Equal(t, "This is an updated test service description", updatedService.Description, "Service description should be updated")
		assert.Equal(t, org.ID, updatedService.OrganizationID, "Service should still belong to the organization")
	})

	t.Run("NotFound", func(t *testing.T) {
		// Setup test user and organization
		_, token := helpers.CreateTestUser("test2@example.com", "Test User 2", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")

		payload := map[string]interface{}{
			"name":        "Updated Service",
			"description": "Updated description with enough length",
		}

		nonExistentID := "non-existent-id"
		resp, err := helpers.MakeAuthenticatedRequest("PATCH", fmt.Sprintf("/v1/orgs/%s/services/%s", org.ID, nonExistentID), payload, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusNotFound)
	})

	t.Run("ValidationErrors", func(t *testing.T) {
		// Setup test user and organization
		_, token := helpers.CreateTestUser("test3@example.com", "Test User 3", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")

		// Create test service
		service := helpers.CreateTestService(token, org.ID, "Integration Test Service", "Service for integration testing")

		testCases := []struct {
			name    string
			payload map[string]interface{}
		}{
			{
				name:    "Invalid name",
				payload: map[string]interface{}{"name": "AB", "description": "Valid description with enough length"},
			},
			{
				name:    "Invalid description",
				payload: map[string]interface{}{"name": "Valid Name", "description": "Short"},
			},
			{
				name:    "Empty body - no fields provided",
				payload: map[string]interface{}{},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := helpers.MakeAuthenticatedRequest("PATCH", fmt.Sprintf("/v1/orgs/%s/services/%s", org.ID, service.ID), tc.payload, token)
				if err != nil {
					t.Fatalf("Failed to make request: %v", err)
				}

				helpers.AssertStatusCode(resp, http.StatusBadRequest)
				helpers.AssertErrorResponseNotEmpty(resp)
			})
		}
	})
}

// TestDeleteService tests DELETE /v1/orgs/{orgId}/services/{serviceId} endpoint
func TestDeleteService(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	t.Run("Success", func(t *testing.T) {
		// Setup test user and organization
		_, token := helpers.CreateTestUser("test@example.com", "Test User", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")

		// Create test service
		service := helpers.CreateTestService(token, org.ID, "Integration Test Service", "Service for integration testing")

		resp, err := helpers.MakeAuthenticatedRequest("DELETE", fmt.Sprintf("/v1/orgs/%s/services/%s", org.ID, service.ID), nil, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusNoContent)

		// Verify service is deleted by trying to get it
		getResp, err := helpers.MakeAuthenticatedRequest("GET", fmt.Sprintf("/v1/orgs/%s/services/%s", org.ID, service.ID), nil, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(getResp, http.StatusNotFound)
	})

	t.Run("NotFound", func(t *testing.T) {
		// Setup test user and organization
		_, token := helpers.CreateTestUser("test2@example.com", "Test User 2", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test org description")

		nonExistentID := "non-existent-id"
		resp, err := helpers.MakeAuthenticatedRequest("DELETE", fmt.Sprintf("/v1/orgs/%s/services/%s", org.ID, nonExistentID), nil, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusNotFound)
	})
}