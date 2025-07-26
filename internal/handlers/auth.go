package handlers

import (
	"net/http"

	"github.com/caner-cetin/hospital-tracker/internal/errors"
	"github.com/caner-cetin/hospital-tracker/internal/models"
	"github.com/caner-cetin/hospital-tracker/internal/services"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login godoc
// @Summary User login
// @Description Login with email/phone and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.LoginResponse "Login successful"
// @Failure 401 {object} models.ErrorResponse "Invalid credentials"
// @Router /login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErr := errors.NewValidationError("request", err.Error())
		errors.HandleAppError(c, validationErr)
		return
	}

	user, token, err := h.authService.Login(req.Identifier, req.Password)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	response := models.LoginResponse{
		Token:    token,
		UserType: string(user.UserType),
		User:     *user,
	}

	c.JSON(http.StatusOK, response)
}
