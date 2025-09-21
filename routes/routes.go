package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/thilak009/kong-assignment/controllers"
)

// SetupRoutes configures all API routes for the given router
func SetupRoutes(r *gin.Engine) {
	v1 := r.Group("/v1")
	{
		/*** Service ***/
		service := new(controllers.ServiceController)

		v1.POST("/services", service.Create)
		v1.GET("/services", service.All)
		v1.GET("/services/:id", service.One)
		v1.PUT("/services/:id", service.Update)
		v1.DELETE("/services/:id", service.Delete)

		/*** Service Version ***/
		serviceVersion := new(controllers.ServiceVersionController)

		v1.POST("/services/:id/versions", serviceVersion.Create)
		v1.GET("/services/:id/versions", serviceVersion.All)
		v1.GET("/services/:id/versions/:versionId", serviceVersion.One)
		v1.PATCH("/services/:id/versions/:versionId", serviceVersion.Update)
		v1.DELETE("/services/:id/versions/:versionId", serviceVersion.Delete)
	}
}