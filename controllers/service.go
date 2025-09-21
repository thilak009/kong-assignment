package controllers

import (
	"github.com/thilak009/kong-assignment/forms"
	"github.com/thilak009/kong-assignment/models"

	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ServiceController struct{}

var serviceModel = new(models.ServiceModel)
var serviceForm = new(forms.ServiceForm)

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

// Create Service godoc
// @Summary Create a service
// @Schemes
// @Description Creates a service
// @Tags Service
// @Accept json
// @Produce json
// @Param service body forms.CreateServiceForm true "Service"
// @Success 	 200  {object}  models.Service
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object}	models.ErrorResponse
// @Router /services [post]
func (ctrl ServiceController) Create(c *gin.Context) {
	var form forms.CreateServiceForm
	if validationErr := c.ShouldBindJSON(&form); validationErr != nil {
		message := serviceForm.Create(validationErr)
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponse{
			Message: message,
		})
		return
	}

	service, err := serviceModel.Create(form)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Service could not be created",
		})
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
// @Param	q	query   string	false	"Service name, supports searching the passed string in the name of the service"
// @Param	sort	query   string	false	"Sort order for the list of services. Accepted values are asc and desc. Default is desc(assumes default on invalid values as well)"
// @Param	sort_by	query   string	false	"The field on which sorting to be applied, supports name, created_at, updated_at. Default is updated_at(assumes default on invalid values as well)"
// @Param	page	query   int	false	"Page number for pagination (0-based). Default is 0"
// @Param	per_page	query   int	false	"Number of items per page. Default is 10, max is 100, assumes 100 if >100 is passed"
// @Param	include	query   string	false	"Additional data to include (comma-separated). Supported values: versionCount"
// @Success 	 200  {object}  models.PaginatedResult[models.Service]
// @Failure      500  {object}	models.ErrorResponse
// @Router /services [GET]
func (ctrl ServiceController) All(c *gin.Context) {
	q := c.Query("q")
	sortBy, sort := models.ParseSortParams(c, models.GetServiceValidSortFields(), "updated_at")
	page, perPage := models.ParsePaginationParams(c)

	// Parse include parameter for multiple values
	include := c.Query("include")
	includeVersionCount := parseIncludeParams(include)

	results, err := serviceModel.All(q, sortBy, sort, page, perPage, includeVersionCount)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Could not get services",
		})
		return
	}

	c.JSON(http.StatusOK, results)
}

// Get One Service godoc
// @Summary Get a service
// @Schemes
// @Description Gets the specified service
// @Tags Service
// @Accept json
// @Produce json
// @Param	id	path	string	true	"Service ID"
// @Param	include	query   string	false	"Additional data to include (comma-separated). Supported values: versionCount"
// @Success 	 200  {object}  models.Service
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router /services/{id} [GET]
func (ctrl ServiceController) One(c *gin.Context) {
	id := c.Param("id")

	// Parse include parameter for multiple values
	include := c.Query("include")
	includeVersionCount := parseIncludeParams(include)

	data, isFound, err := serviceModel.One(id, includeVersionCount)
	if err != nil {
		if !isFound {
			c.AbortWithStatusJSON(http.StatusNotFound, models.ErrorResponse{
				Message: "Service not found",
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Could not get service",
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// Update Service godoc
// @Summary Update a service
// @Schemes
// @Description Updates the specified service
// @Tags Service
// @Accept json
// @Produce json
// @Param	id	path	string	true	"Service ID"
// @Param service body forms.CreateServiceForm true "Service"
// @Success 	 200  {object}  models.Service
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router /services/{id} [PUT]
func (ctrl ServiceController) Update(c *gin.Context) {
	var form forms.CreateServiceForm
	if validationErr := c.ShouldBindJSON(&form); validationErr != nil {
		message := serviceForm.Create(validationErr)
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponse{
			Message: message,
		})
		return
	}

	id := c.Param("id")
	_, isFound, err := serviceModel.One(id, false)
	if err != nil {
		if !isFound {
			c.AbortWithStatusJSON(http.StatusNotFound, models.ErrorResponse{
				Message: "Service not found",
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Could not get service",
		})
		return
	}

	service, err := serviceModel.Update(id, form)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Service could not be updated",
		})
		return
	}
	c.JSON(http.StatusOK, service)
}

// Delete Service godoc
// @Summary Delete a service
// @Schemes
// @Description Deletes the specified service
// @Tags Service
// @Accept json
// @Produce json
// @Param	id	path	string	true	"Service ID"
// @Success 	 204  ""
// @Failure 	 404  {object} models.ErrorResponse
// @Failure      500  {object} models.ErrorResponse
// @Router /services/{id} [DELETE]
func (ctrl ServiceController) Delete(c *gin.Context) {
	id := c.Param("id")
	_, isFound, err := serviceModel.One(id, false)
	if err != nil {
		if !isFound {
			c.AbortWithStatusJSON(http.StatusNotFound, models.ErrorResponse{
				Message: "Service not found",
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Could not get service",
		})
		return
	}

	err = serviceModel.Delete(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Service could not be deleted",
		})
		return
	}

	c.JSON(http.StatusNoContent, "")
}
