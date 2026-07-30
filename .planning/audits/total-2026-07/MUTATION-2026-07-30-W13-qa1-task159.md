# MUTATION W13 slot 4 refill — TASK-159 (qa-auto1) · 2026-07-30

> Checker: **qa-auto1** · Env: localhost `POSTGRES_TEST_URL` / `MONGODB_TEST_URL` · Go 1.25  
> Criteria: `.planning/backlog/TASK-159-2026-07-29-a3-6-3-rbac.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**3/3 PROVEN** · **0 NOT_PROVEN** · TASK-159 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| A3W6-3-001 | **PROVEN** | `TestDeleteCustomRole_CacheFailNotError_TASK159` + `TestDeleteCustomRole_CacheFailHTTP200_TASK159` | `return invErr` after `InvalidateProject` (hard-fail) | service + HTTP |
| A3W6-3-002 | **PROVEN** | `TestRequireSystemRole_AdminAllowsFounder_TASK159` | pure `string(u.SystemRole) != role` без `IsSystemAdmin` | middleware |
| A3W6-3-003 | **PROVEN** | `TestInvalidateByPatternScript_DELNotUNLINK_TASK159` | comment `SCAN+UNLINK` при Lua `DEL` | cache Lua + godoc |

Production restored (`git checkout --`) после каждого BREAK. Leftover MUTATED: **нет**.  
`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)`.

## Per-ID evidence

### A3W6-3-001 — DeleteCustomRole cache-fail best-effort

```text
BASELINE → PASS (~6.8s)

BREAK: custom_roles.go — после InvalidateProject fail → return invErr
→ FAIL: TestDeleteCustomRole_CacheFailNotError_TASK159
        cache-only fail must not surface after successful Tx delete
→ FAIL: TestDeleteCustomRole_CacheFailHTTP200_TASK159
        expected HTTP 200 (delete must succeed after Tx even if Redis down)
RESTORE → PASS (~6.2s)
```

### A3W6-3-002 — RequireSystemRole("admin") allows founder

```text
BASELINE → PASS (~7.1s)

BREAK: middleware.go — убрать IsSystemAdmin branch; pure string equality
→ FAIL: founder_allowed — expected 200, got 403 "insufficient system role"
RESTORE → PASS (~7.7s)
```

### A3W6-3-003 — Lua DEL ↔ comment SCAN+DEL sync

```text
BASELINE → PASS (~6.1s)

BREAK: cache.go — comment/godoc SCAN+UNLINK при Lua redis.call("DEL", …)
→ FAIL: should not contain "SCAN+UNLINK"
→ FAIL: does not contain "SCAN+DEL"
RESTORE → PASS (~6.6s)
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -race -count=1 -timeout=180s ./testframework/rbac/ `
  -run "TestDeleteCustomRole_CacheFail(NotError|HTTP200)_TASK159|TestRequireSystemRole_AdminAllowsFounder_TASK159|TestInvalidateByPatternScript_DELNotUNLINK_TASK159"
```

Runner: `.planning/audits/total-2026-07/_run_mutations_w13_task159.py` · JSON: `_mut-batch-w13-task159.json`.

## Gaps

- Stale elevated bitmask до TTL при best-effort invalidate — accepted residual (не kill target).
- 001 HTTP assert: expected 200 vs non-200 (handler maps returned invErr → error status); service assert также RED.

## Handoff

- Evidence: этот файл · TASK-159 **Закрыта** · README sync  
- be-ver → code-reviewer (fix code already FIXED; mutation PROVEN)  
- Коммит: не создавался
