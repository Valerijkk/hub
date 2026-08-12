---
status: closed_superseded_by_founder_greenfield_cutover
trigger: "TASK-203 canonical baseline: diagnose 12 backend Docker Go 1.25 -tags hubtest -race failures without modifying code or tests"
created: 2026-08-07T00:00:00+03:00
updated: 2026-08-07T20:45:00+03:00
---

# Debug: TASK-203 canonical test baseline failures

## Closure — 2026-08-07

This file is retained as historical diagnostic evidence. Backend and E2E root
causes were closed and the final Docker E2E run passed with retries disabled.
The founder then superseded further legacy remediation: genuine native units
are retained, all legacy non-unit suites are frozen for exact-manifest deletion,
and new non-unit coverage is designed greenfield in Python. No action should be
taken from the stale `Current Focus` section below.

## Current Focus

hypothesis: Backend and frontend deterministic failures have evidence-backed classifications; NATS remains a non-reproducing full-suite flake.
test: Apply disjoint TASK-221/TASK-222 allowlists, run focused checks, then repeat the full canonical backend/frontend baseline. Handle Biome and E2E prerequisites as TASK-223/TASK-224.
expecting: Focused suites become green without backend production changes and with only the confirmed frontend semantic-token production correction; the next full backend run either passes NATS or captures the missing lifecycle state evidence.
next_action: Complete TASK-221/TASK-222, verify independently, then execute TASK-223/TASK-224 before accepting the TASK-203 rerun.

## Symptoms

expected: Canonical backend Docker baseline (`go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...`) passes all packages and produces a valid complete coverage profile.
actual: 40 packages pass, 10 packages fail, 3 packages have no tests; 12 named tests fail. Coverage profile is partial and invalid for publication.
errors: Nine or ten source/schema guards appear sensitive to Docker `/app`; GitLab route-registration test times out at one second; NATS integration connect/close assertion fails under the race detector. Exact raw messages pending from TASK-203 executor.
reproduction: Run the documented canonical backend Docker command from AGENTS.md against the healthy test service stack.
started: Observed during TASK-203 baseline on 2026-08-07; prior trusted Docker baseline dated 2026-06-22 was green.

## Eliminated

- hypothesis: NATS embedded server was not ready or NewClient failed to connect.
  evidence: `NewClient`, non-nil client/Conn/JetStream, `IsConnected`, and the connection log all succeeded before the line-43 close assertion.
  timestamp: 2026-08-07T00:27:00+03:00
- hypothesis: NATS connect/close failure is deterministic in the focused test or package.
  evidence: Docker Go 1.25.11 `-race -count=20` focused run passed 20/20; whole platform/nats package passed 5/5.
  timestamp: 2026-08-07T00:36:00+03:00
- hypothesis: GitLab merely needs a longer Fiber test timeout.
  evidence: The test actually calls external `https://gitlab.com`; commit `a9f931c7` documents the identical historic failure and the need for a deterministic stub. A larger timeout would retain network nondeterminism.
  timestamp: 2026-08-07T00:22:00+03:00

## Evidence

- timestamp: 2026-08-07T00:00:00+03:00
  checked: GSD debugger required-reference paths
  found: `mandatory-initial-read.md`, `common-bug-patterns.md`, `project-skills-discovery.md`, and `debugger-philosophy.md` are absent under both `C:/Users/Valerii/.Codex/get-shit-done/references/` and `C:/Users/Valerii/.agents/`; the complete debugger protocol supplied to this agent and `.cursor/skills/test-triage/SKILL.md` remain available.
  implication: Continue with the supplied scientific-method protocol and Hub test-triage; record this tooling-documentation gap rather than inventing missing guidance.
