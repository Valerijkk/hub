# PLAN: Phase 136 — Python API и integration

**Статус:** PLANNED · **Оценка:** 2–3 дня

Канон: [`06-TEST-ARCHITECTURE-REBASE.md`](../../active/august-2026/06-TEST-ARCHITECTURE-REBASE.md).

## Цель (goal)

Расширить Python scaffold до API/integration contract, заменить все принятые
backend black-box candidates достаточным parity evidence и не выдавать Python
execution за Go coverage.

## Depends on

- Phase 129 PASS (Python scaffold/API slice).
- Phase 135 PASS (закрытый backend disposition ledger).

## Success criteria

- [ ] Core clients/settings/redaction/data isolation поддерживают API и services.
- [ ] PostgreSQL, MongoDB, Redis и NATS profiles имеют fail-closed health checks.
- [ ] Каждая `python-api|integration|contract|full` backend candidate row имеет
  replacement, parity result и expected skip policy; `PARITY_REQUIRED = 0` в scope.
- [ ] Go coverage публикуется отдельно и остаётся в принятом contract.

## Задачи (атомарные)

| Task ID | Outcome | Owner | Verifier | Files | Depends | Estimate |
|---|---|---|---|---|---|---:|
| TASK-199 | Core/API pilot: typed client, factories, schema/assertion helpers, isolation и representative API parity | qa-auto1 | qa-auto-ver | `testframework/python/{core,fixtures,api}/**`, `.planning/evidence/test-rebase/136/api/**` | Phase 129, Phase 135 | 1–1.5д |
| TASK-216 | Infra integration expansion для PostgreSQL/Mongo/Redis/NATS и всех ledger candidates | qa-auto2 | qa-auto-ver | `testframework/python/{integration,full}/**`, `.planning/audits/test-rebase/inventory.json`, `.planning/evidence/test-rebase/136/integration/**` | TASK-199 | 1–1.5д |

## Волны

- Wave 1: TASK-199.
- Wave 2: TASK-216 по service-isolated markers; затем общий parity reconciliation.

## Риски

| Риск | Mitigation |
|---|---|
| Shared DB collision под xdist | Run-id namespace, worker-owned data, deterministic cleanup, serial marker allowlist |
| Недоступный service превращается в skip | Required profiles fail setup; optional skips только allowlisted |
| Python replacement снижает Go coverage | Retain/add Go units или founder-approved gate; метрики не смешивать |

## Verification

- `qa-auto-ver`: два последовательных и parallel-safe runs, exact counts,
  unexpected skip=0, sentinel red/green, redaction scan.
- `be-ver`: Go coverage comparison и подтверждение, что Python не включён в profile.
