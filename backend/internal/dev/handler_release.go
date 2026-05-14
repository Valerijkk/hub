//go:build release

// Release-build stub for the dev package.
//
// ME-09: when the binary is built with `-tags release`, all dev endpoints
// (seed, login, status) are physically compiled out — the only way to
// "enable" them is to recompile without the `release` build tag, which
// is impossible to do at runtime. This is strictly stronger than gating
// on the DEV_MODE env var (which can be flipped by anyone who can edit
// the deployment manifest).
//
// The dev.Handler type and its NewHandler / RegisterRoutes methods are
// re-declared here as no-ops so that cmd/server/main.go continues to
// compile under both tag sets without code changes.

package dev

import (
	"github.com/gofiber/fiber/v3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/auth"
)

// Handler is a release-build stub. All methods are intentionally no-ops.
type Handler struct{}

// NewHandler returns a stub handler. db and authSvc are accepted to keep
// the API identical to the dev-build NewHandler, but neither is stored
// — the release build deliberately has no dev functionality.
func NewHandler(_ *ent.Client, _ *auth.AuthService) *Handler {
	return &Handler{}
}

// RegisterRoutes is a no-op in release builds. /api/dev/* simply does
// not exist.
func (h *Handler) RegisterRoutes(_ fiber.Router, _ bool) {
	// no-op
}
