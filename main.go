package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	db "github.com/thilak009/kong-assignment/db"
	"github.com/thilak009/kong-assignment/models"
	"github.com/thilak009/kong-assignment/routes"
	"github.com/thilak009/kong-assignment/utils"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/thilak009/kong-assignment/docs"
	"github.com/thilak009/kong-assignment/forms"
)

// CORSMiddleware ...
// CORS (Cross-Origin Resource Sharing)
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Origin, Authorization, Accept, Client-Security-Token, Accept-Encoding, x-access-token")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			fmt.Println("OPTIONS")
			c.AbortWithStatus(200)
		} else {
			c.Next()
		}
	}
}

// RequestIDMiddleware ...
// Generate a unique ID and attach it to each request for future reference or use
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set(string(utils.RequestIDKey), requestID)
		c.Next()
	}
}

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
		log.Fatal("error: failed to load the env file")
	}

	if os.Getenv("ENV") == "PRODUCTION" {
		gin.SetMode(gin.ReleaseMode)
	}

	//Start the default gin server
	r := gin.Default()

	//Custom form validator
	binding.Validator = new(forms.DefaultValidator)

	r.Use(CORSMiddleware())
	r.Use(RequestIDMiddleware())
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

	port := os.Getenv("PORT")

	log.Printf("\n\n PORT: %s \n ENV: %s \n Version: %s \n\n", port, os.Getenv("ENV"), os.Getenv("API_VERSION"))

	r.Run(":" + port)
}
