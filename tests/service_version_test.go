package tests

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thilak009/kong-assignment/models"
)

// TestCreateServiceVersion tests POST /v1/services/{serviceId}/versions endpoint
func TestCreateServiceVersion(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	// Create a test service first
	service := helpers.CreateTestService("Test Service", "Service for version testing")

	t.Run("Success", func(t *testing.T) {
		payload := map[string]interface{}{
			"version":          "1.0.0",
			"description":      "Initial version of the service",
			"releaseTimestamp": time.Now().Format(time.RFC3339),
		}

		resp, err := helpers.MakeRequest("POST", fmt.Sprintf("/v1/services/%s/versions", service.ID), payload)
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
				resp, err := helpers.MakeRequest("POST", fmt.Sprintf("/v1/services/%s/versions", service.ID), tc.payload)
				if err != nil {
					t.Fatalf("Failed to make request: %v", err)
				}

				helpers.AssertStatusCode(resp, tc.expectedCode)

				var errorResp models.ErrorResponse
				helpers.AssertJSONResponse(resp, &errorResp)

				assert.NotEmpty(t, errorResp.Message, "Error message should not be empty")
			})
		}
	})

	t.Run("ServiceNotFound", func(t *testing.T) {
		payload := map[string]interface{}{
			"version":     "1.0.0",
			"description": "Valid description with enough length",
		}

		nonExistentServiceID := "non-existent-service-id"
		resp, err := helpers.MakeRequest("POST", fmt.Sprintf("/v1/services/%s/versions", nonExistentServiceID), payload)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusNotFound)
	})
}

