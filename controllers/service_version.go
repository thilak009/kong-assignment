package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thilak009/kong-assignment/forms"
	"github.com/thilak009/kong-assignment/models"
	"github.com/thilak009/kong-assignment/pkg/log"
)

type ServiceVersionController struct{}

var serviceVersionModel = new(models.ServiceVersionModel)
var serviceVersionForm = new(forms.ServiceVersionForm)

// CreateServiceVersion creates a new service version
// @Summary Create a version for a service
// @Schemes
// @Description Creates a version for the specified service
// @Description version value must be a semantic version
// @Tags ServiceVersion
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID"
// @Param	serviceId	path	string	true	"Service ID"
// @Param serviceVersion body forms.CreateServiceVersionForm true "ServiceVersion"
// @Success 	 200  {object}  models.ServiceVersion
// @Failure      400  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      500  {object} models.ErrorResponse
// @Security BearerAuth
// @Router /orgs/{orgId}/services/{serviceId}/versions [post]
func (ctrl ServiceVersionController) CreateServiceVersion(c *gin.Context) {
	_, orgID, hasAccess := checkOrganizationAccess(c)
	if !hasAccess {
		return
	}

	var form forms.CreateServiceVersionForm
	if validationErr := c.ShouldBindJSON(&form); validationErr != nil {
		log.With(c.Request.Context()).Debugf("Validation failed for service version creation: %v", validationErr)
		message := serviceVersionForm.Create(validationErr)
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
		models.AbortWithError(c, http.StatusInternalServerError, "Could not get versions")
		return
	}

	// TODO: handle same version tag creation by returning a bad request maybe
	version, err := serviceVersionModel.Create(c.Request.Context(), serviceID, form)
	if err != nil {
		models.AbortWithError(c, http.StatusInternalServerError, "Service version could not be created")
		return
	}

	c.JSON(http.StatusOK, version)
}

// GetServiceVersions gets all versions for a service
// @Summary Get All versions of a service
// @Schemes
// @Description Gets all the versions available for the specified service
// @Tags ServiceVersion
// @Accept json
// @Produce json
// @Param	q	query   string	false	"version, supports searching with version prefix, for example: passing 1 would return versions like 1.0.1,1.1.4 etc, passing 1.0 would return 1.0.3,1.0.7 etc"
// @Param	sort	query   string	false	"Sort order for the list of service versions. Accepted values are asc and desc. Default is desc(assumes default on invalid values as well)" Enums(asc, desc)
// @Param	sort_by	query   string	false	"The field on which sorting to be applied, supports version, created_at, updated_at. Default is updated_at(assumes default on invalid values as well)" Enums(version, created_at, updated_at)
// @Param	page	query   int	false	"Page number for pagination (0-based). Default is 0"
// @Param	per_page	query   int	false	"Number of items per page. Default is 10, max is 100, assumes 100 if >100 is passed"
// @Param	orgId path string true "Organization ID"
// @Param	serviceId	path	string	true	"Service ID"
// @Success 	 200  {object}  models.PaginatedResult[models.ServiceVersion]
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object} models.ErrorResponse
// @Security BearerAuth
// @Router /orgs/{orgId}/services/{serviceId}/versions [GET]
func (ctrl ServiceVersionController) GetServiceVersions(c *gin.Context) {
	_, orgID, hasAccess := checkOrganizationAccess(c)
	if !hasAccess {
		return
	}

	serviceID := c.Param("serviceId")

	_, isFound, err := serviceModel.One(c.Request.Context(), serviceID, orgID, false)
	if err != nil {
		if !isFound {
			models.AbortWithError(c, http.StatusNotFound, "Service not found")
			return
		}
		models.AbortWithError(c, http.StatusInternalServerError, "Could not get versions")
		return
	}
	q := c.Query("q")
	sortBy, sort := models.ParseSortParams(c, models.GetServiceVersionValidSortFields(), "updated_at")
	page, perPage := models.ParsePaginationParams(c)

	versions, err := serviceVersionModel.All(c.Request.Context(), serviceID, orgID, q, sortBy, sort, page, perPage)
	if err != nil {
		models.AbortWithError(c, http.StatusInternalServerError, "Could not get service versions")
		return
	}

	c.JSON(http.StatusOK, versions)
}

