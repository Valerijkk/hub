//go:build hubtest

package health

import (
	"context"

	"github.com/gofiber/fiber/v3"
)

// ExportProbeParentContext — W4-A4-012: Ready берёт parent из c.Context(), не Background.
func ExportProbeParentContext(c fiber.Ctx) context.Context {
	return probeParentContext(c)
}
