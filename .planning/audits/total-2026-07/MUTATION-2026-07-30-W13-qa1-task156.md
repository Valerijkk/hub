# MUTATION W13 slot 4 refill — TASK-156 (qa-auto1) · 2026-07-30

> Checker: **qa-auto1** · Env: localhost `POSTGRES_TEST_URL` / `MONGODB_TEST_URL` · Go 1.25  
> Criteria: `.planning/backlog/TASK-156-2026-07-29-a3-5-4-search-core.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**3/3 PROVEN** · **0 NOT_PROVEN** · TASK-156 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| A3W5-4-001 | **PROVEN** | `TestNewMeiliIndexerFromEnv_OperatorLoopbackDial_A3W54_001` | drop `NewOperatorMeiliHTTPClient` branch → only `NewSafeHTTPClient` | Meili client / SafeDial |
| A3W5-4-002 | **PROVEN** | `TestParseJQL_UnsupportedComparison_A3W54_002` | skip `validateJQLCondition` in Parse+ToEnt + `return nil` for priority `>` | JQL parse/predicate |
| A3W5-4-003 | **PROVEN** | `TestIndexAllIssues_PaginatesBeyondBatch_A3W54_003` | single `Limit(batchSize)` without offset-loop | indexer Ent query |

Production restored (`git checkout --`) после каждого BREAK. Leftover MUTATED: **нет**.  
`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)`.

## Per-ID evidence

### A3W5-4-001 — operator Meili dial allowlist

```text
BASELINE → PASS (~11.1s)

BREAK: meilisearch_indexer.go — убрать meiliAllowsOperatorPrivate → NewOperatorMeiliHTTPClient;
       остаётся только NewSafeHTTPClient
→ FAIL: operator Meili dial must allow loopback, got ErrBlockedAddress (127.0.0.1)
→ FAIL: operator Meili dial must allow RFC1918, got ErrBlockedAddress (10.0.0.5)
RESTORE → PASS (~7.8s)
```

Metadata / public SafeDial asserts остаются на restore path (не открывали SSRF).

### A3W5-4-002 — JQL unsupported comparison fail-closed

```text
BASELINE → PASS (~7.1s)

BREAK: jql.go — skip validateJQLCondition in Parse + ToEntPredicates;
       priority `>` → return nil (silent-nil)
→ FAIL: ParseJQL(`priority > medium`) — An error is expected but got nil (validate-first)
RESTORE → PASS (~6.9s)
```

Validate-first RED per criteria Gaps; silent-nil path также был сломан в том же BREAK (hand-built ToEnt never-match).

### A3W5-4-003 — IndexAllIssues pagination

```text
BASELINE → PASS (~7.2s)

BREAK: IndexAllIssues — один Limit(batchSize) без offset-loop
       (симуляция голого Limit(5000); batch=2 через ExportSetMeiliIndexBatchSize)
→ FAIL: count expected 3, got 2 (pagination must index all…)
→ FAIL: posts 1 is not >= 2
RESTORE → PASS (~7.2s)
```

Ent sqlite + fake Meili HTTP — **Ent OK**.

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -race -count=1 -timeout=180s ./testframework/search/ `
  -run "TestNewMeiliIndexerFromEnv_OperatorLoopbackDial_A3W54_001|TestParseJQL_UnsupportedComparison_A3W54_002|TestIndexAllIssues_PaginatesBeyondBatch_A3W54_003"
```

Runner: `.planning/audits/total-2026-07/_run_mutations_w13_task156.py` · JSON: `_mut-batch-w13-task156.json`.

## Gaps

- 002: criteria допускает mutate validate **или** silent-nil; прогон сломал оба, RED сработал на Parse error assert (prefer validate-first).
- 001: Hostname→RFC1918 без IP-литерала — ops residual, не в kill recipe.
- 003: kill через `Limit(batchSize)` once (тест overrides batch=2), эквивалент «голый Limit без pagination».

## Handoff

- Evidence: этот файл · TASK-156 **Закрыта** · README sync  
- be-ver → code-reviewer (fix code already FIXED; mutation PROVEN)  
- Коммит: не создавался
