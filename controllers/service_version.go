package controllers

import (
	"github.com/thilak009/kong-assignment/forms"
	"github.com/thilak009/kong-assignment/models"

	"net/http"

	"github.com/gin-gonic/gin"
)

type ServiceVersionController struct{}

var serviceVersionModel = new(models.ServiceVersionModel)
var serviceVersionForm = new(forms.ServiceVersionForm)

// Create ServiceVersion godoc
// @Summary Create a version for a service
// @Schemes
// @Description Creates a version for the specified service
// @Description version value must be a semantic version and releaseTimestamp must be valid RFC3339 timestamp
// @Tags ServiceVersion
// @Accept json
// @Produce json
// @Param	serviceId	path	string	true	"Service ID"
// @Param serviceVersion body forms.CreateServiceVersionForm true "ServiceVersion"
// @Success 	 200  {object}  models.ServiceVersion
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object} models.ErrorResponse
// @Router /services/{serviceId}/versions [post]
func (ctrl ServiceVersionController) Create(c *gin.Context) {
	var form forms.CreateServiceVersionForm
	if validationErr := c.ShouldBindJSON(&form); validationErr != nil {
		message := serviceVersionForm.Create(validationErr)
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponse{
			Message: message,
		})
		return
	}

	serviceID := c.Param("id")
	_, isFound, err := serviceModel.One(serviceID)
	if err != nil {
		if !isFound {
			c.AbortWithStatusJSON(http.StatusNotFound, models.ErrorResponse{
				Message: "Service not found",
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Could not get versions",
		})
		return
	}

	// TODO: handle same version tag creation by returning a bad request maybe
	version, err := serviceVersionModel.Create(serviceID, form)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Service version could not be created",
		})
		return
	}

	c.JSON(http.StatusOK, version)
}

// Get All ServiceVersions godoc
// @Summary Get All versions of a service
// @Schemes
// @Description Gets all the versions available for the specified service
// @Tags ServiceVersion
// @Accept json
// @Produce json
// @Param	q	query   string	false	"version, supports searching with version prefix, for example: passing 1 would return versions like 1.0.1,1.1.4 etc, passing 1.0 would return 1.0.3,1.0.7 etc"
// @Param	sort	query   string	false	"Sort order for the list of service versions. Accepted values are asc and desc. Default is desc(assumes default on invalid values as well)"
// @Param	sort_by	query   string	false	"The field on which sorting to be applied, supports version, created_at, updated_at. Default is updated_at(assumes default on invalid values as well)"
// @Param	page	query   int	false	"Page number for pagination (0-based). Default is 0"
// @Param	per_page	query   int	false	"Number of items per page. Default is 10, max is 100, assumes 100 if >100 is passed"
// @Param	serviceId	path	string	true	"Service ID"
// @Success 	 200  {object}  models.PaginatedResult[models.ServiceVersion]
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object} models.ErrorResponse
// @Router /services/{serviceId}/versions [GET]
func (ctrl ServiceVersionController) All(c *gin.Context) {
	serviceID := c.Param("id")

	_, isFound, err := serviceModel.One(serviceID)
	if err != nil {
		if !isFound {
			c.AbortWithStatusJSON(http.StatusNotFound, models.ErrorResponse{
				Message: "Service not found",
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Could not get versions",
		})
		return
	}
	q := c.Query("q")
	sortBy, sort := models.ParseSortParams(c, models.GetServiceVersionValidSortFields(), "updated_at")
	page, perPage := models.ParsePaginationParams(c)

	versions, err := serviceVersionModel.All(serviceID, q, sortBy, sort, page, perPage)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Could not get service versions",
		})
		return
	}

	c.JSON(http.StatusOK, versions)
}

// Get One ServiceVersion godoc
// @Summary Get a version of a service
// @Schemes
// @Description Get particular version by id for the specified service
// @Tags ServiceVersion
// @Accept json
// @Produce json
// @Param	serviceId	path	string	true	"Service ID"
// @Param	id	path	string	true	"Service Version ID"
// @Success 	 200  {object}  models.ServiceVersion
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object} models.ErrorResponse
// @Router /services/{serviceId}/versions/{id} [GET]
func (ctrl ServiceVersionController) One(c *gin.Context) {
	serviceID := c.Param("id")
	id := c.Param("versionId")

	version, isFound, err := serviceVersionModel.One(serviceID, id)
	if err != nil {
		if !isFound {
			c.AbortWithStatusJSON(http.StatusNotFound, models.ErrorResponse{
				Message: "Service version not found",
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Could not get version",
		})
		return
	}

	c.JSON(http.StatusOK, version)
}

// Update ServiceVersion godoc
// @Summary Update a version for a service
// @Schemes
// @Description Updates the specified version of a service, version tag cannot be updated
// @Tags ServiceVersion
// @Accept json
// @Produce json
// @Param	serviceId	path	string	true	"Service ID"
// @Param	id	path	string	true	"Service Version ID"
// @Param serviceVersion body forms.UpdateServiceVersionForm true "ServiceVersion"
// @Success 	 200  {object}  models.ServiceVersion
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object} models.ErrorResponse
// @Router /services/{serviceId}/versions/{id} [PATCH]
func (ctrl ServiceVersionController) Update(c *gin.Context) {
	var form forms.UpdateServiceVersionForm
	if validationErr := c.ShouldBindJSON(&form); validationErr != nil {
		message := serviceVersionForm.Update(validationErr)
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponse{
			Message: message,
		})
		return
	}

	serviceID := c.Param("id")
	id := c.Param("versionId")

	_, isFound, err := serviceVersionModel.One(serviceID, id)
	if err != nil {
		if !isFound {
			c.AbortWithStatusJSON(http.StatusNotFound, models.ErrorResponse{
				Message: "Service version not found",
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Could not get version",
		})
		return
	}

	version, err := serviceVersionModel.Update(serviceID, id, form)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Service version could not be updated",
		})
		return
	}
	c.JSON(http.StatusOK, version)
}

// Delete ServiceVersion godoc
// @Summary Delete a version for a service
// @Schemes
// @Description Deletes the specified version of a service
// @Tags ServiceVersion
// @Accept json
// @Produce json
// @Param	serviceId	path	string	true	"Service ID"
// @Param	id	path	string	true	"Service Version ID"
// @Success 	 204  ""
// @Success 	 404  {object} models.ErrorResponse
// @Failure      500  {object} models.ErrorResponse
// @Router /services/{serviceId}/versions/{id} [DELETE]
func (ctrl ServiceVersionController) Delete(c *gin.Context) {
	serviceID := c.Param("id")
	id := c.Param("versionId")

	_, isFound, err := serviceVersionModel.One(serviceID, id)
	if err != nil {
		if !isFound {
			c.AbortWithStatusJSON(http.StatusNotFound, models.ErrorResponse{
				Message: "Service version not found",
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Could not get version",
		})
		return
	}

	err = serviceVersionModel.Delete(serviceID, id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Service version could not be deleted",
		})
		return
	}

	c.JSON(http.StatusNoContent, "")
}
