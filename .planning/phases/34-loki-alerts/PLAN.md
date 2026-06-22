# Phase 34 — Loki + Prometheus Alerts (M2-057 / M2-058 partial)

**Goal:** Log aggregation via Loki/Promtail + Prometheus alert rules for Hub SLOs.

## Wave 1 — Loki stack

- [x] `deployments/monitoring/loki/loki-config.yml`
- [x] `deployments/monitoring/promtail/promtail-config.yml` — Docker SD + JSON pipeline
- [x] Loki + Promtail services in `docker-compose.monitoring.yml`
- [x] Grafana Loki datasource (`hub-loki`)

## Wave 2 — Alerts + log correlation

- [x] `deployments/monitoring/prometheus/alerts/hub.yml` — 5xx rate, p95 latency, operation errors
- [x] Prometheus `rule_files` in `prometheus.yml`
- [x] Dashboard logs panel + datasource UIDs
- [x] `slog` default logger includes `service` label (Loki JSON filter)

## Alert rules (M2-058 partial)

| Alert | Threshold | Duration |
|-------|-----------|----------|
| HubHighHTTP5xxRate | > 1% | 5m |
| HubHighHTTPLatencyP95 | > 2000ms | 5m |
| HubOperationErrorRate | > 5% per operation | 10m |

View firing alerts: Prometheus UI → Alerts tab.

## Usage

```bash
cd deployments
docker compose -f docker-compose.prod.yml -f docker-compose.monitoring.yml up -d
# Grafana → Explore → Loki: {compose_service="backend"}
# Prometheus → Alerts
```

## Backlog (Phase 35+)

- ~~Alertmanager + notification routes~~ → Phase 35
- Hub One v4 features
- PG-02 Biome lint cleanup

## Verification

```bash
cd backend && go build ./cmd/server/...
# Manual: Loki ready http://localhost:3100/ready
```
