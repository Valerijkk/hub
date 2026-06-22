# backend-test

CI job: `backend-test` · Stage: `unit`

Go-тесты по доменам функционала. Код в **`backend/testframework/`** (ограничение Go `internal/`).

```bash
cd backend && go test -tags hubtest -race ./testframework/auth/...
cd backend && go test -tags hubtest -race ./testframework/...
```

Манифест доменов: [`../../pipeline/jobs/backend-test.yaml`](../../pipeline/jobs/backend-test.yaml)

Shared fixtures: `backend/testframework/env/`, `backend/testframework/factory/`
