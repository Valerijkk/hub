# TASK-203 — canonical baseline attempt (2026-08-07)

**Verdict:** `BLOCKED`  
**Phase gate:** Phase 128 does not pass; phases 129–141 remain blocked by the
current contract.  
**Reason:** neither backend nor frontend produced a zero-failure baseline.
The Go coverage profile is therefore rejected and no coverage percentage from
this run is published in `.planning/COVERAGE.md`.

This is evidence-only. No product code, tests, CI, Compose files, thresholds,
or legacy paths were changed.

## Revision and environment contract

| Item | Recorded value |
|---|---|
| Branch | `cursor-grok-4.5` |
| Git HEAD | `f7fbdd7ebb23bfaba2aee71d503a7c0280846023` |
| Host OS | Microsoft Windows NT `10.0.26200.0` |
| Docker Desktop | `4.40.0 (187762)` |
| Docker Engine | `28.0.4`, API `1.48`, linux/amd64 |
| Docker Compose | `v2.34.0-desktop.1` |
| Compose source SHA-256 | `5675b61c21c38718d6499ee1dbdde88891c7440cdae7166685218afae8712825` |
| Resolved Compose SHA-256 | `ca116beb3ce89c0a86e4279629402c66e35e5b2400e04cd74a8a47d6470d1408` |
| Backend runner | `golang:1.25-bookworm`, Go `1.25.11` linux/amd64 |
| Backend image | `sha256:1dd0b0cab49abf58ae61bd6509b455375952dc17b7d8a5796905221b78e3531c` |
| Backend image digest | `golang@sha256:bbb255b0e131db500cf0520adc97441d2260cf629c7fa7e39e025ddf53995a24` |
| Local Go (not used for gate) | `go1.24.0 windows/amd64` |
| Node / npm | `v22.11.0` / `10.9.0` |
| TypeScript / Biome | `5.7.3` / `2.4.8` |
| Vitest / Playwright | `3.2.4` / `1.61.0` |

The worktree already contained unrelated uncommitted work when the baseline
started. The revision above identifies executable committed code; this report
does not claim a clean worktree.

### Test services

Only the four services from `deployments/docker-compose.test.yml` were
started. No container or volume was removed.

| Service | Image | Image ID | Observed state |
|---|---|---|---|
| PostgreSQL | `postgres:17` | `sha256:2d6ecff94c7417f93799039c2884c5d86515ec6e8289b1141ce6773de40f538d` | running, healthy |
| MongoDB | `mongo:8` | `sha256:b7e9809b5a1bf31577123d6704296cb5881617cf226251ef9afe4f6f1719b818` | running; no Compose healthcheck |
| Redis | `redis:7.4-alpine` | `sha256:487efc0616382465781b8fdc3d6d1db449e6fd80ae23bf48432a2da6b6929908` | running; no Compose healthcheck |
| NATS | `nats:2.11` | `sha256:8296541e734afc4b6f168616aecb56cd7168616d55b63d5eea8ed82ccb6411d6` | running; no Compose healthcheck |

Environment variables passed to the backend runner: `POSTGRES_TEST_URL` and
`MONGODB_TEST_URL`. Values are intentionally omitted. No production secret was
used.

## Backend canonical attempt

### Command contract

```text
docker compose -f deployments/docker-compose.test.yml up -d postgres mongo redis nats

docker run --rm --network deployments_default \
  --mount type=bind,source=<repo>/backend,target=/app \
  --mount type=bind,source=<temp-evidence>,target=/evidence \
  -w /app \
  -e POSTGRES_TEST_URL=<redacted> \
  -e MONGODB_TEST_URL=<redacted> \
  golang:1.25-bookworm sh -c \
  'go test -tags hubtest -race -count=1 -timeout=600s \
     -covermode=atomic -coverpkg=./internal/... \
     -coverprofile=/evidence/backend.cover.out ./testframework/...'
```

An earlier operator attempt used `sh -lc`, which reset the image `PATH` and
ended before `go test` started (`go: not found`). It is not counted as a product
run. The command above uses the image's normal non-login shell.

### Result

| Metric | Value |
|---|---:|
| Exit code | `1` |
| Wall time | `194 s` |
| Packages PASS | `40` |
| Packages FAIL | `10` |
| Packages with no test files | `3` |
| Failing tests | `12` |
| Coverage verdict | **REJECTED — failed packages** |

The partial profile exists, begins with `mode: atomic`, contains 72,324 lines,
and is parseable by `go tool cover`. It is not a baseline because the run is
red. No overall or per-package coverage is reported from it.

| Temporary artifact | Size | SHA-256 | Status |
|---|---:|---|---|
| `C:\tmp\hub-task203\backend-go-test.log` | 29,117 bytes | `b7f0d2cb6d9f83fdb9f47291fc06b1d1f28f3fc279815cd5ebcd0aa9fdeee7ad` | diagnostic log |
| `C:\tmp\hub-task203\backend.cover.out` | 5,763,165 bytes | `8ed0838c4aef01441c54b83ba39750ddaffaa767251e55f7975181a1a6083cef` | partial; rejected |

### Failing backend tests

