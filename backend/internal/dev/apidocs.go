package dev

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// routeEntry представляет один маршрут API для документации.
type routeEntry struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

// collectRoutes собирает все зарегистрированные маршруты из Fiber-приложения.
// Исключает HEAD-маршруты и внутренние маршруты Fiber.
func collectRoutes(app *fiber.App) []routeEntry {
	seen := make(map[string]bool)
	var routes []routeEntry

	for _, r := range app.GetRoutes(true) {
		if r.Method == "HEAD" || r.Method == "" {
			continue
		}
		key := r.Method + " " + r.Path
		if seen[key] {
			continue
		}
		seen[key] = true
		routes = append(routes, routeEntry{Method: r.Method, Path: r.Path})
	}

	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path == routes[j].Path {
			return methodOrder(routes[i].Method) < methodOrder(routes[j].Method)
		}
		return routes[i].Path < routes[j].Path
	})

	return routes
}

// methodOrder возвращает порядковый номер HTTP-метода для сортировки.
func methodOrder(m string) int {
	switch m {
	case "GET":
		return 0
	case "POST":
		return 1
	case "PUT":
		return 2
	case "PATCH":
		return 3
	case "DELETE":
		return 4
	default:
		return 9
	}
}

// methodColor возвращает CSS-цвет для HTTP-метода (для HTML-документации).
func methodColor(m string) string {
	switch m {
	case "GET":
		return "#61affe"
	case "POST":
		return "#49cc90"
	case "PUT":
		return "#fca130"
	case "PATCH":
		return "#50e3c2"
	case "DELETE":
		return "#f93e3e"
	default:
		return "#999"
	}
}

// guessModule определяет модуль по пути маршрута.
func guessModule(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) < 2 {
		return "core"
	}
	// /api/{module}/...
	seg := parts[1]
	switch {
	case seg == "auth":
		return "Auth"
	case seg == "projects":
		return "Projects"
	case seg == "issues":
		return "Tracker"
	case seg == "messenger" || seg == "channels":
		return "Messenger"
	case seg == "github":
		return "GitHub"
	case seg == "calls":
		return "Calls"
	case seg == "search":
		return "Search"
	case seg == "notifications":
		return "Notifications"
	case seg == "files":
		return "Storage"
	case seg == "dashboards":
		return "Dashboard"
	case seg == "sprints":
		return "Sprints"
	case seg == "admin":
		return "Admin"
	case seg == "api-keys" || seg == "webhooks":
		return "API Gateway"
	case seg == "backups":
		return "Backup"
	case seg == "import" || seg == "export":
		return "Import/Export"
	case seg == "dev":
		return "Dev"
	case seg == "docs" || seg == "health" || seg == "live" || seg == "ready" || seg == "status" || seg == "metrics":
		return "System"
	case seg == "teams":
		return "Teams"
	case seg == "templates":
		return "Templates"
	case seg == "roles":
		return "RBAC"
	case seg == "link-preview":
		return "GitHub"
	default:
		return seg
	}
}

// BuildAPIDocsHTML генерирует HTML-страницу с документацией всех API-эндпоинтов.
// Маршруты группируются по модулям и отображаются с цветовой кодировкой методов.
func BuildAPIDocsHTML(app *fiber.App) string {
	routes := collectRoutes(app)

	// Группируем по модулям.
	groups := make(map[string][]routeEntry)
	var order []string
	for _, r := range routes {
		mod := guessModule(r.Path)
		if _, exists := groups[mod]; !exists {
			order = append(order, mod)
		}
		groups[mod] = append(groups[mod], r)
	}

	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="utf-8">
<title>Hub Platform API</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;background:#fafafa;color:#333;padding:20px 40px}
h1{font-size:28px;margin-bottom:4px}
.subtitle{color:#666;margin-bottom:24px;font-size:14px}
.module{margin-bottom:24px}
.module h2{font-size:18px;color:#333;border-bottom:2px solid #eee;padding-bottom:6px;margin-bottom:8px}
.route{display:flex;align-items:center;padding:6px 12px;border-radius:4px;margin-bottom:2px;font-family:"SF Mono",Monaco,Consolas,monospace;font-size:13px}
.route:hover{background:#f0f0f0}
.method{display:inline-block;width:70px;font-weight:700;font-size:12px;text-align:center;padding:2px 0;border-radius:3px;color:#fff;margin-right:12px;flex-shrink:0}
.path{color:#333}
.count{color:#999;font-size:12px;margin-left:8px}
</style>
</head>
<body>
<h1>Hub Platform API</h1>
<p class="subtitle">`)
	sb.WriteString(fmt.Sprintf("%d endpoints", len(routes)))
	sb.WriteString(`</p>
`)

	for _, mod := range order {
		rr := groups[mod]
		sb.WriteString(fmt.Sprintf(`<div class="module"><h2>%s <span class="count">(%d)</span></h2>`, mod, len(rr)))
		for _, r := range rr {
			sb.WriteString(fmt.Sprintf(
				`<div class="route"><span class="method" style="background:%s">%s</span><span class="path">%s</span></div>`,
				methodColor(r.Method), r.Method, r.Path,
			))
		}
		sb.WriteString("</div>\n")
	}

	sb.WriteString("</body></html>")
	return sb.String()
}

// BuildAPIDocsJSON генерирует JSON-список всех API-эндпоинтов.
// Формат: массив объектов {method, path}.
// W7-A5-015 / P3-BE-01: marshal err → error (handler отдаёт 500), не пустое тело.
func BuildAPIDocsJSON(app *fiber.App) (string, error) {
	routes := collectRoutes(app)
	data, err := json.MarshalIndent(routes, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal api docs: %w", err)
	}
	return string(data), nil
}
