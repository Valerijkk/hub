# W18-C2b — Stage C mutation gap-fill (be1 runtime)

> 2026-07-28 · agent **W18-C2b** · branch `cursor-grok-4.5` · **no push/merge master**  
> Tasks missed in W18-C2: **036, 037, 039, 043, 045** (5) · Stage B author: **be2**  
> Env: `POSTGRES_TEST_URL` / `MONGODB_TEST_URL` → localhost hub fixtures

## Summary

| Metric | Count |
|--------|------:|
| **PASS** (→ «Исправлена») | **4** |
| **FAIL** (→ «Открыта») | **1** |
| FAIL IDs | **TASK-045** |

Protocol per ClassID: break production guard → `go test -tags hubtest -race -count=1` RED + name → restore → GREEN.  
All production mutations restored (no leftover `W18-C2b` diffs in `backend/internal/**`).

---

## TASK-036 — DTO StructValidator + TZ error contract · **PASS** → Исправлена

| ClassID | Mutation | RED test | Layer |
|---------|----------|----------|-------|
| P2-BE-01 | `Validate` → always `nil` | `TestStructValidator_EmptyTitleEmail_Mutation/with_validator_rejected` (400→**201**); `TestCreateIssue_EmptyTitle_StructValidatorMutation/with_validator_400` (400→**500**); `TestCreateComment_MissingRequired_StructValidatorMutation/with_validator_400` (400→**201**) | production `middleware.NewStructValidator` |
| P2-BE-07 | `ErrorHandler` AppError → JSON with `code`/`message` | `TestErrorHandler_ForbiddenNoInternals`; `TestCreateInviteLink_ForbiddenHTTP_NoInternals` | production ErrorHandler |

Mock-only? **no** — tracker suite wires real Fiber + Ent CreateIssue/CreateComment.  
GREEN restore: middleware + tracker empty-title + messenger Forbidden → ok.

---

## TASK-037 — Ent-negative authz · **PASS** → Исправлена

| ClassID | Mutation | RED test | Layer |
|---------|----------|----------|-------|
| P2-BE-02 | drop `HasAuthorWith` in `DeleteComment` | `TestDeleteComment_RejectsForeignAuthor_Ent_TASK037` (expected NotFound, got nil) | **Ent repo** |
| P2-BE-02 | drop `HasProjectWith` in `FindIssueByKey` | `TestDeleteComment_ServiceRejectsForeignProject_Ent_TASK037`; `TestUpdateIssue_RejectsForeignProjectKey_Ent_TASK037` | **Ent + service** |

Mock-only? **no** — Ent sqlite fixtures, not handler_flow mocks.  
GREEN restore → ok.

---

## TASK-039 — Soft-delete + secrets · **PASS** → Исправлена

| ClassID | Mutation | RED test | Layer |
|---------|----------|----------|-------|
| P2-BE-08 | `DeleteSuite` hard `Delete()` | `TestSoftDeleteSuite_KeepsRowWithDeletedAt` (row gone) | Ent TMS |
| P2-BE-08 | `DeleteAPIKey` hard delete | `TestSoftDeleteAPIKey_RevokeKeepsRow` | Ent apigateway |
| P2-BE-09 | `cipher==nil` → plaintext ok (AI + apigateway) | `TestEntRepository_CreateProvider_NilCipherRejectsAPIKey`; `TestWebhookSecret_NilCipherRejectsCreate` | Ent repo |
| P2-BE-09 | `Organization.String()` leak invite_token | `TestSensitiveFields_HiddenInString` | Ent generated Sensitive mask |

Mock-only? **no**. GREEN restore → ok.

---

## TASK-043 — Platform boot/outbox/OTEL · **PASS** → Исправлена

| ClassID | Mutation | RED test | Layer |
|---------|----------|----------|-------|
| P2-BE-16 | skip DeriveDevKey prod reject | `TestLoad_productionRejectsDeriveDevKey_W4A4009` | `config.validate` |
| P2-BE-16 | never `attempts≥10` → failed (+ skip sweep) | `TestRecordPublishFailure_AttemptsGTE10_SetsFailed`; `TestProcessBatch_MarksExhaustedPendingAsFailed` | outbox Worker |
| P2-BE-16 | `httpRouteLabel` → raw `c.Path()` | `TestHTTPTraceMiddleware_NoRawPathTarget_W4A3001`; `TestHTTPMetricsMiddleware_UnmatchedNoUUID_W4A3002` | observability |