- timestamp: 2026-08-07T00:05:00+03:00
  checked: TASK-203 executor raw evidence handoff
  found: Raw log is `C:/tmp/hub-task203/backend-go-test.log`; rejected partial profile is `C:/tmp/hub-task203/backend.cover.out`. Ten exact file-open errors target `/backend/internal/{accounting,calls,github,gitlab,hr,intranet,sprint,tracker}/...` or `/.planning/decisions/ADR-ENT-CHANNELMEMBER-LWW-2026.md`. GitLab handler fails at line 120 with `test: timeout error after 1s`. NATS fails at line 43 (`Should be true`) despite a nearby `connected to NATS` log.
  implication: Correct the initial count from nine to ten path guards; investigate three mechanism clusters totaling 12 tests.
- timestamp: 2026-08-07T00:10:00+03:00
  checked: Ten path-guard implementations and canonical Docker mount
  found: Nine soft-delete source guards compute `root = Dir(testFile)/../../..`. For a Docker test file under `/app/testframework/<domain>`, this resolves to `/`, after which each appends `backend/internal/...`, producing the observed `/backend/internal/...`. Their runtime soft-delete assertions complete successfully before the source-file read. The nested TASK-167 guard correctly reaches `/app/ent/schema` with three ascents, but separately ascends four levels to `/` and attempts `/.planning/...`; AGENTS.md mounts only `backend` at `/app`, so `.planning` is not present.
  implication: All ten are `ROOT=test`: functional code paths passed, and only environment-invalid white-box/doc guards failed. Fix test path resolution/document coupling, not production behavior.
- timestamp: 2026-08-07T00:12:00+03:00
  checked: GitLab route test and service call path
  found: The test creates an `httptest.Server` at lines 64-71 and injects its client at line 75, but deliberately omits `gitlab_url` at lines 93-95. Production defaults it to `https://gitlab.com` (`service.go:71-75`), and `ListPipelines` calls that stored URL (`service.go:265-300`). The stub handler is therefore never selected by URL; `Fiber.App.Test(req)` at line 119 uses its default one-second timeout and the raw log shows the request later returning GitLab 401.
  implication: `ROOT=test`: the route test leaks to an external service instead of using its intended deterministic stub. Increasing the Fiber timeout alone would mask, not remove, the root cause.
- timestamp: 2026-08-07T00:14:00+03:00
  checked: Hub NATS `Client.Close` and nats.go v1.48.0 lifecycle contract
  found: Hub starts a goroutine that calls `Conn.Drain()` and waits for that goroutine (`client.go:89-104`), but upstream `Conn.Drain()` itself only transitions to DRAINING, starts `drainConnection()` asynchronously, and returns immediately (`nats.go:5874-5900`). Hub then conditionally calls synchronous `Conn.Close()` if still not closed. The observed assertion at test line 43 conflicts with the apparent fallback and needs focused repeated reproduction before classifying code versus flaky.
  implication: NATS root cause is not yet proven; test the exact lifecycle repeatedly under race detection rather than changing timing or assertions speculatively.
- timestamp: 2026-08-07T00:16:00+03:00
  checked: First host-side focused NATS reproduction
  found: Test execution did not start because the sandbox denied Go writes to the default cache under `AppData/Local/go-build`.
  implication: This result neither confirms nor refutes the NATS hypothesis; retry with a task-scoped writable cache, without changing dependencies or source.
- timestamp: 2026-08-07T00:18:00+03:00
  checked: Focused NATS retry with `GOCACHE=C:/tmp/hub-debug-go-cache`
  found: The sandbox also denied creating that cache directory. No test process started. The TASK-203 executor was asked to run focused Docker repetitions under its existing execution authority.
  implication: Continue read-only differential analysis; do not treat an infrastructure-denied command as test evidence.
- timestamp: 2026-08-07T00:22:00+03:00
  checked: Git history for GitLab route timeout
  found: Commit `a9f931c7` explicitly fixed the identical CI failure (`ListPipelines hit real gitlab.com and fiber.Test timed out after 1s`) by binding the integration URL to the test server. Commit `e495f186` later removed that URL because the new SSRF validator rejects loopback, but left the stub client/server and a no-op `_ = env.gitlabURL`; this exactly reintroduced the external call.
  implication: GitLab `ROOT=test` is independently confirmed by historical differential debugging, not merely inferred from timing.
