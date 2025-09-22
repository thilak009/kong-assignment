package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/thilak009/kong-assignment/forms"
	"github.com/thilak009/kong-assignment/models"
	"github.com/thilak009/kong-assignment/utils"
)

type UserController struct{}

var userModel = models.UserModel{}
var userForm = forms.UserForm{}

// Register creates a new user account
// @Summary Register a new user
// @Description Register a new user account. Password must be at least 8 characters and contain at least one uppercase letter, one lowercase letter, and one special character.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body forms.CreateUserForm true "User registration data"
// @Success 201 {object} models.User
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/register [post]
func (ctrl UserController) Register(c *gin.Context) {
	var form forms.CreateUserForm

	if err := c.ShouldBindJSON(&form); err != nil {
		message := userForm.Create(err)
		models.AbortWithError(c, http.StatusBadRequest, message)
		return
	}

	// Check if user already exists
	_, exists, err := userModel.FindByEmail(c.Request.Context(), form.Email)
	if err != nil {
		models.AbortWithError(c, http.StatusInternalServerError, "Failed to check user existence")
		return
	}
	if exists {
		// TODO: avoid username enumeration
		// ideally there should be a email verification flow so that all register calls
		// return something like check your email for link kind of response
		models.AbortWithError(c, http.StatusConflict, "User with this email already exists")
		return
	}

	// Create user
	user, err := userModel.Create(c.Request.Context(), form)
	if err != nil {
		models.AbortWithError(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	c.JSON(http.StatusCreated, user)
}

// Login authenticates a user and returns a JWT token
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body forms.LoginForm true "User login credentials"
// @Success 200 {object} map[string]interface{} "Contains user info and JWT token"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/login [post]
func (ctrl UserController) Login(c *gin.Context) {
	var form forms.LoginForm

	if err := c.ShouldBindJSON(&form); err != nil {
		message := userForm.Create(err)
		models.AbortWithError(c, http.StatusBadRequest, message)
		return
	}

	// Find user by email
	user, exists, err := userModel.FindByEmail(c.Request.Context(), form.Email)
	if err != nil {
		if !exists {
			models.AbortWithError(c, http.StatusUnauthorized, "Invalid email/password")
			return
		}
		models.AbortWithError(c, http.StatusInternalServerError, "Failed to find user")
		return
	}

	// Check password
	if !user.CheckPassword(form.Password) {
		models.AbortWithError(c, http.StatusUnauthorized, "Invalid email/password")
		return
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Email)
	if err != nil {
		models.AbortWithError(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	c.JSON(http.StatusOK, models.TokenResponse{
		AccessToken: token,
	})
}

// Logout invalidates the JWT token by adding it to blacklist
// @Summary Logout user
// @Description Invalidate user JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Success 204 ""
// @Failure 401 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /users/logout [post]
func (ctrl UserController) Logout(c *gin.Context) {
	// Extract token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		models.AbortWithError(c, http.StatusUnauthorized, "Authorization header required")
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Get token claims to extract user ID and expiration
	claims, err := utils.GetTokenClaims(tokenString)
	if err != nil {
		models.AbortWithError(c, http.StatusUnauthorized, "Invalid token")
		return
	}

	// Add token to blacklist
	blacklistModel := models.BlacklistedTokenModel{}
	tokenHash := utils.HashToken(tokenString)

	if err := blacklistModel.Create(c.Request.Context(), tokenHash, claims.UserID, claims.ExpiresAt.Time); err != nil {
		models.AbortWithError(c, http.StatusInternalServerError, "Failed to logout")
		return
	}

	c.Status(http.StatusNoContent)
}
