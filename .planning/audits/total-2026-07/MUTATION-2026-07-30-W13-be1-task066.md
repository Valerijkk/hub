# MUTATION W13 slot 1 refill — TASK-066 (be1) · 2026-07-30

> Checker: **be1** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/archive/legacy/TASK-066-be-p3-ai-auth-rbac-audit.md` `[MUTATION CRITERIA]`  
> Driver: `.planning/audits/total-2026-07/_run_mutations_w13_task066.py`  
> Raw: `_mut-batch-w13-task066.json` (+ `_mut-batch-w13-task066-rerun.json` for W5-A2-012 / W6-A4-010)  
> Коммит не создавался (инструкция слота).

## Verdict

**9/10 PROVEN** (из них **1 PROVEN + GAP_NOTED** mock≠Ent) · **1 NOT_PROVEN** (W6-A4-010) · leftover MUTATED: **нет**  
TASK-066 → **Открыта** (NOT_PROVEN residual)

| ID | Verdict | RED |
|----|---------|-----|
| W5-A1-010 | **PROVEN** + **GAP_NOTED** | `TestAdminUpdate_UserScopeProvider_404_TASK066` |
| W5-A2-012 | **PROVEN** | `TestIsYandexProvider` |
| W5-A3-013 | **PROVEN** | `TestFetchReview_GitLabURL_NilGitLab_TASK066` |
| W5-A3-014 | **PROVEN** | `TestTrackerIssueLookup_NilSafe` |
| W5-A3-015 | **PROVEN** | `TestEnsureBotUser_Parallel_OneUser_TASK066` |
| W6-A1-010 | **PROVEN** | `TestCompleteOnboarding_TxRollback_TASK066` |
| W6-A3-012 | **PROVEN** | `TestUnassignRole_WrongRoleID_404_TASK066` |
| W6-A3-014 | **PROVEN** | `TestInvalidateProject_LuaClearsKeys_TASK066` |
| W6-A4-010 | **NOT_PROVEN** | `TestAlertWebhook_CTCFixedDigest_TASK066` |
| W6-A4-012 | **PROVEN** | `TestUpsertSystemSetting_Concurrent_OneRow_TASK066` |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: W5-A1-010` · `docs-only IDs: (none)`

## Per-ID evidence

### W5-A1-010 — requireSystemProvider (admin BYOK 404)

```text
BASELINE TestAdminUpdate_UserScopeProvider_404_TASK066 -> PASS (~7.2s)
BREAK: service.go requireSystemProvider — accept any Scope
-> FAIL expected 404 got 409 (PATCH) / 200 (test)
RESTORE -> PASS
```

Guard: `backend/internal/ai/service.go:277-287`. Primary RED via Fiber+**mockRepo** → **GAP_NOTED** mock≠Ent (Ent twin желателен).

### W5-A2-012 — IsYandexProvider host allowlist

```text
BASELINE TestIsYandexProvider -> PASS
BREAK: return strings.Contains(..., "yandex") (url.Parse retained for compile)
-> FAIL assert.False yandexgpt / evil.example/yandex
RESTORE -> PASS
```

### W5-A3-013 — GitLab nil → explicit error

```text
BASELINE TestFetchReview_GitLabURL_NilGitLab_TASK066 -> PASS
BREAK: drop gitlab==nil in FetchReview + authorizeReview (github fallback)
-> FAIL err does not contain "GitLab not configured"
RESTORE -> PASS
```

### W5-A3-014 — nil IssueContext

```text
BASELINE TestTrackerIssueLookup_NilSafe -> PASS
BREAK: return "", "", nil on nil deps
-> FAIL require.Error got nil
RESTORE -> PASS
```

### W5-A3-015 — bot Create conflict re-query

```text
BASELINE TestEnsureBotUser_Parallel_OneUser_TASK066 -> PASS
BREAK: drop re-query on Create err
-> FAIL UNIQUE constraint / NoError
RESTORE -> PASS
```

### W6-A1-010 — CompleteOnboarding WithEntTx

```text
BASELINE TestCompleteOnboarding_TxRollback_TASK066 -> PASS
BREAK: WithEntTx = fn(ctx, r) without Tx
-> FAIL DisplayName Equal("before") got "ShouldNotPersist"
RESTORE -> PASS
```

### W6-A3-012 — Unassign roleId match

```text
BASELINE TestUnassignRole_WrongRoleID_404_TASK066 -> PASS
BREAK: ClearCustomRole skip custom.ID != roleID
-> FAIL expected 404 got 200; Has(PermTaskCreate) false
RESTORE -> PASS
```

### W6-A3-014 — InvalidateProject Lua DEL

```text
BASELINE TestInvalidateProject_LuaClearsKeys_TASK066 -> PASS
BREAK: Lua SCAN without DEL
-> FAIL assert.False ok (keys remain)
RESTORE -> PASS
```

### W6-A4-010 — SHA-256 digest CTC — NOT_PROVEN

```text
BASELINE TestAlertWebhook_CTCFixedDigest_TASK066 -> PASS
BREAK: subtle.ConstantTimeCompare([]byte(a),[]byte(b)) without digest (_ = sha256.Size)
-> stayed GREEN (~7.0s) — unequal-length still CTC=0; Bearer x still 401
RESTORE -> PASS
```

Named test asserts only equality outcome / 401, not fixed-size digest property. Kill per recipe does not trip RED → **NOT_PROVEN**. Need digest-sensitive assert (или probe length-oracle path) для закрытия.

### W6-A4-012 — Upsert constraint retry

```text
BASELINE TestUpsertSystemSetting_Concurrent_OneRow_TASK066 -> PASS
BREAK: Create without constraint retry
-> FAIL concurrent UNIQUE constraint / NoError
RESTORE -> PASS
```

## Residual

- **W6-A4-010 NOT_PROVEN** — блокирует «Закрыта»; статус задачи **Открыта**.
- W5-A1-010: mock≠Ent GAP_NOTED (не блокирует сам по себе при PROVEN).
- leftover `MUTATED` в prod: **нет** (`git checkout` после каждого ID).
