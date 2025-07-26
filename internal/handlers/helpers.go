package handlers

import (
	"strconv"

	"github.com/caner-cetin/hospital-tracker/internal/errors"
	"github.com/gin-gonic/gin"
)

func parseUintParam(c *gin.Context, paramName, errorMessage string) (uint, bool) {
	idParam := c.Param(paramName)
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		errors.RespondWithValidationError(c, paramName, errorMessage)
		return 0, false
	}
	return uint(id), true
}