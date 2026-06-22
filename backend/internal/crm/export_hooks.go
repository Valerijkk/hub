//go:build hubtest

package crm

import "github.com/gofiber/fiber/v3"

func ExportMapServiceError(err error) error {
	return mapServiceError(err)
}

func ExportFiberStatusFromError(err error) int {
	if fe, ok := err.(*fiber.Error); ok {
		return fe.Code
	}
	return 0
}
