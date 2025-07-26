package handlers

import (
	"net/http"
	"strconv"

	"github.com/caner-cetin/hospital-tracker/internal/errors"
	"github.com/caner-cetin/hospital-tracker/internal/models"
	"github.com/caner-cetin/hospital-tracker/internal/services"
	"github.com/gin-gonic/gin"
)

type ClinicHandler struct {
	clinicService *services.ClinicService
}

func NewClinicHandler(clinicService *services.ClinicService) *ClinicHandler {
	return &ClinicHandler{
		clinicService: clinicService,
	}
}

// CreateClinic godoc
// @Summary Create a new clinic
// @Description Create a new clinic in the hospital (requires authorization)
// @Tags Clinics
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.CreateClinicRequest true "Clinic creation data"
// @Success 201 {object} models.Clinic "Clinic created successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Forbidden"
// @Router /clinics [post]
func (h *ClinicHandler) CreateClinic(c *gin.Context) {
	var req models.CreateClinicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.RespondWithValidationError(c, "request", err.Error())
		return
	}

	hospitalID := c.GetUint("hospital_id")

	clinic, err := h.clinicService.CreateClinic(&req, hospitalID)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"clinic":  clinic,
		"message": "Clinic created successfully",
	})
}

// GetClinics godoc
// @Summary Get all clinics
// @Description Get all clinics in the hospital
// @Tags Clinics
// @Produce json
// @Security Bearer
// @Success 200 {array} models.Clinic "List of clinics"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /clinics [get]
func (h *ClinicHandler) GetClinics(c *gin.Context) {
	hospitalID := c.GetUint("hospital_id")

	clinics, err := h.clinicService.GetClinics(hospitalID)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"clinics": clinics,
	})
}

// GetClinicTypes godoc
// @Summary Get all clinic types
// @Description Get all available clinic types for creating clinics
// @Tags Reference Data
// @Produce json
// @Success 200 {array} models.ClinicType "List of clinic types"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /clinic-types [get]
func (h *ClinicHandler) GetClinicTypes(c *gin.Context) {
	clinicTypes, err := h.clinicService.GetClinicTypes()
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"clinic_types": clinicTypes,
	})
}

// DeleteClinic godoc
// @Summary Delete a clinic
// @Description Delete a clinic from the hospital (requires authorization)
// @Tags Clinics
// @Produce json
// @Security Bearer
// @Param id path int true "Clinic ID"
// @Success 200 {object} map[string]string "Clinic deleted successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Forbidden"
// @Failure 404 {object} models.ErrorResponse "Clinic not found"
// @Router /clinics/{id} [delete]
func (h *ClinicHandler) DeleteClinic(c *gin.Context) {
	clinicIDParam := c.Param("id")
	clinicID, err := strconv.ParseUint(clinicIDParam, 10, 32)
	if err != nil {
		errors.RespondWithValidationError(c, "id", "invalid clinic ID")
		return
	}

	hospitalID := c.GetUint("hospital_id")

	err = h.clinicService.DeleteClinic(uint(clinicID), hospitalID)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Clinic deleted successfully",
	})
}