- timestamp: 2026-08-07T00:25:00+03:00
  checked: Git history and accepted task contract for NATS close
  found: Commit `e164e633` changed `Client.Close` from direct close to goroutine-wrapped `Conn.Drain()` plus timeout. The accepted TASK-065 record requires `Close` via `Conn.Drain()+timeout` but explicitly flags W4-A2-006 as a weak/layer-mismatch test. Upstream nats.go v1.48.0 documents and implements `Drain()` as asynchronous: it starts `drainConnection()` and returns immediately; CLOSED is signalled later. Hub's `done` channel therefore proves only that Drain was initiated, not completed.
  implication: The implementation's synchronization mechanism is definitely wrong even though the exact one-off false-state interleaving still needs focused observation. Classify `ROOT=code` (lifecycle synchronization); strengthen the regression test while fixing, rather than adding sleeps or weakening the assertion.
- timestamp: 2026-08-07T00:27:00+03:00
  checked: Alternative NATS hypotheses
  found: Embedded-server readiness is contradicted by successful `NewClient`, non-nil JetStream, the passing `IsConnected` assertion, and the connection log before the sole failure. Cross-test hook interference is contradicted by absence of `t.Parallel` and by the failing test running before hook-setting tests. The assertion reflects the public `Close` postcondition and is not an obsolete expectation.
  implication: `ROOT=test` and simple readiness failure are eliminated; the code lifecycle defect is the smallest evidence-supported root.
- timestamp: 2026-08-07T00:36:00+03:00
  checked: Focused NATS Docker repetitions supplied by TASK-203 executor
  found: Go 1.25.11 `go test -tags hubtest -race -count=20 -run '^TestNewClient_ConnectAndClose$' ./testframework/platform/nats` passed 20/20; whole package with `-race -count=5` passed 5/5. Logs: `C:/tmp/hub-task203/nats-focused-count20.log`, `C:/tmp/hub-task203/nats-package-count5.log`.
  implication: The original one-off failure is `ROOT=flaky`, not a confirmed deterministic code regression. The incorrect Drain-completion synchronization remains a real code-review finding/high-probability candidate, but scientific-method criteria do not permit claiming it caused this occurrence without a reproducing interleaving.

## Resolution

root_cause: Ten guards are environment-invalid tests; GitLab route test reintroduced live external I/O; NATS is a non-reproducing full-suite flake whose exact trigger remains unconfirmed.
fix: No code or test fix applied in this diagnose-only turn. Apply the confirmed test-only corrections below; investigate NATS separately only on recurrence or with deterministic instrumentation.
verification: Diagnosis uses raw full-suite evidence, exact source/history traces, and independent NATS focused 20/20 plus package 5/5 Docker repetitions.
files_changed:
  - .planning/debug/test-baseline-failures.md

## Cluster ledger

