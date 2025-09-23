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

		v1.POST("/users/register", userController.Register)
		v1.POST("/users/login", userController.Login)

		/*** Protected routes - require authentication ***/
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			/*** User Authentication - Auth required ***/
			protected.POST("/users/logout", userController.Logout)

			/*** Organizations ***/
			orgController := new(controllers.OrganizationController)

			protected.POST("/orgs", orgController.CreateOrganization)
			protected.GET("/orgs", orgController.GetOrganizations)
			/*** Organization routes - require organization access ***/
			protected.GET("/orgs/:orgId", middleware.OrganizationAccessMiddleware(), orgController.GetOrganization)
			protected.PUT("/orgs/:orgId", middleware.OrganizationAccessMiddleware(), orgController.UpdateOrganization)
			protected.DELETE("/orgs/:orgId", middleware.OrganizationAccessMiddleware(), orgController.DeleteOrganization)

			/*** Organization Services - require organization access ***/
			orgServiceController := new(controllers.ServiceController)

			protected.POST("/orgs/:orgId/services", middleware.OrganizationAccessMiddleware(), orgServiceController.CreateService)
			protected.GET("/orgs/:orgId/services", middleware.OrganizationAccessMiddleware(), orgServiceController.GetServices)
			protected.GET("/orgs/:orgId/services/:serviceId", middleware.OrganizationAccessMiddleware(), orgServiceController.GetService)
			protected.PATCH("/orgs/:orgId/services/:serviceId", middleware.OrganizationAccessMiddleware(), orgServiceController.UpdateService)
			protected.DELETE("/orgs/:orgId/services/:serviceId", middleware.OrganizationAccessMiddleware(), orgServiceController.DeleteService)

			/*** Organization Service Versions - require organization access ***/
			orgServiceVersionController := new(controllers.ServiceVersionController)

			protected.POST("/orgs/:orgId/services/:serviceId/versions", middleware.OrganizationAccessMiddleware(), orgServiceVersionController.CreateServiceVersion)
			protected.GET("/orgs/:orgId/services/:serviceId/versions", middleware.OrganizationAccessMiddleware(), orgServiceVersionController.GetServiceVersions)
			protected.GET("/orgs/:orgId/services/:serviceId/versions/:versionId", middleware.OrganizationAccessMiddleware(), orgServiceVersionController.GetServiceVersion)
			protected.PATCH("/orgs/:orgId/services/:serviceId/versions/:versionId", middleware.OrganizationAccessMiddleware(), orgServiceVersionController.UpdateServiceVersion)
			protected.DELETE("/orgs/:orgId/services/:serviceId/versions/:versionId", middleware.OrganizationAccessMiddleware(), orgServiceVersionController.DeleteServiceVersion)
		}
	}
}
