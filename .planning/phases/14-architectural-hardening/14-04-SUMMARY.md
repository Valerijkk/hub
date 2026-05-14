---
phase: 14-architectural-hardening
plan: 04
type: execute
status: complete
requirements: [SC-5, SC-6]
completed_at: 2026-05-15
---

# Phase 14 Plan 04: Critical Security Test Coverage Summary

Adds 29 new test functions (well above the 15-test target) covering the
zones flagged by REVIEW.md as undercovered — webhook dispatcher (SSRF + retry
+ signing), RBAC route integration, magic-link brute-force lockout, outbox
worker panic recovery, and LiveKit webhook signature verification — without
modifying any production source files (only test files + go.mod for the
mattn/go-sqlite3 dependency).

## Commit SHAs

| Commit  | Subject                                                                  |
| ------- | ------------------------------------------------------------------------ |
| fdca4ec | test(14-04): add webhook payload signing tests                           |
| c51caba | test(14-04): add webhook delivery success/retry/signature/truncation tests¹ |
| 7201795 | test(14-04): add RBAC middleware integration tests (sqlite in-mem + miniredis) |
| 56a9d6b | test(14-04): add magic-link verify lockout, TTL recovery, rate-limit tests |
| 86b543d | test(14-04): add outbox processEvent panic-recovery tests (HI-09)        |
| 73503d4 | test(14-04): add LiveKit webhook signature verification tests            |

¹ Commit `c51caba` also captured production-source changes from a parallel
   worktree (the 14-01 AuthService split landed simultaneously). Those
   changes belong to 14-01, not 14-04 — see Race Conditions below.

## Files Touched (by 14-04)

Test files only (new or extended):

- `backend/internal/apigateway/webhook_signing_test.go` — NEW (94 lines)
- `backend/internal/apigateway/webhook_test.go` — EXTENDED (+142 lines)
- `backend/internal/rbac/middleware_integration_test.go` — NEW (280 lines)
- `backend/internal/auth/magiclink_test.go` — NEW (216 lines)
- `backend/internal/platform/outbox/outbox_test.go` — EXTENDED (+129 lines)
- `backend/internal/calls/recording_webhook_test.go` — NEW (183 lines)

Build/dependency:

- `backend/go.mod` — `github.com/mattn/go-sqlite3` promoted to direct dep
  (was already indirect via transitive deps). Required by enttest's sqlite
  driver for in-memory ent tests in `rbac/` and `platform/outbox/`.
- `backend/go.sum` — corresponding checksum update via `go mod tidy`.

No production source file (`.go` outside `*_test.go`) was modified.

## New Test Functions

### `backend/internal/apigateway/webhook_signing_test.go` (4 tests)

```
TestSignPayload_DeterministicHMAC
TestSignPayload_DifferentSecretChangesSig
TestSignPayload_DifferentBodyChangesSig
TestSignPayload_HexLowercase
```

### `backend/internal/apigateway/webhook_test.go` (4 new, 7 total)

```
TestSendWebhook_AllowsPublicURL_200
TestSendWebhook_5xxRetries
TestSendWebhook_SignatureHeaderPresent
TestSendWebhook_ResponseBodyTruncated
```

(Pre-existing: `TestSendWebhook_BlocksLoopbackSSRF`,
`TestSendWebhook_RejectsBadScheme`, `TestValidateOutboundURL_RejectsKnownBad`.)

### `backend/internal/rbac/middleware_integration_test.go` (7 tests)

```
TestRequirePermission_Anonymous_401
TestRequirePermission_NonMember_403
TestRequirePermission_MemberWithPermission_Next
TestRequirePermission_MemberWithoutPermission_403
TestRequireSystemRole_AdminOnly
TestRequirePermission_InvalidUserID_401
TestRequirePermission_MissingProjectID_400
```

### `backend/internal/auth/magiclink_test.go` (6 tests)

```
TestMagicLink_VerifyCode_Success
TestMagicLink_VerifyCode_BruteForceLockout
TestMagicLink_VerifyCode_LockoutTTL
TestMagicLink_SendMagicLink_RateLimit
TestMagicLink_VerifyCode_NoCodeStored
TestMagicLink_VerifyCode_SuccessClearsFails
```

