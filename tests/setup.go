package tests

import (
	"log"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/thilak009/kong-assignment/db"
	"github.com/thilak009/kong-assignment/forms"
	"github.com/thilak009/kong-assignment/models"
	"github.com/thilak009/kong-assignment/routes"
	"gorm.io/gorm"
)

var (
	testDB     *gorm.DB
	testServer *httptest.Server
	testRouter *gin.Engine
)

// TestMain runs before any tests and sets up the test environment
func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

// setup initializes the test environment
func setup() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Setup test database
	setupTestDatabase()

	// Setup test router
	setupTestRouter()

	// Setup test server
	testServer = httptest.NewServer(testRouter)
}

// teardown cleans up the test environment
func teardown() {
	if testServer != nil {
		testServer.Close()
	}

	if testDB != nil {
		sqlDB, _ := testDB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

// setupTestDatabase initializes a test database connection using existing db package
func setupTestDatabase() {
	// Set test environment variables
	os.Setenv("DB_HOST", getEnv("TEST_DB_HOST", "localhost:5433"))
	os.Setenv("DB_USER", getEnv("TEST_DB_USER", "admin"))
	os.Setenv("DB_PASS", getEnv("TEST_DB_PASS", "admin"))
	os.Setenv("DB_NAME", getEnv("TEST_DB_NAME", "konnect"))

	// Initialize database using existing db package
	db.Init()

	// Get the initialized database instance
	testDB = db.GetDB()

	// Run migrations using existing function
	err := db.RunMigrations(&models.User{}, &models.Organization{}, &models.Service{}, &models.ServiceVersion{}, &models.UserOrganizationMap{})
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
}

// setupTestRouter creates a test router reusing main.go setup
func setupTestRouter() {
	testRouter = gin.New()

	// Add basic middleware (minimal for testing)
	testRouter.Use(gin.Logger())
	testRouter.Use(gin.Recovery())

	// Use the same form validator as main app
	binding.Validator = new(forms.DefaultValidator)

	// Setup routes using the same function as main.go
	routes.SetupRoutes(testRouter)
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// GetTestDB returns the test database instance
func GetTestDB() *gorm.DB {
	return testDB
}

// GetTestRouter returns the test router instance
func GetTestRouter() *gin.Engine {
	return testRouter
}

// GetTestServer returns the test server instance
func GetTestServer() *httptest.Server {
	return testServer
}
