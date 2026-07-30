# MUTATION W13 slot 4 refill — TASK-155 (qa-auto1) · 2026-07-30

> Checker: **qa-auto1** · Env: localhost `POSTGRES_TEST_URL` / `MONGODB_TEST_URL` · Go 1.25  
> Criteria: `.planning/backlog/TASK-155-2026-07-29-a3-5-3-ai-service.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**8/9 PROVEN** · **1 PROVEN + GAP_NOTED** (A3W5-3-009 mock≠Ent) · **0 NOT_PROVEN** · TASK-155 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| A3W5-3-001 | **PROVEN** | `TestNormalizeComposerCompletion_UTF8CaseFold_TASK155` | case-sensitive `HasPrefix` (EqualFold removed) | service pure |
| A3W5-3-002 | **PROVEN** | `TestIncrementUsage_ParallelRace_TASK155` | Query→Create/Update check-then-act | repository Ent |
| A3W5-3-003 | **PROVEN** | `TestEntSettingsStore_ConcurrentUpsert_TASK155` | Update-n==0→Create, no conflict retry | repository Ent settings |
| A3W5-3-004 | **PROVEN** | `TestTrackerIssueLookup_NilAccessDeny_TASK155` | fail-open `if access != nil` | service lookup |
| A3W5-3-005 | **PROVEN** | `TestSaveUserPreference_ModelOnlyAllowFalse_TASK155` | nil → `SetAllowExternalAi(true)` | repository Ent |
| A3W5-3-006 | **PROVEN** | `TestUpdateProvider_VersionConflict_TASK155` (+ `TestSaveUserPreference_StaleVersion_TASK155`) | drop `Where(VersionEQ)` | repository Ent |
| A3W5-3-007 | **PROVEN** | `TestStartChatStream_ErrorSanitized_TASK155` | `publish(..., err.Error())` | service WS |
| A3W5-3-008 | **PROVEN** | compile gate (`TestEnsureBotUser_*` / call sites) | `EnsureBotUser(ctx, *ent.Client)` | layering |
| A3W5-3-009 | **PROVEN** + **GAP_NOTED** | `TestMessengerBot_OnMessageSent_DMStreamReply` (+ Help/Docs) | final reply → `WRONG_REPLY_MUTATED` | bot flow / HookTestRepo |

Production restored после каждого BREAK. Leftover MUTATED: **нет** (`git diff --stat backend/internal/ai/` empty).  
`layer_mismatch IDs: (none)` · `mock≠Ent IDs: A3W5-3-009` (GAP_NOTED, не NOT_PROVEN).

## Per-ID evidence

### A3W5-3-001 — EqualFold / UTF-8 trim

```text
BASELINE → PASS

Literal kill (ToLower + s[len(prefix):]) → PASS на текущих фикстурах
  (кириллица/ASCII: len(prefix) == байты EqualFold-префикса в s)

BREAK: case-sensitive HasPrefix only (EqualFold removed)
→ FAIL: expect "продолжение"/"ок"/"world", got untrimmed mixed-case
RESTORE → PASS
```

### A3W5-3-002 — IncrementUsage race

```text
BREAK: Query→Create/Update check-then-act
→ FAIL: database table is locked (goroutine 0) under -race
RESTORE → PASS (~0.04s)
```

### A3W5-3-003 — settings Upsert race

```text
BREAK: Update without VersionEQ/retry; Create on miss
→ FAIL: UNIQUE constraint failed: system_settings.key
RESTORE → PASS (~0.10s)
```

### A3W5-3-004 — issue lookup fail-closed

```text
BREAK: if access != nil { CanView… } (skip when nil)
→ FAIL/panic: nil TrackerService.GetIssue (fail-open reached GetIssue)
RESTORE → PASS
```

### A3W5-3-005 — consent default false on create

```text
Note: schema Default(false) already set — bare remove SetAllowExternalAi(false)
would stay GREEN via Ent default.

BREAK: nil path → SetAllowExternalAi(true) (Default true path)
→ FAIL: Should be false (AllowExternalAI)
RESTORE → PASS
```

### A3W5-3-006 — VersionEQ optimistic lock

```text
BREAK: drop Where(VersionEQ) on provider (+ preference VersionEQ;
       preference equality pre-check temporarily relaxed to expose Ent guard)
→ FAIL: both stale tests — Expected ErrVersionConflict, got nil
RESTORE → PASS
```

### A3W5-3-007 — WS stream sanitize

```text
BREAK: publish error message = err.Error()
→ FAIL: expected "AI stream failed", got "driver: llm-secret-dsn=internal"
RESTORE → PASS
```

### A3W5-3-008 — BotUserRepository layering (compile gate)

```text
BREAK: EnsureBotUser(ctx, *ent.Client)
→ build failed: cannot use *EntBotUserRepository as *ent.Client
  (bot_user_test.go:22/26/39, scope_routing_task066, task113_policy_rpm)
RESTORE → PASS (runtime LayeringNoEntClient)
```

Не classic `-run` RED — **PROVEN** как compile/type gate per criteria Gaps.

### A3W5-3-009 — bot flow body asserts · mock≠Ent

```text
BREAK: messenger_bot.streamReply final answer = "WRONG_REPLY_MUTATED"
→ FAIL: DMStreamReply (no "Привет"); HelpCommand (no "commands");
        HandleCommand_Docs (no "answer from docs")
RESTORE → PASS
```

**GAP_NOTED:** asserts идут через `HookTestRepo` in-memory messenger, не Ent messenger persistence. Body assert **не hollow** (PROVEN), Ent layer **не доказан**. PublicMention остаётся NoError-only residual (criteria Gaps).

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -race -count=1 -timeout=180s ./testframework/ai/ `
  -run "TestNormalizeComposerCompletion_UTF8CaseFold_TASK155|TestIncrementUsage_ParallelRace_TASK155|TestEntSettingsStore_ConcurrentUpsert_TASK155|TestTrackerIssueLookup_NilAccessDeny_TASK155|TestSaveUserPreference_ModelOnlyAllowFalse_TASK155|TestUpdateProvider_VersionConflict_TASK155|TestSaveUserPreference_StaleVersion_TASK155|TestStartChatStream_ErrorSanitized_TASK155|TestEnsureBotUser_LayeringNoEntClient|TestMessengerBot_OnMessageSent_DMStreamReply|TestMessengerBot_OnMessageSent_HelpCommand|TestMessengerBot_HandleCommand_Docs"
```

Final GREEN after restore: `ok …/testframework/ai 2.338s`.

## Gaps

- 001: literal `ToLower+len(prefix)` kill не RED на текущих фикстурах; EqualFold-removal — RED → PROVEN.
- 005: schema `Default(false)` уже в коде; kill через invert create-path.
- 008: compile gate, не runtime RED.
- 009: **GAP_NOTED** mock≠Ent; PublicMention hollow residual.
- 005 Atlas migrate DEFAULT — owner follow-up (не блокирует mutation close).

## Handoff

- Evidence: этот файл · TASK-155 **Закрыта** · README sync  
- be-ver → code-reviewer (fix code already FIXED; mutation PROVEN / GAP_NOTED only)  
- Коммит: не создавался
