# Phase 31 — OBS-01 Wave 3 (Prometheus Export)

**Goal:** Expose Timed() and HTTP metrics for Prometheus scrape (M2-053 partial).

## Deliverables

- [x] `hub_operation_calls_total{operation,result}` counter
- [x] `hub_operation_duration_milliseconds{operation}` histogram
- [x] `hub_http_requests_total{method,route,status}` counter
- [x] `hub_http_request_duration_milliseconds{method,route}` histogram
- [x] `GET /metrics` — promhttp scrape endpoint (unauthenticated, infra-scoped)
- [x] `HTTPMetricsMiddleware` in `main.go`
- [x] E2E `metrics.spec.ts`
- [x] Unit tests `prometheus_test.go`

## Verification

```bash
cd backend && go test ./internal/platform/observability/... -count=1
curl -s localhost:3000/metrics | findstr hub_operation
```

## Backlog (Phase 32+)

- OTel traces (Grafana Tempo)
- Auth/network policy for `/metrics` in prod K3s
- Grafana dashboards in deployments/

## Success criteria

- Scrape returns valid Prometheus text with hub_* metrics
- Timed() operations appear in both admin UI and `/metrics`
- Full `go test ./...` green
