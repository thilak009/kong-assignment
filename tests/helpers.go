package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thilak009/kong-assignment/models"
)

// TestHelpers provides utility functions for testing
type TestHelpers struct {
	t *testing.T
}

// NewTestHelpers creates a new TestHelpers instance
func NewTestHelpers(t *testing.T) *TestHelpers {
	return &TestHelpers{t: t}
}

// ensureTestEnvironment ensures the test environment is initialized
func (h *TestHelpers) ensureTestEnvironment() {
	if GetTestDB() == nil || GetTestRouter() == nil {
		setup()
	}
}

// MakeRequest makes an HTTP request to the test server
func (h *TestHelpers) MakeRequest(method, path string, body interface{}) (*httptest.ResponseRecorder, error) {
	h.ensureTestEnvironment()
	var reqBody io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	recorder := httptest.NewRecorder()
	GetTestRouter().ServeHTTP(recorder, req)

	return recorder, nil
}

// AssertStatusCode checks if the response has the expected status code
func (h *TestHelpers) AssertStatusCode(recorder *httptest.ResponseRecorder, expectedStatus int) {
	assert.Equal(h.t, expectedStatus, recorder.Code, "Response body: %s", recorder.Body.String())
}

// AssertJSONResponse unmarshals response JSON into the provided interface
func (h *TestHelpers) AssertJSONResponse(recorder *httptest.ResponseRecorder, v interface{}) {
	err := json.Unmarshal(recorder.Body.Bytes(), v)
	assert.NoError(h.t, err, "Failed to unmarshal response JSON. Response body: %s", recorder.Body.String())
}

// AssertErrorResponse checks if the response contains an error with expected message
func (h *TestHelpers) AssertErrorResponse(recorder *httptest.ResponseRecorder, expectedMessage string) {
	var errorResp models.ErrorResponse
	h.AssertJSONResponse(recorder, &errorResp)
	assert.Equal(h.t, expectedMessage, errorResp.Message, "Error message mismatch")
}

// AssertErrorResponseNotEmpty checks if the response contains a non-empty error message
func (h *TestHelpers) AssertErrorResponseNotEmpty(recorder *httptest.ResponseRecorder) {
	var errorResp models.ErrorResponse
	h.AssertJSONResponse(recorder, &errorResp)
	assert.NotEmpty(h.t, errorResp.Message, "Error message should not be empty")
}

// AssertServiceFields validates all service fields using testify
func (h *TestHelpers) AssertServiceFields(service models.Service, expectedName, expectedDescription string) {
	assert.NotEmpty(h.t, service.ID, "Service ID should not be empty")
	assert.Equal(h.t, expectedName, service.Name, "Service name mismatch")
	assert.Equal(h.t, expectedDescription, service.Description, "Service description mismatch")
	assert.False(h.t, service.CreatedAt.IsZero(), "CreatedAt should not be zero")
	assert.False(h.t, service.UpdatedAt.IsZero(), "UpdatedAt should not be zero")
}

// AssertServiceVersionFields validates all service version fields using testify
func (h *TestHelpers) AssertServiceVersionFields(version models.ServiceVersion, expectedServiceID, expectedVersion, expectedDescription string) {
	assert.NotEmpty(h.t, version.ID, "Version ID should not be empty")
	assert.Equal(h.t, expectedServiceID, version.ServiceID, "Service ID mismatch")
	assert.Equal(h.t, expectedVersion, version.Version, "Version mismatch")
	assert.Equal(h.t, expectedDescription, version.Description, "Version description mismatch")
	assert.False(h.t, version.CreatedAt.IsZero(), "CreatedAt should not be zero")
	assert.False(h.t, version.UpdatedAt.IsZero(), "UpdatedAt should not be zero")
}

// AssertVersionCountIncluded checks if version count is included and has expected value
func (h *TestHelpers) AssertVersionCountIncluded(service models.Service, expectedCount int) {
	assert.NotNil(h.t, service.Metadata.VersionCount, "Version count should be included in metadata")
	if service.Metadata.VersionCount != nil {
		assert.Equal(h.t, expectedCount, *service.Metadata.VersionCount, "Version count mismatch")
	}
}

// CleanupDatabase cleans all test data from the database
func (h *TestHelpers) CleanupDatabase() {
	h.ensureTestEnvironment()
	testDB := GetTestDB()
	if testDB == nil {
		h.t.Fatal("Test database not initialized")
	}

	// Clean tables in reverse order of dependencies
	testDB.Exec("DELETE FROM service_versions")
	testDB.Exec("DELETE FROM services")
}

// CreateTestService creates a test service in the database
func (h *TestHelpers) CreateTestService(name, description string) *models.Service {
	h.ensureTestEnvironment()
	service := &models.Service{
		Name:        name,
		Description: description,
	}

	testDB := GetTestDB()
	if err := testDB.Create(service).Error; err != nil {
		h.t.Fatalf("Failed to create test service: %v", err)
	}

	return service
}

// CreateTestServiceVersion creates a test service version in the database
func (h *TestHelpers) CreateTestServiceVersion(serviceID, version, description string) *models.ServiceVersion {
	h.ensureTestEnvironment()
	serviceVersion := &models.ServiceVersion{
		ServiceID:   serviceID,
		Version:     version,
		Description: description,
	}

	testDB := GetTestDB()
	if err := testDB.Create(serviceVersion).Error; err != nil {
		h.t.Fatalf("Failed to create test service version: %v", err)
	}

	return serviceVersion
}

// GetTestServerURL returns the test server URL
func (h *TestHelpers) GetTestServerURL() string {
	return GetTestServer().URL
}

// LogResponse logs the response for debugging purposes
func (h *TestHelpers) LogResponse(recorder *httptest.ResponseRecorder) {
	h.t.Logf("Response Status: %d", recorder.Code)
	h.t.Logf("Response Headers: %v", recorder.Header())
	h.t.Logf("Response Body: %s", recorder.Body.String())
}