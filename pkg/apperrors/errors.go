package apperrors

import "net/http"

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(code int, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// Error Codes
const (
	ErrCodeEmailAlreadyExists = 1001
	ErrCodeUserNotFound       = 1002
)

func NewServiceDisabledError(message string) *AppError {
	return NewAppError(http.StatusBadRequest, message, http.StatusBadRequest)
}
