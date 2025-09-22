package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thilak009/kong-assignment/models"
)

// TestCreateOrganization tests POST /v1/orgs endpoint
func TestCreateOrganization(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	t.Run("Success", func(t *testing.T) {
		// Create test user
		_, token := helpers.CreateTestUser("test@example.com", "Test User", TestPassword)

		payload := map[string]interface{}{
			"name":        "Test Organization",
			"description": "This is a test organization",
		}

		resp, err := helpers.MakeAuthenticatedRequest("POST", "/v1/orgs", payload, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusCreated)

		var org models.Organization
		helpers.AssertJSONResponse(resp, &org)

		assert.NotEmpty(t, org.ID, "Organization ID should not be empty")
		assert.Equal(t, "Test Organization", org.Name, "Organization name should match")
		assert.Equal(t, "This is a test organization", org.Description, "Organization description should match")
		assert.False(t, org.CreatedAt.IsZero(), "CreatedAt should not be zero")
		assert.False(t, org.UpdatedAt.IsZero(), "UpdatedAt should not be zero")
	})

	t.Run("ValidationErrors", func(t *testing.T) {
		// Create test user
		_, token := helpers.CreateTestUser("test2@example.com", "Test User 2", TestPassword)

		testCases := []struct {
			name         string
			payload      map[string]interface{}
			expectedCode int
		}{
			{
				name:         "Missing name",
				payload:      map[string]interface{}{"description": "Valid description"},
				expectedCode: http.StatusBadRequest,
			},
			{
				name:         "Missing description",
				payload:      map[string]interface{}{"name": "Valid Name"},
				expectedCode: http.StatusBadRequest,
			},
			{
				name:         "Name too short",
				payload:      map[string]interface{}{"name": "AB", "description": "Valid description"},
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
				resp, err := helpers.MakeAuthenticatedRequest("POST", "/v1/orgs", tc.payload, token)
				if err != nil {
					t.Fatalf("Failed to make request: %v", err)
				}

				helpers.AssertStatusCode(resp, tc.expectedCode)
				helpers.AssertErrorResponseNotEmpty(resp)
			})
		}
	})

	t.Run("Unauthorized", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":        "Test Organization",
			"description": "This is a test organization",
		}

		// Test without token
		resp, err := helpers.MakeRequest("POST", "/v1/orgs", payload)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusUnauthorized)
	})
}

// TestGetOrganizations tests GET /v1/orgs endpoint
func TestGetOrganizations(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	t.Run("EmptyList", func(t *testing.T) {
		// Create test user
		_, token := helpers.CreateTestUser("test@example.com", "Test User", TestPassword)

		resp, err := helpers.MakeAuthenticatedRequest("GET", "/v1/orgs", nil, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusOK)

		var result models.PaginatedResult[models.Organization]
		helpers.AssertJSONResponse(resp, &result)

		assert.Len(t, result.Data, 0, "Expected empty organizations data")
		assert.Equal(t, 0, result.Meta.TotalCount, "Expected total count to be 0")
	})

	t.Run("WithData", func(t *testing.T) {
		// Create test user and organization
		_, token := helpers.CreateTestUser("test2@example.com", "Test User 2", TestPassword)
		helpers.CreateTestOrganization(token, "Test Organization", "Test organization description")

		resp, err := helpers.MakeAuthenticatedRequest("GET", "/v1/orgs", nil, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusOK)

		var result models.PaginatedResult[models.Organization]
		helpers.AssertJSONResponse(resp, &result)

		assert.Len(t, result.Data, 1, "Expected 1 organization in data")
		assert.Equal(t, 1, result.Meta.TotalCount, "Expected total count to be 1")
		assert.Equal(t, "Test Organization", result.Data[0].Name, "Organization name should match")
	})

	t.Run("WithQueryParameters", func(t *testing.T) {
		// Create test user and organizations
		_, token := helpers.CreateTestUser("test3@example.com", "Test User 3", TestPassword)
		helpers.CreateTestOrganization(token, "Alpha Organization", "First organization")
		helpers.CreateTestOrganization(token, "Beta Organization", "Second organization")

		testCases := []struct {
			name       string
			query      string
			shouldFind bool
		}{
			{
				name:       "Search by name - exact match",
				query:      "?q=Alpha Organization",
				shouldFind: true,
			},
			{
				name:       "Search by name - partial match",
				query:      "?q=Alpha",
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
				name:       "Pagination - page 0",
				query:      "?page=0&per_page=1",
				shouldFind: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := helpers.MakeAuthenticatedRequest("GET", fmt.Sprintf("/v1/orgs%s", tc.query), nil, token)
				if err != nil {
					t.Fatalf("Failed to make request: %v", err)
				}

				helpers.AssertStatusCode(resp, http.StatusOK)

				var result models.PaginatedResult[models.Organization]
				helpers.AssertJSONResponse(resp, &result)

				if tc.shouldFind {
					assert.NotEmpty(t, result.Data, "Expected to find organizations but got empty result")
				} else {
					assert.Empty(t, result.Data, "Expected empty result")
				}
			})
		}
	})

	t.Run("Unauthorized", func(t *testing.T) {
		resp, err := helpers.MakeRequest("GET", "/v1/orgs", nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusUnauthorized)
	})
}

