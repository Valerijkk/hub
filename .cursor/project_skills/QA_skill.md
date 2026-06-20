# QA — Test Engineer

## Роль

Сквозное качество Hub: test plan, регрессия, покрытие, баг-репорты.

## Обязательные docs/

- `docs/test_plan.md` — основной артефакт QA
- `docs/vision_and_scope.md` — acceptance criteria
- `docs/system_specification.md` — матрица требований
- `docs/security_guidelines.md` — security tests

## Шаги работы

1. Загрузить все файлы из `docs/`
2. Актуализировать `docs/test_plan.md`
3. Контролировать покрытие: unit ≥80% security-critical
4. Интеграционные тесты на стыках сервисов
5. Smoke + регрессионный набор (`/qa-gate`) — **обязательно после каждого изменения** (REGULATION §8)
6. E2E по POM для критических путей
7. Баг-репорты в формате BUG-<module>-<NN>
8. Передать QA_VER

## GSD Skills (использовать по ситуации)

| Задача | GSD Skill |
|--------|-----------|
| Тесты для фазы | `gsd-add-tests` |
| Валидация покрытия | `gsd-validate-phase` |
| Аудит → автофикс | `gsd-audit-fix` |
| UAT gaps | `gsd-audit-uat` |
| Quality gate | `/qa-gate` (Hub) |

Матрица: `.cursor/GSD-INTEGRATION.md`

## Test Triage (REGULATION §10) — ОБЯЗАТЕЛЬНО

При любом FAIL теста:

1. **Сначала** проверить: функционал работает правильно по `docs/system_specification.md`?
2. Код верен → исправить тест (mock, assertion, fixture)
3. Код неверен → исправить код, обновить тест если контракт изменился
4. Skill: `.cursor/skills/test-triage/SKILL.md`

**Не чинить тесты «вслепую» чтобы стали зелёными.**

## Техники тест-дизайна (обязательны)

- Классы эквивалентности
- Граничные значения
- Попарное тестирование
- Таблицы решений

## Верификация

| Статус | Дата | Верификатор |
|--------|------|-------------|
| verified: false | — | QA_VER |

## Лог изменений

| Дата | Изменение | Файл |
|------|-----------|------|
| 2026-06-17 | Инициализация test_plan под регламент | `docs/test_plan.md` |
| 2026-06-17 | Wave 1 security tests: audit +20.8%, auth +23.3%, ws +28.8%, rbac +38% pkg cov | `service_security_test.go`, etc. |
| 2026-06-17 | GSD integration + test-triage protocol | `.cursor/GSD-INTEGRATION.md`, REGULATION §9-§10 |
| 2026-06-17 | Wave 2–4: Vitest **17 tests** (6 files), Playwright **5 specs**, BE coverage **~13%** | `frontend/src/**/*.test.ts(x)`, `frontend/e2e/` |

### Test triage note (Wave 1)

Новые тесты `audit/service_security_test.go`, `auth/service_session_test.go` — **ROOT=code** проверки (BanUser reason, refresh hash). При падении: сначала сверить с `docs/security_guidelines.md` NF-12, потом тест.

## QA Session Log — 2026-06-17

### Commands executed

```bash
cd backend && go test -race -count=1 ./...
cd backend && go test -cover ./internal/...
cd frontend && npx tsc --noEmit
cd frontend && npx vitest run
```

### Backend results

| Metric | Value |
|--------|-------|
| Test packages | 27 ok |
| Test files | 42 (`internal/**/*_test.go`) |
| Test functions | 323 (`^func Test`) |
| FAIL | 0 |
| Data races | 0 |
| Overall statement coverage | 11.6% |
| Longest package | `internal/calls` ~14s |
| Total wall time | ~68s |

### Frontend results

| Check | Result |
|-------|--------|
| `npx tsc --noEmit` | ✅ pass |
| `npx vitest run` | ❌ exit 1 — 0 test files (Vitest v3.2.4 configured) |
| Playwright E2E | ❌ not set up (no `tests/e2e/`) |

### Packages with zero `_test.go`

- `internal/platform/config`
- `internal/platform/cron`
- `internal/platform/database`
- `internal/platform/redis`

