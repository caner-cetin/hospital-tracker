package handlers

import (
	"net/http"

	"github.com/caner-cetin/hospital-tracker/internal/models"
	"github.com/caner-cetin/hospital-tracker/internal/services"
	"github.com/gin-gonic/gin"
)

type PasswordResetHandler struct {
	passwordResetService *services.PasswordResetService
}

func NewPasswordResetHandler(passwordResetService *services.PasswordResetService) *PasswordResetHandler {
	return &PasswordResetHandler{
		passwordResetService: passwordResetService,
	}
}

// RequestReset godoc
// @Summary Request password reset
// @Description Request a password reset code via phone number
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.PasswordResetRequest true "Phone number for password reset"
// @Success 200 {object} models.PasswordResetResponse "Reset code sent"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Router /password-reset/request [post]
func (h *PasswordResetHandler) RequestReset(c *gin.Context) {
	var req models.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation error",
			Message: err.Error(),
		})
		return
	}

	code, err := h.passwordResetService.RequestPasswordReset(req.Phone)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "reset request failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.PasswordResetResponse{
		Code: code,
	})
}

// ConfirmReset godoc
// @Summary Confirm password reset
// @Description Reset password using the verification code
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.PasswordResetConfirmRequest true "Password reset confirmation data"
// @Success 204 "password reset successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Router /password-reset/confirm [post]
func (h *PasswordResetHandler) ConfirmReset(c *gin.Context) {
	var req models.PasswordResetConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation error",
			Message: err.Error(),
		})
		return
	}

	err := h.passwordResetService.ResetPassword(req.Phone, req.Code, req.NewPassword, req.ConfirmPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "password reset failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
