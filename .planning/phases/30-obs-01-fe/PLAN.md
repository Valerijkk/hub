# Phase 30 — OBS-01 Wave 2 (Frontend)

**Goal:** Admin UI for in-memory operation metrics from Phase 29.

## Wave 1 — API + component

- [x] `OperationMetric` type + `getOperationMetrics()` in `frontend/src/api/admin.ts`
- [x] `OperationMetricsPanel` table component
- [x] Tab «Метрики» on `/admin/dashboard`

## Wave 2 — Tests + backend instrumentation

- [x] Vitest: `operation-metrics.test.tsx`, extend `admin.test.ts`
- [x] E2E: `admin-metrics.spec.ts` — metrics tab smoke
- [x] Backend: `Timed()` on `project.CreateProject`, `tms.CreateTestRun`

## Verification

```bash
cd backend && go test ./internal/project/... ./internal/tms/... -count=1
cd frontend && npx vitest run src/components/admin/operation-metrics.test.tsx src/api/admin.test.ts
```

## Success criteria

- Admin sees operation metrics table with count/errors/avg/max/last seen
- Refresh reloads from `GET /api/admin/observability/ops`
- Vitest green; E2E smoke when backend+DEV_MODE available

## Backlog (Phase 31+)

- OTel/Prometheus export
- Auto-refresh interval
- Slow-operation highlight threshold
