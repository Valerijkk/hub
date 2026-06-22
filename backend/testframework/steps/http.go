package steps

// Тесты модуля steps: интеграционные и unit-сценарии testframework.

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

// HTTPAct отправляет запрос и возвращает ответ (шаг Act).
func HTTPAct(t *testing.T, app *fiber.App, req *http.Request) *http.Response {
	t.Helper()
	resp, err := app.Test(req)
	require.NoError(t, err)
	return resp
}

// AssertStatus — переиспользуемый шаг Assert (AAA).
func AssertStatus(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	require.Equal(t, want, resp.StatusCode)
}