| # | Package / test | Exact primary failure |
|---:|---|---|
| 1 | `accounting::TestEntRepository_JournalLineJoinFilter_A2W8_3_004` | `open /backend/internal/accounting/repository.go: no such file or directory` |
| 2 | `calls::TestEntRepository_RecordingSoftDelete_A2W8_4_005` | `open /backend/internal/calls/repository.go: no such file or directory` |
| 3 | `ent/task167::TestSchema_TASK167_ChannelMemberLWW_NoVersionMixin` | `open /.planning/decisions/ADR-ENT-CHANNELMEMBER-LWW-2026.md: no such file or directory` |
| 4 | `github::TestEntRepository_IntegrationSoftDelete_W7_A3_010` | `open /backend/internal/github/repository.go: no such file or directory` |
| 5 | `gitlab::TestHandler_RegisterRoutes_WebhookAndIntegrations` | `test: timeout error after 1s` |
| 6 | `gitlab::TestEntRepository_IntegrationSoftDelete_W7_A3_010` | `open /backend/internal/gitlab/repository.go: no such file or directory` |
| 7 | `hr::TestEntRepository_LeaveSoftDelete_A2W8_3_001` | `open /backend/internal/hr/repository.go: no such file or directory` |
| 8 | `intranet::TestEntRepository_KudoSoftDelete_A2W8_3_004` | `open /backend/internal/intranet/repository.go: no such file or directory` |
| 9 | `platform/nats::TestNewClient_ConnectAndClose` | assertion at `client_integration_test.go:43`: `Should be true` |
| 10 | `sprint::TestEntRepository_SprintSoftDelete_A2W8_4_004` | `open /backend/internal/sprint/repository.go: no such file or directory` |
| 11 | `tracker::TestEntRepository_LabelSoftDelete_A2W8_3_005` | `open /backend/internal/tracker/repo_labels.go: no such file or directory` |
| 12 | `tracker::TestEntRepository_CustomFieldSoftDelete_A2W8_2_004` | `open /backend/internal/tracker/repo_custom_fields.go: no such file or directory` |

Ten failures are path-sensitive guards: nine assume a `/backend` mount and one
assumes `/.planning`. The documented canonical Docker command mounts only
`backend/` at `/app`, so these tests cannot pass under the current environment
contract. GitLab timeout and NATS assertion are separate failures and require
test-triage against the implementation contract before any fix.

## Frontend canonical attempt

Commands were executed from `frontend/` in the documented order:

```text
node scripts/link-testframework-node-modules.mjs
npx tsc --noEmit
npm run lint
npm run test:run
```

The link command first hit host sandbox `EPERM` on `C:\Users\Valerii`; the same
documented command succeeded outside the sandbox. The sandbox error is not
counted as a product failure.

| Step | Result | Wall time / counts |
|---|---|---|
| link testframework modules | PASS | 60 ms |
| TypeScript `--noEmit` | PASS | 21.766 s |
| Biome `check ./src` | FAIL | 1.755 s; 370 files, 310 errors, 4 warnings |
| Vitest default reporter | FAIL | 41.2 s |
| Vitest JSON evidence rerun | FAIL | 38.704 s; same suite semantics, reporter only |

Vitest JSON result:

| Metric | Value |
|---|---:|
| Test files | `180` |
| Files PASS / FAIL / pending | `176 / 4 / 0` |
| Tests | `517` |
| Tests PASS / FAIL / skipped / todo | `510 / 7 / 0 / 0` |
| Nested suites PASS / FAIL / pending | `371 / 8 / 0` |
| JSON artifact | `C:\tmp\hub-task203\vitest.json` |
| JSON size / SHA-256 | 221,134 bytes / `d6c2ea41de9dfc2b3a0dd3ad55d3fd89c2d359e65d23a2b2366414c5b80ef9a5` |

### Failing frontend tests

| # | File / test | Primary failure |
|---:|---|---|
| 1 | `calls/upcoming-calls-join.test.tsx` — `calls joinCall(call.id) on Войти, not startCall` | accessible button named `Войти` not found |
| 2 | `calls/upcoming-calls-join.test.tsx` — `joins even when channel_id is missing from dashboard DTO` | accessible button named `Войти` not found |
| 3 | `dashboard/system-health.test.tsx` — `renders service names and healthy status` | text `postgres` not found; empty state rendered |
| 4 | `dashboard/system-health.test.tsx` — `renders unhealthy service label` | text `nats` not found; empty state rendered |
| 5 | `dashboard/system-health.test.tsx` — `renders degraded service label` | text `mongodb` not found; empty state rendered |
| 6 | `platform/error-empty-auth-p1.test.tsx` — `ErrorBoundary resets when resetKey changes` | `toHaveTextContent` assertion failed |
| 7 | `tracker/pipeline-status-sanitize.test.tsx` — `source guard: hub tokens` | source still matches `text-(green|red)-600` |

## E2E decision

`e2e-full` is `NOT_RUN`, not PASS and not skipped. Read-only
`docker compose -f docker-compose.local.yml ps --format json` returned no Hub
application services. The test Compose stack provides databases/brokers only;
there was no documented running backend, frontend, LiveKit, or authenticated
seed. Starting Playwright in that state would measure missing prerequisites,
not the product flow.

## Sentinel and handoff

- Empty or malformed collection/profile is rejected.
- This Go profile is non-empty and syntactically valid but still rejected
  because ten packages failed.
- Required frontend lint and Vitest are red; required E2E is `NOT_RUN`.
- `.planning/COVERAGE.md` remains unchanged because TASK-203 permits updating
  it only after a valid zero-failure result.
- `qa-auto-ver` should independently rerun after test triage resolves the path
  contract, GitLab timeout, NATS assertion, frontend failures, and E2E stack
  prerequisites.

**Skill log:** `gsd-execute-phase` contract read; its referenced external
workflow file is absent in this environment, so the task's embedded gates were
executed directly. Result: `FAIL/BLOCKED`.
