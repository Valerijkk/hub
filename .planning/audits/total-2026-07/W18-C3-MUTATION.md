# W18-C3 — Mutation (be2 · independent of Stage B be1)

> 2026-07-28 · Agent: W18-C3 (runtime be2) · Branch: `cursor-grok-4.5`  
> Scope: TASK-023, 025, 038, 040, 041, 042, 044, 046  
> Protocol: break production guard → hubtest RED → restore GREEN · mock-only vs Ent-repo

## Summary

| Metric | Count |
|--------|------:|
| **PASS** | **4** |
| **FAIL** | **4** |
| FAIL IDs | TASK-041, TASK-042, TASK-044, TASK-046 |

## Per-task results

| ID | Verdict | Mock-only? | Key RED test(s) | Notes |
|----|---------|------------|-----------------|-------|
| TASK-023 | **PASS** → Исправлена | GDPR: service+stub repo (OK — guard in service) | `TestDemoteAdmin_RequirePermission403_P1BE33`; `TestDeleteUserGDPR_FilesFailAbortsHardDelete_P1BE34` | Demote: drop full-mask recheck → 200 not 403. GDPR: ignore cascade errors → HardDelete called. RBAC demote uses Ent+miniredis. |
| TASK-025 | **PASS** → Исправлена | No — Ent migrate/runtime | `TestCoffeeEnrollment_UniquePerUser_TASK025`; `TestNotificationPreference_PartialUnique_TASK025`; `TestJournalEntry_InvoiceO2O_TASK025` | Mutated `ent/migrate/schema.go` Unique flags (generated runtime DDL; schema/*.go alone without regen does not affect enttest). All three RED then GREEN. |
| TASK-038 | **PASS** → Исправлена | No — Ent repo | `TestEntRepository_MoveIssue_OptimisticLock_W1A3004`; `TestEntRepository_UpdateIssue_SoftDeleted_W1A3005`; `TestUpdateBase_OptimisticLock_VersionEQ_W2A5004`; `TestGetTableView_HasMore_W2A5005` | Drop VersionEQ / DeletedAtIsNil / HasMore → RED. Layer = Ent repository. Residual Source IDs (migration/out-of-scope) unchanged. |
| TASK-040 | **PASS** → Исправлена | No (miniredis / Fiber / Ent settings) | `TestWSHandler_NilAuthorizer_RefusesUpgrade`; `TestRateLimitRedis_RejectedDoesNotInflateZSET`; `TestCheckRegistrationAllowed_SettingsFail_InviteOnly` | authz==nil→allowAll; skip ZRem→card 7; settings fail→open. All RED. |
| TASK-041 | **FAIL** → Открыта | Dual-Start Ent test insufficient | Vote: `TestPoker_VoteAfterReveal_Rejected` **RED** ✓ | **Fail:** remove `ErrActiveSprintConflict` in `StartSprintAtomic` → `TestStartSprintAtomic_Ent_Concurrent` stays **GREEN** (SQLite serializes `client.Tx`; conflict guard never proven). Mock concurrent test embeds guard in mock. |
| TASK-042 | **FAIL** → Открыта | via_invite assertion static | CreateBoard / revoke / guest RED ✓ | **Fail:** `via_invite: true→false` in production → `TestAcceptInvite_PayloadViaInviteBool_P2_BE_13` stays GREEN (asserts handmade JSON map, not NATS payload from AcceptInvite). |
| TASK-044 | **FAIL** → Открыта | Prefer default: service+mock (OK) | Prefer + OR RED ✓ | **Fail:** no hubtest for Meili SSRF (`validateMeiliHost` / `169.254…`); no hubtest for stream cancel (`ctx.Err` in adapter_stream). OR: only RED if `issue.Or` wrap removed (len→2); ignore-OR-flush alone stays green. |
| TASK-046 | **FAIL** → Открыта | **Yes-analogue:** schema source grep, not Ent-runtime | `TestSchema_P2BE22_UniqueFKCheck` RED on string | ACCEPTANCE wants duplicate Create→unique / XOR CHECK at DB. Test only `assert.Contains` on `ent/schema/*.go` text — not enttest Create. Insufficient layer vs TASK-025. |

## Commands (canonical)

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"
go test -tags hubtest -count=1 -timeout=90s -run '<TestName>' ./testframework/<pkg>/
```

## Status updates

- PASS → README + task file: **Исправлена** + W18-C3 evidence  
- FAIL → **Открыта** + W18-C3 FAIL reason  

No production code left mutated after restore. Artifact-only commit for this report + backlog status sync.
