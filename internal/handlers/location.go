package handlers

import (
	"net/http"

	"github.com/caner-cetin/hospital-tracker/internal/models"
	"github.com/caner-cetin/hospital-tracker/internal/services"
	"github.com/gin-gonic/gin"
)

type LocationHandler struct {
	locationService *services.LocationService
}

func NewLocationHandler(locationService *services.LocationService) *LocationHandler {
	return &LocationHandler{
		locationService: locationService,
	}
}

// GetProvinces godoc
// @Summary Get all provinces
// @Description Get all provinces in the country
// @Tags Reference Data
// @Produce json
// @Success 200 {array} models.Province "List of provinces"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /provinces [get]
func (h *LocationHandler) GetProvinces(c *gin.Context) {
	provinces, err := h.locationService.GetProvinces()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "failed to fetch provinces",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"provinces": provinces,
	})
}

// GetDistricts godoc
// @Summary Get districts
// @Description Get all districts or districts filtered by province
// @Tags Reference Data
// @Produce json
// @Param province_id query string false "Province ID to filter districts"
// @Success 200 {array} models.District "List of districts"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /districts [get]
func (h *LocationHandler) GetDistricts(c *gin.Context) {
	provinceID := c.Query("province_id")
	if provinceID == "" {
		districts, err := h.locationService.GetAllDistricts()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "failed to fetch districts",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"districts": districts,
		})
		return
	}

	districts, err := h.locationService.GetDistrictsByProvince(provinceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "failed to fetch districts",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"districts": districts,
	})
}
