package errors_test
import (
	"github.com/valerii/hub/backend/internal/platform/errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRateLimitedAfter(t *testing.T) {
	e := errors.NewRateLimitedAfter("wait", 120)
	assert.Equal(t, 429, e.Code)
	assert.Equal(t, "wait", e.Message)
	assert.Equal(t, 120, e.RetryAfter)
}

func TestNewRateLimitedAfter_MinimumOneSecond(t *testing.T) {
	e := errors.NewRateLimitedAfter("slow down", 0)
	assert.Equal(t, 1, e.RetryAfter)
	e = errors.NewRateLimitedAfter("slow down", -5)
	assert.Equal(t, 1, e.RetryAfter)
}
