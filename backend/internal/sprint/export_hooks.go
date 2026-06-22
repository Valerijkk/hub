//go:build hubtest

package sprint

import (
	"context"

	"github.com/valerii/hub/backend/ent"
)

// ExportSprintServiceToSprintResponse конвертирует ent.Sprint в SprintResponse для hubtest.
func ExportSprintServiceToSprintResponse(s *SprintService, ctx context.Context, sp *ent.Sprint) *SprintResponse {
	return s.toSprintResponse(ctx, sp)
}

// ExportNilIfEmpty возвращает nil для пустой строки для hubtest.
func ExportNilIfEmpty(s string) *string {
	return nilIfEmpty(s)
}