// TestGetOrganization tests GET /v1/orgs/{orgId} endpoint
func TestGetOrganization(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	t.Run("Success", func(t *testing.T) {
		// Create test user and organization
		_, token := helpers.CreateTestUser("test@example.com", "Test User", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test organization description")

		resp, err := helpers.MakeAuthenticatedRequest("GET", fmt.Sprintf("/v1/orgs/%s", org.ID), nil, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusOK)

		var retrievedOrg models.Organization
		helpers.AssertJSONResponse(resp, &retrievedOrg)

		assert.Equal(t, org.ID, retrievedOrg.ID, "Organization ID should match")
		assert.Equal(t, "Test Organization", retrievedOrg.Name, "Organization name should match")
		assert.Equal(t, "Test organization description", retrievedOrg.Description, "Organization description should match")
	})

	t.Run("NotFound", func(t *testing.T) {
		// Create test user
		_, token := helpers.CreateTestUser("test2@example.com", "Test User 2", TestPassword)

		nonExistentID := "non-existent-id"
		resp, err := helpers.MakeAuthenticatedRequest("GET", fmt.Sprintf("/v1/orgs/%s", nonExistentID), nil, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusForbidden)
		helpers.AssertErrorResponse(resp, "You are not authorized to perform the request")
	})

	t.Run("Forbidden", func(t *testing.T) {
		// Create first user and organization
		_, token1 := helpers.CreateTestUser("test3@example.com", "Test User 3", TestPassword)
		org1 := helpers.CreateTestOrganization(token1, "Test Organization 1", "Test org description")

		// Create second user (different user)
		_, token2 := helpers.CreateTestUser("test4@example.com", "Test User 4", TestPassword)

		// Try to access org1 using token2 (user4 is not member of org1)
		resp, err := helpers.MakeAuthenticatedRequest("GET", fmt.Sprintf("/v1/orgs/%s", org1.ID), nil, token2)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusForbidden)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		// Create test user and organization
		_, token := helpers.CreateTestUser("test5@example.com", "Test User 5", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test organization description")

		resp, err := helpers.MakeRequest("GET", fmt.Sprintf("/v1/orgs/%s", org.ID), nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusUnauthorized)
	})
}

// TestUpdateOrganization tests PUT /v1/orgs/{orgId} endpoint
func TestUpdateOrganization(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	t.Run("Success", func(t *testing.T) {
		// Create test user and organization
		_, token := helpers.CreateTestUser("test@example.com", "Test User", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test organization description")

		payload := map[string]interface{}{
			"name":        "Updated Organization",
			"description": "Updated organization description",
		}

		resp, err := helpers.MakeAuthenticatedRequest("PUT", fmt.Sprintf("/v1/orgs/%s", org.ID), payload, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusOK)

		var updatedOrg models.Organization
		helpers.AssertJSONResponse(resp, &updatedOrg)

		assert.Equal(t, "Updated Organization", updatedOrg.Name, "Organization name should be updated")
		assert.Equal(t, "Updated organization description", updatedOrg.Description, "Organization description should be updated")
	})

	t.Run("NotFound", func(t *testing.T) {
		// Create test user
		_, token := helpers.CreateTestUser("test2@example.com", "Test User 2", TestPassword)

		payload := map[string]interface{}{
			"name":        "Updated Organization",
			"description": "Updated description",
		}

		nonExistentID := "non-existent-id"
		resp, err := helpers.MakeAuthenticatedRequest("PUT", fmt.Sprintf("/v1/orgs/%s", nonExistentID), payload, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusForbidden)
	})

	t.Run("ValidationErrors", func(t *testing.T) {
		// Create test user and organization
		_, token := helpers.CreateTestUser("test3@example.com", "Test User 3", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test organization description")

		testCases := []struct {
			name    string
			payload map[string]interface{}
		}{
			{
				name:    "Invalid name",
				payload: map[string]interface{}{"name": "AB", "description": "Valid description"},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := helpers.MakeAuthenticatedRequest("PUT", fmt.Sprintf("/v1/orgs/%s", org.ID), tc.payload, token)
				if err != nil {
					t.Fatalf("Failed to make request: %v", err)
				}

				helpers.AssertStatusCode(resp, http.StatusBadRequest)
				helpers.AssertErrorResponseNotEmpty(resp)
			})
		}
	})
}

// TestDeleteOrganization tests DELETE /v1/orgs/{orgId} endpoint
func TestDeleteOrganization(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	t.Run("Success", func(t *testing.T) {
		// Create test user and organization
		_, token := helpers.CreateTestUser("test@example.com", "Test User", TestPassword)
		org := helpers.CreateTestOrganization(token, "Test Organization", "Test organization description")

		resp, err := helpers.MakeAuthenticatedRequest("DELETE", fmt.Sprintf("/v1/orgs/%s", org.ID), nil, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusNoContent)

		// Verify organization is deleted by trying to get it
		getResp, err := helpers.MakeAuthenticatedRequest("GET", fmt.Sprintf("/v1/orgs/%s", org.ID), nil, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(getResp, http.StatusForbidden)
	})

	t.Run("NotFound", func(t *testing.T) {
		// Create test user
		_, token := helpers.CreateTestUser("test2@example.com", "Test User 2", TestPassword)

		nonExistentID := "non-existent-id"
		resp, err := helpers.MakeAuthenticatedRequest("DELETE", fmt.Sprintf("/v1/orgs/%s", nonExistentID), nil, token)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusForbidden)
	})
}