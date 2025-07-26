package errors

import (
	"errors"
	"fmt"
	"net/http"
)

type ErrorCode string

const (
	ErrCodeValidation    ErrorCode = "VALIDATION_ERROR"
	ErrCodeInvalidInput  ErrorCode = "INVALID_INPUT"
	ErrCodeMissingField  ErrorCode = "MISSING_FIELD"
	ErrCodeInvalidFormat ErrorCode = "INVALID_FORMAT"

	ErrCodeNotFound         ErrorCode = "NOT_FOUND"
	ErrCodeUserNotFound     ErrorCode = "USER_NOT_FOUND"
	ErrCodeStaffNotFound    ErrorCode = "STAFF_NOT_FOUND"
	ErrCodeClinicNotFound   ErrorCode = "CLINIC_NOT_FOUND"
	ErrCodeHospitalNotFound ErrorCode = "HOSPITAL_NOT_FOUND"

	ErrCodeConflict            ErrorCode = "CONFLICT"
	ErrCodeDuplicateEmail      ErrorCode = "DUPLICATE_EMAIL"
	ErrCodeDuplicatePhone      ErrorCode = "DUPLICATE_PHONE"
	ErrCodeDuplicateNationalID ErrorCode = "DUPLICATE_NATIONAL_ID"
	ErrCodeDuplicateTaxID      ErrorCode = "DUPLICATE_TAX_ID"

	ErrCodeBusinessRule       ErrorCode = "BUSINESS_RULE_VIOLATION"
	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS" //nolint:gosec // this is just a constant identifier, not actual credentials
	ErrCodeExpiredToken       ErrorCode = "EXPIRED_TOKEN"
	ErrCodeInvalidToken       ErrorCode = "INVALID_TOKEN"

	ErrCodeInternal        ErrorCode = "INTERNAL_ERROR"
	ErrCodeDatabase        ErrorCode = "DATABASE_ERROR"
	ErrCodeExternalService ErrorCode = "EXTERNAL_SERVICE_ERROR"
)

type AppError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	StatusCode int                    `json:"-"`
	Err        error                  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewValidationError(field, message string) *AppError {
	return &AppError{
		Code:       ErrCodeValidation,
		Message:    "Validation failed",
		Details:    message,
		StatusCode: http.StatusBadRequest,
		Context: map[string]interface{}{
			"field": field,
		},
	}
}

func NewNotFoundError(resource string, identifier interface{}) *AppError {
	return &AppError{
		Code:       ErrCodeNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		StatusCode: http.StatusNotFound,
		Context: map[string]interface{}{
			"resource":   resource,
			"identifier": identifier,
		},
	}
}

func NewUserNotFoundError() *AppError {
	return &AppError{
		Code:       ErrCodeUserNotFound,
		Message:    "User not found",
		StatusCode: http.StatusNotFound,
	}
}

func NewStaffNotFoundError() *AppError {
	return &AppError{
		Code:       ErrCodeStaffNotFound,
		Message:    "Staff member not found",
		StatusCode: http.StatusNotFound,
	}
}

func NewClinicNotFoundError() *AppError {
	return &AppError{
		Code:       ErrCodeClinicNotFound,
		Message:    "Clinic not found",
		StatusCode: http.StatusNotFound,
	}
}

func NewConflictError(field, value, resource string) *AppError {
	return &AppError{
		Code:       ErrCodeConflict,
		Message:    fmt.Sprintf("%s already exists", field),
		StatusCode: http.StatusConflict,
		Context: map[string]interface{}{
			"field":    field,
			"value":    value,
			"resource": resource,
		},
	}
}

func NewDuplicateEmailError(email string) *AppError {
	return &AppError{
		Code:       ErrCodeDuplicateEmail,
		Message:    "Email address already exists",
		StatusCode: http.StatusConflict,
		Context: map[string]interface{}{
			"email": email,
		},
	}
}

func NewDuplicatePhoneError(phone string) *AppError {
	return &AppError{
		Code:       ErrCodeDuplicatePhone,
		Message:    "Phone number already exists",
		StatusCode: http.StatusConflict,
		Context: map[string]interface{}{
			"phone": phone,
		},
	}
}

func NewDuplicateNationalIDError(nationalID string) *AppError {
	return &AppError{
		Code:       ErrCodeDuplicateNationalID,
		Message:    "National ID already exists",
		StatusCode: http.StatusConflict,
		Context: map[string]interface{}{
			"national_id": nationalID,
		},
	}
}

func NewDuplicateTaxIDError(taxID string) *AppError {
	return &AppError{
		Code:       ErrCodeDuplicateTaxID,
		Message:    "Tax ID already exists",
		StatusCode: http.StatusConflict,
		Context: map[string]interface{}{
			"tax_id": taxID,
		},
	}
}

func NewBusinessRuleError(rule string, context map[string]interface{}) *AppError {
	return &AppError{
		Code:       ErrCodeBusinessRule,
		Message:    "Business rule violation",
		Details:    rule,
		StatusCode: http.StatusBadRequest,
		Context:    context,
	}
}

func NewInvalidCredentialsError() *AppError {
	return &AppError{
		Code:       ErrCodeInvalidCredentials,
		Message:    "Invalid credentials",
		StatusCode: http.StatusUnauthorized,
	}
}

func NewExpiredTokenError() *AppError {
	return &AppError{
		Code:       ErrCodeExpiredToken,
		Message:    "Token has expired",
		StatusCode: http.StatusUnauthorized,
	}
}

func NewInvalidTokenError() *AppError {
	return &AppError{
		Code:       ErrCodeInvalidToken,
		Message:    "Invalid token",
		StatusCode: http.StatusUnauthorized,
	}
}

func NewDatabaseError(operation string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeDatabase,
		Message:    "Database operation failed",
		Details:    operation,
		StatusCode: http.StatusInternalServerError,
		Context: map[string]interface{}{
			"operation": operation,
		},
		Err: err,
	}
}

func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeInternal,
		Message:    "Internal server error",
		Details:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

func NewExternalServiceError(service string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeExternalService,
		Message:    fmt.Sprintf("%s service error", service),
		StatusCode: http.StatusServiceUnavailable,
		Context: map[string]interface{}{
			"service": service,
		},
		Err: err,
	}
}

func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

func GetStatusCode(err error) int {
	if appErr, ok := IsAppError(err); ok {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}
