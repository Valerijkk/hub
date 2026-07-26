# Hub — Post-Phase-14 Cross-Cutting Audit Summary

> **ИСТОРИЧЕСКИЙ BASELINE (2026-05-15).** Не использовать для текущего статуса продакшена.  
> Актуально: COV **83%** (2026-06-22 Docker baseline), CRITICAL **0/12**, product score **87/100** — `.planning/audits/2026-06/AUDIT-SCORECARD.md`, Wave C FA-02.

**Date:** 2026-05-15
**Scope:** 6 параллельных аудитов после закрытия milestone v2.0 (Hub Redesign) + Phase 14 (Architectural Hardening) + 6 волн audit-фиксов (CR-01..LO-08).
**Method:** 6 независимых агентов в параллель, отчёты в `.planning/code-review-final/`.

## Score-card

| Срез | Score | Артефакт |
|------|-------|----------|
| Backend Go code | 78/100 | [REVIEW-BACKEND.md](REVIEW-BACKEND.md) |
| Security (OWASP/ASVS) | — | [SECURITY.md](SECURITY.md) |
| Tests + race + coverage | 0 races, 11.5% cov | [TESTS.md](TESTS.md) |
| Frontend (TS/React/A11y) | 76/100 | [FRONTEND.md](FRONTEND.md) |
| Architecture | 72/100 | [ARCHITECTURE.md](ARCHITECTURE.md) |
| Infra/Docker/CI | 35/100 | [INFRA.md](INFRA.md) |

**Aggregate verdict:** код-уровень чистый (build/test/race зелёные), но **production cutover blocked** тремя категориями:
1. **Middleware-ordering bug** обнуляет API-key auth.
2. **Production deployment** содержит committed devkey, форсит `DEV_MODE=true`, не доставляет LiveKit fix.
3. **SSRF/IDOR/coverage gap** — Phase 14 закрыл только 30% surface; ещё 7 пакетов без Repository, GitLab/WebPush без SafeDialer, search без channel-scope.

## Unified Severity Stack-Rank

### CRITICAL (12) — блокеры прода

| ID | Source | Что | Файл | Effort |
|----|--------|-----|------|--------|
| SEC-01 | NF-01 | `APIKeyAuth` зарегистрирован после routes → dead code | `cmd/server/main.go:554` | 5 мин |
| SEC-02 | NF-04 | Webhook signature: body-hash optional (no `sha256` claim → skip) | `calls/handler.go:256-265` | 30 мин |
| SEC-03 | NF-05 | `livekit-prod.yaml` содержит `devkey:secret` в репо | `deployments/livekit-prod.yaml:8-11` | 15 мин + rotation |
| SEC-04 | C-01/NF-06 | GitLab SSRF — `http.DefaultClient` на user-controlled URL | `gitlab/service.go:265,367,453,508` | 1 час |
| SEC-05 | C-02 | Send-on-closed-channel в WS Hub → паника в проде | `ws/hub.go` (removeClient) | 2 часа |
| SEC-06 | C-03 | First-admin race — `IsFirstUser`+`User.Create` без транзакции | `auth/service_registration.go` | 30 мин |
| SEC-07 | C-04 | Timing attack в magic-link verify (`!=` vs `subtle.ConstantTimeCompare`) | `auth/magiclink.go` | 5 мин |
| INF-01 | (Infra) | LiveKit `departure_timeout=120` не доедет до прода — heredoc в compose.prod игнорит yaml | `docker-compose.prod.yml:113-134` | 30 мин |
| INF-02 | (Infra) | CI форсит `DEV_MODE=true` в проде → `/api/dev/*` открыт + rate-limit off | `.github/workflows/deploy.yml:64` | 5 мин |
| INF-03 | (Infra) | Plaintext credentials в `docker-compose.local.yml:115-117` — Resend + Telegram | rotate + git-purge | 1 час |
| INF-04 | (Infra) | `BackupService` получает `s3Client=nil` → бекапы не отправляются | `cmd/server/main.go:557` | 1 час |
| FE-01 | (Frontend) | Rules of Hooks violation — 8 хуков после condit return, костыль `ChatErrorBoundary` | `message-item.tsx:167-260` | 2 часа |

### HIGH (16)

| ID | Source | Что | Effort |
|----|--------|-----|--------|
| SEC-H1 | NF-02 | WS `CheckOrigin=true` + `Hub.Subscribe` без authz → broadcast IDOR | 2 часа |
| SEC-H2 | NF-03 | Mongo `$text` search без channel-фильтра — утечка между каналами | 1 час |
| SEC-H3 | NF-07 | WebPush endpoint user-controlled без `ValidateOutboundURL` (SSRF sink) | 30 мин |
| BE-H1 | H-01 | Unbounded `io.ReadAll` (GitLab/Resend) — memory DoS | 30 мин |
| BE-H2 | H-02 | `respBuf.ReadFrom(resp.Body)` в webhook читает всё перед обрезкой | 20 мин |
| BE-H3 | H-03 | Dashboard `Where()` без user-фильтра — чужие звонки | 30 мин |
| BE-H4 | H-04 | Goroutine-per-request в apikey path | 1 час |
| BE-H5 | H-05 | `/health` создаёт DB-conn на каждый probe | 15 мин |
| BE-H6 | H-06 | GDPR-delete глотает ошибки реакций/тайм-логов | 30 мин |
| BE-H7 | H-07 | Retention recordings не удаляет S3-объект | 1 час |
| FE-H1 | (Frontend) | Refresh-token race (нет single-flight mutex) | 1 час |
| FE-H2 | (Frontend) | Нет multi-tab auth sync (BroadcastChannel/storage events) | 2 часа |
| FE-H3 | (Frontend) | Нет `POST /auth/logout` → refresh жив после клиент-logout | 1 час |
| FE-H4 | (Frontend) | Нет Error Boundaries вокруг calls/dashboard/tracker/TMS | 2 часа |
| FE-H5 | (Frontend) | Bundle 1.1 MB + livekit 478 KB eager в 6 файлах | 3 часа |
| ARCH-H1 | (Architecture) | `auth/handler.go:82,503` — direct DB access in handler (layering) | 1 час |