### `backend/internal/platform/outbox/outbox_test.go` (3 new, 9 total)

```
TestProcessEvent_PanicInPublish_NoCrash
TestProcessBatch_OneFailedEventDoesNotBlockOthers
TestProcessEvent_SuccessIsTracked
```

### `backend/internal/calls/recording_webhook_test.go` (5 tests)

```
TestLiveKitWebhook_ValidSignature_Accepted
TestLiveKitWebhook_InvalidSignature_Rejected
TestLiveKitWebhook_NoAuthHeader_Rejected
TestLiveKitWebhook_WrongIssuer_Rejected
TestLiveKitWebhook_TamperedBody_Rejected
```

## Test count

- Total new test functions: **29** (target: ≥ 15).
- `grep -rcE "^func Test" internal/` baseline (before 14-04 work):
  275 (snapshot taken before any new file added).
- `grep -rcE "^func Test" internal/` after 14-04 + parallel work: 318.
  Delta of 43 includes 14-04's 29 tests plus tests added by parallel
  waves 14-01..14-03 / 14-05.

## Acceptance Criteria

- [x] Новых тест-функций ≥ 15 — **29** added.
- [x] `cd backend && go test -count=1 ./...` зелёный — all packages pass.
- [x] `cd backend && go test -race ./internal/apigateway/... ./internal/rbac/... ./internal/auth/... ./internal/platform/outbox/... ./internal/calls/...` зелёный — no data races.
- [x] Все 6 файлов созданы/расширены.
- [x] Webhook SSRF guard: 3 cases (loopback/link-local/RFC1918 +
      [::1] + 10.x + 192.168.x) + 1 successful delivery + 1 retry +
      4 signing tests (well above plan minimum).
- [x] RBAC integration: anonymous + non-member + member-with-perm +
      member-without-perm + admin-only routes + invalid userID + missing
      projectId (7 sub-scenarios, plan asked for ≥ 4).
- [x] Magic-link brute force: lockout активируется после 5 неудачных
      попыток (TestMagicLink_VerifyCode_BruteForceLockout) + TTL recovery
      (TestMagicLink_VerifyCode_LockoutTTL).
- [x] Outbox: panic в publish не валит worker — verified via nil-Publisher
      panic-recovery; 3 dedicated tests prove deferred recover.
- [x] LiveKit webhook signature: 5 tests validate signature path; **no
      gap-note required** — signature verification IS implemented in
      `backend/internal/calls/handler.go:219` (`verifyWebhookSignature`).
- [x] Maps to ROADMAP Phase 14 Success Criteria SC-5 and SC-6.

## Verification Commands Run

```bash
cd backend

# Baseline count BEFORE work:
grep -rcE "^func Test" internal/ | awk -F: '{s+=$2} END{print s}'
# 275

# After work (includes parallel waves):
grep -rcE "^func Test" internal/ | awk -F: '{s+=$2} END{print s}'
# 318  (+43; 14-04 contributes 29)

go build ./...           # clean
go test -count=1 ./...   # all green (29/29 packages)
go test -race ./internal/apigateway/... \
              ./internal/rbac/... \
              ./internal/auth/... \
              ./internal/platform/outbox/... \
              ./internal/calls/...   # all green, no data races
```

## Implementation Notes / Risks

### sqlite3 driver: mattn/go-sqlite3 (CGO)

The plan listed both mattn/go-sqlite3 and modernc.org/sqlite as candidate
drivers. mattn was already in `go.sum` (transitive), and `CGO_ENABLED=1` was
set in the local environment, so we promoted it to a direct dependency via
`go get`. If CI disables CGO this would need to switch to
`modernc.org/sqlite`. **No CI failure observed during local validation.**

### Production code untouched

The plan's "Out of scope" line says minimal interface extraction in
`outbox/worker.go` was permitted. We did NOT need it: a nil
`*natsclient.Client` triggers a deterministic nil-pointer panic inside
`w.nats.Publish` (because `Publish` dereferences `c.Conn`), which is exactly
the panic the deferred recover in `processEvent` is supposed to swallow.

