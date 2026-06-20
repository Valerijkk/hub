# ARCH — Architect

## Роль

Проектирование системы Hub, нефункциональные характеристики, стек.

## Обязательные docs/

- `docs/vision_and_scope.md`
- `docs/system_specification.md`
- `docs/architecture.md` — основной артефакт
- `docs/stack_and_tools.md`
- `docs/code_style_guide.md`
- `docs/security_guidelines.md`

## Шаги работы

1. Загрузить все файлы из `docs/`
2. Обновить `docs/architecture.md`: C4, deployment, patterns
3. При изменении стека → `docs/stack_and_tools.md`
4. При изменении стандартов → `docs/code_style_guide.md`
5. Проверить соответствие security_guidelines
6. Передать ARCH_VER с описанием решений

## GSD Skills

`gsd-ai-integration-phase` | `gsd-ui-phase` | `gsd-research-phase` | `.cursor/GSD-INTEGRATION.md`

## Зона ответственности

**Делаю:** Ent схемы (design), REST контракты, module structure, security architecture
**НЕ делаю:** реализация кода (→ BE/FE)

## Верификация

| Статус | Дата | Верификатор |
|--------|------|-------------|
| verified: true | 2026-06-17 | ARCH_VER |

## Лог изменений

| Дата | Изменение | Файл |
|------|-----------|------|
| 2026-06-17 | Инициализация architecture под регламент | `docs/architecture.md` |
| 2026-06-17 | **ARCH audit vs codebase** — score 75/100, 10 TD items, top-5 violations | `docs/architecture.md`, `docs/stack_and_tools.md` |

---

## Session 2026-06-17 — Architecture Audit

### Score: **75 / 100** (+3 vs May 2025 audit 72/100)

| Категория | Вес | Балл | Δ | Обоснование |
|-----------|-----|------|---|-------------|
| Layering integrity | 20 | 15 | +1 | auth handler fixed; остаются middleware/dev/handler bypass |
| Ent leakage | 20 | 10 | 0 | 7 domain-пакетов без Repository |
| Dep cycles | 15 | 15 | 0 | `go build ./...` clean |
| DI / singletons | 10 | 10 | +1 | `platform/cron` вынесен из main.go |
| Interface boundaries | 10 | 7 | 0 | repo OK; handler→service concrete |
| Repository coverage | 15 | 9 | -1 | 11/18 = 61% (было 65% при другом denominator) |
| Package/file size | 5 | 4 | 0 | main.go 630 LOC |
| go.mod / build tags | 5 | 5 | +1 | Go 1.25, dev build tags OK |
| **Итого** | **100** | **75** | **+3** | |

### Inaccuracies fixed (docs vs codebase)

| Было | Стало | Evidence |
|------|-------|----------|
| 18 backend modules | 20 domain + `platform/` | `backend/internal/` — 21 dir |
| `main.go` ~629 LOC, cron inline | 630 LOC, `platform/cron` | `main.go:19`, `cron.Schedule` |
| Go 1.24 | Go 1.25 | `backend/go.mod:3` |
| 39 TanStack routes | 47 routes | `routeTree.gen.ts` — 47 `Route as` |
| Tremor charts | Recharts 3.x | `frontend/package.json` |
| ESLint frontend lint | Biome | `npm run lint` → biome |
| `middleware.Recover()` | `fiberrecover.New()` | `main.go:189` |
| WS NF-02 «в работе» | `ws.Authorizer` deployed | `main.go:575-605` |
| K3s in deployments/ | Compose only; K3s planned | no `deployments/k3s/` |
| Loki/Prometheus/Grafana in prod | Not in compose | `docker-compose.prod.yml` |
| Dev compose in deployments/ | `docker-compose.local.yml` at repo root | root file exists |
| Ent «latest» | ent v0.14.1 | `go.mod` |
| LiveKit SDK 1.8 | server-sdk-go/v2 2.4.0 | `go.mod:22` |
| nats.go 2.11 | client v1.48.0, server image 2.11 | `go.mod` vs compose |
| Missing `dev/` module | Added to C4 | `internal/dev/` build tag `!release` |
| Repository 7 packages (TD-02) | 7 confirmed + escape hatch TD-03 | grep `*ent.Client` |

### Top-5 arch violations (file:line)

1. `backend/internal/tracker/handler_issues.go:269` — Handler → Ent via `GetDB()`
2. `backend/internal/tracker/repository.go:153` — `GetDB() *ent.Client` escape hatch
3. `backend/internal/auth/middleware.go:120` — middleware → `svc.db.User.Get`
4. `backend/internal/rbac/middleware.go:46,114` + `cmd/server/main.go:540` — RBAC on `*ent.Client`
5. `backend/internal/dev/handler.go:102` — dev handler → `h.db.User.Query` (!release)

### Tech debt registry (TD-*)

| ID | Summary | Key evidence |
|----|---------|--------------|
| TD-01 | main.go DI bloat | `cmd/server/main.go` — 630 LOC |
| TD-02 | 7 packages без Repository | auth, dashboard, backup, imex, storage, apigateway, rbac |
| TD-03 | GetDB() escape hatch | `tracker/repository.go:153`, `handler_issues.go:269` |
| TD-04 | No K3s manifests | `deployments/` |
| TD-05 | No observability stack | `docker-compose.prod.yml` |
| TD-06 | No security headers middleware | `main.go:185-194` |
| TD-07 | No handler→service interfaces | all handlers |
| TD-08 | audit/retention outside repo | `audit/retention.go:41` |
| TD-09 | notification/push on Ent | `notification/push.go:45` |
| TD-10 | CRITICAL security backlog | `.planning/code-review-final/SUMMARY.md` |
| TD-11 | Playwright/MSW/ESLint documented but absent | `frontend/package.json` |

### Positive deltas since May 2025 audit

- SEC-01: `APIKeyAuth` before routes — `main.go:250`
- INF-04: backup uses real S3 client — `main.go:531-537`
- SEC-05 partial: WS `client.done` guard — `ws/hub.go:244-248`
- auth handler no longer holds `h.db` — layering fix in auth package
- github/gitlab services fully on Repository — no `entDB` in service
- `platform/cron` package — cron boilerplate extracted

### Next ARCH actions

1. Phase 19: Repository for auth (highest smell — 5 service files + middleware)
2. Remove `GetDB()` — move JQL issue search to tracker service
3. RBAC middleware → depend on RBACService interface, not `*ent.Client`
4. Add K3s + observability manifests or mark roadmap-only in README
5. Register security headers middleware in main.go (v2.5 Phase 16)
