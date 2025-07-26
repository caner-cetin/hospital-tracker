package errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Code    ErrorCode              `json:"code,omitempty"`
	Details string                 `json:"details,omitempty"`
	Context map[string]interface{} `json:"context,omitempty"`
}

func HandleError(c *gin.Context, err error) {
	if appErr, ok := IsAppError(err); ok {
		HandleAppError(c, appErr)
		return
	}

	log.Error().Err(err).Msg("Unhandled error")
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error:   "Internal Server Error",
		Message: "An unexpected error occurred",
		Code:    ErrCodeInternal,
	})
}

func HandleAppError(c *gin.Context, appErr *AppError) {
	if appErr.StatusCode >= 500 {
		logger := log.Error().
			Str("code", string(appErr.Code)).
			Str("message", appErr.Message).
			Int("status_code", appErr.StatusCode)

		if appErr.Err != nil {
			logger = logger.Err(appErr.Err)
		}

		if appErr.Details != "" {
			logger = logger.Str("details", appErr.Details)
		}

		if len(appErr.Context) > 0 {
			logger = logger.Interface("context", appErr.Context)
		}

		logger.Msg("Internal error")
	}

	response := ErrorResponse{
		Error:   http.StatusText(appErr.StatusCode),
		Message: appErr.Message,
		Code:    appErr.Code,
		Details: appErr.Details,
		Context: appErr.Context,
	}

	c.JSON(appErr.StatusCode, response)
}

func AbortWithError(c *gin.Context, err error) {
	HandleError(c, err)
	c.Abort()
}

func AbortWithAppError(c *gin.Context, appErr *AppError) {
	HandleAppError(c, appErr)
	c.Abort()
}

func RespondWithValidationError(c *gin.Context, field, message string) {
	err := NewValidationError(field, message)
	HandleAppError(c, err)
}

func RespondWithNotFound(c *gin.Context, resource string, identifier interface{}) {
	err := NewNotFoundError(resource, identifier)
	HandleAppError(c, err)
}

func RespondWithConflict(c *gin.Context, field, value, resource string) {
	err := NewConflictError(field, value, resource)
	HandleAppError(c, err)
}

func RespondWithBusinessRuleError(c *gin.Context, rule string, context map[string]interface{}) {
	err := NewBusinessRuleError(rule, context)
	HandleAppError(c, err)
}

func RespondWithInvalidCredentials(c *gin.Context) {
	err := NewInvalidCredentialsError()
	HandleAppError(c, err)
}
