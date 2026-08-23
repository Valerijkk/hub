// Package errors определяет структурированные ошибки приложения с HTTP-кодами и сообщениями.
package errors

import "fmt"

// AppError — структурированная ошибка приложения с HTTP-кодом, сообщением и ID запроса.
//
// Публичный JSON (NFR-13 / TZ): только {"error","request_id"}.
// HTTP-статус берётся из Code; code/message/details в теле не отдаются.
type AppError struct {
	Code       int    `json:"-"`
	Message    string `json:"error"`
	RequestID  string `json:"request_id,omitempty"`
	Details    any    `json:"-"`
	RetryAfter int    `json:"-"` // секунды для заголовка Retry-After (429)
}

// PublicErrorBody is the redacted API error response body.
type PublicErrorBody struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id"`
}

// Error возвращает строковое представление ошибки в формате "[код] сообщение".
func (e *AppError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// NewBadRequest создаёт ошибку 400 Bad Request.
func NewBadRequest(msg string) *AppError {
	return &AppError{Code: 400, Message: msg}
}

// NewUnauthorized создаёт ошибку 401 Unauthorized.
func NewUnauthorized(msg string) *AppError {
	return &AppError{Code: 401, Message: msg}
}

// NewForbidden создаёт ошибку 403 Forbidden.
func NewForbidden(msg string) *AppError {
	return &AppError{Code: 403, Message: msg}
}

// NewNotFound создаёт ошибку 404 Not Found.
func NewNotFound(msg string) *AppError {
	return &AppError{Code: 404, Message: msg}
}

// NewConflict создаёт ошибку 409 Conflict.
func NewConflict(msg string) *AppError {
	return &AppError{Code: 409, Message: msg}
}

// NewRateLimited создаёт ошибку 429 Too Many Requests.
func NewRateLimited(msg string) *AppError {
	return &AppError{Code: 429, Message: msg}
}

// NewRateLimitedAfter создаёт 429 с Retry-After (секунды).
func NewRateLimitedAfter(msg string, retryAfterSec int) *AppError {
	if retryAfterSec < 1 {
		retryAfterSec = 1
	}
	return &AppError{Code: 429, Message: msg, RetryAfter: retryAfterSec}
}

// NewInternal создаёт ошибку 500 Internal Server Error.
func NewInternal(msg string) *AppError {
	return &AppError{Code: 500, Message: msg}
}