### Packages with tests but 0% statement coverage

- `internal/dashboard` — DTO-only tests
- `internal/dev` — seed struct validation only
- `internal/platform/nats` — constants only
- `internal/storage` — URL TTL constants only

### Security-critical coverage gaps (priority)

| Priority | Module | Finding ID | Action |
|----------|--------|------------|--------|
| P0 | platform/nats | — | embedded NATS: connect, publish, JetStream init |
| P0 | storage | — | presign upload/download, orphan cleanup |
| P0 | calls | SEC-02 | webhook sha256 mandatory rejection tests |
| P0 | auth | SEC-06/07 | expand race + timing tests; refresh rotation/replay |
| P1 | ws | SEC-05, NF-02 | hub concurrency + authorizer membership |
| P1 | search | NF-03 | user-scope filter in search results |
| P1 | notification | NF-07 | WebPush endpoint URL validation |
| P2 | gitlab | SEC-04 | SSRF on outbound URL |
| P2 | apigateway | SEC-01 | APIKeyAuth middleware integration |
| P2 | frontend | FE-01 | Vitest: auth refresh mutex, Rules of Hooks |
| P3 | frontend | — | Playwright POM E2E-01..06 |

### Delta vs `.planning/code-review-final/TESTS.md` (2026-05-15)

- Test functions: 318 → **323** (+5)
- Test files: 41 → **42**
- New test files: `ws/hub_test.go`, `auth/service_registration_race_test.go`, `apigateway/webhook_test.go`, `apigateway/webhook_signing_test.go`
- Coverage: 11.5% → **11.6%**
- Pass rate: still **0 fail, 0 races**

### Next QA actions

1. Track P0 test additions in v2.0.1 hotfix sprint
2. Add first frontend Vitest spec before v2.5 Phase 17
3. Re-run after CRITICAL fixes merged — expect coverage bump in auth/calls/ws

## QA Session Log — 2026-06-17 (Waves 2–4)

### Commands executed

```bash
cd backend && go test -race -count=1 ./...
cd backend && go test -cover ./internal/...
cd frontend && npx vitest run
cd frontend && npx playwright test   # scaffold only; needs running stack
```

### Backend results (Wave 2)

| Metric | Value |
|--------|-------|
| Test packages | 27 ok |
| Test functions | 323+ |
| FAIL | 0 |
| Data races | 0 |
| Overall statement coverage | **~13%** (up from 11.6% Wave 1) |
| New coverage | calls webhook, messenger rate, storage presign, platform/nats, apigateway SEC-01 |

### Frontend results (Wave 3)

| Check | Result |
|-------|--------|
| `npx vitest run` | ✅ **17 passed** — 6 files |
| Test files | `auth-store`, `client`, `admin`, `use-auth`, `error-boundary`, `audit-log` |
| Playwright E2E | ✅ **5 specs** in `frontend/e2e/` (login, navigation, logout, admin-audit, health) |
| ErrorBoundary | ✅ tracker, tms, calls, dashboard routes wrapped |

### Observability + docs (Wave 4)

| Check | Result |
|-------|--------|
| Audit `slow=1` filter | ✅ backend query param + admin UI checkbox |
| `docs/wiki/API.md` | ✅ +15 routes synced |
| `.planning/COVERAGE.md` | ✅ refreshed — COV-01 ~13% |

### Exit criteria progress (PLAN-CLOSE-GAPS)

| Criterion | Target | Current |
|-----------|--------|---------|
| FE-TEST | ≥10 Vitest smoke | ✅ **17 tests** |
| E2E-01 | 5 Playwright paths | ✅ **5 specs** (CI gate pending) |
| COV-01 | ≥50% stmt | ❌ **~13%** |
| REGRESSION | qa-gate all green | 🔄 Wave 5 |

### Next QA actions (Wave 5)

1. Run full `/qa-gate` — build + test -race + tsc + lint + vitest
2. Wire Playwright into CI with `PLAYWRIGHT_SKIP_WEBSERVER=1`
3. Verifier acceptance against PLAN-CLOSE-GAPS success criteria
4. Uplift security-critical packages toward 80% (auth, ws, rbac, calls, audit)
