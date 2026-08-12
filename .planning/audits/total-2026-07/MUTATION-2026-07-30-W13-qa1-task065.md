# MUTATION W13 slot 4 refill — TASK-065 (qa-auto1) · 2026-07-30

> Checker: **qa-auto1** · Env: localhost `POSTGRES_TEST_URL` / `MONGODB_TEST_URL` · Go 1.25  
> Criteria: `.planning/backlog/archive/legacy/TASK-065-be-p3-platform-http-obs.md` `[MUTATION CRITERIA]`  
> Classes: **P3-BE-11**, **P3-BE-12** · Коммит: **не создавался** (инструкция слота).

## Verdict

**7/9 PROVEN** · **2 N/A** (layer_mismatch weak) · **1 GAP_NOTED** · **0 NOT_PROVEN** · TASK-065 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| W4-A1-008 | **PROVEN** | `TestParseOpenGraphHTML_ClearsUnsafeImage_P3BE11` | og:image assign без `ValidateOutboundURL` | HTTP parse (OG) |
| W4-A1-009 | **PROVEN** | `TestValidateOutboundURL` (blocked_* IP literal) | убрать IP-literal `isBlockedIP` block | HTTP ValidateOutboundURL |
| W4-A1-011 | **PROVEN** | `TestParseCORSOrigins_TrimsAndRejectsWildcardAndEmpty` (null/path/query/ftp) | `if false && !isAbsoluteCORSOrigin` | middleware CORS |
| W4-A2-006 | **N/A** | `TestClient_Close_Drains_P3BE12` stays GREEN | bare `Conn.Close` без Drain | NATS client (**layer_mismatch**/weak) |
| W4-A2-009 | **PROVEN** | `TestProcessBatch_CancelMidBatch_Stops_P3BE12` | убрать `select ctx.Done()` mid-batch | outbox worker |
| W4-A3-005 | **N/A** | `TestInitTracing_PropagatorWithoutBaggage_P3BE12` stays GREEN | `InitTracing` + Baggage | observability (**layer_mismatch**) |
| W4-A3-006 | **GAP_NOTED** | — (нет Exact RED) | `Debug→Info` «operation completed» | observability Timed |
| W4-A3-008 | **PROVEN** | `TestSanitizeOperationLabel_UnknownToOther_P3BE12`; `TestRecordOperation_UnknownBucketsToOther_P3BE12` | allowlist → pass-through | observability metrics |
| W4-A4-012 | **PROVEN** | `TestReady_ProbeParentIsRequestContext_P3BE12` | `probeParentContext` → `Background` | health Ready |

Production restored (`git checkout --` touched guards) после каждого BREAK. Leftover MUTATED: **нет**.  
`layer_mismatch IDs: W4-A2-006, W4-A3-005` · `mock≠Ent IDs: (none)` · `GAP IDs: W4-A3-006`.

## Per-ID evidence

### W4-A1-008 — og:image ValidateOutboundURL

```text
BASELINE TestParseOpenGraphHTML_ClearsUnsafeImage_P3BE11 → PASS (~5.0s)

BREAK: backend/internal/platform/http/opengraph.go
  ValidateOutboundURL gate → meta.ImageURL = content // MUTATED W4-A1-008
→ FAIL file_scheme / link-local / loopback / rfc1918
        unsafe og:image must be cleared (ImageURL populated)
RESTORE → PASS (~5.5s)
```

### W4-A1-009 — IP-literal block in ValidateOutboundURL

```text
BASELINE TestValidateOutboundURL → PASS (~5.0s)

BREAK: backend/internal/platform/http/safedialer.go
  ParseIP+isBlockedIP → removed // MUTATED W4-A1-009
→ FAIL blocked_link-local_IP_literal / blocked_loopback / blocked_IPv4-mapped
        ErrorIs ErrBlockedAddress got nil
RESTORE → PASS (~5.2s)
```

### W4-A1-011 — isAbsoluteCORSOrigin reject

```text
BASELINE TestParseCORSOrigins_TrimsAndRejectsWildcardAndEmpty → PASS (~5.1s)

BREAK: backend/internal/platform/middleware/cors.go parseCORSOrigins
  if !isAbsoluteCORSOrigin(o) → if false && !isAbsoluteCORSOrigin(o) // MUTATED W4-A1-011
  (compile-safe; gutting isAbsoluteCORSOrigin alone unused-imports → build fail ≠ assertion RED)
→ FAIL null_string / path_rejected / query_rejected / ftp_rejected / mixed_valid_invalid
        bad origins accepted into wantOrigins; wantProbs=0
RESTORE → PASS (~4.7s)
```

