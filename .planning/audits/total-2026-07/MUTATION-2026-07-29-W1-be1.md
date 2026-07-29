# MUTATION — 2026-07-29 · W1 · be1-checker (независимый от авторов)

> Задача 2 / TASK-008 · Проверяющий: **be1** · Авторы rework: TASK-014/019/045 = be2  
> Контекст FAIL: `.planning/audits/total-2026-07/W18-C2-MUTATION.md` (014, 019), `W18-C2b-MUTATION.md` (045)  
> Env: `POSTGRES_TEST_URL` / `MONGODB_TEST_URL` (localhost hub:hub) · `CGO_ENABLED=1` · `-tags hubtest`

## Итог

| ID | ClassIDs | Verdict | Test (RED) | Mock-gap? | README |
|----|----------|---------|------------|-----------|--------|
| TASK-014 | P1-BE-05..09 | **PASS** | см. ниже | нет (Ent mid-fail + service ACL) | Закрыта |
| TASK-019 | P1-BE-23..26 | **PASS** | см. ниже | нет | Закрыта |
| TASK-045 | P2-BE-19..21 | **PASS** | см. ниже | нет (Confirm = service+S3 Head) | Закрыта |

Production после restore: мутации в `backend/internal/**` откатены (нет leftover W1-be1 guards).

---

## TASK-014 — Project ACL + CompleteSprint atomic · **PASS**

### ACCEPTANCE
revoke чужого invite / demote last owner / stranger AddTeamToProject / fail mid CompleteSprint

### Мутации → RED → restore GREEN

| Guard | Mutation | RED test |
|-------|----------|----------|
| P1-BE-06 DeleteInvite scope | drop `HasProjectWith` | `TestRevokeInvite_Ent_ScopedByProject` |
| P1-BE-07 last-owner demote | skip demote guard | `TestUpdateMemberRole_Ent_LastOwnerGuard`, `TestUpdateMemberRole_LastOwnerDemote_Forbidden` |
| P1-BE-08 stranger AddTeamToProject | drop `requireTeamMember` | `TestTeams_StrangerCannotAddTeamToProject`, `TestAddTeamToProject_Ent_RejectsStranger` |
| P1-BE-05 CompleteSprintAtomic tx | non-tx client ops | `TestCompleteSprintAtomic_Ent_FailAfterMove_RollsBack` (issue count 1→0) |

### Notes
- Mock `TestCompleteSprint_Atomic_FailAfterMove_NoPartialState` по-прежнему embeds rollback (W18 gap) — **не** единственный proof: Ent `TestCompleteSprintAtomic_Ent_FailAfterMove_RollsBack` ловит drop tx.
- Stranger ACL на service + Ent, не RBAC-only.

### GREEN restore
```
go test -tags hubtest -count=1 -timeout=180s ./testframework/project/ -run 'TestRevokeInvite_Ent_ScopedByProject|TestUpdateMemberRole_Ent_LastOwnerGuard|TestUpdateMemberRole_LastOwnerDemote_Forbidden|TestTeams_StrangerCannotAddTeamToProject|TestAddTeamToProject_Ent_RejectsStranger'
→ ok
go test -tags hubtest -count=1 -timeout=180s ./testframework/sprint/ -run 'TestCompleteSprintAtomic_Ent_FailAfterMove_RollsBack|TestCompleteSprintAtomic_Ent_MovesAndCompletes'
→ ok
```

### Mock-gap?
**Нет** для ClassIDs ACCEPTANCE (Ent mid-fail + Ent stranger).

---

## TASK-019 — Platform SSRF/outbox/pool/config · **PASS**

### ACCEPTANCE
`::ffff:169.254.169.254` blocked; drop до Flush → pending; hub:hub@ в prod fail (+ pool ClassID)

### Мутации → RED → restore GREEN

| Guard | Mutation | RED test |
|-------|----------|----------|
| P1-BE-23 To4 normalize | drop `To4()` in `normalizeIPForBlocklist` | `TestNormalizeIPForBlocklist_IPv4MappedFourBytes_P1BE23` (len 4→16) |
| P1-BE-24 Flush | drop `FlushTimeout` in `PublishMsg` | `TestPublishWithTraceContext_FlushAfterPublish_A3W4_2_003` |
| P1-BE-24 PubAck | ignore PubAck err → always nil | `TestProcessEvent_PubAckFailure_StaysPending_A2W4_2_002` (pending→published) |
| P1-BE-25 pool | empty `applyPoolLimits` | `TestApplyPoolLimits_MaxOpenConns_P1_BE_25` (25→0) |
| P1-BE-26 hub:hub@ | skip `weakDataStoreURLProblem` | `TestLoad_rejectsWeakHubCreds_P1_BE_26`, `TestLoad_productionRejectsDefaultDataStoreURLs_P1_BE_26` |

### Notes
- W18 FAIL на P1-BE-23: `isBlockedIP` без To4 оставался GREEN (Go 1.25 `Contains`). Rework `len==4` assert — мутация детектируется.
- `TestIsBlockedIP` / `TestSafeDialContext_BlocksIPv4MappedMetadata` без To4 остаются GREEN — критерий Stage C закрыт отдельным normalize-тестом.

### GREEN restore
все перечисленные `-run` → ok

### Mock-gap?
**Нет**

---

## TASK-045 — RBAC Bit=0 + CTC + Confirm ghost · **PASS**

### ACCEPTANCE
Bit=0→Has false; CTC unequal reject; Confirm без object→404

### Мутации → RED → restore GREEN

| Guard | Mutation | RED test |
|-------|----------|----------|
| P2-BE-19 Bit=0 | drop `if perm == 0` in `Has` | `TestPermission_Has/bit_zero_check_denied` (false→true) |
| P2-BE-20 CTC | `constantTimeEqualDigest` always true | `TestConstantTimeEqualDigest_UnequalLength` |
| P2-BE-21 Confirm ghost | drop `isS3NotFound`→`ErrObjectNotFound` | `TestConfirmUpload_MissingS3Object_NotFound` (want ErrObjectNotFound, got wrapped Head 404) |

### Notes
- W18-C2b FAIL: Confirm был helper/tautology. Rework IT: pending FileMetadata + httptest S3 404 → `ConfirmUpload` → `ErrObjectNotFound`.

### GREEN restore
```
./testframework/rbac/ -run TestPermission_Has$ → ok
./testframework/gitlab/ -run TestConstantTimeEqualDigest_UnequalLength → ok
./testframework/storage/ -run TestConfirmUpload_MissingS3Object_NotFound → ok
```

### Mock-gap?
**Нет**

---

## Production integrity

После всех restore: целевые файлы мутаций совпадают с pre-check (кроме unrelated dirty tree вне scope). Сломанный код не оставлен.
