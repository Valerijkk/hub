package ws

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestCheckOrigin_AllowedList(t *testing.T) {
	t.Setenv("CORS_ORIGINS", "https://hub.example.com,https://app.example.com")
	t.Setenv("DEV_MODE", "false")
	resetOriginCheck()

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.Set("Origin", "https://hub.example.com")
	assert.True(t, checkOrigin(ctx))

	ctx.Request.Header.Set("Origin", "https://evil.example.com")
	assert.False(t, checkOrigin(ctx))
}

func TestCheckOrigin_EmptyOriginAllowed(t *testing.T) {
	t.Setenv("CORS_ORIGINS", "https://hub.example.com")
	t.Setenv("DEV_MODE", "false")
	resetOriginCheck()

	ctx := &fasthttp.RequestCtx{}
	assert.True(t, checkOrigin(ctx))
}

func TestCheckOrigin_DevModeAllowsAll(t *testing.T) {
	t.Setenv("CORS_ORIGINS", "")
	t.Setenv("DEV_MODE", "true")
	resetOriginCheck()

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.Set("Origin", "https://any.example.com")
	assert.True(t, checkOrigin(ctx))
}

func TestCheckOrigin_ProductionRejectsWhenUnset(t *testing.T) {
	t.Setenv("CORS_ORIGINS", "")
	t.Setenv("DEV_MODE", "false")
	resetOriginCheck()

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.Set("Origin", "https://hub.example.com")
	assert.False(t, checkOrigin(ctx))
}
