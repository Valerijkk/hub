# MUTATION W13 slot 4 refill — TASK-153 (qa-auto1) · 2026-07-30

> Checker: **qa-auto1** · Env: localhost `POSTGRES_TEST_URL` / `MONGODB_TEST_URL` · Go 1.25  
> Criteria: `.planning/backlog/TASK-153-2026-07-29-a3-5-1-ai-handlers.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**3/3 PROVEN** · TASK-153 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| A3W5-1-001 | **PROVEN** | `TestHandler_ChatCompletions_Stream_NoProvider` | drop stream NoProvider/Disabled→503 + AllFailed→502 → catch-all 400+`err.Error()` | handler HTTP map |
| A3W5-1-002 | **PROVEN** | `TestHandler_AssistJQL_EmptyQuery` | drop `ErrQueryRequired`→400 (empty → catch-all 502) | handler+service sentinel |
| A3W5-1-003 | **PROVEN** | `…_Stream_DailyQuotaExceeded` + `…_Stream_ConsentRequired` + `…_Stream_NoProvider` | invert stream quota≠429 / consent≠403 / no-provider≠503 | handler HTTP map |

Production restored (`git checkout -- internal/ai/handler.go`) после каждого BREAK. Leftover MUTATED: **нет**.  
`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)`.

## Per-ID evidence

### A3W5-1-001 — stream NoProvider→503

```text
BASELINE → PASS (~2.1s)

BREAK: handler.go stream — remove ErrNoProvider/Disabled→503 + AllFailed→502;
       catch-all → StatusBadRequest + err.Error()
→ FAIL: TestHandler_ChatCompletions_Stream_NoProvider (expect 503, got 400)
RESTORE → PASS (~1.9s)
```

### A3W5-1-002 — AssistJQL empty→400

```text
BASELINE → PASS (~2.0s)

BREAK: handler.go AssistJQL — remove errors.Is(ErrQueryRequired)→400
→ FAIL: TestHandler_AssistJQL_EmptyQuery (expect 400, got 502) ×3 query variants
RESTORE → PASS (~2.0s)
```

### A3W5-1-003 — stream quota/consent/no-provider matrix

```text
BASELINE → PASS (~2.1s)

BREAK: handler.go stream — quota 429→400, consent 403→400, no-provider 503→400
→ FAIL: Stream_DailyQuotaExceeded (429→400)
→ FAIL: Stream_ConsentRequired (403→400)
→ FAIL: Stream_NoProvider (503→400)
RESTORE → PASS (~2.1s)
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -race -count=1 -timeout=120s ./testframework/ai/ `
  -run "TestHandler_ChatCompletions_Stream_|TestHandler_AssistJQL_EmptyQuery"
```

Final GREEN after restore: `ok …/testframework/ai 2.110s`.

## Handoff

- Evidence: этот файл · TASK-153 **Закрыта** · README sync  
- be-ver → code-reviewer (fix code already FIXED; mutation PROVEN)  
- Коммит: не создавался
