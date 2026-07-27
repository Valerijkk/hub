# FIX-W16 — Block E ops (observability / deploy)

**Ветка:** `cursor-grok-4.5`  
**Дата:** 2026-07-27  
**Агент:** devops-engineer  
**Правила:** 1 commit / finding · no push master · no real secrets · templating/docs/config preferred

Источник: `.planning/audits/grok-2026-07/14-observability.md`

---

## Итог

| ID | Verified | Status | SHA | Notes |
|----|----------|--------|-----|-------|
| A14-001 | OPEN→FIXED | FIXED | `4aa77f5` | `/api/live` + `/api/ready`; `/health`=ready alias; Dockerfile liveness; Compose readiness |
| A14-002 | OPEN→FIXED | FIXED* | `7c03388` | *Repo fixed; **MANUAL** rotate S3 if `hub`/`hubsecret` ever deployed |
| A14-003 | OPEN→FIXED | FIXED* | `53f0436` | *Compose+redact; **MANUAL** set `NATS_USER`/`NATS_PASSWORD` on existing prod + recreate |
| A14-004 | OPEN→FIXED | FIXED | `14fc94f` | Monitoring ports → `127.0.0.1` only |
| A14-005 | OPEN→FIXED | FIXED | `cb40522` | Compose default CIDRs + `config.validate` requires non-empty in production |
| A14-009 | OPEN→FIXED | FIXED* | `68cbd14` | *OTEL_* wired; Tempo `--profile otel`; **MANUAL** set endpoint to enable spans |

\* = config/code PASS, live prod enablement remains ops.

---

## Detail

### A14-001 — health ≠ liveness

- Package `backend/internal/platform/health`: `Live` always 200; `Ready` pings PG/Redis/Mongo/NATS → 200/503.
- Compose `healthcheck` → `/api/ready`; Dockerfile `HEALTHCHECK` → `/api/live`.
- Tests: `testframework/platform/health/probes_test.go`.

### A14-002 — SeaweedFS secrets

- Deleted committed `seaweedfs-s3.json` (`hub`/`hubsecret`); example with placeholders.
- Prod generates `/tmp/s3.json` from `S3_*` env (LiveKit-style heredoc).
- `.env.example`: `S3_ACCESS_KEY=CHANGE_ME_NON_DEFAULT_KEY`.
- **MANUAL:** if prod ever ran with git defaults — rotate S3 keys (PG-03-style).

### A14-003 — NATS auth

- Prod NATS: `--user` / `--pass` from env; `NATS_URL=nats://USER:PASS@nats:4222`.
- Connect log redacts userinfo.
- **MANUAL:** existing deployments without auth must set creds and `--force-recreate nats backend`.

### A14-004 — monitoring ports

- Prometheus/Alertmanager/Loki/Grafana bind `127.0.0.1:*` (SSH tunnel / VPN).

### A14-005 — open `/metrics`

- Compose default: `172.16.0.0/12,10.0.0.0/8,192.168.0.0/16`.
- `ENVIRONMENT=production` + empty CIDRs → `config.Load` error.
- Optional: `kubectl apply -f deployments/k3s/monitoring/networkpolicy-metrics.yaml`.

### A14-009 — OTEL in prod

- Backend env: `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_SERVICE_NAME`.
- Optional Tempo: `deployments/monitoring/tempo/tempo.yaml` + compose `--profile otel`.
- Without endpoint — no-op (by design). **MANUAL** to turn tracing on in a given env.

---

## Tests / validation

```
go test -tags hubtest ./testframework/platform/health/     # A14-001 PASS
go test -tags hubtest ./testframework/platform/nats/ -run Redact  # A14-003 PASS
go test -tags hubtest ./testframework/platform/config/     # A14-005 PASS
docker compose -f deployments/docker-compose.prod.yml \
  -f deployments/docker-compose.monitoring.yml \
  --env-file deployments/.env.example config --quiet   # OK
```

---

## Out of scope (same audit zone, not Block E list)

A14-006…A14-008, A14-010…A14-019 — not in this wave task.
