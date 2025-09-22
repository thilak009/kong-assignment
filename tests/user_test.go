package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thilak009/kong-assignment/models"
)

// TestUserRegistration tests POST /v1/users/register endpoint
func TestUserRegistration(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	t.Run("Success", func(t *testing.T) {
		payload := map[string]interface{}{
			"email":    "test@example.com",
			"name":     "Test User",
			"password": "password123",
		}

		resp, err := helpers.MakeRequest("POST", "/v1/users/register", payload)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp, http.StatusCreated)

		var user models.User
		helpers.AssertJSONResponse(resp, &user)

		assert.NotEmpty(t, user.ID, "User ID should not be empty")
		assert.Equal(t, "test@example.com", user.Email, "Email should match")
		assert.Equal(t, "Test User", user.Name, "Name should match")
		assert.Empty(t, user.Password, "Password should not be returned")
		assert.False(t, user.CreatedAt.IsZero(), "CreatedAt should not be zero")
		assert.False(t, user.UpdatedAt.IsZero(), "UpdatedAt should not be zero")
	})

	t.Run("ValidationErrors", func(t *testing.T) {
		testCases := []struct {
			name         string
			payload      map[string]interface{}
			expectedCode int
		}{
			{
				name:         "Missing email",
				payload:      map[string]interface{}{"name": "Test User", "password": "password123"},
				expectedCode: http.StatusBadRequest,
			},
			{
				name:         "Missing name",
				payload:      map[string]interface{}{"email": "validation1@example.com", "password": "password123"},
				expectedCode: http.StatusBadRequest,
			},
			{
				name:         "Missing password",
				payload:      map[string]interface{}{"email": "validation2@example.com", "name": "Test User"},
				expectedCode: http.StatusBadRequest,
			},
			{
				name:         "Invalid email format",
				payload:      map[string]interface{}{"email": "invalid-email", "name": "Test User", "password": "password123"},
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
				resp, err := helpers.MakeRequest("POST", "/v1/users/register", tc.payload)
				if err != nil {
					t.Fatalf("Failed to make request: %v", err)
				}

				helpers.AssertStatusCode(resp, tc.expectedCode)
				helpers.AssertErrorResponseNotEmpty(resp)
			})
		}
	})

	t.Run("DuplicateEmail", func(t *testing.T) {
		// Create first user
		payload1 := map[string]interface{}{
			"email":    "duplicate@example.com",
			"name":     "User One",
			"password": "password123",
		}

		resp1, err := helpers.MakeRequest("POST", "/v1/users/register", payload1)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		helpers.AssertStatusCode(resp1, http.StatusCreated)

		// Try to create second user with same email
		payload2 := map[string]interface{}{
			"email":    "duplicate@example.com",
			"name":     "User Two",
			"password": "password123",
		}

		resp2, err := helpers.MakeRequest("POST", "/v1/users/register", payload2)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		helpers.AssertStatusCode(resp2, http.StatusConflict)
		helpers.AssertErrorResponse(resp2, "User with this email already exists")
	})
}

// TestUserLogin tests POST /v1/users/login endpoint
func TestUserLogin(t *testing.T) {
	helpers := NewTestHelpers(t)

	// Clean database before and after test
	helpers.CleanupDatabase()
	t.Cleanup(func() {
		helpers.CleanupDatabase()
	})

	t.Run("Success", func(t *testing.T) {
		// Create test user
		_, token := helpers.CreateTestUser("test@example.com", "Test User", "password123")

		assert.NotEmpty(t, token, "Token should not be empty")
	})

	t.Run("InvalidCredentials", func(t *testing.T) {
		// Create test user
		helpers.CreateTestUser("test2@example.com", "Test User 2", "password123")

		testCases := []struct {
			name     string
			email    string
			password string
		}{
			{
				name:     "Wrong password",
				email:    "test2@example.com",
				password: "wrongpassword",
			},
			{
				name:     "Non-existent email",
				email:    "nonexistent@example.com",
				password: "password123",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				payload := map[string]interface{}{
					"email":    tc.email,
					"password": tc.password,
				}

				resp, err := helpers.MakeRequest("POST", "/v1/users/login", payload)
				if err != nil {
					t.Fatalf("Failed to make request: %v", err)
				}

				helpers.AssertStatusCode(resp, http.StatusUnauthorized)
				helpers.AssertErrorResponse(resp, "Invalid email/password")
			})
		}
	})

	t.Run("ValidationErrors", func(t *testing.T) {
		testCases := []struct {
			name         string
			payload      map[string]interface{}
			expectedCode int
		}{
			{
				name:         "Missing email",
				payload:      map[string]interface{}{"password": "password123"},
				expectedCode: http.StatusBadRequest,
			},
			{
				name:         "Missing password",
				payload:      map[string]interface{}{"email": "test@example.com"},
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
				resp, err := helpers.MakeRequest("POST", "/v1/users/login", tc.payload)
				if err != nil {
					t.Fatalf("Failed to make request: %v", err)
				}

				helpers.AssertStatusCode(resp, tc.expectedCode)
				helpers.AssertErrorResponseNotEmpty(resp)
			})
		}
	})
}
