package errors

import "fmt"

type AppError struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
	Details   any    `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func NewBadRequest(msg string) *AppError {
	return &AppError{Code: 400, Message: msg}
}

func NewUnauthorized(msg string) *AppError {
	return &AppError{Code: 401, Message: msg}
}

func NewForbidden(msg string) *AppError {
	return &AppError{Code: 403, Message: msg}
}

func NewNotFound(msg string) *AppError {
	return &AppError{Code: 404, Message: msg}
}

func NewConflict(msg string) *AppError {
	return &AppError{Code: 409, Message: msg}
}

func NewRateLimited(msg string) *AppError {
	return &AppError{Code: 429, Message: msg}
}

func NewInternal(msg string) *AppError {
	return &AppError{Code: 500, Message: msg}
}
