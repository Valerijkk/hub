# Hub — План тестирования

> Единая база знаний QA. Обновляется при изменении стратегии, CI или coverage floor.  
> Детали TC-*: `docs/TESTING.md` · Матрица: `.planning/COVERAGE.md` · Triage: `.cursor/skills/test-triage/SKILL.md`

---

## Пирамида тестирования

```
        ┌─────────────┐
        │     E2E     │  Playwright — критические user paths (CI e2e-full)
        ├─────────────┤
        │ Integration │  httptest handlers, enttest/sqlite, Mongo в CI
        ├─────────────┤
        │    Unit     │  Go testify (-tags hubtest) + Vitest/RTL
        └─────────────┘
```

| Уровень | Инструмент | Расположение | Gate |
|---------|------------|--------------|------|
| Unit (BE) | Go testing + testify | `backend/internal/*_test.go`, `backend/testframework/` | CI `backend-test` |
| Unit (FE) | Vitest + RTL | `frontend/testframework/unit/`, `frontend/src/**/*.test.ts` | CI `frontend-test` |
| Integration | httptest, enttest, miniredis | `backend/internal/*/` | В составе unit |
| API smoke | httptest | auth, health, critical handlers | e2e-full |
| E2E | Playwright + POM + Steps | `testframework/e2e/specs/` | `e2e-smoke` + `e2e-full` |
| Build | `tsc`, `vite build`, `-tags release` | CI `build` job | 0 errors |
| Lint | Biome | `npm run lint` | **gate** (не continue-on-error) |
| Security | grep + audit docs | `.planning/code-review-final/` | 0 CRITICAL |

---

## Test Triage Protocol (REGULATION §10)

При падении любого теста (Go / Vitest / Playwright):

1. **Воспроизвести функционал** по `docs/system_specification.md`, не только assert
2. **Классифицировать:** `ROOT=code` (баг в prod) | `ROOT=test` (устаревший тест)
3. **Чинить минимальный слой** — код или тест, не оба «наугад»
4. **Записать** обоснование в QA log / STATE.md

**Запрещено:** зелёные тесты без проверки соответствия спецификации.

---

## Текущие метрики (2026-06-21)

Источник: `.planning/COVERAGE.md` Phase 120+, локальный прогон Vitest 2026-06-21.

| Метрика | Значение |
|---------|----------|
| Backend `*_test.go` | **200+** файлов |
| Statement coverage `internal/` | **~82.1%** (floor Phase 120) |
| `internal/auth/` | **~81.6%** |
| `internal/tracker/` | **~81.2%** |
| `internal/tms/` | **~80.6%** |
| `internal/search/` | **~92.2%** |
| `internal/storage/` | **~86.8%** |
| Security-critical packages ≥80% | **6/6** ✅ |
| Frontend Vitest | **25** files, **56** tests ✅ (локальный прогон) |
| Playwright specs | **20** files в `testframework/e2e/specs/` |
| TypeScript | `tsc --noEmit` — 0 errors (CI) |
| Data races | 0 (CI `-race`) |

### Команды регрессии

```bash
# Backend (как CI)
cd backend
go generate ./ent
go vet ./...
go test -tags hubtest -race -count=1 -timeout=600s ./internal/... ./testframework/...

# Prod build check
go build -tags release -o bin/server ./cmd/server

# Frontend (как CI)
cd frontend
npx tsc --noEmit
npm run lint
npx vitest run
npm run build

# E2E локально (из frontend/, нужен backend + infra)
cd frontend
npm run test:e2e   # link-e2e-node-modules.mjs + playwright.config.ts → testframework/e2e/specs/
```

---

## CI Pipeline (`.github/workflows/ci.yml`)

| Job | Зависимости | Содержание |
|-----|-------------|------------|
| `backend-test` | — | Go **1.25**, Mongo **8** service, `go generate`, `go vet`, `go test -tags hubtest -race` 600s |
| `frontend-test` | — | Node **22**, `tsc`, `vitest run`, Biome lint |
| `build` | backend + frontend test | `go build -tags release`, `npm run build` |
| `e2e-smoke` | frontend-test | Playwright без backend: `login`, `a11y`, `critical-gate` |
| `e2e-full` | backend + frontend test | Infra script + backend `DEV_MODE` + 20 spec files |

**Deploy** (`.github/workflows/deploy.yml`): только после успешного CI на push `master`; образы GHCR по SHA; **без** `DEV_MODE=true`.

