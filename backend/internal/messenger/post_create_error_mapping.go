package messenger

import (
	"errors"

	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ExportPhase151MapPostCreatePipelineError maps the pure coordinator's domain
// outcomes to the stable public service contract without exposing internal
// repository or authorization details. Live writers remain separately bound.
func ExportPhase151MapPostCreatePipelineError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, ErrPostCreatePipelineNonceConflict):
		return apperrors.NewConflict("client_nonce уже использован для другого сообщения")
	case errors.Is(err, ErrPostCreatePipelineNonceDeleted):
		return apperrors.NewConflict("client_nonce уже принадлежит удалённому сообщению")
	default:
		return apperrors.NewInternal("ошибка сохранения сообщения")
	}
}
