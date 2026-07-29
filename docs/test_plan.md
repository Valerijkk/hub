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
| Unit (BE) | Go testing + testify | `backend/testframework/{domain}/` (0 в `internal/`) | CI `backend-test` |
| Unit (FE) | Vitest + RTL | `testframework/stages/unit/frontend-test/domains/{domain}/` | CI `frontend-test` |
| Integration | httptest, enttest, miniredis | `backend/testframework/` | В составе unit |
| API smoke | httptest | auth, health, critical handlers | e2e-full |
| E2E | Playwright + POM + Steps | `testframework/stages/e2e/domains/{domain}/` | `e2e-smoke` + `e2e-full` |
| Build | `tsc`, `vite build`, `-tags release` | CI `build` job | 0 errors |
| Lint | Biome | `npm run lint` | **gate** (не continue-on-error) |
| Security | grep + audit docs | `.planning/audits/code-review-final/` | 0 CRITICAL |

---

## Test Triage Protocol (REGULATION §10)

При падении любого теста (Go / Vitest / Playwright):

1. **Воспроизвести функционал** по `docs/system_specification.md`, не только assert
2. **Классифицировать:** `ROOT=code` (баг в prod) | `ROOT=test` (устаревший тест)
3. **Чинить минимальный слой** — код или тест, не оба «наугад»
4. **Записать** обоснование в QA log / STATE.md

**Запрещено:** зелёные тесты без проверки соответствия спецификации.

---

## Текущие метрики (2026-06-22)

Источник: `.planning/COVERAGE.md` · `.planning/audits/2026-06/AUDIT-SCORECARD.md`

| Метрика | Значение |
|---------|----------|
| Backend `testframework/**/*_test.go` | **805** файлов (measured 2026-07-29; was 515 → 580 → 652) |
| Backend `internal/*_test.go` | **0** (TF-M08 ✅) |
| Backend `export_hooks.go` (`//go:build hubtest`) | **39** файлов (measured 2026-07-29) |
| Statement coverage `internal/` | **83.0%** (2026-06-22 Docker baseline) |
| Security-critical packages ≥80% | **6/6** ✅ (auth, ws, rbac, audit, calls, platform/middleware) |
| Frontend Vitest | **167** files в `stages/unit/frontend-test/` (measured 2026-07-29; 61 `.ts` + 106 `.tsx`) |
| Playwright specs | **20** files в `testframework/stages/e2e/domains/` |
| Lighthouse CI | artifact job (master push, non-blocking) |
| TypeScript | `tsc --noEmit` — 0 errors (CI) |
| Data races | 0 (CI `-race`) |

### Команды регрессии

```bash
# Backend (как CI — только testframework после TF-M08)
cd backend
go generate ./ent
go vet -tags hubtest ./...
go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...

# Prod build check
go build -tags release -o bin/server ./cmd/server

# Frontend (как CI)
cd frontend
node scripts/link-testframework-node-modules.mjs
npx tsc --noEmit
npm run lint
npm run test:run
npm run build

# E2E локально (из frontend/, нужен backend + infra)
cd frontend
npm run test:e2e:smoke   # job e2e-smoke (--project=chromium-smoke)
npm run test:e2e         # job e2e-full (--project=chromium-full)
```

---

## CI Pipeline (`.github/workflows/ci.yml`)

| Job | Зависимости | Содержание |
|-----|-------------|------------|
| `backend-test` | — | Go **1.25**, Mongo **8** service, `go generate`, `go vet -tags hubtest`, `go test -tags hubtest -race ./testframework/...` 600s |
| `frontend-test` | — | Node **22**, link testframework deps, `tsc`, `vitest run`, Biome lint |
| `build` | backend + frontend test | `go build -tags release`, `npm run build` |
| `e2e-smoke` | frontend-test | Playwright без backend: `login`, `a11y`, `critical-gate` |
| `e2e-full` | backend + frontend test | Infra script + backend `DEV_MODE` + 20 spec files |
| `coverage-gate` | master push only | Docker infra, ≥82% statement coverage |
| `lighthouse` | master push only | Login page LH audit, artifact (non-blocking) |

**Deploy** (`.github/workflows/deploy.yml`): только после успешного CI на push `master`; образы GHCR по SHA; **без** `DEV_MODE=true`.

E2E конфиг: `frontend/playwright.config.ts` (спеки: `testframework/stages/e2e/domains/{domain}/`)  
Module resolution: `frontend/scripts/link-testframework-node-modules.mjs`  
Pipeline manifests: `testframework/pipeline/jobs/*.yaml`  
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

Покрытие unit — **2026-06-22 Docker baseline** (`.planning/COVERAGE.md`).

| Модуль | Unit | E2E | Security |
|--------|------|-----|----------|
| auth | ✅ 82.4% (2026-06-22) | login, logout | SEC-06, SEC-07, magic-link timing |
| tracker | ✅ 82.4% (2026-06-22) | project-lifecycle | optimistic lock |
| messenger | ✅ 82.3% (2026-06-22) | messenger.spec | NF-02 WS authz |
| tms | ✅ 82.0% (2026-06-22) | tms.spec | webhook API key |
| calls | ✅ | partial | SEC-02 webhook sha256 |
| ws | ✅ | real-time | SEC-05, NF-02 |
| search | ✅ ~92% (2026-06-22) | command-palette | NF-03 user scope |
| apigateway | ✅ | — | SEC-01 API key order |
| notification | ✅ ~82% (2026-06-22) | notifications | NF-07 WebPush URL |
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

POM/steps: `testframework/stages/e2e/shared/pages/`, `testframework/stages/e2e/shared/steps/`.

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

Трекинг: `.planning/audits/code-review-final/SUMMARY.md`

---

## Связанные документы

| Документ | Содержание |
|----------|------------|
| `docs/TESTING.md` | TC-AUTH-01…, API code matrix, manual checklist |
| `docs/stack_and_tools.md` | Версии, CI jobs |
| `.cursor/rules/56-regression-verify.mdc` | Обязательный цикл qa-gate |
| `.cursor/skills/qa-gate/SKILL.md` | Полный gate |

---
*Последнее обновление: 2026-07-29 — inventory **805** BE `*_test.go` / **39** export_hooks / Vitest **167** files; COV 83% остаётся baseline прогона 2026-06-22*