E2E конфиг: `frontend/playwright.config.ts` (спеки: `testframework/e2e/specs/`)  
Module resolution: `frontend/scripts/link-e2e-node-modules.mjs` (symlink `node_modules` для specs вне `frontend/`)  
Infra: `scripts/ci-e2e-backend.sh` (postgres:17, redis:7.4, mongo:8, nats:2.11).

### e2e-full spec list (CI)

`login`, `critical-gate`, `navigation`, `logout`, `health`, `admin-audit`, `admin-ai`, `admin-metrics`, `ai-hub`, `automation`, `metrics`, `command-palette`, `shortcuts`, `onboarding-profile`, `notifications`, `project-lifecycle`, `messenger`, `tms`, `a11y`, `v4-superapp`.

---

## Критерии приёмки (общие)

1. Все CI jobs green на PR
2. Новый код покрыт unit-тестами (граничные случаи, error paths)
3. Нет регрессии по `docs/security_guidelines.md`
4. API changes → `docs/system_specification.md` + `docs/wiki/API.md`
5. User-facing critical path → E2E spec или расширение существующего
6. Docs sync: `STATE.md` + матрица из rule 55-docs-sync

---

## Матрица модулей (краткая)

| Модуль | Unit | E2E | Security |
|--------|------|-----|----------|
| auth | ✅ ~82% | login, logout | SEC-06, SEC-07, magic-link timing |
| tracker | ✅ ~81% | project-lifecycle | optimistic lock |
| messenger | ✅ ~76% | messenger.spec | NF-02 WS authz |
| tms | ✅ ~81% | tms.spec | webhook API key |
| calls | ✅ | partial | SEC-02 webhook sha256 |
| ws | ✅ | real-time | SEC-05, NF-02 |
| search | ✅ ~92% | command-palette | NF-03 user scope |
| apigateway | ✅ | — | SEC-01 API key order |
| notification | ✅ ~82% | notifications | NF-07 WebPush URL |
| v4 (crm, hr, intranet…) | ✅ | v4-superapp | RBAC |

Полная матрица: `.planning/COVERAGE.md`.

---

## E2E — критические пути

| ID | Сценарий | Spec | Статус |
|----|----------|------|--------|
| E2E-01 | Magic Link / dev login | `login.spec.ts` | ✅ |
| E2E-02 | Create project + issue | `project-lifecycle.spec.ts` | ✅ |
| E2E-03 | Messenger send | `messenger.spec.ts` | ✅ |
| E2E-04 | TMS flow | `tms.spec.ts` | ✅ |
| E2E-05 | Voice call | — | ⚠️ partial (LiveKit infra) |
| E2E-06 | Ctrl+K search | `command-palette.spec.ts` | ✅ |
| E2E-07 | a11y axe | `a11y.spec.ts` | ✅ |

POM/steps: `testframework/e2e/pages/`, `testframework/e2e/steps/`.

---

## Техники тест-дизайна

- **Классы эквивалентности:** valid/invalid email, UUID, issue types
- **Граничные значения:** rate limits (4-й auth → 429), WIP limits, file 25MB
- **Попарное тестирование:** filter combinations (status × priority × assignee)
- **Таблицы решений:** workflow transitions, RBAC permission matrix

---

## Security-critical тесты (приоритет)

| ID | Модуль | Требуемый тест |
|----|--------|----------------|
| SEC-01 | apigateway | API key middleware order → 401 without key |
| SEC-02 | calls | Webhook reject missing/tampered sha256 |
| SEC-05 | ws | send after disconnect — no panic |
| SEC-06 | auth | concurrent register / first-admin lock |
| SEC-07 | auth | `ConstantTimeCompare` on magic link |
| NF-02 | ws | non-member subscribe → reject |
| NF-03 | search | results scoped to user projects/channels |
| NF-07 | notification | WebPush invalid endpoint → 400 |

Статус fixes: `docs/security_guidelines.md` (12/12 CRITICAL fixed).

---

## Баг-репорты

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

---

## Связанные документы

| Документ | Содержание |
|----------|------------|
| `docs/TESTING.md` | TC-AUTH-01…, API code matrix, manual checklist |
| `docs/stack_and_tools.md` | Версии, CI jobs |
| `.cursor/rules/56-regression-verify.mdc` | Обязательный цикл qa-gate |
| `.cursor/skills/qa-gate/SKILL.md` | Полный gate |

---
*Последнее обновление: 2026-06-22 — e2e: frontend/playwright.config.ts, 20 specs в testframework/e2e/specs/*
