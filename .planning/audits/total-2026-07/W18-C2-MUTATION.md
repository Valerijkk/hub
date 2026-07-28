# W18-C2 — Stage C mutation (be1, independent of Stage B author be2)

> 2026-07-28 · agent W18-C2 · branch `cursor-grok-4.5` · **no push/merge master**  
> Tasks: **014, 015, 016, 019, 020, 022, 024** (7)  
> Env: `POSTGRES_TEST_URL=postgresql://hub:hub@localhost:5432/hub?sslmode=disable`  
> `MONGODB_TEST_URL=mongodb://hub:hub@localhost:27017/?authSource=admin`

## Summary

| Metric | Count |
|--------|------:|
| **PASS** (→ «Исправлена») | **5** |
| **FAIL** (→ «Открыта») | **2** |
| FAIL IDs | **TASK-014**, **TASK-019** |

Protocol per ClassID: break production guard → `go test -tags hubtest -race -count=1` RED + name → restore → GREEN. All production mutations restored after runs (no leftover `W18-C2` diffs in `backend/internal/**` from this agent).

---

## TASK-014 — Project ACL + CompleteSprint atomic · **FAIL** → Открыта

| ClassID | Mutation | RED test | Layer | Verdict |
|---------|----------|----------|-------|---------|
| P1-BE-06 | drop `HasProjectWith` in `DeleteInvite` | `TestRevokeInvite_Ent_ScopedByProject` | Ent repo | PASS |
| P1-BE-07 | drop last-owner demote in `UpdateMemberRole` | `TestUpdateMemberRole_Ent_LastOwnerGuard`, `TestUpdateMemberRole_LastOwnerDemote_Forbidden` | service + Ent | PASS |
| P1-BE-08 | drop `requireTeamMember` in `AddTeamMember` | `TestTeams_StrangerCannotAddMember`, `TestAddTeamMember_RejectsStranger_Ent_TASK037` | service + Ent | **PARTIAL** |
| P1-BE-09 | drop ownership/admin in `DeleteTemplate` | `TestDeleteTemplate_Ent_Ownership`, `TestDeleteTemplate_ForeignUser_Forbidden` | service + Ent | PASS |
| P1-BE-05 | non-atomic `CompleteSprintAtomic` (no tx) | **no RED** — `TestCompleteSprint_Atomic_FailAfterMove_NoPartialState` + `TestCompleteSprintAtomic_Ent_MovesAndCompletes` stay GREEN | mock embeds rollback; Ent = happy path only | **FAIL** |

### Notes (FAIL)

1. **P1-BE-05:** mid-fail mutation criterion not covered on Ent. Mock `failStatusAfterMoveRepo` reimplements rollback inside the mock; removing production tx does not turn tests red.
2. **P1-BE-08:** ACCEPTANCE says *stranger `AddTeamToProject` → 403*. Production `AddTeamToProject` has **no** service ownership ACL (only RBAC middleware). Denial tests cover `AddTeamMember` / `ListTeams`, not `AddTeamToProject`.

GREEN restore: `TestRevokeInvite_Ent_ScopedByProject|TestUpdateMemberRole_Ent_LastOwnerGuard|TestTeams_StrangerCannotAddMember|TestDeleteTemplate_Ent_Ownership|TestCompleteSprint_*` → ok.

---

## TASK-015 — TMS suite/env IDOR + FK · **PASS** → Исправлена

| ClassID | Mutation | RED test | Layer |
|---------|----------|----------|-------|
| P1-BE-10 | drop `HasProjectWith` in `DeleteSuite` | `TestDeleteSuite_RejectsForeignProject_P1BE10` | Ent |
| P1-BE-11 | drop `HasProjectWith` in `GetEnvironmentInProject` | `TestGetEnvironmentInProject_RejectsForeign_P1BE11` | Ent |
| P1-BE-12 | drop `GetSuiteInProject` in `CreateTestCase` | `TestCreateTestCase_RejectsForeignSuite_P1BE12` | service + Ent |

GREEN: `-run P1BE10|P1BE11|P1BE12 ./testframework/tms/` → ok.

---

## TASK-016 — TMS import limits + steps · **PASS** → Исправлена

| ClassID | Mutation | RED test | Layer |
|---------|----------|----------|-------|
| P1-BE-13 | drop `LimitReader` + size→413 in `handleImport` | `TestHandler_ImportTestCases_OversizedBody_413` (expected 413, got 200) | handler production |
| P1-BE-14 | drop `case "steps"` in `applyFieldToInput` | `TestImportFromCSV_StepsColumn_Mapped` | importer |

GREEN: oversized + Steps tests → ok.

---