| # | Test | Failing line / exact mechanism | Contract evidence | Triage |
|---|------|--------------------------------|-------------------|--------|
| 1 | `accounting/...::TestEntRepository_JournalLineJoinFilter_A2W8_3_004` | test `:77-79` resolves `/backend/internal/accounting/repository.go`; all runtime assertions through `:73` passed | ADR soft-delete `:44-48`; production `repository.go:180-203` | `ROOT=test` |
| 2 | `calls/...::TestEntRepository_RecordingSoftDelete_A2W8_4_005` | test `:57-59` resolves `/backend/internal/calls/repository.go`; runtime get/list/raw assertions passed | ADR table A; production `repository.go:355-426` | `ROOT=test` |
| 3 | `ent/task167/...::TestSchema_TASK167_ChannelMemberLWW_NoVersionMixin` | schema assertions `:44-48` passed; `repoRoot` `:24-28` resolves `/`, then ADR read `:50-52` targets absent `/.planning` | ADR exists and is accepted at `.planning/decisions/ADR-ENT-CHANNELMEMBER-LWW-2026.md:1-6,49-52`; backend-only Docker mount is canonical | `ROOT=test` |
| 4 | `github/...::TestEntRepository_IntegrationSoftDelete_W7_A3_010` | test `:43-45` resolves `/backend/internal/github/repository.go`; runtime delete/list/get/raw assertions passed | ADR table `:33`; production `repository.go:172-229` | `ROOT=test` |
| 5 | `gitlab/soft_delete...::TestEntRepository_IntegrationSoftDelete_W7_A3_010` | test `:39-41` resolves `/backend/internal/gitlab/repository.go`; runtime assertions passed | ADR table `:33`; production `repository.go:120-175` | `ROOT=test` |
| 6 | `hr/...::TestEntRepository_LeaveSoftDelete_A2W8_3_001` | test `:59-61` resolves `/backend/internal/hr/repository.go`; runtime assertions passed | ADR table `:27`; production `repository.go:55-59,174-177` | `ROOT=test` |
| 7 | `intranet/...::TestEntRepository_KudoSoftDelete_A2W8_3_004` | test `:52-54` resolves `/backend/internal/intranet/repository.go`; runtime assertions passed | global soft-delete contract `docs/tz_constraints.md:37`; production `repository.go:176-183,230-233` | `ROOT=test` |
| 8 | `sprint/...::TestEntRepository_SprintSoftDelete_A2W8_4_004` | test `:45-47` resolves `/backend/internal/sprint/repository.go`; runtime assertions passed | ADR table A; production `repository.go:141-161` | `ROOT=test` |
| 9 | `tracker/...::TestEntRepository_LabelSoftDelete_A2W8_3_005` | helper `:68-70` resolves `/backend/internal/tracker/repo_labels.go`; runtime assertions passed | ADR `:28`; production `repo_labels.go:33-53` | `ROOT=test` |
| 10 | `tracker/...::TestEntRepository_CustomFieldSoftDelete_A2W8_2_004` | helper `:81-83` resolves `/backend/internal/tracker/repo_custom_fields.go`; runtime assertions passed | ADR soft-delete decision; production `repo_custom_fields.go:33-53` | `ROOT=test` |
| 11 | `gitlab/handler_extended...::TestHandler_RegisterRoutes_WebhookAndIntegrations` | payload `:93-95` stores default `https://gitlab.com`; GET `:118-120` exceeds Fiber's 1 s test deadline; raw log then records external 401 | NFR-21/SEC-04 requires SSRF guard (`docs/system_specification.md:67`); tests must not bypass it or call live providers | `ROOT=test` |
| 12 | `platform/nats/...::TestNewClient_ConnectAndClose` | one full-suite false at test `:43`; not reproduced in focused 20/20 or package 5/5 | `Client.Close` postcondition is valid; TASK-065 marks current Drain test weak/layer-mismatch | `ROOT=flaky` (exact cause `unclear`) |

## Confirmed fix worker allowlist

Only these test files are needed for the eleven deterministic failures:

- `backend/testframework/accounting/soft_delete_journalline_task097_test.go`
- `backend/testframework/calls/soft_delete_recording_task097_test.go`
- `backend/testframework/ent/task167/schema_guard_test.go`
- `backend/testframework/github/soft_delete_task097_test.go`
- `backend/testframework/gitlab/soft_delete_task097_test.go`
- `backend/testframework/gitlab/handler_extended_test.go`
- `backend/testframework/hr/soft_delete_task097_test.go`
- `backend/testframework/intranet/soft_delete_kudo_task097_test.go`
- `backend/testframework/sprint/soft_delete_task097_test.go`
- `backend/testframework/tracker/soft_delete_task097_test.go`

Do not include production repositories or NATS files in this worker.

### Smallest safe corrections

