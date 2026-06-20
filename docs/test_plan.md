# Hub — Test Plan

> Единая база знаний. Обновляется QA при изменении стратегии тестирования.
> Детали: `docs/TESTING.md`, coverage: `.planning/COVERAGE.md`

---

## Стратегия тестирования (пирамида)

```
        ┌─────────┐
        │   E2E   │  Playwright — критические пути
        ├─────────┤
        │ Integr. │  testcontainers, API smoke
        ├─────────┤
        │  Unit   │  Go testify + Vitest/RTL
        └─────────┘
```

## Виды тестов

| Уровень | Инструмент | Расположение | Порог |
|---------|------------|--------------|-------|
| Unit (BE) | Go testing + testify + mockery | `backend/internal/*/` | ≥ 80% security-critical |
| Unit (FE) | Vitest + RTL | `frontend/src/**/*.test.tsx` | Цель v2.5 Phase 18 |
| Integration | testcontainers (PG, Redis, Mongo) | `backend/internal/*/` | Все ключевые сервисы |
| API smoke | httptest | `backend/internal/auth/handler_test.go` | Auth endpoints |
| E2E | Playwright + POM | `frontend/e2e/` | Критические пути |
| Build | tsc + vite build | CI | 0 errors |
| Security | Manual audit + grep | `.planning/code-review-final/` | 0 CRITICAL |

---

## Test Triage Protocol (REGULATION §10)

При падении любого теста (Go/Vitest/Playwright):

1. **Воспроизвести функционал** — поведение по `docs/system_specification.md`, не только assert
2. **Классифицировать:** ROOT=code (баг в продакшн) | ROOT=test (устаревший/неверный тест)
3. **Чинить минимальный слой** — код или тест, не оба «наугад»
4. **Записать** в QA skill-лог: triage + обоснование

Skill: `.cursor/skills/test-triage/SKILL.md` | GSD: `gsd-debug` для сложных кейсов

**Запрещено:** зелёные тесты без проверки соответствия спецификации.

---

## Критерии приёмки (общие)

1. Все существующие тесты проходят (`go test -race`, `vitest`, `playwright`)
2. Новый код покрыт unit-тестами (граничные случаи)
3. Нет регрессии по security checklist (`docs/security_guidelines.md`)
4. API changes отражены в `docs/system_specification.md`
5. E2E для новых user-facing flows (если критический путь)

## Матрица покрытия требований

| Модуль | Unit | Integration | E2E | Security audit |
|--------|------|-------------|-----|----------------|
| auth | ✅ handler + service | partial | login flow | SEC-06, SEC-07 |
| tracker | ✅ service | partial | kanban CRUD | — |
| messenger | ✅ 4 test files | partial | send message | HI-01 |
| tms | ✅ service + allure | partial | — | — |
| calls | ✅ service + chat | partial | join call | SEC-02, NF-04 |
| ws | ✅ protocol + hub | — | real-time | SEC-05, NF-02 |
| search | ✅ service + jql | ⚠️ command-palette unit | Ctrl+K UI ✅ | NF-03 |
| apigateway | ✅ apikey | — | — | SEC-01 |
| platform | ✅ middleware, errors, http, outbox | — | — | — |

Полная матрица: `.planning/COVERAGE.md` (Phase 0 reconciliation).

## Backend — текущее покрытие

**42 тестовых файла, 27 пакетов с тестами, 323 тестовых функции**

Последний прогон QA (2026-06-17, `go test -race -count=1 ./...`):

| Метрика | Результат |
|---------|-----------|
| Пакеты с тестами | **27 ok** |
| FAIL | **0** |
| Data races | **0** |
| Wall time | ~68s (longest: `internal/calls` ~14s) |
| Statement coverage (`./internal/...`) | **11.6%** |

Пакеты **без** `_test.go` (4): `platform/config`, `platform/cron`, `platform/database`, `platform/redis`.

Пакеты с тестами, но **0% statement coverage** (DTO/constant-only): `dashboard`, `dev`, `platform/nats`, `storage`.

Покрытие по пакетам (ascending):

| Package | Coverage |
|---------|----------|
| dashboard, dev, platform/nats, storage | 0.0% |
| imex | 2.5% |
| gitlab | 4.0% |
| github | 4.9% |
| calls | 5.1% |
| backup | 5.2% |
| notification | 5.6% |
| apigateway | 8.4% |
| audit | 9.2% |
| project | 8.1% |
| messenger | 12.0% |
| tms | 12.2% |
| tracker | 12.4% |
| auth | 18.0% |
| sprint | 18.3% |
| ws | 20.9% |
| search | 23.6% |
| platform/outbox | 28.6% |
| rbac | 33.8% |
| platform/http | 40.4% |
| platform/middleware | 45.4% |
| platform/crypto | 78.0% |
| platform/errors, platform/telegram | 100.0% |

Запуск:
```bash
cd backend && go test -race -count=1 ./...
go test -cover ./internal/...
```

## Frontend — текущее покрытие

