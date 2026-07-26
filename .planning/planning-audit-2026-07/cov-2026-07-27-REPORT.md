# Coverage gate — measured 2026-07-27 (Docker, full hubtest)

**Method:** `docker run --network hub_default golang:1.25-bookworm`  
`go test -tags hubtest -count=1 -timeout=900s -covermode=atomic -coverpkg=./internal/... ./testframework/...`  
then `go tool cover -func` → statement **overall**.

| Metric | Value | Date |
|--------|------:|------|
| Overall `internal/` statements | **81.9%** | **2026-07-27** |
| Baseline overall | 83.0% | 2026-06-22 |
| **Delta** | **-1.1 pp** | — |

## Comparable decline (statement overall only)

The **only** figure that is method-comparable to the 2026-06-22 COVERAGE.md package table is the **overall** statement total from `go tool cover`.

A second Docker attempt to persist a full `coverprofile` for per-package statement totals failed mid-suite (`proxy.golang.org` EOF). A third run with host `GOMODCACHE` mounted was started to obtain package statement %; until that artifact lands, **do not treat func-mean package numbers as comparable to the 2026-06-22 statement baselines**.

## Packages — provisional func-mean (NOT comparable to 2026-06-22 stmt %)

Source: mean of per-function statement percentages in `go tool cover -func` output from the successful 2026-07-27 run. Useful for relative ranking inside this run only.

See: `cov-2026-07-27-func-avg.tsv`, full table in git history of this file's companion parse.

Notable within-run: `messenger` func-mean **74.6%** (lowest among large domains in this proxy).

## Command to re-measure

```bash
docker run --rm --network hub_default \
  -v "$PWD/backend:/app" -v "$TMP/hub-cov-out:/out" -w /app \
  -e POSTGRES_TEST_URL='postgresql://hub:hub@postgres:5432/hub?sslmode=disable' \
  -e MONGODB_TEST_URL='mongodb://hub:hub@mongo:27017/?authSource=admin' \
  golang:1.25-bookworm \
  sh -c 'go test -tags hubtest -count=1 -timeout=900s -covermode=atomic -coverpkg=./internal/... -coverprofile=/out/hub-cov.out ./testframework/... && go tool cover -func=/out/hub-cov.out | tee /out/hub-cov-func.txt'
```
