# PLAN: Phase 139 — Locust local harness и baseline

**Статус:** PLANNED · **Оценка:** 1–2 дня

Канон: [`06-TEST-ARCHITECTURE-REBASE.md`](../../active/august-2026/06-TEST-ARCHITECTURE-REBASE.md).

## Цель (goal)

Создать безопасный Locust harness для локального/scheduled non-production run и
получить первый воспроизводимый hot-path baseline, не превращая его в PR gate.

## Depends on

- Phase 129 PASS.
- Phase 136 PASS для shared API/auth clients и environment profiles.

## Success criteria

- [ ] Harness hard-fail отклоняет production-like URL и требует test marker.
- [ ] Auth, feed, search, issue/board/sprint/bases/dashboard flows используют
  отдельный user/data pool и deterministic cleanup.
- [ ] Smoke baseline содержит users/ramp/duration, p50/p95, error rate,
  throughput и наблюдаемую точку деградации либо честный NOT_RUN blocker.
- [ ] Секреты/PII не попадают в CSV, logs и reports.

## Задачи (атомарные)

| Task ID | Outcome | Owner | Verifier | Files | Depends | Estimate |
|---|---|---|---|---|---|---:|
| TASK-200 | Locust config/tasks, target allowlist, seeded user pool, smoke/soak profiles и dated local baseline | qa-auto2 | qa-auto-ver | `testframework/load/locustfile.py`, `testframework/load/tasks/**`, `testframework/load/data/users.example.csv`, `testframework/load/README.md`, `.planning/evidence/test-rebase/139/**` | Phase 136 | 1–2д |

## Волны

- Wave 1: safety gate + auth/user pool.
- Wave 2: hot paths + smoke baseline.

## Риски

| Риск | Mitigation |
|---|---|
| Случайная нагрузка на production | Host allowlist + non-production marker + fail before user spawn |
| Shared data искажает baseline | Dedicated project/users, warm-up, fixed dataset and cleanup |
| Docker unavailable | Record NOT_RUN; Phase 139 не PASS без approved executable run |

## Verification

- `qa-auto-ver`: production URL rejection, non-zero requests, known failure
  sensitivity, metrics artifact, redaction scan и повторяемость smoke.
