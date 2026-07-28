//go:build hubtest

package ai_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/ai"
)

// TestGetUserPreference_DefaultAllowExternalAIFalse — W5-A3-012 / P2-BE-17.
func TestGetUserPreference_DefaultAllowExternalAIFalse(t *testing.T) {
	svc := ai.NewService(&mockRepo{}, &mockAdapter{}, nil)
	resp, err := svc.GetUserPreference(context.Background(), uuid.New())
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.AllowExternalAI)
}