## TASK-019 — Platform SSRF/outbox/pool/config · **FAIL** → Открыта

| ClassID | Mutation | RED test | Verdict |
|---------|----------|----------|---------|
| P1-BE-23 | drop `ip.To4()` in `isBlockedIP` | **no RED** — `TestIsBlockedIP`, `TestSafeDialContext_BlocksIPv4MappedMetadata` stay GREEN | **FAIL** |
| P1-BE-24 | drop `FlushTimeout` in `PublishMsg` | `TestProcessEvent_DropBeforeFlush_StaysPending_P1_BE_24` (pending→published) | PASS |
| P1-BE-25 | empty `applyPoolLimits` | `TestApplyPoolLimits_MaxOpenConns_P1_BE_25` (want 25, got 0) | PASS |
| P1-BE-26 | skip `weakDataStoreURLProblem` hub:hub@ | `TestLoad_productionRejectsDefaultDataStoreURLs_P1_BE_26` | PASS |

### Notes (FAIL)

**P1-BE-23:** on Go 1.25, `net.IPNet.Contains` for IPv4 CIDR already matches IPv4-mapped IPv6 (`::ffff:169.254.169.254` → contains=true without `To4()`). Removing the normalize guard does not change observed block behavior; mutation is undetectable. Criterion «`::ffff:169.254.169.254` blocked» still holds at runtime, but Stage C requires RED on breaking *that* guard.

GREEN restore: pool / outbox / config / `TestIsBlockedIP` → ok.

---

## TASK-020 — Webhook Create contract · **PASS** → Исправлена

| ClassID | Mutation | RED test |
|---------|----------|----------|
| P1-BE-27 | `CreateWebhookResponse.Secret = ""` | `TestWebhookCreate_SecretOnce_P1BE27` |
| P1-BE-28 | allow empty events in `validateWebhookEvents` | `TestWebhookCreate_EventsAllowlist_P1BE28/empty_events_rejected` (400→201) |

Service-layer + HTTP against real Ent wiring. GREEN restore → ok.

---

## TASK-022 — Search soft-delete + private · **PASS** → Исправлена

| ClassID | Mutation | RED test |
|---------|----------|----------|
| P1-BE-32 | drop `deleted_at IS NULL` + membership OR in `searchChannels` | `TestSearchSQL_ExcludesSoftDeleted_A4_002` (static SQL) + `TestPostgresSearch_Channels_SoftDeleteAndPrivateScope_P1BE32` (runtime Postgres) |

Guard is production SQL in `SearchService`, not mock-only. GREEN restore → ok.

---

## TASK-024 — Notif / Git* / storage / backup / coffee · **PASS** → Исправлена

| ClassID | Mutation | RED test | Layer note |
|---------|----------|----------|------------|
| P1-BE-35 | `repoLookupErr` always nil | `TestNotificationConsumer_NakOnRepoErr_P1_BE_35` | production consumer + closed DB |
| P1-BE-36 | `redactPipelineVariables` returns vars as-is | `TestRedactPipelineVariables_P1_BE_36` | production helper |
| P1-BE-37 | skip `IsProjectMember` deny | `TestListGitHubLinks_StrangerForbidden_P1_BE_37` | **service + mock repo** (guard in service, not Ent filter) |
| P1-BE-38 size | drop `maxUploadBytes` | `TestGenerateUploadURL_RejectsOversized_P1_BE_38` | storage service |
| P1-BE-38 entity | allow cross-project issue | `TestConfirmUpload_RejectsForeignIssue_P1_BE_38` | service + Ent |
| P1-BE-39 | `MarkCompleted` when s3 nil | `TestPerformBackup_FailsClosedWithoutS3_P1_BE_39` | backup service |
| P1-BE-43 | odd N → pair+unpaired not triple | `TestBuildCoffeeMatches_OddNTriple_P1_BE_43` | coffee matcher |

P1-BE-37: not Ent-repo mock mismatch (ACL is intentionally service-level). Optional follow-up: Ent integration denial for `ListGitHubLinks`.

GREEN restore of all listed tests → ok (notification package alone after parallel dirty `search` compile noise cleared for that package).

---

## FAIL IDs (action for Stage B rework)

1. **TASK-014** — add Ent mid-fail proof for `CompleteSprintAtomic`; add denial for stranger `AddTeamToProject` (or fix ACCEPTANCE if RBAC-only is intended).
2. **TASK-019** — P1-BE-23: either strengthen mutation-sensitive assert (e.g. unit that *requires* `To4` path / pre-Go-1.25 behavior), or document Go 1.25 `Contains` as making `To4` redundant and adjust registry mutation criterion.
)
