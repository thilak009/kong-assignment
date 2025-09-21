package controllers

import (
	"github.com/thilak009/kong-assignment/forms"
	"github.com/thilak009/kong-assignment/models"

	"net/http"

	"github.com/gin-gonic/gin"
)

type ServiceController struct{}

var serviceModel = new(models.ServiceModel)
var serviceForm = new(forms.ServiceForm)

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
// @Param	sort	query   string	false	"Sort order for the list of services. Accepted values are asc and desc. Default is desc"
// @Param	sort_by	query   string	false	"The field on which sorting to be applied, supports name, created_at, updated_at. Default is updated_at"
// @Success 	 200  {object}  []models.Service
// @Failure      500  {object}	models.ErrorResponse
// @Router /services [GET]
func (ctrl ServiceController) All(c *gin.Context) {
	q := c.Query("q")
	sortBy := c.DefaultQuery("sort_by", "updated_at")
	sort := c.DefaultQuery("sort", "desc")

	results, err := serviceModel.All(q, sortBy, sort)
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
// @Success 	 200  {object}  models.Service
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router /services/{id} [GET]
func (ctrl ServiceController) One(c *gin.Context) {
	id := c.Param("id")
	data, isFound, err := serviceModel.One(id)
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
	_, isFound, err := serviceModel.One(id)
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
	_, isFound, err := serviceModel.One(id)
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
