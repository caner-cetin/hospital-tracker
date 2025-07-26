package handlers

import (
	"net/http"
	"strconv"

	"github.com/caner-cetin/hospital-tracker/internal/models"
	"github.com/caner-cetin/hospital-tracker/internal/services"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user in the hospital (requires authorization)
// @Tags Users
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.CreateUserRequest true "User creation data"
// @Success 201 {object} models.User "User created successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Forbidden"
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation error",
			Message: err.Error(),
		})
		return
	}

	userID := c.GetUint("user_id")
	hospitalID := c.GetUint("hospital_id")

	user, err := h.userService.CreateUser(&req, userID, hospitalID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "user creation failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":    user,
		"message": "User created successfully",
	})
}

// UpdateUser godoc
// @Summary Update a user
// @Description Update user information (requires authorization)
// @Tags Users
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "User ID"
// @Param request body models.UpdateUserRequest true "User update data"
// @Success 200 {object} models.User "User updated successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Forbidden"
// @Failure 404 {object} models.ErrorResponse "User not found"
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userIDToUpdate, ok := parseUintParam(c, "id", "invalid user ID")
	if !ok {
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation error",
			Message: err.Error(),
		})
		return
	}

	hospitalID := c.GetUint("hospital_id")

	user, err := h.userService.UpdateUser(userIDToUpdate, &req, hospitalID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "user update failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":    user,
		"message": "User updated successfully",
	})
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Delete a user from the hospital (requires authorization)
// @Tags Users
// @Produce json
// @Security Bearer
// @Param id path int true "User ID"
// @Success 200 {object} map[string]string "User deleted successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 403 {object} models.ErrorResponse "Forbidden"
// @Failure 404 {object} models.ErrorResponse "User not found"
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userIDParam := c.Param("id")
	userIDToDelete, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid user ID",
		})
		return
	}

	hospitalID := c.GetUint("hospital_id")

	err = h.userService.DeleteUser(uint(userIDToDelete), hospitalID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "user deletion failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// GetUsers godoc
// @Summary Get all users
// @Description Get all users in the hospital
// @Tags Users
// @Produce json
// @Security Bearer
// @Success 200 {array} models.User "List of users"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
	hospitalID := c.GetUint("hospital_id")

	users, err := h.userService.GetUsers(hospitalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "failed to fetch users",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
	})
}

// GetUser godoc
// @Summary Get a user by ID
// @Description Get detailed information about a specific user
// @Tags Users
// @Produce json
// @Security Bearer
// @Param id path int true "User ID"
// @Success 200 {object} models.User "User information"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 404 {object} models.ErrorResponse "User not found"
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	userIDParam := c.Param("id")
	userIDToGet, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid user ID",
		})
		return
	}

	hospitalID := c.GetUint("hospital_id")

	user, err := h.userService.GetUser(uint(userIDToGet), hospitalID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "user not found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}
