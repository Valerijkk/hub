# coverage-gate

CI job: `coverage-gate` · Stage: `quality`

Запускается только на push в `master` (не на PR). Порог: **≥82%** statement coverage по `./internal/...`.

```bash
cd backend && go test -tags hubtest -coverprofile=cov.out -coverpkg=./internal/... ./testframework/...
```

Манифест: [`../../pipeline/jobs/coverage-gate.yaml`](../../pipeline/jobs/coverage-gate.yaml)