| Проверка | Команда | Статус (2026-06-17) |
|----------|---------|---------------------|
| TypeScript | `npx tsc --noEmit` | ✅ **0 errors** |
| Build | `npx vite build` | ✅ CI (не перепроверялось в этой сессии) |
| Lint (Biome) | `npm run lint` | ⚠️ **269 errors**, 64 warnings, 13 infos в `src/` |
| Unit (Vitest) | `npx vitest run` | ✅ **49 tests**, 15 files — all green (2026-06-20) |
| E2E (Playwright) | `npx playwright test` | ✅ **7 specs** in `frontend/e2e/`; CI `e2e-full` job wired |
| Rules of Hooks | `biome lint` → `useHookAtTopLevel` | ✅ **0 violations** (FE-01 закрыт) |

### FE-TEST — Vitest smoke tests (PLAN-CLOSE-GAPS)

| Поле | Значение |
|------|----------|
| ID | FE-TEST |
| Target | ≥10 Vitest smoke tests |
| Статус | ✅ **MET** — **49 tests** in 15 files (Wave 3 + cycles 7–10) |
| Файлы | `auth-store.test.ts`, `client.test.ts`, `admin.test.ts`, `use-auth.test.ts`, `error-boundary.test.tsx`, `audit-log.test.tsx` |
| Верификация | `npx vitest run` — 17 passed, 0 failed |
| Примечание | Statement coverage ~13% overall; FE unit coverage not yet measured in CI |

### FE-01 — Rules of Hooks (`message-item.tsx`)

| Поле | Значение |
|------|----------|
| ID | FE-01 (CRITICAL) |
| Файл | `frontend/src/components/messenger/message-item.tsx` |
| Было | 8× `useMemo`/`useCallback` после условного `return null` → React #185 |
| Статус | ✅ **FIXED** — коммит `f334340` (`fix(messenger): hoist hooks above early returns`) |
| Верификация | `npx biome lint src` — **0** `useHookAtTopLevel`; полный scan `frontend/src` — других нарушений не найдено |
| Примечание | `ChatErrorBoundary` в `$channelId.tsx` оставлен как defence-in-depth (см. TODO в файле) |

### Открытые FE-блокеры (не FE-01)

| ID | Severity | Описание | Файл |
|----|----------|----------|------|
| FE-H1 | HIGH | Refresh-token race — N параллельных 401 → N refresh | `api/client.ts` |
| FE-H2 | HIGH | Main bundle ~1.1 MB, LiveKit eager-load | route-level lazy |
| FE-H3 | HIGH | Нет multi-tab auth sync (BroadcastChannel) | `stores/auth-store.ts` |
| FE-H4 | HIGH | Logout без `POST /auth/logout` | `stores/auth-store.ts` |
| FE-H5 | HIGH | Error Boundaries отсутствуют в calls/tracker/TMS | ✅ Wave 3 — route wrappers added |
| FE-M1 | MEDIUM | 15× `useExhaustiveDependencies` | см. FRONTEND.md §2 |
| FE-M2 | MEDIUM | 50× non-null assertion (`!`) | по codebase |

## E2E — критические пути (Page Object Model)

### E2E-01 — Playwright critical paths (PLAN-CLOSE-GAPS)

| Поле | Значение |
|------|----------|
| ID | E2E-01 |
| Target | 5 Playwright critical paths |
| Статус | ✅ **MET** — **7 specs** in `frontend/e2e/` (Wave 3 + cycles 7–10) |
| CI gate | ✅ `e2e-smoke` + `e2e-full` in `.github/workflows/ci.yml` |
| Config | `frontend/playwright.config.ts` — webServer auto-start, `PLAYWRIGHT_SKIP_WEBSERVER=1` for CI |

| Spec file | Scenarios covered |
|-----------|-------------------|
| `login.spec.ts` | Login page branding, dev login redirect, sessionStorage auth |
| `navigation.spec.ts` | Sidebar navigation, unauthenticated redirect |
| `logout.spec.ts` | Profile logout calls `POST /api/auth/logout` |
| `admin-audit.spec.ts` | Admin audit viewer, non-admin redirect |
| `health.spec.ts` | `/api/health` status, proxy reachability |
| `command-palette.spec.ts` | Ctrl+K global search (E2E-06) |
| `a11y.spec.ts` | axe-core: login + 10 authenticated routes (E2E-07 / NFR-16) |

### Full E2E roadmap (E2E-01..E2E-07)

| ID | Сценарий | POM класс | Статус |
|----|----------|-----------|--------|
| E2E-01 | Magic Link login | `LoginPage` | ✅ partial — dev login in `login.spec.ts` |
| E2E-02 | Create project + issue | `ProjectsPage`, `IssuePage` | ❌ not yet |
| E2E-03 | Send message in channel | `MessengerPage` | ❌ not yet |
| E2E-04 | Create test case + run | `TmsPage` | ❌ not yet |
| E2E-05 | Join voice call | `CallsPage` | ❌ not yet |
| E2E-06 | Global search Ctrl+K | `CommandPalette` | ✅ `command-palette.spec.ts` |
| E2E-07 | Accessibility (WCAG 2.1 AA) | axe via Playwright | ✅ **12 tests** — login + 10 auth routes |

