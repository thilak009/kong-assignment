package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thilak009/kong-assignment/models"
)

// TestCreateService tests POST /v1/services endpoint
func TestCreateService(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	t.Run("Success", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":        "Test Service",
			"description": "This is a test service for integration testing",
		}

		resp, err := helpers.MakeRequest("POST", "/v1/services", payload)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusOK)

		var service models.Service
		helpers.AssertJSONResponse(resp, &service)

		// Validate response structure with helper
		helpers.AssertServiceFields(service, "Test Service", "This is a test service for integration testing")
	})

	t.Run("ValidationErrors", func(t *testing.T) {
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
				name:         "Missing description",
				payload:      map[string]interface{}{"name": "Valid Name"},
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
				resp, err := helpers.MakeRequest("POST", "/v1/services", tc.payload)
				if err != nil {
					t.Fatalf("Failed to make request: %v", err)
				}

				helpers.AssertStatusCode(resp, tc.expectedCode)
				helpers.AssertErrorResponseNotEmpty(resp)
			})
		}
	})
}

// TestGetAllServices tests GET /v1/services endpoint
func TestGetAllServices(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	t.Run("EmptyList", func(t *testing.T) {
		resp, err := helpers.MakeRequest("GET", "/v1/services", nil)
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
		// Create test service
		helpers.CreateTestService("Integration Test Service", "Service for integration testing")

		resp, err := helpers.MakeRequest("GET", "/v1/services", nil)
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
		// Create test service
		helpers.CreateTestService("Integration Test Service", "Service for integration testing")

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
				resp, err := helpers.MakeRequest("GET", "/v1/services"+tc.query, nil)
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
		// Create test service
		helpers.CreateTestService("Integration Test Service", "Service for integration testing")

		resp, err := helpers.MakeRequest("GET", "/v1/services?include=versionCount", nil)
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

// TestGetService tests GET /v1/services/{id} endpoint
func TestGetService(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	t.Run("Success", func(t *testing.T) {
		// Create test service
		service := helpers.CreateTestService("Integration Test Service", "Service for integration testing")

		resp, err := helpers.MakeRequest("GET", fmt.Sprintf("/v1/services/%s", service.ID), nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusOK)

		var retrievedService models.Service
		helpers.AssertJSONResponse(resp, &retrievedService)

		assert.Equal(t, service.ID, retrievedService.ID, "Service ID should match")
		assert.Equal(t, "Integration Test Service", retrievedService.Name, "Service name should match")
	})

	t.Run("NotFound", func(t *testing.T) {
		nonExistentID := "non-existent-id"
		resp, err := helpers.MakeRequest("GET", fmt.Sprintf("/v1/services/%s", nonExistentID), nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusNotFound)
		helpers.AssertErrorResponseNotEmpty(resp)
	})

	t.Run("WithIncludeVersionCount", func(t *testing.T) {
		// Create test service
		service := helpers.CreateTestService("Integration Test Service", "Service for integration testing")

		resp, err := helpers.MakeRequest("GET", fmt.Sprintf("/v1/services/%s?include=versionCount", service.ID), nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusOK)

		var retrievedService models.Service
		helpers.AssertJSONResponse(resp, &retrievedService)

		helpers.AssertVersionCountIncluded(retrievedService, 0)
	})
}

// TestUpdateService tests PUT /v1/services/{id} endpoint
func TestUpdateService(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	t.Run("Success", func(t *testing.T) {
		// Create test service
		service := helpers.CreateTestService("Integration Test Service", "Service for integration testing")

		payload := map[string]interface{}{
			"name":        "Updated Test Service",
			"description": "This is an updated test service description",
		}

		resp, err := helpers.MakeRequest("PUT", fmt.Sprintf("/v1/services/%s", service.ID), payload)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusOK)

		var updatedService models.Service
		helpers.AssertJSONResponse(resp, &updatedService)

		assert.Equal(t, "Updated Test Service", updatedService.Name, "Service name should be updated")
		assert.Equal(t, "This is an updated test service description", updatedService.Description, "Service description should be updated")
	})

	t.Run("NotFound", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":        "Updated Service",
			"description": "Updated description with enough length",
		}

		nonExistentID := "non-existent-id"
		resp, err := helpers.MakeRequest("PUT", fmt.Sprintf("/v1/services/%s", nonExistentID), payload)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusNotFound)
	})

	t.Run("ValidationErrors", func(t *testing.T) {
		// Create test service
		service := helpers.CreateTestService("Integration Test Service", "Service for integration testing")

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
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := helpers.MakeRequest("PUT", fmt.Sprintf("/v1/services/%s", service.ID), tc.payload)
				if err != nil {
					t.Fatalf("Failed to make request: %v", err)
				}

				helpers.AssertStatusCode(resp, http.StatusBadRequest)
				helpers.AssertErrorResponseNotEmpty(resp)
			})
		}
	})
}

// TestDeleteService tests DELETE /v1/services/{id} endpoint
func TestDeleteService(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	t.Run("Success", func(t *testing.T) {
		// Create test service
		service := helpers.CreateTestService("Integration Test Service", "Service for integration testing")

		resp, err := helpers.MakeRequest("DELETE", fmt.Sprintf("/v1/services/%s", service.ID), nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusNoContent)

		// Verify service is deleted by trying to get it
		getResp, err := helpers.MakeRequest("GET", fmt.Sprintf("/v1/services/%s", service.ID), nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(getResp, http.StatusNotFound)
	})

	t.Run("NotFound", func(t *testing.T) {
		nonExistentID := "non-existent-id"
		resp, err := helpers.MakeRequest("DELETE", fmt.Sprintf("/v1/services/%s", nonExistentID), nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusNotFound)
		helpers.AssertErrorResponseNotEmpty(resp)
	})
}