// GetServiceVersion gets a specific service version
// @Summary Get a version of a service
// @Schemes
// @Description Get particular version by id for the specified service
// @Tags ServiceVersion
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID"
// @Param	serviceId	path	string	true	"Service ID"
// @Param	versionId	path	string	true	"Service Version ID"
// @Success 	 200  {object}  models.ServiceVersion
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object} models.ErrorResponse
// @Security BearerAuth
// @Router /orgs/{orgId}/services/{serviceId}/versions/{versionId} [GET]
func (ctrl ServiceVersionController) GetServiceVersion(c *gin.Context) {
	_, orgID, hasAccess := checkOrganizationAccess(c)
	if !hasAccess {
		return
	}

	serviceID := c.Param("serviceId")
	id := c.Param("versionId")

	version, isFound, err := serviceVersionModel.One(c.Request.Context(), serviceID, orgID, id)
	if err != nil {
		if !isFound {
			models.AbortWithError(c, http.StatusNotFound, "Service version not found")
			return
		}
		models.AbortWithError(c, http.StatusInternalServerError, "Could not get version")
		return
	}

	c.JSON(http.StatusOK, version)
}

// UpdateServiceVersion updates a service version
// @Summary Update a version for a service
// @Schemes
// @Description Updates the specified version of a service, version tag cannot be updated
// @Tags ServiceVersion
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID"
// @Param	serviceId	path	string	true	"Service ID"
// @Param	versionId	path	string	true	"Service Version ID"
// @Param serviceVersion body forms.UpdateServiceVersionForm true "ServiceVersion"
// @Success 	 200  {object}  models.ServiceVersion
// @Failure      400  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object} models.ErrorResponse
// @Security BearerAuth
// @Router /orgs/{orgId}/services/{serviceId}/versions/{versionId} [PATCH]
func (ctrl ServiceVersionController) UpdateServiceVersion(c *gin.Context) {
	_, orgID, hasAccess := checkOrganizationAccess(c)
	if !hasAccess {
		return
	}

	var form forms.UpdateServiceVersionForm
	if validationErr := c.ShouldBindJSON(&form); validationErr != nil {
		message := serviceVersionForm.Update(validationErr)
		models.AbortWithError(c, http.StatusBadRequest, message)
		return
	}

	// Validate that at least one field is provided
	if message := serviceVersionForm.ValidateUpdate(form); message != "" {
		models.AbortWithError(c, http.StatusBadRequest, message)
		return
	}

	serviceID := c.Param("serviceId")
	id := c.Param("versionId")

	_, isFound, err := serviceVersionModel.One(c.Request.Context(), serviceID, orgID, id)
	if err != nil {
		if !isFound {
			models.AbortWithError(c, http.StatusNotFound, "Service version not found")
			return
		}
		models.AbortWithError(c, http.StatusInternalServerError, "Could not get version")
		return
	}

	version, err := serviceVersionModel.Update(c.Request.Context(), serviceID, orgID, id, form)
	if err != nil {
		models.AbortWithError(c, http.StatusInternalServerError, "Service version could not be updated")
		return
	}
	c.JSON(http.StatusOK, version)
}

// DeleteServiceVersion deletes a service version
// @Summary Delete a version for a service
// @Schemes
// @Description Deletes the specified version of a service
// @Tags ServiceVersion
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID"
// @Param	serviceId	path	string	true	"Service ID"
// @Param	versionId	path	string	true	"Service Version ID"
// @Success 	 204  ""
// @Failure      403  {object}  models.ErrorResponse
// @Success 	 404  {object} models.ErrorResponse
// @Failure      500  {object} models.ErrorResponse
// @Security BearerAuth
// @Router /orgs/{orgId}/services/{serviceId}/versions/{versionId} [DELETE]
func (ctrl ServiceVersionController) DeleteServiceVersion(c *gin.Context) {
	_, orgID, hasAccess := checkOrganizationAccess(c)
	if !hasAccess {
		return
	}

	serviceID := c.Param("serviceId")
	id := c.Param("versionId")

	_, isFound, err := serviceVersionModel.One(c.Request.Context(), serviceID, orgID, id)
	if err != nil {
		if !isFound {
			models.AbortWithError(c, http.StatusNotFound, "Service version not found")
			return
		}
		models.AbortWithError(c, http.StatusInternalServerError, "Could not get version")
		return
	}

	err = serviceVersionModel.Delete(c.Request.Context(), id)
	if err != nil {
		models.AbortWithError(c, http.StatusInternalServerError, "Service version could not be deleted")
		return
	}

	c.JSON(http.StatusNoContent, "")
}
