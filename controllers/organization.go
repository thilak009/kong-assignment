package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thilak009/kong-assignment/forms"
	"github.com/thilak009/kong-assignment/middleware"
	"github.com/thilak009/kong-assignment/models"
	"github.com/thilak009/kong-assignment/utils"
)

type OrganizationController struct{}

var organizationModel = models.OrganizationModel{}

// GetOrganizations returns all organizations the user belongs to
// @Summary Get user's organizations
// @Description Get all organizations that the authenticated user belongs to
// @Tags Organizations
// @Accept json
// @Produce json
// @Param q query string false "Search query"
// @Param sort_by query string false "Sort field" Enums(name, created_at, updated_at)
// @Param sort query string false "Sort direction" Enums(asc, desc)
// @Param page query int false "Page number" default(0)
// @Param per_page query int false "Items per page" default(10)
// @Success 200 {object} models.PaginatedResult[models.Organization]
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /orgs [get]
func (ctrl OrganizationController) GetOrganizations(c *gin.Context) {
	userID := middleware.GetUserID(c)

	q := c.Query("q")
	sortBy, sort := models.ParseSortParams(c, models.GetOrganizationValidSortFields(), "updated_at")
	page, perPage := models.ParsePaginationParams(c)

	result, err := organizationModel.GetUserOrganizations(userID, q, sortBy, sort, page, perPage)
	if err != nil {
		utils.AbortWithError(c, http.StatusInternalServerError, "Failed to fetch organizations")
		return
	}

	c.JSON(http.StatusOK, result)
}

// CreateOrganization creates a new organization
// @Summary Create a new organization
// @Description Create a new organization for the authenticated user
// @Tags Organizations
// @Accept json
// @Produce json
// @Param organization body forms.CreateOrganizationForm true "Organization data"
// @Success 201 {object} models.Organization
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /orgs [post]
func (ctrl OrganizationController) CreateOrganization(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var form forms.CreateOrganizationForm

	if err := c.ShouldBindJSON(&form); err != nil {
		utils.AbortWithError(c, http.StatusBadRequest, "Invalid request data")
		return
	}

	organization, err := organizationModel.Create(form, userID)
	if err != nil {
		utils.AbortWithError(c, http.StatusInternalServerError, "Failed to create organization")
		return
	}

	c.JSON(http.StatusCreated, organization)
}

// GetOrganization returns a specific organization
// @Summary Get organization by ID
// @Description Get a specific organization by ID
// @Tags Organizations
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID"
// @Success 200 {object} models.Organization
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /orgs/{orgId} [get]
func (ctrl OrganizationController) GetOrganization(c *gin.Context) {
	userID := middleware.GetUserID(c)
	orgID := c.Param("orgId")

	// Check if user is member of organization
	isMember, err := organizationModel.IsUserMember(orgID, userID)
	if err != nil {
		utils.AbortWithError(c, http.StatusInternalServerError, "Failed to check organization access")
		return
	}

	if !isMember {
		utils.AbortWithError(c, http.StatusForbidden, "You are not authorized to perform the request")
		return
	}

	organization, exists, err := organizationModel.One(orgID)
	if err != nil {
		utils.AbortWithError(c, http.StatusInternalServerError, "Failed to fetch organization")
		return
	}

	if !exists {
		utils.AbortWithError(c, http.StatusNotFound, "Organization not found")
		return
	}

	c.JSON(http.StatusOK, organization)
}

// UpdateOrganization updates an organization
// @Summary Update organization
// @Description Update an existing organization
// @Tags Organizations
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID"
// @Param organization body forms.CreateOrganizationForm true "Organization update data"
// @Success 200 {object} models.Organization
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /orgs/{orgId} [put]
func (ctrl OrganizationController) UpdateOrganization(c *gin.Context) {
	userID := middleware.GetUserID(c)
	orgID := c.Param("orgId")
	var form forms.CreateOrganizationForm

	if err := c.ShouldBindJSON(&form); err != nil {
		utils.AbortWithError(c, http.StatusBadRequest, "Invalid request data")
		return
	}

	// Check if user is member of organization
	isMember, err := organizationModel.IsUserMember(orgID, userID)
	if err != nil {
		utils.AbortWithError(c, http.StatusInternalServerError, "Failed to check organization access")
		return
	}

	if !isMember {
		utils.AbortWithError(c, http.StatusForbidden, "You are not authorized to perform the request")
		return
	}

	organization, err := organizationModel.Update(orgID, form)
	if err != nil {
		utils.AbortWithError(c, http.StatusInternalServerError, "Failed to update organization")
		return
	}

	c.JSON(http.StatusOK, organization)
}

// DeleteOrganization deletes an organization
// @Summary Delete organization
// @Description Delete an organization
// @Tags Organizations
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID"
// @Success 204 "No Content"
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /orgs/{orgId} [delete]
func (ctrl OrganizationController) DeleteOrganization(c *gin.Context) {
	userID := middleware.GetUserID(c)
	orgID := c.Param("orgId")

	// Check if user is member of organization
	isMember, err := organizationModel.IsUserMember(orgID, userID)
	if err != nil {
		utils.AbortWithError(c, http.StatusInternalServerError, "Failed to check organization access")
		return
	}

	if !isMember {
		utils.AbortWithError(c, http.StatusForbidden, "You are not authorized to perform the request")
		return
	}

	err = organizationModel.Delete(orgID)
	if err != nil {
		utils.AbortWithError(c, http.StatusInternalServerError, "Failed to delete organization")
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