// TestGetAllServiceVersions tests GET /v1/services/{serviceId}/versions endpoint
func TestGetAllServiceVersions(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	// Create a test service first
	service := helpers.CreateTestService("Test Service", "Service for version testing")

	t.Run("EmptyList", func(t *testing.T) {
		resp, err := helpers.MakeRequest("GET", fmt.Sprintf("/v1/services/%s/versions", service.ID), nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusOK)

		var result models.PaginatedResult[models.ServiceVersion]
		helpers.AssertJSONResponse(resp, &result)

		assert.Empty(t, result.Data, "Expected empty service versions data")
		assert.Equal(t, 0, result.Meta.TotalCount, "Expected total count to be 0")
	})

	t.Run("WithData", func(t *testing.T) {
		// Create test service version
		helpers.CreateTestServiceVersion(service.ID, "1.0.0", "Initial version for testing")

		resp, err := helpers.MakeRequest("GET", fmt.Sprintf("/v1/services/%s/versions", service.ID), nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusOK)

		var result models.PaginatedResult[models.ServiceVersion]
		helpers.AssertJSONResponse(resp, &result)

		assert.Len(t, result.Data, 1, "Expected 1 service version")
		assert.Equal(t, 1, result.Meta.TotalCount, "Expected total count 1")
		assert.Equal(t, 0, result.Meta.CurrentPage, "Expected current page 0")
	})

	t.Run("WithQueryParameters", func(t *testing.T) {
		// Clean up any existing data first to ensure test isolation
		helpers.CleanupDatabase()

		// Create test service for this subtest
		queryTestService := helpers.CreateTestService("Query Test Service", "Service for query parameter testing")

		// Create multiple test service versions
		helpers.CreateTestServiceVersion(queryTestService.ID, "1.0.0", "Initial version for query testing")
		helpers.CreateTestServiceVersion(queryTestService.ID, "1.0.1", "Patch version for query testing")
		helpers.CreateTestServiceVersion(queryTestService.ID, "1.1.0", "Minor version for query testing")
		helpers.CreateTestServiceVersion(queryTestService.ID, "2.0.0", "Major version for query testing")

		testCases := []struct {
			name           string
			query          string
			expectedCount  int
			shouldContain  string
		}{
			{
				name:           "Search by version prefix - 1",
				query:          "?q=1",
				expectedCount:  3, // 1.0.0, 1.0.1, 1.1.0
				shouldContain:  "1.",
			},
			{
				name:           "Search by version prefix - 1.0",
				query:          "?q=1.0",
				expectedCount:  2, // 1.0.0, 1.0.1
				shouldContain:  "1.0",
			},
			{
				name:           "Search by version prefix - 2",
				query:          "?q=2",
				expectedCount:  1, // 2.0.0
				shouldContain:  "2.0.0",
			},
			{
				name:           "Search by version - no match",
				query:          "?q=3",
				expectedCount:  0,
				shouldContain:  "",
			},
			{
				name:           "Sort by version ascending",
				query:          "?sort_by=version&sort=asc",
				expectedCount:  4, // Total: 1.0.0, 1.0.1, 1.1.0, 2.0.0
				shouldContain:  "",
			},
			{
				name:           "Sort by created_at descending",
				query:          "?sort_by=created_at&sort=desc",
				expectedCount:  4,
				shouldContain:  "",
			},
			{
				name:           "Pagination - page 0",
				query:          "?page=0&per_page=2",
				expectedCount:  2,
				shouldContain:  "",
			},
			{
				name:           "Pagination - page 1",
				query:          "?page=1&per_page=2",
				expectedCount:  2,
				shouldContain:  "",
			},
			{
				name:           "Pagination - page 2",
				query:          "?page=2&per_page=2",
				expectedCount:  0, // No more versions on page 2 (4 total, 2 per page = 2 full pages)
				shouldContain:  "",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := helpers.MakeRequest("GET", fmt.Sprintf("/v1/services/%s/versions%s", queryTestService.ID, tc.query), nil)
				if err != nil {
					t.Fatalf("Failed to make request: %v", err)
				}

				helpers.AssertStatusCode(resp, http.StatusOK)

				var result models.PaginatedResult[models.ServiceVersion]
				helpers.AssertJSONResponse(resp, &result)

				assert.Len(t, result.Data, tc.expectedCount, "Expected %d service versions", tc.expectedCount)

				if tc.shouldContain != "" && len(result.Data) > 0 {
					found := false
					for _, version := range result.Data {
						if strings.Contains(version.Version, tc.shouldContain) {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected to find version containing '%s', but found versions: %v", tc.shouldContain, getVersionStrings(result.Data))
				}
			})
		}
	})

	t.Run("ServiceNotFound", func(t *testing.T) {
		nonExistentServiceID := "non-existent-service-id"
		resp, err := helpers.MakeRequest("GET", fmt.Sprintf("/v1/services/%s/versions", nonExistentServiceID), nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusNotFound)

		var errorResp models.ErrorResponse
		helpers.AssertJSONResponse(resp, &errorResp)

		assert.NotEmpty(t, errorResp.Message, "Error message should not be empty")
	})
}

// TestGetServiceVersion tests GET /v1/services/{serviceId}/versions/{id} endpoint
func TestGetServiceVersion(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	// Create test service and version
	service := helpers.CreateTestService("Test Service", "Service for version testing")
	serviceVersion := helpers.CreateTestServiceVersion(service.ID, "1.0.0", "Initial version for testing")

	t.Run("Success", func(t *testing.T) {
		resp, err := helpers.MakeRequest("GET", fmt.Sprintf("/v1/services/%s/versions/%s", service.ID, serviceVersion.ID), nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusOK)

		var retrievedVersion models.ServiceVersion
		helpers.AssertJSONResponse(resp, &retrievedVersion)

		assert.Equal(t, serviceVersion.ID, retrievedVersion.ID, "Version ID should match")
		assert.Equal(t, service.ID, retrievedVersion.ServiceID, "Service ID should match")
		assert.Equal(t, "1.0.0", retrievedVersion.Version, "Version should match")
	})

	t.Run("ServiceNotFound", func(t *testing.T) {
		nonExistentServiceID := "non-existent-service-id"
		resp, err := helpers.MakeRequest("GET", fmt.Sprintf("/v1/services/%s/versions/%s", nonExistentServiceID, serviceVersion.ID), nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusNotFound)

		var errorResp models.ErrorResponse
		helpers.AssertJSONResponse(resp, &errorResp)

		assert.NotEmpty(t, errorResp.Message, "Error message should not be empty")
	})

	t.Run("VersionNotFound", func(t *testing.T) {
		nonExistentVersionID := "non-existent-version-id"
		resp, err := helpers.MakeRequest("GET", fmt.Sprintf("/v1/services/%s/versions/%s", service.ID, nonExistentVersionID), nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusNotFound)

		var errorResp models.ErrorResponse
		helpers.AssertJSONResponse(resp, &errorResp)

		assert.NotEmpty(t, errorResp.Message, "Error message should not be empty")
	})
}

// TestUpdateServiceVersion tests PATCH /v1/services/{serviceId}/versions/{id} endpoint
func TestUpdateServiceVersion(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	// Create test service and version
	service := helpers.CreateTestService("Test Service", "Service for version testing")
	serviceVersion := helpers.CreateTestServiceVersion(service.ID, "1.0.0", "Initial version for testing")

	t.Run("Success", func(t *testing.T) {
		payload := map[string]interface{}{
			"description":      "Updated version description",
			"releaseTimestamp": time.Now().Add(time.Hour).Format(time.RFC3339),
		}

		resp, err := helpers.MakeRequest("PATCH", fmt.Sprintf("/v1/services/%s/versions/%s", service.ID, serviceVersion.ID), payload)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusOK)

		var updatedVersion models.ServiceVersion
		helpers.AssertJSONResponse(resp, &updatedVersion)

		assert.Equal(t, "Updated version description", updatedVersion.Description, "Description should be updated")
		// Version should remain unchanged
		assert.Equal(t, "1.0.0", updatedVersion.Version, "Version should remain unchanged")
	})

	t.Run("ValidationErrors", func(t *testing.T) {
		testCases := []struct {
			name    string
			payload map[string]interface{}
		}{
			{
				name:    "Description too short",
				payload: map[string]interface{}{"description": "Short"},
			},
			{
				name:    "Description too long",
				payload: map[string]interface{}{"description": string(make([]byte, 1001))},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := helpers.MakeRequest("PATCH", fmt.Sprintf("/v1/services/%s/versions/%s", service.ID, serviceVersion.ID), tc.payload)
				if err != nil {
					t.Fatalf("Failed to make request: %v", err)
				}

				helpers.AssertStatusCode(resp, http.StatusBadRequest)

				var errorResp models.ErrorResponse
				helpers.AssertJSONResponse(resp, &errorResp)

				assert.NotEmpty(t, errorResp.Message, "Error message should not be empty")
			})
		}
	})

	t.Run("ServiceNotFound", func(t *testing.T) {
		payload := map[string]interface{}{
			"description": "Updated description with enough length",
		}

		nonExistentServiceID := "non-existent-service-id"
		resp, err := helpers.MakeRequest("PATCH", fmt.Sprintf("/v1/services/%s/versions/%s", nonExistentServiceID, serviceVersion.ID), payload)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusNotFound)
	})

	t.Run("VersionNotFound", func(t *testing.T) {
		payload := map[string]interface{}{
			"description": "Updated description with enough length",
		}

		nonExistentVersionID := "non-existent-version-id"
		resp, err := helpers.MakeRequest("PATCH", fmt.Sprintf("/v1/services/%s/versions/%s", service.ID, nonExistentVersionID), payload)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusNotFound)
	})
}

