//go:build hubtest

package gitlab_test

import (
	"crypto/sha256"
	"crypto/subtle"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valerii/hub/backend/internal/gitlab"
)

// TestConstantTimeEqualDigest_UnequalLength — W7-A3-003 / P2-BE-20.
func TestConstantTimeEqualDigest_UnequalLength(t *testing.T) {
	assert.False(t, gitlab.ExportConstantTimeEqualDigest("short", "much-longer-secret-value"))
	assert.False(t, gitlab.ExportConstantTimeEqualDigest("abc", "xyz"))
	assert.True(t, gitlab.ExportConstantTimeEqualDigest("same-token", "same-token"))

	// Дайджесты равной длины — CTC не early-exit по len(token).
	a := sha256.Sum256([]byte("a"))
	b := sha256.Sum256([]byte("bb"))
	assert.Equal(t, 1, subtle.ConstantTimeCompare(a[:], a[:]))
	assert.Equal(t, 0, subtle.ConstantTimeCompare(a[:], b[:]))
}
