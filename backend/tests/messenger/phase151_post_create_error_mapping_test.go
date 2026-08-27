package messenger_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

func TestPhase151MapsTypedPostCreateNonceOutcomesToStableConflicts(t *testing.T) {
	for _, scenario := range []struct {
		name    string
		err     error
		message string
	}{
		{
			name:    "direct fingerprint conflict",
			err:     messenger.ErrPostCreatePipelineNonceConflict,
			message: "client_nonce уже использован для другого сообщения",
		},
		{
			name:    "wrapped fingerprint conflict",
			err:     fmt.Errorf("post-create failed: %w", messenger.ErrPostCreatePipelineNonceConflict),
			message: "client_nonce уже использован для другого сообщения",
		},
		{
			name:    "direct legacy reservation",
			err:     messenger.ErrPostCreatePipelineNonceLegacyReserved,
			message: "client_nonce уже использован для другого сообщения",
		},
		{
			name:    "wrapped legacy reservation",
			err:     fmt.Errorf("post-create failed: %w", messenger.ErrPostCreatePipelineNonceLegacyReserved),
			message: "client_nonce уже использован для другого сообщения",
		},
		{
			name:    "direct deleted reservation",
			err:     messenger.ErrPostCreatePipelineNonceDeleted,
			message: "client_nonce уже принадлежит удалённому сообщению",
		},
		{
			name:    "wrapped deleted reservation",
			err:     fmt.Errorf("post-create failed: %w", messenger.ErrPostCreatePipelineNonceDeleted),
			message: "client_nonce уже принадлежит удалённому сообщению",
		},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			mapped := messenger.ExportPhase151MapPostCreatePipelineError(scenario.err)

			var appError *apperrors.AppError
			require.ErrorAs(t, mapped, &appError)
			assert.Equal(t, 409, appError.Code)
			assert.Equal(t, scenario.message, appError.Message)
			assert.Empty(t, appError.RequestID)
			assert.Nil(t, appError.Details)
			assert.Zero(t, appError.RetryAfter)
		})
	}
}

func TestPhase151RedactsNonUserPostCreateFailuresToStableInternalError(t *testing.T) {
	for _, scenario := range []struct {
		name string
		err  error
	}{
		{name: "invalid canonical result", err: messenger.ErrPostCreatePipelineInvalid},
		{
			name: "wrapped invalid canonical result",
			err:  fmt.Errorf("canonical payload leaked: %w", messenger.ErrPostCreatePipelineInvalid),
		},
		{name: "producer scope denied", err: messenger.ErrTrustedMessageProducerDenied},
		{name: "repository detail", err: errors.New("mongo host secret: duplicate key")},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			mapped := messenger.ExportPhase151MapPostCreatePipelineError(scenario.err)

			var appError *apperrors.AppError
			require.ErrorAs(t, mapped, &appError)
			assert.Equal(t, 500, appError.Code)
			assert.Equal(t, "ошибка сохранения сообщения", appError.Message)
			assert.NotContains(t, appError.Error(), "canonical payload leaked")
			assert.NotContains(t, appError.Error(), "mongo host secret")
			assert.Empty(t, appError.RequestID)
			assert.Nil(t, appError.Details)
			assert.Zero(t, appError.RetryAfter)
		})
	}
}

func TestPhase151PostCreateErrorMapperPreservesNil(t *testing.T) {
	assert.NoError(t, messenger.ExportPhase151MapPostCreatePipelineError(nil))
}
