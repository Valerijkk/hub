package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppError_Error(t *testing.T) {
	e := &AppError{Code: 400, Message: "bad request"}
	assert.Equal(t, "[400] bad request", e.Error())
}

func TestAppError_ImplementsError(t *testing.T) {
	var err error = &AppError{Code: 500, Message: "fail"}
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestAppError_Fields(t *testing.T) {
	e := &AppError{
		Code:      422,
		Message:   "validation",
		RequestID: "req-123",
		Details:   map[string]string{"field": "name"},
	}
	assert.Equal(t, 422, e.Code)
	assert.Equal(t, "validation", e.Message)
	assert.Equal(t, "req-123", e.RequestID)
	assert.NotNil(t, e.Details)
}

func TestAppError_OmitEmptyFields(t *testing.T) {
	e := &AppError{Code: 404, Message: "not found"}
	assert.Empty(t, e.RequestID)
	assert.Nil(t, e.Details)
}

func TestNewBadRequest(t *testing.T) {
	e := NewBadRequest("invalid input")
	assert.Equal(t, 400, e.Code)
	assert.Equal(t, "invalid input", e.Message)
	assert.Equal(t, "[400] invalid input", e.Error())
}

func TestNewUnauthorized(t *testing.T) {
	e := NewUnauthorized("no token")
	assert.Equal(t, 401, e.Code)
	assert.Equal(t, "no token", e.Message)
	assert.Equal(t, "[401] no token", e.Error())
}

func TestNewForbidden(t *testing.T) {
	e := NewForbidden("access denied")
	assert.Equal(t, 403, e.Code)
	assert.Equal(t, "access denied", e.Message)
	assert.Equal(t, "[403] access denied", e.Error())
}

func TestNewNotFound(t *testing.T) {
	e := NewNotFound("resource missing")
	assert.Equal(t, 404, e.Code)
	assert.Equal(t, "resource missing", e.Message)
	assert.Equal(t, "[404] resource missing", e.Error())
}

func TestNewConflict(t *testing.T) {
	e := NewConflict("duplicate entry")
	assert.Equal(t, 409, e.Code)
	assert.Equal(t, "duplicate entry", e.Message)
	assert.Equal(t, "[409] duplicate entry", e.Error())
}

func TestNewRateLimited(t *testing.T) {
	e := NewRateLimited("slow down")
	assert.Equal(t, 429, e.Code)
	assert.Equal(t, "slow down", e.Message)
	assert.Equal(t, "[429] slow down", e.Error())
}

func TestNewInternal(t *testing.T) {
	e := NewInternal("server error")
	assert.Equal(t, 500, e.Code)
	assert.Equal(t, "server error", e.Message)
	assert.Equal(t, "[500] server error", e.Error())
}

func TestConstructors_Table(t *testing.T) {
	tests := []struct {
		name    string
		fn      func(string) *AppError
		code    int
		message string
	}{
		{"BadRequest", NewBadRequest, 400, MsgBadRequest},
		{"Unauthorized", NewUnauthorized, 401, MsgUnauthorized},
		{"Forbidden", NewForbidden, 403, MsgForbidden},
		{"NotFound", NewNotFound, 404, MsgNotFound},
		{"Conflict", NewConflict, 409, MsgConflict},
		{"RateLimited", NewRateLimited, 429, "rate limited"},
		{"Internal", NewInternal, 500, MsgInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := tt.fn(tt.message)
			require.NotNil(t, e)
			assert.Equal(t, tt.code, e.Code)
			assert.Equal(t, tt.message, e.Message)
			assert.Empty(t, e.RequestID, "RequestID should be empty by default")
			assert.Nil(t, e.Details, "Details should be nil by default")
		})
	}
}

func TestMessages_Constants(t *testing.T) {
	assert.NotEmpty(t, MsgBadRequest)
	assert.NotEmpty(t, MsgUnauthorized)
	assert.NotEmpty(t, MsgForbidden)
	assert.NotEmpty(t, MsgNotFound)
	assert.NotEmpty(t, MsgInternal)
	assert.NotEmpty(t, MsgConflict)
}

func TestNewBadRequest_WithStandardMessage(t *testing.T) {
	e := NewBadRequest(MsgBadRequest)
	assert.Equal(t, 400, e.Code)
	assert.Equal(t, MsgBadRequest, e.Message)
}

func TestNewNotFound_WithStandardMessage(t *testing.T) {
	e := NewNotFound(MsgNotFound)
	assert.Equal(t, 404, e.Code)
	assert.Equal(t, MsgNotFound, e.Message)
}

func TestNewInternal_WithStandardMessage(t *testing.T) {
	e := NewInternal(MsgInternal)
	assert.Equal(t, 500, e.Code)
	assert.Equal(t, MsgInternal, e.Message)
}

func TestAppError_EmptyMessage(t *testing.T) {
	e := NewBadRequest("")
	assert.Equal(t, 400, e.Code)
	assert.Equal(t, "", e.Message)
	assert.Equal(t, "[400] ", e.Error())
}
