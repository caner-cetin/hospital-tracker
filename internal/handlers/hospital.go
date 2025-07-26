package handlers

import (
	"net/http"

	"github.com/caner-cetin/hospital-tracker/internal/errors"
	"github.com/caner-cetin/hospital-tracker/internal/models"
	"github.com/caner-cetin/hospital-tracker/internal/services"
	"github.com/gin-gonic/gin"
)

type HospitalHandler struct {
	hospitalService *services.HospitalService
}

func NewHospitalHandler(hospitalService *services.HospitalService) *HospitalHandler {
	return &HospitalHandler{
		hospitalService: hospitalService,
	}
}

// Register godoc
// @Summary Register a new hospital
// @Description Register a new hospital with authorized user
// @Tags Hospital
// @Accept json
// @Produce json
// @Param request body models.HospitalRegistrationRequest true "Hospital registration data"
// @Success 201 {object} map[string]interface{} "Hospital registered successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Router /register [post]
func (h *HospitalHandler) Register(c *gin.Context) {
	var req models.HospitalRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.RespondWithValidationError(c, "request", err.Error())
		return
	}

	hospital, user, err := h.hospitalService.RegisterHospital(&req)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"hospital": hospital,
		"user":     user,
		"message":  "Hospital registered successfully",
	})
}