### MEDIUM (21)

Backend: 9 (см. REVIEW-BACKEND.md §MEDIUM)
Frontend: 6 (50 non-null asserts, 16 index keys, a11y, button type, MediaRecorder cleanup)
Architecture: 6 (Repository pattern на 7 пакетов, ent escape-hatches, handler→service interfaces, notification cross-package events, rbac.RequireAdmin needs ent.Client refactor, `*ent.Client` в 10 файлах)

### LOW/INFO (8)

- `tms/ent_repository.go` 932 LOC — единственный оставшийся god-object
- `cmd/server/main.go` 629 LOC + 7 копий cron-boilerplate → вынести в `platform/cron`
- 10-параметровая `Dispatch(...)` в notification consumer
- Docs: README/wiki ссылаются на несуществующий `docker-compose.demo.yml`
- 19 frontend файлов > 300 строк (top: `message-composer.tsx` 742)
- Containers как root, `seaweedfs:latest` unpinned
- Mongo:7 (prod) vs Mongo:8 (local) drift
- `SeaweedFS` creds mismatch prod (`seaweedfs-s3.json` hardcoded vs env)

## Coverage Reality

`go test -race ./...` зелёный — но **statement coverage = 11.5%**. 4 пакета имеют 0% coverage при наличии test-файлов (ассертят только DTO-форму). Phase 14 добавил 29 функций по security-critical путям, но проект остаётся под-тестирован.

Топ-7 на покрытие: `platform/nats`, `storage`, `auth refresh rotation`, `calls recording webhook`, `messenger fan-out`, `rbac resolver`, `platform/outbox worker`.

## Рекомендуемые следующие фазы

**v2.0.1 (hotfix, 1 день):** SEC-01..SEC-07 + INF-01..INF-04 + FE-01 — 12 CRITICAL за 1 sprint-day. Обязательно перед prod cutover.

**v2.5 — Production Hardening (милстоун, ~2 недели):**
- Phase 15: SSRF + WS authz + search scope (NF-02/03/07, BE-H1..H7) — 5 plans
- Phase 16: Security headers + refresh-token hashing + storage validation (NF-08/09/12) — 3 plans
- Phase 17: Frontend refresh-mutex + bundle splitting + a11y + multi-tab + error boundaries (FE-H1..H5) — 5 plans
- Phase 18: Test coverage uplift (target 50% overall, 80% для security-critical) — 7 plans

**v2.6 — Architectural Completion (милстоун, ~1 неделя):**
- Phase 19: Repository pattern на 7 оставшихся пакетов (auth, dashboard, backup, imex, storage, apigateway, rbac) — 4 plans
- Phase 20: Escape-hatch removal (`EntClient()`, `GetDB()`) + handler→service interfaces — 2 plans
- Phase 21: Cron boilerplate в `platform/cron`, observability (metrics + OTel) — 3 plans

## Сильные стороны (что работает)

- Build/test/race зелёные, `go vet` clean, `go build -tags release` clean
- SRP-сплиты Auth/Calls аккуратные, циклов нет, `errors.Is(ent.IsNotFound)` повсеместно
- TypeScript: `tsc --noEmit` = 0 errors
- Phase 8 CSS clean — `!important` остался только в `@media (prefers-reduced-motion)`
- SafeDialer в webhook работает, sync.Once в Stop, корректный WS Hub через chan, graceful shutdown консистентный
- 22 integration-тестов на login/refresh/RBAC middleware/platform middleware
- Repository pattern на 11 пакетов, layering правильный для 14 из 17
- Phase 14 закрыт по metricам: оба god-object < 250 LOC, ent.Client удалён из 3 сервисов (полная очистка — milestone v2.6)

## Total finding count

| Severity | Count | Source breakdown |
|----------|-------|------------------|
| CRITICAL | 12 | Sec 3 + Backend 4 + Infra 4 + Frontend 1 |
| HIGH | 16 | Sec 3 + Backend 7 + Frontend 5 + Architecture 1 |
| MEDIUM | 21 | Backend 9 + Frontend 6 + Architecture 6 |
| LOW/INFO | 8 | Cross-cutting |
| **Total** | **57** | (vs 41 в первой волне до Phase 14) |

Phase 14 + 6 audit-волн закрыли 41 findings, но runtime audit выявил **57 новых/остаточных** — это нормально, потому что:
- Глубже копались (frontend bundle, infra deploy, layering)
- Phase 14 рефакторил структуру, не security — security surface шире
- Multi-pass audit находит то, что single-pass пропустил

Готовность к продакшену: **NO**, пока не закрыты 12 CRITICAL. После v2.0.1 hotfix — **YES** при условии rotation скомпрометированных секретов.
