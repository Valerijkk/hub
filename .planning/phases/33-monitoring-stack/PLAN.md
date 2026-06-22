# Phase 33 — Monitoring Stack (M2-056 partial)

**Goal:** Grafana dashboard provisioning + Prometheus scrape + `/metrics` hardening.

## Wave 1 — Deploy artifacts

- [x] `deployments/monitoring/prometheus.yml` — scrape `backend:3000/metrics`
- [x] `deployments/monitoring/grafana/` — datasource + dashboard provisioning
- [x] `deployments/monitoring/grafana/dashboards/hub-observability.json`
- [x] `deployments/docker-compose.monitoring.yml` — Prometheus + Grafana overlay

## Wave 2 — Security

- [x] `METRICS_ALLOWED_CIDRS` env + `MetricsAllowMiddleware`
- [x] `deployments/k3s/monitoring/networkpolicy-metrics.yaml` template
- [x] `deployments/.env.example` observability vars

## Usage

```bash
cd deployments
docker compose -f docker-compose.prod.yml -f docker-compose.monitoring.yml up -d
open http://localhost:3001   # Grafana
open http://localhost:9090   # Prometheus
```

## Verification

```bash
cd backend && go test ./internal/platform/observability/... ./internal/platform/config/... -count=1
curl -s http://localhost:3000/metrics | findstr hub_http_requests_total
```

## Backlog (Phase 34+)

- ~~Loki log aggregation (M2-057)~~ → Phase 34
- ~~Alert rules (M2-058)~~ → Phase 34 (Alertmanager backlog)
- Hub One v4 features
