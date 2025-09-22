package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/thilak009/kong-assignment/controllers"
	"github.com/thilak009/kong-assignment/pkg/middleware"
)

// SetupRoutes configures all API routes for the given router
func SetupRoutes(r *gin.Engine) {
	v1 := r.Group("/v1")
	{
		/*** User Authentication - No auth required ***/
		userController := new(controllers.UserController)

		v1.POST("/user/register", userController.Register)
		v1.POST("/user/login", userController.Login)
		// v1.POST("/user/logout", userController.Logout)

		/*** Protected routes - require authentication ***/
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			/*** Organizations ***/
			orgController := new(controllers.OrganizationController)

			protected.POST("/orgs", orgController.CreateOrganization)
			protected.GET("/orgs", orgController.GetOrganizations)
			protected.GET("/orgs/:orgId", orgController.GetOrganization)
			protected.PUT("/orgs/:orgId", orgController.UpdateOrganization)
			protected.DELETE("/orgs/:orgId", orgController.DeleteOrganization)

			/*** Organization Services ***/
			orgServiceController := new(controllers.ServiceController)

			protected.POST("/orgs/:orgId/services", orgServiceController.CreateService)
			protected.GET("/orgs/:orgId/services", orgServiceController.GetServices)
			protected.GET("/orgs/:orgId/services/:serviceId", orgServiceController.GetService)
			protected.PUT("/orgs/:orgId/services/:serviceId", orgServiceController.UpdateService)
			protected.DELETE("/orgs/:orgId/services/:serviceId", orgServiceController.DeleteService)

			/*** Organization Service Versions ***/
			orgServiceVersionController := new(controllers.ServiceVersionController)

			protected.POST("/orgs/:orgId/services/:serviceId/versions", orgServiceVersionController.CreateServiceVersion)
			protected.GET("/orgs/:orgId/services/:serviceId/versions", orgServiceVersionController.GetServiceVersions)
			protected.GET("/orgs/:orgId/services/:serviceId/versions/:versionId", orgServiceVersionController.GetServiceVersion)
			protected.PATCH("/orgs/:orgId/services/:serviceId/versions/:versionId", orgServiceVersionController.UpdateServiceVersion)
			protected.DELETE("/orgs/:orgId/services/:serviceId/versions/:versionId", orgServiceVersionController.DeleteServiceVersion)
		}
	}
}
