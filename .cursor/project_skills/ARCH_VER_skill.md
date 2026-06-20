# ARCH_VER — Верификатор Architect

## Роль

Проверка архитектурных решений Hub на соответствие требованиям и KISS.

## Чек-лист

- [x] Архитектура покрывает все NFR из system_specification
- [x] SOLID, Clean Architecture в module structure
- [x] Нет неоправданной сложности (KISS)
- [x] Реализуемо выбранным стеком (stack_and_tools.md)
- [x] Security на уровне архитектуры (WS, search, webhooks)
- [x] Нет запрещённых замен (Fiber, NATS, SeaweedFS, TanStack Router)
- [ ] Описание от ARCH соответствует codebase (2 stale items — см. лог 2026-06-17)

## Верификация

| Статус | Дата |
|--------|------|
| verified: true | 2026-06-17 |

## Лог проверок

| Дата | Артефакт | Результат | Замечания |
|------|----------|-----------|-----------|
| 2026-06-17 | architecture.md init | pending | — |
| 2026-06-17 | architecture.md + stack_and_tools.md audit | **FAIL** | TD-06 stale (SecurityHeaders wired `main.go:190`); middleware order omits SecurityHeaders. 28/30 claims verified. |

### 2026-06-17 — Audit spot-check detail

| Claim | Status | Evidence |
|-------|--------|----------|
| 20 domain modules + platform | ✅ | 21 dirs in `backend/internal/` |
| Repository 11/18 (61%) | ✅ | 11 `repository.go`; 7 without: auth, dashboard, backup, imex, storage, apigateway, rbac |
| Ent 59 schemas | ✅ | 60 files − `helpers.go` |
| main.go ~630 LOC | ✅ (631) | line count |
| Go 1.25, Fiber v3 beta.3, ent v0.14.1 | ✅ | `go.mod` |
| 47 TanStack routes, 11 Zustand stores | ✅ | `routeTree.gen.ts`, `src/stores/` |
| Top-5 violations (file:line) | ✅ | all 5 confirmed at documented lines |
| SEC-01 APIKeyAuth before routes | ✅ | `main.go:251` before `RegisterRoutes` |
| INF-04 backup S3 client | ✅ | `main.go:532-538` |
| WS Authorizer | ✅ | `main.go:576-605` |
| No K3s / no observability in prod | ✅ | `deployments/` |
| Recharts, Biome, no Playwright/MSW/ESLint | ✅ | `package.json`, `ci.yml` |
| TD-06 security headers not wired | ❌ | `middleware.SecurityHeaders()` at `main.go:190` |
| Middleware order (HTTP Request) | ❌ | docs omit SecurityHeaders between Recover and Logger |
