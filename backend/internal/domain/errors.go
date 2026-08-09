package domain

import "net/http"

type ErrorDetail struct {
	Field string `json:"field"`
	Issue string `json:"issue"`
}

type AppError struct {
	Status  int
	Code    string
	Message string
	Details []ErrorDetail
}

func (e *AppError) Error() string { return e.Message }

func NewError(status int, code, message string, details ...ErrorDetail) *AppError {
	return &AppError{Status: status, Code: code, Message: message, Details: details}
}

func BadRequest(message string) *AppError { return NewError(http.StatusBadRequest, "BAD_REQUEST", message) }
func Unauthorized(code, message string) *AppError { return NewError(http.StatusUnauthorized, code, message) }
func Forbidden(feature string) *AppError {
	return NewError(http.StatusForbidden, "INSUFFICIENT_PERMISSIONS", "No tiene acceso a este recurso", ErrorDetail{Field: "required_feature", Issue: feature})
}
func NotFound(code, message string) *AppError { return NewError(http.StatusNotFound, code, message) }
func Conflict(code, message string) *AppError { return NewError(http.StatusConflict, code, message) }
func Validation(message string) *AppError { return NewError(http.StatusUnprocessableEntity, "VALIDATION_ERROR", message) }