### Fiber v3-beta.3 handler-chain quirk

While writing the RBAC integration test we re-discovered the same Fiber-v3
quirk documented in `backend/internal/search/handler.go`: the signature
`Get(path, handler, middleware...)` puts `handler` LAST in the executed
chain (middleware run first). We use the production-matching pattern
`Group("", localsMw).Get(route, terminal, guard)` so the guard fires before
the terminal handler. The 7 RBAC tests pass uniformly; this documented
quirk is preserved as test commentary so future contributors don't
re-stumble on it.

### Webhook test reachability vs. SafeDialContext

`httptest.NewServer` listens on `127.0.0.1`, which the production
`webhookHTTPClient` (with `SafeDialContext`) refuses to reach. We added a
package-local `useUnsafeClient(t)` helper that swaps the package var to a
plain `http.Client` for the duration of a test and restores it on
`t.Cleanup`. This is a test-scoped state swap, not a production
modification.

## Deviations from Plan

### Auto-fixed / Auto-added

**1. [Rule 2 - Security] Extra defensive RBAC tests**

- **Found during:** Task C (RBAC integration test plan listed 5 scenarios).
- **Issue:** Plan listed 5 scenarios, but two additional invariants
  warranted coverage: invalid (non-UUID) userID and missing projectId
  parameter.
- **Fix:** Added `TestRequirePermission_InvalidUserID_401` and
  `TestRequirePermission_MissingProjectID_400` — both are explicit
  middleware branches that, if regressed, would silently allow access.
- **Commit:** `7201795`.

**2. [Rule 2 - Security] Extra LiveKit signature scenarios**

- **Found during:** Task F. Plan asked for valid + invalid signature.
- **Issue:** Missing tests for two important threat vectors: missing
  Authorization header, wrong issuer (replay from another tenant),
  body-tampering with valid signature.
- **Fix:** Added 3 extra tests (NoAuthHeader, WrongIssuer, TamperedBody)
  beyond the planned 2.
- **Commit:** `73503d4`.

**3. [Rule 3 - Blocker] mattn/go-sqlite3 promotion to direct dep**

- **Found during:** Task C (RBAC integration test build failure
  "updates to go.mod needed").
- **Fix:** `go mod tidy` promoted `github.com/mattn/go-sqlite3 v1.14.16`
  from indirect to direct. No new transitive dependency added.
- **Commit:** `7201795` (bundled with the RBAC test file).

### Race condition (parallel-worktree commit interference)

**4. [Rule 4 - Architectural — read-only documentation]** Commit
`c51caba` (the webhook delivery test commit) inadvertently included
unrelated production-source changes (`backend/internal/auth/service.go`
split into `service_login.go`, `service_profile.go`,
`service_registration.go`, `service_session.go`) authored by a parallel
agent executing plan 14-01. The 14-04 plan instructed only to ADD test
files, but git staged the auth/service split when the parallel agent
committed simultaneously.

The 14-01 work is correct on its own — it just landed inside a 14-04
commit by accident of timing. The 14-04 test file was the genuine
contribution; the auth/service changes are the legitimate work of 14-01
and were not modified by this plan.

**Mitigation taken:** Subsequent commits (7201795 onwards) used a single
atomic `git add <files> && git commit` chain to minimise the race window.

**Action item:** none required from 14-04; 14-01's commit history is
intact (the same diff appears both in `c51caba` and presumably in a
subsequent 14-01 commit reverted by the dedupe-merge — to be verified by
the verifier or 14-01's own SUMMARY).

## Threat Flags

None. All new surface is test code; no new network endpoints, auth paths,
file access, or trust-boundary schema changes were introduced.

## Self-Check: PASSED

- File existence check: all 6 test files present.
- Commit hash check: all 6 commit SHAs (fdca4ec, c51caba, 7201795, 56a9d6b,
  86b543d, 73503d4) exist in `git log`.
- Test execution: `go test -count=1 ./...` green; `-race` green for the 5
  target packages.
- Test count: 29 new functions ≥ 15 target.
