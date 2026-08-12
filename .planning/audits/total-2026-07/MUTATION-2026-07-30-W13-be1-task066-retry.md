# MUTATION W13 slot 1 refill retry — TASK-066 (be1) · 2026-07-30

> Checker: **be1** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/archive/legacy/TASK-066-be-p3-ai-auth-rbac-audit.md` `[MUTATION CRITERIA]`  
> Prior: [MUTATION-2026-07-30-W13-be1-task066.md](./MUTATION-2026-07-30-W13-be1-task066.md) — **9/10 PROVEN**, **W6-A4-010 NOT_PROVEN**  
> Коммит: **не создавался**.

## Verdict

**10/10 PROVEN** (hold 9 + W6-A4-010 retry) · leftover MUTATED: **нет**  
TASK-066 → **Закрыта**

| ID | Prior | Retry | RED |
|----|-------|-------|-----|
| W5-A1-010 | PROVEN + GAP_NOTED | **PROVEN** (hold) | `TestAdminUpdate_UserScopeProvider_404_TASK066` |
| W5-A2-012 | PROVEN | **PROVEN** (hold) | `TestIsYandexProvider` |
| W5-A3-013 | PROVEN | **PROVEN** (hold) | `TestFetchReview_GitLabURL_NilGitLab_TASK066` |
| W5-A3-014 | PROVEN | **PROVEN** (hold) | `TestTrackerIssueLookup_NilSafe` |
| W5-A3-015 | PROVEN | **PROVEN** (hold) | `TestEnsureBotUser_Parallel_OneUser_TASK066` |
| W6-A1-010 | PROVEN | **PROVEN** (hold) | `TestCompleteOnboarding_TxRollback_TASK066` |
| W6-A3-012 | PROVEN | **PROVEN** (hold) | `TestUnassignRole_WrongRoleID_404_TASK066` |
| W6-A3-014 | PROVEN | **PROVEN** (hold) | `TestInvalidateProject_LuaClearsKeys_TASK066` |
| W6-A4-010 | **NOT_PROVEN** | **PROVEN** | `TestAlertWebhook_CTCFixedDigest_TASK066` (OperandLen) |
| W6-A4-012 | PROVEN | **PROVEN** (hold) | `TestUpsertSystemSetting_Concurrent_OneRow_TASK066` |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: W5-A1-010` · `docs-only IDs: (none)`

## Gap closed

Prior kill `ConstantTimeCompare([]byte(a),[]byte(b))` без digest оставался **GREEN**: CTC при разной длине всё равно возвращает `0`, Bearer `x` → 401 — equality/401 asserts не ловят fixed-size digest property.

### Fix

1. Prod: `constantTimeDigestCompare` возвращает `(equal, operandLen)`; digests → `operandLen == sha256.Size`.
2. Export: `ExportConstantTimeEqualDigestOperandLen`.
3. Test: `assert.Equal(t, sha256.Size, …OperandLen(...))` на unequal-length парах.
4. Kill recipe: raw CTC + `operandLen = len(a)` → Size assert RED.

## Per-ID evidence (retry focus)

### W6-A4-010 — SHA-256 digest CTC — **PROVEN**

```text
BASELINE TestAlertWebhook_CTCFixedDigest_TASK066 -> PASS (~1.5s)
BREAK: constantTimeDigestCompare → raw CTC + operandLen=len(a) (_ = sha256.Size)
-> FAIL expected 32 actual 5 / 1 / 6 (OperandLen asserts)
RESTORE -> PASS (~1.0s; + TestUpsertSystemSetting_Concurrent_OneRow_TASK066 ok)
```

Production restored (no `MUTATED` leftover). `git grep MUTATED` empty on `handler_alerts.go`.

## Residual

- W5-A1-010: mock≠Ent GAP_NOTED (не блокирует Закрыта при PROVEN).
- leftover `MUTATED` в prod: **нет**.
