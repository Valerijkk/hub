# TASK-203 — canonical baseline attempt (2026-08-07)

**Verdict:** `RECORDED — LEGACY DISCOVERY BASELINE; MIGRATION GATE SUPERSEDED`
**Phase gate:** none. On 2026-08-07 the founder approved an immediate
greenfield cutover: retain only genuine native units, delete legacy non-unit
suites by exact manifest, and design all new non-unit suites in Python from
requirements and risks. Row parity and the historical 82% statement target do
not block that cutover.
**Final evidence:** backend is zero-failure with exact merged coverage
**82.0373%** and all **6/6** security packages at or above 80%; Docker frontend
typecheck and Biome pass; the frozen legacy Vitest diagnostic is **510/519**;
Docker E2E smoke is **3/3 PASS** and full with retries disabled is **48 PASS / 5
expected skip / 0 fail**.

This file preserves the initial red attempts below as historical triage
evidence. Product/test repairs were performed and committed by their own
TASK-221…225 atomic scopes; the baseline measurement itself did not change CI,
Compose, thresholds or legacy paths.

## Final backend baseline after remediation

The canonical Docker Go 1.25 command was rerun after TASK-225 and the
deterministic cron cancellation repair. The named task runner completed with
exit code `0`; cron and NATS both passed inside the full race-enabled suite.

| Metric | Final value |
|---|---:|
| Runtime | `119 s` (`2026-08-07T15:21:16Z` → `15:23:15Z`) |
| Packages PASS / FAIL / no-test-files | `50 / 0 / 3` |
| Overall exact source-range coverage | **82.0373%** (`21,269 / 25,926`) |
| Overall floor `>=82%` | **PASS (+0.0373 pp)** |
| Security packages `>=80%` | **6/6 PASS** |
| Profile SHA-256 | `7C20BF7452AD9387D96F548368830282FD2A50361A0D0106E7B8069FC4C99856` |

The profile was written outside tracked source at
`.gotmp/task203-rerun-20260807/backend-final.cover.out`. `go tool cover`
rounds the total to `82.0%`; the gate verdict uses exact merged covered/total
source ranges from the same non-empty atomic profile.

| Package | Covered / total | Exact coverage | Floor `>=80%` |
|---|---:|---:|:---:|
| `auth` | 686 / 851 | **80.6110%** | PASS |
| `ws` | 408 / 443 | **92.0993%** | PASS |
| `rbac` | 337 / 387 | **87.0801%** | PASS |
| `audit` | 928 / 1,083 | **85.6879%** | PASS |
| `calls` | 1,059 / 1,281 | **82.6698%** | PASS |
| `platform/middleware` | 207 / 245 | **84.4898%** | PASS |

## Final frontend Docker observation

The resource-limited canonical Docker sequence produced:

| Gate | Final result |
|---|---:|
| TypeScript `tsc --noEmit` | PASS |
| Biome `check ./src` | PASS — 370 files, 0 diagnostics |
| Legacy Vitest | **510/519 PASS, 9 FAIL** |

Six Vitest failures are caused by the Docker harness not mounting `/work/docs`,
two nested `tsc` fixtures time out under the 2-CPU limit, and one assertion uses
a stale i18n key. These failures are frozen discovery evidence. The founder
cancelled further legacy remediation in favor of greenfield Python suites.

## Revision and environment contract

| Item | Recorded value |
|---|---|
| Branch | `cursor-grok-4.5` |
| Git HEAD | `f7fbdd7ebb23bfaba2aee71d503a7c0280846023` |
| Post-triage backend rerun HEAD | `1f41dcbd14366d93fc47acf8da86c97d06c1ab28` |
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

### NATS focused reproduction

The NATS failure was materially refined with two post-baseline diagnostic runs
in fresh Go 1.25.11 containers. Source, tests, services, and Compose state were
not changed.

| Scope | Exact test arguments | Result | Wall time |
|---|---|---|---:|
| Single failing test | `go test -tags hubtest -race -count=20 -run '^TestNewClient_ConnectAndClose$' ./testframework/platform/nats` | PASS, 20/20 | 33 s |
| Whole NATS package | `go test -tags hubtest -race -count=5 ./testframework/platform/nats` | PASS, 5/5 | 74 s |

| Temporary diagnostic log | SHA-256 |
|---|---|
| `C:\tmp\hub-task203\nats-focused-count20.log` | `67d1106134164c4d374f830aaa30089048fbfe0a1ac90e67411ef48abd69e8e3` |
| `C:\tmp\hub-task203\nats-package-count5.log` | `3aec302a1c0be3b847cae6fe2edad0d22df4bf41b97e87d1f2c96261ec2eea14` |

The original NATS assertion is not reproducible in isolated or package-level
repetition. This supports suite-level contention/order as a triage hypothesis;
it does not prove the root cause and does not make the red full-suite baseline
valid.

## Backend canonical rerun after triage

The exact canonical backend command was repeated after commits `04283387`
(portable source guards) and `1f41dcbd` (deterministic GitLab handler fixture).
Both commits are ancestors of the rerun HEAD.

