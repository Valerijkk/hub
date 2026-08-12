# Phase 32 — Verification (PW2 retro)

**Date:** 2026-07-27  
**Agent:** Grok 4.5  
**Verdict:** PASS

## Goal

OTLP trace export when `OTEL_EXPORTER_OTLP_ENDPOINT` is set (FB-21).

## Goal-backward check

| Criterion | Expected | Actual (2026-07-27) | Status |
|-----------|----------|---------------------|--------|
| `observability/otel.go` — `InitTracing`, `HTTPTraceMiddleware` | present | `backend/internal/platform/observability/otel.go` (`InitTracing` :42, `HTTPTraceMiddleware` :97, `TracingEnabled` :36) | PASS |
| Wired in server | init + middleware after RequestID | `wiring_bootstrap.go:81` `InitTracing`; `wiring_fiber.go:42` `HTTPTraceMiddleware` | PASS |
| No endpoint → noop | `TracingEnabled=false` | Covered by `TestInitTracing_NoEndpointIsNoop` | PASS |
| With tracing → spans | Timed + HTTP | `TestTimed_CreatesSpanWhenTracingEnabled`, `TestHTTPTraceMiddleware_CreatesServerSpan` | PASS |
| Env example OTEL vars | documented | `.env.local.example:15–16`; `deployments/.env.example:63–64` | PASS |
| Unit tests | `otel_test.go` | `backend/testframework/platform/observability/otel_test.go` | PASS |

## Tests run (2026-07-27)

```text
cd backend
go test -tags hubtest -count=1 -timeout=120s ./testframework/platform/observability/...
# ok  github.com/valerii/hub/backend/testframework/platform/observability  ~1.4s
```

## Notes

- Live OTLP export to Tempo/Jaeger not exercised (no collector). Unit tests use tracetest recorder.
- OTEL tests relocated from colocated `otel_test.go` to `testframework/`.

## Sign-off

OTel wiring + noop/enabled unit paths **PASS**.
