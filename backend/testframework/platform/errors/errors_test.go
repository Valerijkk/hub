package errors_test
import (
	"github.com/valerii/hub/backend/internal/platform/errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppError_Error(t *testing.T) {
	e := &errors.AppError{Code: 400, Message: "bad request"}
	assert.Equal(t, "[400] bad request", e.Error())
}

func TestAppError_ImplementsError(t *testing.T) {
	var err error = &errors.AppError{Code: 500, Message: "fail"}
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestAppError_Fields(t *testing.T) {
	e := &errors.AppError{
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
	e := &errors.AppError{Code: 404, Message: "not found"}
	assert.Empty(t, e.RequestID)
	assert.Nil(t, e.Details)
}

func TestNewBadRequest(t *testing.T) {
	e := errors.NewBadRequest("invalid input")
	assert.Equal(t, 400, e.Code)
	assert.Equal(t, "invalid input", e.Message)
	assert.Equal(t, "[400] invalid input", e.Error())
}

func TestNewUnauthorized(t *testing.T) {
	e := errors.NewUnauthorized("no token")
	assert.Equal(t, 401, e.Code)
	assert.Equal(t, "no token", e.Message)
	assert.Equal(t, "[401] no token", e.Error())
}

func TestNewForbidden(t *testing.T) {
	e := errors.NewForbidden("access denied")
	assert.Equal(t, 403, e.Code)
	assert.Equal(t, "access denied", e.Message)
	assert.Equal(t, "[403] access denied", e.Error())
}

func TestNewNotFound(t *testing.T) {
	e := errors.NewNotFound("resource missing")
	assert.Equal(t, 404, e.Code)
	assert.Equal(t, "resource missing", e.Message)
	assert.Equal(t, "[404] resource missing", e.Error())
}

func TestNewConflict(t *testing.T) {
	e := errors.NewConflict("duplicate entry")
	assert.Equal(t, 409, e.Code)
	assert.Equal(t, "duplicate entry", e.Message)
	assert.Equal(t, "[409] duplicate entry", e.Error())
}

func TestNewRateLimited(t *testing.T) {
	e := errors.NewRateLimited("slow down")
	assert.Equal(t, 429, e.Code)
	assert.Equal(t, "slow down", e.Message)
	assert.Equal(t, "[429] slow down", e.Error())
}

func TestNewInternal(t *testing.T) {
	e := errors.NewInternal("server error")
	assert.Equal(t, 500, e.Code)
	assert.Equal(t, "server error", e.Message)
	assert.Equal(t, "[500] server error", e.Error())
}

func TestConstructors_Table(t *testing.T) {
	tests := []struct {
		name    string
		fn      func(string) *errors.AppError
		code    int
		message string
	}{
		{"BadRequest", errors.NewBadRequest, 400, errors.MsgBadRequest},
		{"Unauthorized", errors.NewUnauthorized, 401, errors.MsgUnauthorized},
		{"Forbidden", errors.NewForbidden, 403, errors.MsgForbidden},
		{"NotFound", errors.NewNotFound, 404, errors.MsgNotFound},
		{"Conflict", errors.NewConflict, 409, errors.MsgConflict},
		{"RateLimited", errors.NewRateLimited, 429, "rate limited"},
		{"Internal", errors.NewInternal, 500, errors.MsgInternal},
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
	assert.NotEmpty(t, errors.MsgBadRequest)
	assert.NotEmpty(t, errors.MsgUnauthorized)
	assert.NotEmpty(t, errors.MsgForbidden)
	assert.NotEmpty(t, errors.MsgNotFound)
	assert.NotEmpty(t, errors.MsgInternal)
	assert.NotEmpty(t, errors.MsgConflict)
}

func TestNewBadRequest_WithStandardMessage(t *testing.T) {
	e := errors.NewBadRequest(errors.MsgBadRequest)
	assert.Equal(t, 400, e.Code)
	assert.Equal(t, errors.MsgBadRequest, e.Message)
}

func TestNewNotFound_WithStandardMessage(t *testing.T) {
	e := errors.NewNotFound(errors.MsgNotFound)
	assert.Equal(t, 404, e.Code)
	assert.Equal(t, errors.MsgNotFound, e.Message)
}

func TestNewInternal_WithStandardMessage(t *testing.T) {
	e := errors.NewInternal(errors.MsgInternal)
	assert.Equal(t, 500, e.Code)
	assert.Equal(t, errors.MsgInternal, e.Message)
}

func TestAppError_EmptyMessage(t *testing.T) {
	e := errors.NewBadRequest("")
	assert.Equal(t, 400, e.Code)
	assert.Equal(t, "", e.Message)
	assert.Equal(t, "[400] ", e.Error())
}