Mock-only? **no**. GREEN restore → ok.

---

## TASK-045 — RBAC/GDPR + CTC + Confirm · **FAIL** → Открыта

| ClassID | Mutation | RED test | Verdict |
|---------|----------|----------|---------|
| P2-BE-19 | drop Bit=0 guard in `Has` | `TestPermission_Has/bit_zero_check_denied` (want false, got true) | **PASS** |
| P2-BE-20 | `constantTimeEqualDigest` always true | `TestConstantTimeEqualDigest_UnequalLength` | **PASS** |
| P2-BE-21 | remove `isS3NotFound`→`ErrObjectNotFound` in `assertObjectWithinLimit` only | **no RED** — all `Confirm*` / storage suites stay GREEN | **FAIL** |

### Notes (FAIL)

ACCEPTANCE: *Confirm без object→404*. Production guard is mapping HeadObject NotFound → `ErrObjectNotFound` in `storage.ConfirmUpload` path.

- Breaking **only** that mapping leaves `TestIsS3NotFound` / Confirm suites **GREEN**.
- `content_type_confirm_task045_test.go` asserts `require.ErrorIs(ErrObjectNotFound, ErrObjectNotFound)` (tautology) and unit-tests `isS3NotFound` helper — **not** ConfirmUpload with missing S3 object → 404.
- Mutating `isS3NotFound`→always false RED's the helper test, but that is not the Confirm→404 criterion.

**Rework:** add Ent/service test: pending file + S3 Head 404 → `ConfirmUpload` returns `ErrObjectNotFound` / handler 404; then Stage C can RED on removing the mapping.

Bit=0 + CTC guards are mutation-proven; task stays **Открыта** until Confirm ClassID is covered.

GREEN restore of Bit=0 / CTC / storage → ok.

---

## FAIL IDs (Stage B rework)

1. **TASK-045** — add Confirm-without-object integration/unit proof (not helper-only / tautology).

## Commands (evidence)

```bash
cd backend
# 036
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/platform/middleware/ -run 'TestStructValidator_EmptyTitleEmail_Mutation|TestErrorHandler_ForbiddenNoInternals'
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/tracker/ -run 'TestCreateIssue_EmptyTitle_StructValidatorMutation|TestCreateComment_MissingRequired_StructValidatorMutation'
go test -tags hubtest -race -count=1 -timeout=120s ./testframework/messenger/ -run 'TestCreateInviteLink_ForbiddenHTTP_NoInternals'
# 037
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/tracker/ -run 'TestDeleteComment_RejectsForeignAuthor_Ent_TASK037|TestDeleteComment_ServiceRejectsForeignProject_Ent_TASK037|TestUpdateIssue_RejectsForeignProjectKey_Ent_TASK037'
# 039
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/tms/ -run 'TestSoftDeleteSuite_KeepsRowWithDeletedAt'
go test -tags hubtest -race -count=1 -timeout=120s ./testframework/ai/ -run 'TestEntRepository_CreateProvider_NilCipherRejectsAPIKey'
go test -tags hubtest -race -count=1 -timeout=120s ./testframework/platform/ -run 'TestSensitiveFields_HiddenInString'
go test -tags hubtest -race -count=1 -timeout=120s ./testframework/apigateway/ -run 'TestSoftDeleteAPIKey_RevokeKeepsRow|TestWebhookSecret_NilCipherRejectsCreate'
# 043
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/platform/config/ -run 'TestLoad_productionRejectsDeriveDevKey_W4A4009'
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/platform/outbox/ -run 'TestRecordPublishFailure_AttemptsGTE10_SetsFailed|TestProcessBatch_MarksExhaustedPendingAsFailed'
go test -tags hubtest -race -count=1 -timeout=120s ./testframework/platform/observability/ -run 'TestHTTPTraceMiddleware_NoRawPathTarget_W4A3001|TestHTTPMetricsMiddleware_UnmatchedNoUUID_W4A3002'
# 045
go test -tags hubtest -race -count=1 -timeout=120s ./testframework/rbac/ -run 'TestPermission_Has$'
go test -tags hubtest -race -count=1 -timeout=120s ./testframework/gitlab/ -run 'TestConstantTimeEqualDigest_UnequalLength'
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/storage/ -run 'ConfirmUpload|ObjectNotFound|IsS3'
```
