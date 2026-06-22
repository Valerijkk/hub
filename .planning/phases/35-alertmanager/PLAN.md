# Phase 35 — Alertmanager (M2-058 complete)

**Goal:** Route Prometheus alerts through Alertmanager; Grafana AM datasource.

## Deliverables

- [x] `deployments/monitoring/alertmanager/alertmanager.yml` — routes, inhibit rules, null receivers (dev)
- [x] `alertmanager.prod.example.yml` — webhook template for prod ops
- [x] Prometheus `alerting.alertmanagers` → `alertmanager:9093`
- [x] Alertmanager service in `docker-compose.monitoring.yml` (`:9093`)
- [x] Grafana Alertmanager datasource (`hub-alertmanager`)
- [x] New alert: `HubBackendMetricsDown` (critical, `up==0`)

## Usage

```bash
cd deployments
docker compose -f docker-compose.prod.yml -f docker-compose.monitoring.yml up -d
open http://localhost:9093   # Alertmanager UI — active alerts
open http://localhost:9090/alerts  # Prometheus rules status
```

Prod webhooks: copy `alertmanager.prod.example.yml` → `alertmanager.yml`.

## Success criteria

- Prometheus → Alertmanager connectivity configured
- Firing alerts visible in AM UI
- M2-058 **COMPLETE** (rules + routing; notification channel = ops webhook)

## Backlog (Phase 36+)

- Hub One v4
- PG-02 Biome lint
- Native Telegram admin alert hook (optional BE endpoint)
