package status_test
import (
	"github.com/valerii/hub/backend/internal/status"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_GetStatus(t *testing.T) {
	svc := status.NewService(
		func(ctx context.Context) error { return nil },
		func(ctx context.Context) error { return nil },
		func(ctx context.Context) error { return nil },
		func(ctx context.Context) error { return nil },
	)
	h := status.NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api)

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	svcDegraded := status.NewService(
		func(ctx context.Context) error { return errors.New("down") },
		nil, nil, nil,
	)
	h2 := status.NewHandler(svcDegraded)
	app2 := fiber.New()
	api2 := app2.Group("/api")
	h2.RegisterRoutes(api2)

	req = httptest.NewRequest(http.MethodGet, "/api/status", nil)
	resp, err = app2.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}
