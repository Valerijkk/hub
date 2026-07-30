# MUTATION W13 slot 4 refill — TASK-154 (qa-auto1) · 2026-07-30

> Checker: **qa-auto1** · Env: localhost `POSTGRES_TEST_URL` / `MONGODB_TEST_URL` · Go 1.25  
> Criteria: `.planning/backlog/TASK-154-2026-07-29-a3-5-2-ai-adapters.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**3/3 PROVEN** · TASK-154 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| A3W5-2-001 | **PROVEN** | `TestOpenAIAdapter_OfflineStreamClient_NoWholeRequestTimeout_A3W52_001` (+ Delayed) | `offlineStreamClient` → `NewOfflineAIHTTPClient(90s)` | adapter HTTP stream |
| A3W5-2-002 | **PROVEN** | `TestOpenAIAdapter_StreamHTTPError_ReadBodyErr_A3W52_002` | `raw, _ := io.ReadAll(...)` (swallow err) | adapter SSE error-body |
| A3W5-2-003 | **PROVEN** | `TestCheckContentPolicy_HomoglyphAndInsertBypass_A3W52_003` | `policyScanForms` → only raw text (no fold/compact) | content policy |

Production restored (`git checkout --`) после каждого BREAK. Leftover MUTATED: **нет**.  
`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)`.

## Per-ID evidence

### A3W5-2-001 — offline SSE без Client.Timeout

```text
BASELINE → PASS (~2.1s)

BREAK: adapter_openai.go — offlineStreamClient = NewOfflineAIHTTPClient(90s)
→ FAIL: TestOpenAIAdapter_OfflineStreamClient_NoWholeRequestTimeout_A3W52_001
        (expect Timeout=0, got 1m30s) ×2 asserts
→ FAIL: TestOpenAIAdapter_OfflineStream_DelayedFirstEvent_A3W52_001
        (resolved.Timeout expect 0, got 1m30s)
RESTORE → PASS (~2.0s)
```

### A3W5-2-002 — ReadAll err на error-body

```text
BASELINE → PASS (~2.0s)

BREAK: adapter_stream.go — raw, _ := io.ReadAll(...); drop "read error body" wrap
→ FAIL: TestOpenAIAdapter_StreamHTTPError_ReadBodyErr_A3W52_002
        (got "provider HTTP 502: ", missing "read error body")
RESTORE → PASS (~2.0s)
```

### A3W5-2-003 — homoglyph / insert denylist layer

```text
BASELINE → PASS (~2.0s)

BREAK: content_policy.go — policyScanForms returns []string{text} only
→ FAIL: TestCheckContentPolicy_HomoglyphAndInsertBypass_A3W52_003
        (homoglyph → nil, expect ErrContentPolicyViolation)
RESTORE → PASS (~2.9s final suite)
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -race -count=1 -timeout=120s ./testframework/ai/ `
  -run "TestOpenAIAdapter_OfflineStreamClient_NoWholeRequestTimeout_A3W52_001|TestOpenAIAdapter_OfflineStream_DelayedFirstEvent_A3W52_001|TestOpenAIAdapter_StreamHTTPError_ReadBodyErr_A3W52_002|TestCheckContentPolicy_HomoglyphAndInsertBypass_A3W52_003"
```

Final GREEN after restore: `ok …/testframework/ai 2.863s`.

## Gaps

- DelayedFirstEvent secondary (Timeout assert + cancel); канон kill = NoWholeRequestTimeout — **PROVEN**.
- Cloud `streamClient` / ValidateOutboundURL не трогались.

## Handoff

- Evidence: этот файл · TASK-154 **Закрыта** · README sync  
- be-ver → code-reviewer (fix code already FIXED; mutation PROVEN)  
- Коммит: не создавался
