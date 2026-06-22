//go:build hubtest

package bases

import (
	"github.com/gofiber/fiber/v3"
	"github.com/valerii/hub/backend/ent"
)

func ExportMapServiceError(err error) error {
	return mapServiceError(err)
}

func ExportNormalizeCells(cells map[string]interface{}, columns []*ent.HubColumn) map[string]interface{} {
	return normalizeCells(cells, columns)
}

func ExportFiberStatusFromError(err error) int {
	if fe, ok := err.(*fiber.Error); ok {
		return fe.Code
	}
	return 0
}
