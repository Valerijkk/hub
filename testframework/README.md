# Hub — Test Framework (`testframework/`)

Единая инфраструктура тестирования Hub: E2E, frontend unit, shared backend fixtures.

## Структура

```
testframework/                 — E2E + документация (repo root)
├── README.md
├── AUDIT.md
└── e2e/                       — Playwright POM + Steps + specs

frontend/testframework/        — Vitest (внутри frontend для node_modules)
├── render.tsx                 — TestRenderer
└── unit/                      — component tests

backend/testframework/         — Go Builder + Factory + Steps
```

## Паттерны

| Паттерн | Где | Назначение |
|---------|-----|------------|
| **Page Object** | `e2e/pages/` | Инкапсуляция локаторов и UI-действий |
| **Step** | `e2e/steps/` | Бизнес-шаги (login → projects) поверх POM |
| **Builder** | `backend/testframework/env` | Fluent Ent test env (DRY setup) |
| **Factory** | `backend/testframework/factory` | Создание User/Project/Issue |
| **TestRenderer** | `frontend/render.tsx` | QueryClient + i18n (DRY render) |
| **AAA** | все specs | Arrange → Act → Assert в каждом тесте |
| **SOLID** | framework code | SRP: page ≠ step ≠ spec; DIP: specs → steps → pages |

## Go unit-тесты

`*_test.go` остаются рядом с кодом (`backend/internal/*`) — **требование Go** для white-box тестов.
Общий setup выносится в `backend/testframework` (Builder/Factory), тесты импортируют фреймворк.

## Команды

```bash
# Frontend unit (Vitest)
cd frontend && npx vitest run

# E2E (Playwright)
cd frontend && npm run test:e2e

# Backend
cd backend && go test -race -count=1 ./internal/... ./testframework/...
```

## Test design techniques

- **Equivalence partitioning** — valid/invalid email, magic code, RBAC roles
- **Boundary values** — rate limits (3/10m), file 25MB, pagination limits
- **State transition** — issue workflow, onboarding steps
- **Error guessing** — 401/403/404/409 paths in handler tests
- **Pairwise** — TMS environment labels (dev/staging/prod/custom)

См. `AUDIT.md` для матрицы соответствия по модулям.
