# Phase 34 — Loki + Prometheus Alerts — VERIFICATION

**Date:** 2026-07-27  
**Auditor:** Grok 4.5 (PW2 retro-verify)  
**Mode:** documentation-only · deployments presence against HEAD  
**Verdict:** PASS

## Goal (exit criteria)

Loki/Promtail in monitoring compose · Grafana Loki datasource · Prometheus alert rules for Hub SLOs · slog `service` label.

## Measured evidence

| Criterion | Evidence (2026-07-27) | Result |
|-----------|----------------------|--------|
| Loki config | `deployments/monitoring/loki/loki-config.yml` | PASS |
| Promtail config | `deployments/monitoring/promtail/promtail-config.yml` | PASS |
| Compose services | `docker-compose.monitoring.yml` — `loki:` / `promtail:` services + volume mounts | PASS |
| Grafana Loki datasource | `grafana/provisioning/datasources/loki.yml` — `uid: hub-loki` | PASS |
| Alert rules file | `deployments/monitoring/prometheus/alerts/hub.yml` — `HubHighHTTP5xxRate`, `HubHighHTTPLatencyP95`, `HubOperationErrorRate` | PASS |
| `rule_files` wired | `deployments/monitoring/prometheus.yml` includes `rule_files:` | PASS |
| slog `service` label | `wiring_bootstrap.go` — `slog.New(...).With("service", serviceName)` + `slog.SetDefault` | PASS |

## Notes

- Loki ready probe / Prom UI Alerts tab **not** exercised live; structural configs only.
- Alertmanager routing is Phase 35 (not required here).

## Residual

None for PLAN Wave 1–2 deliverables.
