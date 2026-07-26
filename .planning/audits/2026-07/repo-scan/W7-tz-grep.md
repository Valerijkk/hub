# W7 — TZ grep (readonly)

> **Provenance class:** PRE_EXISTING · **Confidence:** MIXED · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> Агент: `security-auditor` (Grok 4.5) · Дата: **2026-07-23** · Scope: grep TZ-forbidden паттернов
> Паттерны: `http.DefaultClient`, `panic(` в handlers/internal, `*ent.Client` в service/handler, `seaweedfs:latest`, `DEV_MODE=true` в prod.
> Код **не менялся**.

## Чек-лист (результат)

| Контроль | Статус | Evidence |
|----------|--------|----------|
| `http.DefaultClient` в prod-коде | PASS | 0 вызовов в `backend/internal/**` и `backend/cmd/**`; только комментарии |
| `panic(` в handlers / internal | PASS | 0 вызовов `panic(` во всём `backend/**/*.go` (prod + testframework) |
| `*ent.Client` в Service | PASS | в `**/service*.go` — только комментарии «не хранит»; полей/параметров нет |
| `*ent.Client` в Handler (prod) | PASS | prod handlers без поля; единственный hit — `dev` (`!release`) |
| `seaweedfs:latest` в compose/K3s | PASS | prod/dev compose: `chrislusf/seaweedfs:3.97`; литерала `:latest` на seaweedfs в yaml нет |
| `DEV_MODE=true` в prod/deploy | PASS | `deploy.yml` — 0 hits; `docker-compose.prod.yml:205` → `"false"`; CI e2e — accepted |

## Findings

| Sev | file:line | Уязвимость / наблюдение | Рекомендация | OWASP/ASVS | Статус |
|-----|-----------|-------------------------|--------------|------------|--------|
| INFO | `backend/internal/dev/handler.go:25,30` | Handler держит `db *ent.Client` — **намеренно**, файл под `//go:build !release` (ME-09). Release-заглушка: `handler_release.go:31` (no-op). | Не переносить паттерн в prod-модули. Оставить как есть. | A01 / layering TZ | accepted |
| INFO | `.github/workflows/ci.yml:225` | `DEV_MODE: 'true'` в job `e2e-full` — тестовый контур, не prod. | Не переносить в `deploy.yml` / `docker-compose.prod.yml`. | A05 Misconfig | accepted |
| INFO | `deployments/docker-compose.prod.yml:172,237,261` | Default image tags `ghcr.io/.../backend:latest` / `frontend:latest` (app GHCR), **не** seaweedfs. Deploy по SHA — см. W7-ci. | Опциональный pinning app-образов (вне scope этого grep). | A06 Components | deferred → W7-ci |
| — | — | CRITICAL/HIGH/MEDIUM по scope **не найдены** | — | — | — |

### Закрытые / подтверждённые инварианты (не findings)

1. **`http.DefaultClient`:**
   - `backend/internal/notification/push.go:21` — комментарий «вместо DefaultClient»; факт: `push.go:25` → `safehttp.NewSafeHTTPClient(10 * time.Second)`.
   - `backend/testframework/github/service_http_test.go:11` — комментарий в тесте SEC-04.
   - Prod usage: **0**.

2. **`panic(`:** grep по `backend/**/*.go` — **0** вызовов `panic(`. Упоминания только в комментариях/recover:
   - `platform/crypto/aesgcm.go:29` (запрет panic),
   - `platform/cron/cron.go:76-78` (recover log),
   - `platform/outbox/worker.go:75-78` (recover log),
   - `ws/hub.go:359` (исторический race).
   Handlers: **0**.

3. **`*ent.Client` в service:** hits только комментарии:
   - `audit/service.go:32`, `messenger/service.go:18`, `github/service.go:41,535`, `gitlab/service.go:48`, `auth/service.go:101` — все «не держит / больше не зависит».
   Реальных полей `*ent.Client` в service: **0**.

4. **`*ent.Client` в handler:**
   - `auth/handler.go:17` — комментарий ARCH-H1.
   - `dev/handler.go:25,30` — accepted (см. INFO).
   - `dev/handler_release.go:31` — сигнатура совместимости, клиент игнорируется.

5. **`seaweedfs:latest`:**
   - `deployments/docker-compose.prod.yml:93-95` — `chrislusf/seaweedfs:3.97` (+ комментарий LO-Inf про бывший `:latest`).
   - `deployments/docker-compose.yml:51-52` — `3.97`.
   - `docker-compose.local.yml.example:70-71` — `3.97`.
   - Литерал `seaweedfs:latest` остаётся только в **исторических** docs/planning (v1.0 milestones, code-review-final INFRA) и agent rules — **не** в runtime compose/K3s.

6. **`DEV_MODE=true` prod:**
   - `.github/workflows/deploy.yml` — совпадений `DEV_MODE` **нет**.
   - `deployments/docker-compose.prod.yml:205` — `DEV_MODE: "false"`.
   - Runtime guard: `backend/internal/platform/config/config.go:211-214` — `DevMode && Environment==production` → ошибка валидации.
   - CI: `ci.yml:225` — `DEV_MODE: 'true'` только в `e2e-full` (accepted).

## Grep-лог (кратко)

```
http.DefaultClient     → comments only (push.go:21, testframework github test)
panic(                 → 0 calls in backend/**/*.go
*ent.Client service*   → comments only (no fields)
*ent.Client handler*   → only dev (!release) + release stub signature
seaweedfs:latest       → 0 in deployments/*.yml; pinned 3.97
DEV_MODE=true prod     → deploy.yml clean; prod compose "false"; ci e2e true (accepted)
```

## Вердикт

- **CRITICAL:** 0
- **HIGH:** 0
- **MEDIUM:** 0
- **LOW:** 0
- **INFO/accepted:** 3 (dev `*ent.Client`, CI `DEV_MODE`, app `:latest` deferred)

**verified: false** · **historical_self_check:** PASS (grep scope only) — 2026-07-23 01:10 UTC+3

Блокеров релиза по TZ-grep нет. Все пять целевых инвариантов — PASS.
`)
