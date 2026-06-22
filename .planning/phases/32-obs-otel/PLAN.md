# Phase 32 — OBS-01 Wave 4 (OpenTelemetry Traces)

**Goal:** OTLP trace export when `OTEL_EXPORTER_OTLP_ENDPOINT` is set (FB-21).

## Deliverables

- [x] `observability/otel.go` — `InitTracing`, `HTTPTraceMiddleware`
- [x] `Timed()` creates child spans when tracing enabled
- [x] W3C trace context propagation on HTTP requests
- [x] `hub.request_id` span attribute from RequestID middleware
- [x] Wired in `main.go` (init + middleware after RequestID)
- [x] Unit tests `otel_test.go` (tracetest recorder)
- [x] `.env.local.example` OTEL vars

## Enable locally

```bash
# Grafana Alloy / Tempo / Jaeger OTLP HTTP receiver on 4318
OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4318
OTEL_SERVICE_NAME=hub-backend
```

## Verification

```bash
cd backend && go test ./internal/platform/observability/... -count=1
```

## Success criteria

- Without OTEL endpoint: zero overhead noop (TracingEnabled=false)
- With endpoint: HTTP spans + Timed() operation spans exported via OTLP HTTP
- OBS-01 **COMPLETE** (metrics + UI + Prometheus + traces)

## Backlog (Phase 33+)

- K3s network policy for `/metrics`
- Grafana dashboards in deployments/
- Hub One v4 features