// TestDeleteServiceVersion tests DELETE /v1/services/{serviceId}/versions/{id} endpoint
func TestDeleteServiceVersion(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	// Create test service and version
	service := helpers.CreateTestService("Test Service", "Service for version testing")
	serviceVersion := helpers.CreateTestServiceVersion(service.ID, "1.0.0", "Initial version for testing")

	t.Run("Success", func(t *testing.T) {
		resp, err := helpers.MakeRequest("DELETE", fmt.Sprintf("/v1/services/%s/versions/%s", service.ID, serviceVersion.ID), nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusNoContent)

		// Verify version is deleted by trying to get it
		getResp, err := helpers.MakeRequest("GET", fmt.Sprintf("/v1/services/%s/versions/%s", service.ID, serviceVersion.ID), nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(getResp, http.StatusNotFound)
	})

	t.Run("ServiceNotFound", func(t *testing.T) {
		// Create another version for this test
		anotherVersion := helpers.CreateTestServiceVersion(service.ID, "1.0.1", "Another version for testing")

		nonExistentServiceID := "non-existent-service-id"
		resp, err := helpers.MakeRequest("DELETE", fmt.Sprintf("/v1/services/%s/versions/%s", nonExistentServiceID, anotherVersion.ID), nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusNotFound)

		var errorResp models.ErrorResponse
		helpers.AssertJSONResponse(resp, &errorResp)

		assert.NotEmpty(t, errorResp.Message, "Error message should not be empty")
	})

	t.Run("VersionNotFound", func(t *testing.T) {
		nonExistentVersionID := "non-existent-version-id"
		resp, err := helpers.MakeRequest("DELETE", fmt.Sprintf("/v1/services/%s/versions/%s", service.ID, nonExistentVersionID), nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusNotFound)

		var errorResp models.ErrorResponse
		helpers.AssertJSONResponse(resp, &errorResp)

		assert.NotEmpty(t, errorResp.Message, "Error message should not be empty")
	})
}
// getVersionStrings extracts version strings from service versions for debugging
func getVersionStrings(versions []*models.ServiceVersion) []string {
	result := make([]string, len(versions))
	for i, version := range versions {
		result[i] = version.Version
	}
	return result
}
