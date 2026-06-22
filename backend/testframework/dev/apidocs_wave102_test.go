package dev_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/dev"
)

func TestMethodOrder(t *testing.T) {
	assert.Equal(t, 0, dev.ExportMethodOrder("GET"))
	assert.Equal(t, 1, dev.ExportMethodOrder("POST"))
	assert.Equal(t, 2, dev.ExportMethodOrder("PUT"))
	assert.Equal(t, 3, dev.ExportMethodOrder("PATCH"))
	assert.Equal(t, 4, dev.ExportMethodOrder("DELETE"))
	assert.Equal(t, 9, dev.ExportMethodOrder("OPTIONS"))
}

func TestMethodColor(t *testing.T) {
	assert.Equal(t, "#61affe", dev.ExportMethodColor("GET"))
	assert.Equal(t, "#49cc90", dev.ExportMethodColor("POST"))
	assert.Equal(t, "#f93e3e", dev.ExportMethodColor("DELETE"))
	assert.Equal(t, "#999", dev.ExportMethodColor("TRACE"))
}

func TestGuessModule(t *testing.T) {
	cases := map[string]string{
		"/health":              "core",
		"/api/auth/login":      "Auth",
		"/api/projects":        "Projects",
		"/api/issues/1":        "Tracker",
		"/api/messenger/ch":    "Messenger",
		"/api/channels/1":      "Messenger",
		"/api/github/repos":    "GitHub",
		"/api/calls/start":     "Calls",
		"/api/search":          "Search",
		"/api/notifications":   "Notifications",
		"/api/files/upload":    "Storage",
		"/api/dashboards/1":    "Dashboard",
		"/api/sprints":         "Sprints",
		"/api/admin/users":     "Admin",
		"/api/api-keys":        "API Gateway",
		"/api/webhooks":        "API Gateway",
		"/api/backups":         "Backup",
		"/api/import/jira":     "Import/Export",
		"/api/export/csv":      "Import/Export",
		"/api/dev/status":      "Dev",
		"/api/docs":            "System",
		"/api/health/live":     "System",
		"/api/teams":           "Teams",
		"/api/templates":       "Templates",
		"/api/roles":           "RBAC",
		"/api/link-preview":    "GitHub",
		"/api/custom-module/x": "custom-module",
	}
	for path, want := range cases {
		assert.Equal(t, want, dev.ExportGuessModule(path), "path %s", path)
	}
}

func TestCollectRoutes_DedupesAndSorts(t *testing.T) {
	app := fiber.New()
	app.Get("/api/projects", func(c fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })
	app.Post("/api/projects", func(c fiber.Ctx) error { return c.SendStatus(fiber.StatusCreated) })
	app.Get("/api/auth/me", func(c fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

	routes := dev.ExportCollectRoutes(app)
	require.NotEmpty(t, routes)

	var keys []string
	for _, r := range routes {
		keys = append(keys, r.Method+" "+r.Path)
	}
	assert.Contains(t, keys, "GET /api/auth/me")
	assert.Contains(t, keys, "GET /api/projects")
	assert.Contains(t, keys, "POST /api/projects")

	for _, r := range routes {
		assert.NotEqual(t, "HEAD", r.Method)
	}
}

func TestBuildAPIDocsHTMLAndJSON(t *testing.T) {
	app := fiber.New()
	app.Get("/api/notifications", func(c fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })
	app.Post("/api/auth/login", func(c fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

	html := dev.BuildAPIDocsHTML(app)
	assert.True(t, strings.HasPrefix(html, "<!DOCTYPE html>"))
	assert.Contains(t, html, "Hub Platform API")
	assert.Contains(t, html, "Notifications")
	assert.Contains(t, html, "GET")

	jsonStr := dev.BuildAPIDocsJSON(app)
	var routes []dev.ExportRouteEntry
	require.NoError(t, json.Unmarshal([]byte(jsonStr), &routes))
	assert.NotEmpty(t, routes)
}