1. In the eight domain source-guard files (nine tests), resolve the backend module root as `Dir(thisFile)/../..` and append `internal/<domain>/...`; do not append another `backend` segment. This works in current `backend/testframework/<domain>` and target `backend/tests/<domain>` layouts.
2. In `ent/task167/schema_guard_test.go`, keep the executable ChannelMember schema assertions but remove the backend unit test's dependency on `.planning/**`, which is outside the canonical backend-only Docker mount. Preserve the ADR link as traceability; validate ADR presence/content in planning/docs validation, not the Go unit job.
3. In the GitLab handler test, preserve the production SSRF rejection. Replace the accidental live call with a deterministic injected `http.RoundTripper` (or equivalent no-network client) that returns the pipeline fixture for the default `https://gitlab.com` request. Do not raise the one-second timeout and do not weaken SEC-04.
4. NATS: make no assertion sleep/timeout relaxation in this worker. If it recurs, capture `Conn.Status()`, `IsDraining()`, and `IsClosed()` immediately before/after `Client.Close` in a separate bounded debug change. Independently backlog the observed `Drain()` completion-wait defect in `backend/internal/platform/nats/client.go:93-104` for a deterministic regression-first fix.

## Verification contract

Focused/package checks after the confirmed test fixes:

```text
go test -tags hubtest -race -count=1 ./testframework/accounting ./testframework/calls ./testframework/ent/task167 ./testframework/github ./testframework/gitlab ./testframework/hr ./testframework/intranet ./testframework/sprint ./testframework/tracker
go test -tags hubtest -race -count=10 -run '^TestHandler_RegisterRoutes_WebhookAndIntegrations$' ./testframework/gitlab
go test -tags hubtest -race -count=20 -run '^TestNewClient_ConnectAndClose$' ./testframework/platform/nats
go test -tags hubtest -race -count=5 ./testframework/platform/nats
```

Final gate: rerun the exact TASK-203 canonical Docker Go 1.25 command with `-tags hubtest -race -count=1 -timeout=600s -covermode=atomic -coverpkg=./internal/...` over `./testframework/...`; accept coverage only if every package passes. No partial profile may replace the trusted baseline.

## Frontend triage

Focused Vitest reproduced exactly 7/7 failures while eight adjacent tests passed.

| Cluster | Evidence | Triage |
|---|---|---|
| Calls, two tests | Fixtures omit the `my_tasks` discriminant required by `isPersonalData()`; the second also omits mandatory `channel_name`. Production `joinCall(call.id)` remains correct. | `ROOT=test` |
| System health, three tests | Partial `AdminStats as unknown` fixtures omit `total_users`; `isAdminStats()` correctly rejects them. A neighboring full typed fixture passes. | `ROOT=test` |
| ErrorBoundary resetKey | Test fails before `rerender` because it hardcodes RU while detected locale is EN; production fallback is intentionally localized. | `ROOT=test` |
| Pipeline status tokens | `pipeline-status.tsx` contains raw green/red/yellow/gray palette despite the accepted FE-TA-005 semantic-token contract; the source guard is valid. | `ROOT=code` |

Frontend worker allowlist:

- `testframework/stages/unit/frontend-test/domains/calls/upcoming-calls-join.test.tsx`
- `testframework/stages/unit/frontend-test/domains/dashboard/system-health.test.tsx`
- `testframework/stages/unit/frontend-test/domains/platform/error-empty-auth-p1.test.tsx`
- `frontend/src/modules/tracker/pipeline-status.tsx`

Minimum corrections: complete typed `PersonalData` and `AdminStats` fixtures;
make the ErrorBoundary precondition locale-independent while retaining the
reset assertion; replace raw pipeline palette with existing semantic tokens
without changing URL sanitization.

## Frontend lint debt

The TASK-203 Biome result is separate from the seven Vitest failures:

- 303 formatting errors;
- 3 `useExhaustiveDependencies` errors;
- 2 `useAriaPropsSupportedByRole` errors;
- 2 `organizeImports` errors;
- 4 `useSemanticElements` warnings.

Classification: `ROOT=code/config-debt`. Do not apply a blind global `--write`:
dependency diagnostics and semantic/a11y findings require manual review.
TASK-223 owns the controlled recovery after TASK-222.
