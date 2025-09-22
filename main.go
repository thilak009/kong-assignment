package main

import (
	stdlog "log"
	"net/http"
	"os"

	db "github.com/thilak009/kong-assignment/db"
	"github.com/thilak009/kong-assignment/models"
	"github.com/thilak009/kong-assignment/pkg/log"
	"github.com/thilak009/kong-assignment/pkg/middleware"
	"github.com/thilak009/kong-assignment/routes"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/joho/godotenv"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/thilak009/kong-assignment/docs"
	"github.com/thilak009/kong-assignment/forms"
)

// @title           Konnect
// @version         1.0
// @description     API server for the Konnect Platform
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @host      localhost:9000
// @BasePath  /v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func main() {
	//Load the .env file
	err := godotenv.Load(".env")
	if err != nil {
		stdlog.Fatal("error: failed to load the env file")
	}

	if os.Getenv("ENV") == "PRODUCTION" {
		gin.SetMode(gin.ReleaseMode)
	}

	//Start the gin server without default middleware
	r := gin.New()

	// Add only the recovery middleware (panics), skip default logger
	r.Use(gin.Recovery())

	//Custom form validator
	binding.Validator = new(forms.DefaultValidator)

	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggingMiddleware())
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	//Start PostgreSQL database
	db.Init()
	// create the https://www.postgresql.org/docs/current/pgtrgm.html extension before doing auto migrate
	// improves like operation efficiency for search
	db.GetDB().Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm;")

	// Run migrations
	db.RunMigrations(
		&models.User{},
		&models.Organization{},
		&models.Service{},
		&models.ServiceVersion{},
		&models.UserOrganizationMap{},
		&models.BlacklistedToken{},
	)

	// Setup API routes
	routes.SetupRoutes(r)

	// Swagger docs
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	r.GET("/", func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"status": "UP",
		})
	})

	// Start periodic cleanup of expired blacklisted tokens
	go models.StartTokenCleanup()

	port := os.Getenv("PORT")

	// Log server startup info using our logger
	logger := log.GetLogger()
	logger.Infof("Starting server on port %s (ENV: %s, Version: %s)", port, os.Getenv("ENV"), os.Getenv("API_VERSION"))

	r.Run(":" + port)
}