```text
go test -tags hubtest -race -count=1 -timeout=600s \
  -covermode=atomic -coverpkg=./internal/... \
  -coverprofile=/evidence/backend-rerun-1f41dcbd.cover.out \
  ./testframework/...
```

| Metric | Value |
|---|---:|
| Exit code | `0` |
| Wall time | `206 s` |
| Packages PASS / FAIL / no-test-files | `50 / 0 / 3` |
| Failing tests | `0` |
| Declared top-level tests in TASK-202 ledger | `3,663` |
| Profile mode / lines | `atomic` / `72,324` |
| Unique coverage blocks | `18,832` |
| Covered / total statements | `21,206 / 25,924` |
| Overall `internal/...` | **81.8%** |
| Overall gate ≥82% | **MISS (-0.2 pp)** |

The canonical non-JSON command does not emit executed/skipped test totals;
`3,663` is the independently inventoried declared top-level count, not a claim
that every declaration executed without a runtime skip. Package and failure
counts above come from this run's result lines.

The full profile is non-empty and parsed successfully with Go 1.25.11
`go tool cover -func`. Repeated source blocks produced by multi-package
`-coverpkg` were merged by source range with covered=true when any counter was
non-zero. The merged overall exactly matches `go tool cover`: 81.8%.

### Security packages from the valid full profile

| Package | Covered / total statements | Coverage | Floor ≥80% |
|---|---:|---:|:---:|
| `auth` | 683 / 851 | **80.3%** | PASS |
| `ws` | 408 / 443 | **92.1%** | PASS |
| `rbac` | 337 / 387 | **87.1%** | PASS |
| `audit` | 861 / 1,083 | **79.5%** | MISS |
| `calls` | 1,059 / 1,281 | **82.7%** | PASS |
| `platform/middleware` | 207 / 245 | **84.5%** | PASS |

The current full-profile security observation is 5/6 at the 80% floor. The
historical 6/6 result remains historical and is not used to override this
dated measurement.

| Temporary artifact | Size | SHA-256 |
|---|---:|---|
| `C:\tmp\hub-task203\backend-rerun-1f41dcbd.log` | 15,500 bytes | `c0b384b859fe3dc3a2dc3d95521dd1cc089d1d7cb0332f8a83d1b61585ccf4f0` |
| `C:\tmp\hub-task203\backend-rerun-1f41dcbd.cover.out` | 5,763,160 bytes | `b415570a86e4d4bbdc9884c8ae6aad4c7b71e40c1cd01d29fa57daf066e367a4` |
| `C:\tmp\hub-task203\backend-rerun-1f41dcbd-coverfunc.txt` | 330,281 bytes | `689f2fa21e5181818d9e1a751a7e05334f81f5854c1348f522afbc24abc824b8` |

Compared with the first attempt, all 12 test failures and all 10 failed
packages are cleared. NATS also passed inside the full suite (`6.410 s`),
consistent with the 20/20 focused and 5/5 package sentinels. Backend execution
is green, but the coverage contract still blocks Phase 128.

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

## Final E2E baseline after TASK-224

An isolated, resource-limited Docker application stack and deterministic dev
seed closed the earlier prerequisite gap. The focused admin-users and Telegram
slices passed **3/3**. The first full run exposed one Messenger retry-only
locator flake; after the semantic locator repair its focused repeat passed
**3/3**, and the final full run with `--retries=0` produced **48 PASS / 5
expected skip / 0 fail** in 1.1 minutes. Smoke produced **3/3 PASS**.

Two failures were classified as product gaps and repaired atomically: the
admin-safe user list endpoint and stable Telegram routes without a bot token.
The Messenger failure was classified as a test locator defect. Task-owned E2E
containers, network, volume, and temporary frontend image were removed after
evidence capture.

## Frozen frontend diagnostic

The final resource-limited Docker legacy Vitest diagnostic produced **510/519
PASS, 9 FAIL**: six harness failures from the absent `/work/docs` mount, two
nested fixture timeouts under the 2-CPU limit, and one stale i18n assertion.
These are recorded honestly and are not remediated: the founder superseded
legacy-suite repair with the exact-manifest greenfield cutover. Docker
TypeScript typecheck and Biome remain green.

## Sentinel and handoff

- Empty or malformed collection/profile is rejected.
- The first Go profile remains rejected because ten packages failed.
- The final Go profile is valid and published in `.planning/COVERAGE.md`;
  exact coverage is 82.0373% and all 6/6 security floors pass.
- Required frontend typecheck and lint are green. Legacy Vitest is frozen at
  510/519 and is not a migration gate.
- TASK-224 E2E is closed: smoke 3/3; full 48 pass / 5 expected skip / 0 fail
  with retries disabled.
- Handoff is the founder cutover contract and exact keep/move/delete manifest,
  not additional legacy remediation.

**Skill log:** `gsd-execute-phase` contract read; its referenced external
workflow file is absent in this environment, so the task's embedded gates were
executed directly. Final result: legacy discovery recorded; migration gate
superseded by the founder-approved greenfield cutover.