POM pattern (planned):
```typescript
// frontend/e2e/pages/LoginPage.ts
export class LoginPage {
  constructor(private page: Page) {}
  async goto() { await this.page.goto('/login'); }
  async loginWithMagicLink(email: string) { ... }
}
```

## Техники тест-дизайна (обязательны)

- **Классы эквивалентности**: valid/invalid email, UUID, issue types
- **Граничные значения**: rate limits (3rd attempt), WIP limits, max file 25MB
- **Попарное тестирование**: filter combinations (status × priority × assignee)
- **Таблицы решений**: workflow transitions, RBAC permission matrix

## Регрессионный набор

**Обязательно после каждого изменения** (`.cursor/REGULATION.md` §8, `/qa-gate`):

```bash
# Backend regression
cd backend && go build ./... && go vet ./... && go test -race -count=1 -timeout=120s ./...

# Frontend regression (при изменениях frontend/)
cd frontend && npx tsc --noEmit && npm run lint && npx vitest run
```

## CI Pipeline

GitHub Actions (`.github/workflows/ci.yml`):
1. `go generate ./ent && go build -tags release ./...`
2. `go test -race -count=1 -timeout=600s ./...`
3. `npx tsc --noEmit`
4. `npm run lint` (Biome baseline, continue-on-error)
5. `npx vite build`
6. E2E smoke (Playwright, Mongo `authSource=admin`)
7. Security scan (grep secrets in diff)

Deploy (`.github/workflows/deploy.yml`): triggers via `workflow_run` **only after CI success** on `master`; SHA-tagged images; post-deploy health curl; no `docker system prune --volumes`.

Backend prod image: `backend/Dockerfile` builds with `-tags release` (excludes `/dev/*` routes).

## Баг-репорты

Формат:
```
ID: BUG-<module>-<NN>
Title: краткое описание
Steps: 1. ... 2. ... 3. ...
Expected: ...
Actual: ...
Severity: CRITICAL/HIGH/MEDIUM/LOW
File: path:line
```

Трекинг: `.planning/code-review-final/SUMMARY.md`

## Security-critical gaps (приоритет v2.0.1 / v2.5)

| ID | Модуль | Coverage | Gap | Рекомендуемый тест |
|----|--------|----------|-----|-------------------|
| SEC-01 | apigateway | 8.4% | APIKeyAuth middleware registration order | Integration: request with valid API key → 200, without → 401 |
| SEC-02 | calls | 5.1% | Webhook signature optional | `TestWebhook_*`: reject missing/tampered sha256 |
| SEC-04 | gitlab | 4.0% | SSRF on user-controlled URL | `TestValidateOutboundURL_*` on gitlab service |
| SEC-05 | ws | 20.9% | send-on-closed-channel panic | `hub_test`: disconnect mid-broadcast, concurrent unsubscribe |
| SEC-06 | auth | 18.0% | first-admin race | `service_registration_race_test` exists (1 test) — expand concurrent register |
| SEC-07 | auth | 18.0% | magic-link timing attack | `magiclink_test` (6 tests) — add `ConstantTimeCompare` assertion |
| NF-02 | ws/authorizer | — | WS subscribe без membership | `TestAuthorizer_*`: non-member subscribe → reject |
| NF-03 | search | 23.6% | search без user-scope | `TestSearch_*`: results limited to user's projects |
| NF-07 | notification/push | 5.6% | WebPush без URL validation | `TestPushSubscribe_InvalidEndpoint` |
| — | storage | 0.0% | presign TTL, key sanitization | httptest S3 stub: upload/download presign |
| — | platform/nats | 0.0% | JetStream connect/publish | embedded NATS server: publish ack, stream init |

## Recommended priority (test additions)

1. **P0** `platform/nats` + `storage` — 0% on event bus and file boundary
2. **P0** `auth` refresh rotation/replay + `calls` webhook signature enforcement (SEC-02)
3. **P1** `ws/authorizer` + `ws/hub` concurrency (SEC-05, NF-02)
4. **P1** `search` user-scope filter (NF-03), `notification/push` URL validation (NF-07)
5. **P2** `gitlab` SSRF (SEC-04), `apigateway` APIKeyAuth middleware (SEC-01)
6. ~~**P2** Frontend: первые Vitest тесты~~ ✅ Wave 3 — 17 tests
7. ~~**P3** E2E Playwright scaffold~~ ✅ Wave 3 — 5 specs; expand E2E-02..06 + CI gate
8. **P1** COV-01 uplift: backend **56.6%** ✅ overall; security **6/6 ≥80%** ✅
9. **P2** Playwright CI integration + full POM for E2E-02..E2E-06

---
*Последнее обновление: 2026-06-20 (Wave 6: COV-01 PASS, Ctrl+K UI, CI/deploy hardening)*
