# Hub — Test Framework (`testframework/`)

> Полная карта: [`PIPELINE.md`](PIPELINE.md)

Структура повторяет CI: **стадия → job → домен**.

```
testframework/
├── pipeline/jobs/           # манифесты CI jobs
└── stages/
    ├── unit/
    │   ├── backend-test/    → backend/testframework/{domain}/
    │   └── frontend-test/domains/{domain}/
    ├── e2e/
    │   ├── shared/          # POM, steps, fixtures
    │   └── domains/{domain}/
    └── quality/coverage-gate/
```

Go-тесты физически в `backend/testframework/` — ограничение Go `internal/`. Логически это job `backend-test`.

```bash
cd frontend && npm run test:run              # job: frontend-test
cd frontend && npm run test:e2e:smoke        # job: e2e-smoke
cd frontend && npm run test:e2e              # job: e2e-full
cd backend && go test -tags hubtest ./testframework/...  # job: backend-test
```
