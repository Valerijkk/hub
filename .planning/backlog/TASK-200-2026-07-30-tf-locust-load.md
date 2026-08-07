# TASK-200 — 🧪 Locust local baseline

> Заведено: 2026-07-30 · Revision: **2026-08-07** · Приоритет: **P2** · Статус: Открыта · Фаза: **139**

## [OUTCOME]

`testframework/load/**` безопасно запускает параметризованный Locust baseline только против allowlisted non-production target и фиксирует воспроизводимые p50/p95/error-rate/throughput результаты.

## [ROLE]

- Owner: `qa-auto2` — load scenarios, target guard, profiles и baseline report.
- Verifier / handoff: `qa-auto-ver` — safety rejection и повторяемость измерения.

## [SCOPE AND CONSTRAINTS]

- Allowed: `testframework/load/**`, Python load config/docs/reports without secrets.
- Do not touch: product code, CI jobs, production config, legacy tests.
- Default target — localhost/explicit CI hostname. Production-like URL, public IP/domain или отсутствующий environment marker must fail before load. Никаких production runs.
- Scenarios: auth, feed, search, issue create и project board; destructive data uniquely namespaced and cleaned.
- Depends on: TASK-193, TASK-216.

## [REFERENCES]

- `.planning/codebase/{CONCERNS,INTEGRATIONS,TESTING}.md`
- `docs/security_guidelines.md`, `docs/testframework_strategy.md`
- Performance findings A13-001…006 referenced in planning audit.

## [ACCEPTANCE]

- Completion command: `cd testframework && locust --headless -f load/locustfile.py --host http://localhost:8080 --users 1 --spawn-rate 1 --run-time 30s`; exit code 0 и заявленные counts обязательны.

- Completion command: `cd testframework && locust --headless -f load/locustfile.py --host http://localhost:8080 --users 1 --spawn-rate 1 --run-time 30s`; exit code 0 и заявленные counts обязательны.

- Headless command accepts host/users/spawn-rate/duration/profile; local smoke completes and emits CSV/summary with p50/p95, throughput, error rate and scenario mix.
- Safety sentinel: production-like host is rejected before first request; evidence includes zero requests.
- Repeatability replaces mutation: two runs on same isolated profile/environment stay within documented tolerance (default ≤20% for p95/throughput unless measured noise justifies tighter/looser approved bound).
- Regression: Python smoke/API/integration green after load; created data cleaned; no CI PR job added.
- Evidence: target fingerprint (no secrets), resources, profile, two result sets, delta, wall time, safety sentinel.
- Stop when: target cannot be proven non-production or environment is shared with users; do not run.

## [OUTPUT SCHEMA]

1. PASS/BLOCKED and summary.
2. Scenario/profile/guard table.
3. Two-run metrics and delta; safety sentinel proof.
4. Cleanup/regression and limitations.
5. Handoff `qa-auto-ver`, risks, skill-log.

## [OPERATING RULES]

- Preflight safety before any request; ramp gradually and stop on error/latency guard.
- No secrets/PII in reports; metrics are environment-specific, not production capacity claims.
- Test triage on functional errors; docs-sync load README/report.
- Atomic commit with TASK link, no AI attribution.