### W4-A2-006 — NATS Close Drain · **N/A** (layer_mismatch)

```text
BASELINE TestClient_Close_Drains_P3BE12 → PASS (~5.3s)

BREAK: backend/internal/platform/nats/client.go Close
  Drain()+timeout → c.Conn.Close() only // MUTATED W4-A2-006
→ GREEN (IsClosed + ExportNATSDrainTimeout still hold; Drain not asserted)
RESTORE → PASS (~4.6s)

N/A — criteria weak/layer_mismatch: named test не убивает Drain-path.
Не NOT_PROVEN (pre-noted weak kill; runnable evidence зафиксирован).
```

### W4-A2-009 — outbox cancel mid-batch

```text
BASELINE TestProcessBatch_CancelMidBatch_Stops_P3BE12 → PASS (~8.0s)

BREAK: backend/internal/platform/outbox/worker.go
  select ctx.Done() mid-batch removed // MUTATED W4-A2-009
→ FAIL expected processed==1, got 5
RESTORE → PASS (~6.4s)
```

### W4-A3-005 — TraceContext without Baggage · **N/A** (layer_mismatch)

```text
BASELINE TestInitTracing_PropagatorWithoutBaggage_P3BE12 → PASS (~5.1s)

BREAK: backend/internal/platform/observability/otel.go InitTracing
  TraceContext{} → Composite(TraceContext, Baggage) // MUTATED W4-A3-005
→ GREEN (тест вручную otel.SetTextMapPropagator(TraceContext{}), не вызывает InitTracing)
RESTORE → PASS (~4.6s)

N/A — layer_mismatch: Exact test не бьёт prod InitTracing+OTLP path.
```

### W4-A3-006 — Timed success log level · **GAP_NOTED**

```text
GAP — нет Exact RED на slog level (criteria).
BREAK Debug→Info «operation completed» // MUTATED W4-A3-006
  package scan TestTimed|TestSanitize|TestInitTracing|TestRecord → GREEN (~5.7s)
RESTORE — clean
```

### W4-A3-008 — sanitize unknown → other

```text
BASELINE Sanitize/Record Unknown*P3BE12 → PASS (~4.6s)

BREAK: backend/internal/platform/observability/operation_label.go
  sanitize allowlist → return TrimSpace(op) // MUTATED W4-A3-008
→ FAIL UUID / user-controlled ≠ "other"; RecordOperation buckets raw UUID
RESTORE → PASS (~4.9s)
```

### W4-A4-012 — Ready probe parent = c.Context()

```text
BASELINE TestReady_ProbeParentIsRequestContext_P3BE12 → PASS (~4.7s)

BREAK: backend/internal/platform/health/probes.go
  probeParentContext → context.Background() // MUTATED W4-A4-012
→ FAIL Export ≠ c.Context()
RESTORE → PASS (~4.9s)
```

```text
leftover_mutated=False
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -race -count=1 -timeout=180s ./testframework/platform/http/ -run TestParseOpenGraphHTML_ClearsUnsafeImage_P3BE11
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/platform/http/ -run TestValidateOutboundURL
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/platform/middleware/ -run TestParseCORSOrigins_TrimsAndRejectsWildcardAndEmpty
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/platform/nats/ -run TestClient_Close_Drains_P3BE12
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/platform/outbox/ -run TestProcessBatch_CancelMidBatch_Stops_P3BE12
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/platform/observability/ -run TestInitTracing_PropagatorWithoutBaggage_P3BE12
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/platform/observability/ -run "TestSanitizeOperationLabel_UnknownToOther_P3BE12|TestRecordOperation_UnknownBucketsToOther_P3BE12"
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/platform/health/ -run TestReady_ProbeParentIsRequestContext_P3BE12
```

Runner: `.planning/audits/total-2026-07/_run_mutations_w13_task065.py`  
JSON: `_mut-batch-w13-task065.json`.  
W4-A1-011 re-prove (compile-safe `if false &&`): inline after gut-function build-fail false-positive.

## Gaps

- **W4-A2-006** — Drain kill weak: тест проверяет IsClosed/timeout export, не факт Drain → **N/A** layer_mismatch (не reopen).
- **W4-A3-005** — тест вручную ставит TraceContext, не InitTracing → **N/A** layer_mismatch (не reopen).
- **W4-A3-006** — нет hubtest slog capture / Exact RED на log level → **GAP_NOTED** (не reopen).

## Handoff

- Evidence: этот файл · TASK-065 **Закрыта** · README sync  
- be-ver → code-reviewer (fix/tests already FIXED; mutation 7 PROVEN + 2 N/A + 1 GAP)  
- Коммит: не создавался
