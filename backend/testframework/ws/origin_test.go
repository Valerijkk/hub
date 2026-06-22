package ws_test

// Тесты модуля WebSocket: интеграционные и unit-сценарии testframework.

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/valerii/hub/backend/internal/ws"
)

// TestCheckOrigin_AllowedList проверяет сценарий: проверка origin allowed список.
func TestCheckOrigin_AllowedList(t *testing.T) {
	t.Setenv("CORS_ORIGINS", "https://hub.example.com,https://app.example.com")
	t.Setenv("DEV_MODE", "false")
	ws.ExportResetOriginCheck()

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.Set("Origin", "https://hub.example.com")
	assert.True(t, ws.ExportCheckOrigin(ctx))

	ctx.Request.Header.Set("Origin", "https://evil.example.com")
	assert.False(t, ws.ExportCheckOrigin(ctx))
}

// TestCheckOrigin_EmptyOriginAllowed проверяет сценарий: проверка origin пустой origin allowed.
func TestCheckOrigin_EmptyOriginAllowed(t *testing.T) {
	t.Setenv("CORS_ORIGINS", "https://hub.example.com")
	t.Setenv("DEV_MODE", "false")
	ws.ExportResetOriginCheck()

	ctx := &fasthttp.RequestCtx{}
	assert.True(t, ws.ExportCheckOrigin(ctx))
}

// TestCheckOrigin_DevModeAllowsAll проверяет сценарий: проверка origin dev mode allows all.
func TestCheckOrigin_DevModeAllowsAll(t *testing.T) {
	t.Setenv("CORS_ORIGINS", "")
	t.Setenv("DEV_MODE", "true")
	ws.ExportResetOriginCheck()

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.Set("Origin", "https://any.example.com")
	assert.True(t, ws.ExportCheckOrigin(ctx))
}

// TestCheckOrigin_ProductionRejectsWhenUnset проверяет сценарий: проверка origin production rejects when unset.
func TestCheckOrigin_ProductionRejectsWhenUnset(t *testing.T) {
	t.Setenv("CORS_ORIGINS", "")
	t.Setenv("DEV_MODE", "false")
	ws.ExportResetOriginCheck()

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.Set("Origin", "https://hub.example.com")
	assert.False(t, ws.ExportCheckOrigin(ctx))
}
