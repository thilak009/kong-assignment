package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/thilak009/kong-assignment/forms"
	"github.com/thilak009/kong-assignment/models"
)

type ServiceController struct{}

var serviceModel = models.ServiceModel{}
var serviceForm = forms.ServiceForm{}
var orgModel = models.OrganizationModel{}

// parseIncludeParams parses comma-separated include parameter and returns flags for each supported field
func parseIncludeParams(include string) (includeVersionCount bool) {
	if include == "" {
		return false
	}

	includeFields := strings.Split(include, ",")
	for _, field := range includeFields {
		if strings.TrimSpace(field) == "versionCount" {
			includeVersionCount = true
		}
	}
	return includeVersionCount
}


// CreateService creates a new service in an organization
// @Summary Create a service
// @Schemes
// @Description Creates a service
// @Tags Service
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID"
// @Param service body forms.CreateServiceForm true "Service"
// @Success 	 200  {object}  models.Service
// @Failure      400  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      500  {object}	models.ErrorResponse
// @Security BearerAuth
// @Router /orgs/{orgId}/services [post]
func (ctrl ServiceController) CreateService(c *gin.Context) {
	orgID := c.Param("orgId")

	var form forms.CreateServiceForm
	if validationErr := c.ShouldBindJSON(&form); validationErr != nil {
		message := serviceForm.Create(validationErr)
		models.AbortWithError(c, http.StatusBadRequest, message)
		return
	}

	service, err := serviceModel.Create(c.Request.Context(), form, orgID)
	if err != nil {
		models.AbortWithError(c, http.StatusInternalServerError, "Service could not be created")
		return
	}

	c.JSON(http.StatusOK, service)
}

// Get All Services godoc
// @Summary Get All services
// @Schemes
// @Description Gets all the services available
// @Tags Service
// @Accept json
// @Produce json
// @Param	orgId path string true "Organization ID"
// @Param	q	query   string	false	"Service name, supports searching the passed string in the name of the service"
// @Param	sort	query   string	false	"Sort order for the list of services. Accepted values are asc and desc. Default is desc(assumes default on invalid values as well)" Enums(asc, desc)
// @Param	sort_by	query   string	false	"The field on which sorting to be applied, supports name, created_at, updated_at. Default is updated_at(assumes default on invalid values as well)" Enums(name, created_at, updated_at)
// @Param	page	query   int	false	"Page number for pagination (0-based). Default is 0"
// @Param	per_page	query   int	false	"Number of items per page. Default is 10, max is 100, assumes 100 if >100 is passed"
// @Param	include	query   string	false	"Additional data to include (comma-separated). Supported values: versionCount"
// @Success 	 200  {object}  models.PaginatedResult[models.Service]
// @Failure      403  {object}	models.ErrorResponse
// @Failure      500  {object}	models.ErrorResponse
// @Security BearerAuth
// @Router /orgs/{orgId}/services [GET]
func (ctrl ServiceController) GetServices(c *gin.Context) {
	orgID := c.Param("orgId")

	q := c.Query("q")
	sortBy, sort := models.ParseSortParams(c, models.GetServiceValidSortFields(), "updated_at")
	page, perPage := models.ParsePaginationParams(c)

	// Parse include parameter for multiple values
	include := c.Query("include")
	includeVersionCount := parseIncludeParams(include)

	results, err := serviceModel.All(c.Request.Context(), orgID, q, sortBy, sort, page, perPage, includeVersionCount)
	if err != nil {
		models.AbortWithError(c, http.StatusInternalServerError, "Could not get services")
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetService gets a specific service by ID
// @Summary Get a service
// @Schemes
// @Description Gets the specified service
// @Tags Service
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID"
// @Param	serviceId	path	string	true	"Service ID"
// @Param	include	query   string	false	"Additional data to include (comma-separated). Supported values: versionCount"
// @Success 	 200  {object}  models.Service
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Security BearerAuth
// @Router /orgs/{orgId}/services/{serviceId} [GET]
func (ctrl ServiceController) GetService(c *gin.Context) {
	orgID := c.Param("orgId")

	serviceID := c.Param("serviceId")
	include := c.DefaultQuery("include", "")
	includeVersionCount := parseIncludeParams(include)

	service, isFound, err := serviceModel.One(c.Request.Context(), serviceID, orgID, includeVersionCount)
	if err != nil {
		if !isFound {
			models.AbortWithError(c, http.StatusNotFound, "Service not found")
			return
		}
		models.AbortWithError(c, http.StatusInternalServerError, "Could not get service")
		return
	}

	c.JSON(http.StatusOK, service)
}

// UpdateService updates a service
// @Summary Update a service
// @Schemes
// @Description Updates the specified service. Both name and description are optional.
// @Tags Service
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID"
// @Param	serviceId	path	string	true	"Service ID"
// @Param service body forms.UpdateServiceForm true "Service"
// @Success 	 200  {object}  models.Service
// @Failure      400  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Security BearerAuth
// @Router /orgs/{orgId}/services/{serviceId} [PATCH]
func (ctrl ServiceController) UpdateService(c *gin.Context) {
	orgID := c.Param("orgId")

	var form forms.UpdateServiceForm
	if validationErr := c.ShouldBindJSON(&form); validationErr != nil {
		message := serviceForm.Update(validationErr)
		models.AbortWithError(c, http.StatusBadRequest, message)
		return
	}

	// Validate that at least one field is provided
	if message := serviceForm.ValidateUpdate(form); message != "" {
		models.AbortWithError(c, http.StatusBadRequest, message)
		return
	}

	serviceID := c.Param("serviceId")
	_, isFound, err := serviceModel.One(c.Request.Context(), serviceID, orgID, false)
	if err != nil {
		if !isFound {
			models.AbortWithError(c, http.StatusNotFound, "Service not found")
			return
		}
		models.AbortWithError(c, http.StatusInternalServerError, "Could not get service")
		return
	}

	service, err := serviceModel.Update(c.Request.Context(), serviceID, orgID, form)
	if err != nil {
		models.AbortWithError(c, http.StatusInternalServerError, "Service could not be updated")
		return
	}
	c.JSON(http.StatusOK, service)
}

// DeleteService deletes a service
// @Summary Delete a service
// @Schemes
// @Description Deletes the specified service
// @Tags Service
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID"
// @Param	serviceId	path	string	true	"Service ID"
// @Success 	 204  ""
// @Failure      403  {object}  models.ErrorResponse
// @Failure 	 404  {object} models.ErrorResponse
// @Failure      500  {object} models.ErrorResponse
// @Security BearerAuth
// @Router /orgs/{orgId}/services/{serviceId} [DELETE]
func (ctrl ServiceController) DeleteService(c *gin.Context) {
	orgID := c.Param("orgId")

	serviceID := c.Param("serviceId")
	_, isFound, err := serviceModel.One(c.Request.Context(), serviceID, orgID, false)
	if err != nil {
		if !isFound {
			models.AbortWithError(c, http.StatusNotFound, "Service not found")
			return
		}
		models.AbortWithError(c, http.StatusInternalServerError, "Could not get service")
		return
	}

	err = serviceModel.Delete(c.Request.Context(), serviceID, orgID)
	if err != nil {
		models.AbortWithError(c, http.StatusInternalServerError, "Service could not be deleted")
		return
	}

	c.JSON(http.StatusNoContent, "")
}
