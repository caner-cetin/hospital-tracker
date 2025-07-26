package handlers

import (
	"net/http"
	"strconv"

	"github.com/caner-cetin/hospital-tracker/internal/models"
	"github.com/caner-cetin/hospital-tracker/internal/services"
	"github.com/gin-gonic/gin"
)

type StaffHandler struct {
	staffService *services.StaffService
}

func NewStaffHandler(staffService *services.StaffService) *StaffHandler {
	return &StaffHandler{
		staffService: staffService,
	}
}


// CreateStaff godoc
// @Summary Create a new staff member
// @Description Create a new staff member in the hospital (requires authorization)
// @Tags Staff
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.CreateStaffRequest true "Staff creation data"
// @Success 201 {object} models.Staff "Staff created successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Forbidden"
// @Router /staff [post]
func (h *StaffHandler) CreateStaff(c *gin.Context) {
	var req models.CreateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation error",
			Message: err.Error(),
		})
		return
	}

	hospitalID := c.GetUint("hospital_id")

	staff, err := h.staffService.CreateStaff(&req, hospitalID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "staff creation failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"staff":   staff,
		"message": "Staff created successfully",
	})
}

// UpdateStaff godoc
// @Summary Update a staff member
// @Description Update staff member information (requires authorization)
// @Tags Staff
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Staff ID"
// @Param request body models.UpdateStaffRequest true "Staff update data"
// @Success 200 {object} models.Staff "Staff updated successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Forbidden"
// @Failure 404 {object} models.ErrorResponse "Staff not found"
// @Router /staff/{id} [put]
func (h *StaffHandler) UpdateStaff(c *gin.Context) {
	staffID, ok := parseUintParam(c, "id", "invalid staff ID")
	if !ok {
		return
	}

	var req models.UpdateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation error",
			Message: err.Error(),
		})
		return
	}

	hospitalID := c.GetUint("hospital_id")

	staff, err := h.staffService.UpdateStaff(staffID, &req, hospitalID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "staff update failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"staff":   staff,
		"message": "Staff updated successfully",
	})
}

// DeleteStaff godoc
// @Summary Delete a staff member
// @Description Delete a staff member from the hospital (requires authorization)
// @Tags Staff
// @Produce json
// @Security Bearer
// @Param id path int true "Staff ID"
// @Success 200 {object} map[string]string "Staff deleted successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Forbidden"
// @Failure 404 {object} models.ErrorResponse "Staff not found"
// @Router /staff/{id} [delete]
func (h *StaffHandler) DeleteStaff(c *gin.Context) {
	staffIDParam := c.Param("id")
	staffID, err := strconv.ParseUint(staffIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid staff ID",
		})
		return
	}

	hospitalID := c.GetUint("hospital_id")

	err = h.staffService.DeleteStaff(uint(staffID), hospitalID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "staff deletion failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Staff deleted successfully",
	})
}

// GetStaff godoc
// @Summary Get staff members with filtering
// @Description Get staff members with optional filtering by clinic, profession group, etc.
// @Tags Staff
// @Produce json
// @Security Bearer
// @Param clinic_id query int false "Filter by clinic ID"
// @Param profession_group_id query int false "Filter by profession group ID"
// @Param page query int false "Page number for pagination"
// @Param limit query int false "Number of items per page"
// @Success 200 {object} models.StaffPaginatedResponse "List of staff members"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /staff [get]
func (h *StaffHandler) GetStaff(c *gin.Context) {
	var filter models.StaffFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation error",
			Message: err.Error(),
		})
		return
	}

	hospitalID := c.GetUint("hospital_id")

	result, err := h.staffService.GetStaff(&filter, hospitalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "failed to fetch staff",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetStaffByID godoc
// @Summary Get a staff member by ID
// @Description Get detailed information about a specific staff member
// @Tags Staff
// @Produce json
// @Security Bearer
// @Param id path int true "Staff ID"
// @Success 200 {object} models.Staff "Staff member information"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 404 {object} models.ErrorResponse "Staff not found"
// @Router /staff/{id} [get]
func (h *StaffHandler) GetStaffByID(c *gin.Context) {
	staffIDParam := c.Param("id")
	staffID, err := strconv.ParseUint(staffIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid staff ID",
		})
		return
	}

	hospitalID := c.GetUint("hospital_id")

	staff, err := h.staffService.GetStaffByID(uint(staffID), hospitalID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "staff not found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"staff": staff,
	})
}

// GetProfessionGroups godoc
// @Summary Get all profession groups
// @Description Get all available profession groups for staff
// @Tags Reference Data
// @Produce json
// @Success 200 {object} map[string][]models.ProfessionGroupResponse "List of profession groups"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /profession-groups [get]
func (h *StaffHandler) GetProfessionGroups(c *gin.Context) {
	professionGroups, err := h.staffService.GetProfessionGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "failed to fetch profession groups",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"profession_groups": professionGroups,
	})
}
