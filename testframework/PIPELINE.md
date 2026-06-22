# Hub — Test Framework

Единая точка входа для тестов Hub. Структура повторяет **CI pipeline**: стадия → job → домен функционала.

## Карта пайплайна

```
testframework/
├── PIPELINE.md              ← вы здесь
├── pipeline/
│   ├── stages.yaml          # порядок стадий CI
│   └── jobs/                # 1 YAML = 1 CI job
│       ├── backend-test.yaml
│       ├── frontend-test.yaml
│       ├── e2e-smoke.yaml
│       ├── e2e-full.yaml
│       └── coverage-gate.yaml
└── stages/
    ├── unit/                # CI stage: unit (parallel)
    │   ├── backend-test/    # README → backend/testframework/{domain}/
    │   └── frontend-test/
    │       └── domains/{auth,tracker,tms,...}/
    ├── build/               # CI stage: build (без тестов)
    ├── e2e/                 # CI stage: e2e
    │   ├── shared/          # POM, steps, fixtures
    │   └── domains/{auth,admin,...}/
    └── quality/
        └── coverage-gate/   # README → backend coverage gate
```

| CI Job | Стадия | Путь к тестам | Домены |
|--------|--------|---------------|--------|
| `backend-test` | unit | `backend/testframework/{domain}/` | auth, tracker, tms, messenger, … |
| `frontend-test` | unit | `stages/unit/frontend-test/domains/{domain}/` | auth, admin, ai, tracker, v4, … |
| `build` | build | — | только сборка |
| `e2e-smoke` | e2e | `stages/e2e/domains/` (3 spec) | auth, platform |
| `e2e-full` | e2e | `stages/e2e/domains/` (20 spec) | auth, navigation, admin, … |
| `coverage-gate` | quality | `backend/testframework/` | все BE домены, ≥82% |

### Почему Go не в `stages/unit/backend-test/`?

Go **запрещает** импорт `backend/internal/*` из другого модуля. Тесты с `//go:build hubtest` и `export_hooks.go` обязаны жить в модуле `github.com/valerii/hub/backend`. Логически это job `backend-test`; физически — `backend/testframework/`.

## Команды (локально)

```bash
# Job: frontend-test
cd frontend && node scripts/link-testframework-node-modules.mjs && npm run test:run

# Job: e2e-smoke
cd frontend && npm run test:e2e:smoke

# Job: e2e-full
cd frontend && npm run test:e2e

# Job: backend-test
cd backend && go test -tags hubtest -race -count=1 ./testframework/...

# Один домен (frontend)
cd frontend && npx vitest run ../testframework/stages/unit/frontend-test/domains/auth

# Один домен (backend)
cd backend && go test -tags hubtest ./testframework/auth/...
```

## Паттерны

| Паттерн | Где |
|---------|-----|
| Page Object | `stages/e2e/shared/pages/` |
| Step | `stages/e2e/shared/steps/` |
| Builder | `backend/testframework/env/` |
| Factory | `backend/testframework/factory/` |
| TestRenderer | `stages/unit/frontend-test/render.tsx` |

Миграция: `MIGRATION.md` · Скрипт реорганизации: `scripts/reorganize-testframework-pipeline.ps1`
