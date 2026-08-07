# PLAN: Phase 137 — Python E2E foundation и smoke parity

**Статус:** PLANNED · **Оценка:** 2–3 дня

Канон: [`06-TEST-ARCHITECTURE-REBASE.md`](../../active/august-2026/06-TEST-ARCHITECTURE-REBASE.md).

## Цель (goal)

Создать shared Python Playwright fixtures/POM/steps и доказать parity
трёх-spec smoke contract на том же immutable application build, что TS legacy.

## Depends on

- Phase 129 PASS (Python scaffold и authenticated API slice).
- Phase 136 не блокирует shared UI foundation, но его core fixtures переиспользуются
  при наличии без дублирования.

## Success criteria

- [ ] Auth/data/browser fixtures изолируют пользователей и чистят mutable data.
- [ ] POM/steps не содержат bare sleeps; traces/screenshots redacted.
- [ ] Python smoke соответствует legacy login + a11y + critical-gate: 3 specs,
  те же обязательные assertions, `unexpected skip = 0`.
- [ ] TS smoke остаётся green и исполняемым.

## Задачи (атомарные)

| Task ID | Outcome | Owner | Verifier | Files | Depends | Estimate |
|---|---|---|---|---|---|---:|
| TASK-194 | Shared fixtures, auth/data factories, POM, steps, a11y helper и artifact policy | qa-auto1 | qa-auto-ver | `testframework/python/e2e/{conftest.py,pages/**,steps/**}`, `testframework/python/fixtures/**` | Phase 129 | 1–1.5д |
| TASK-217 | Login/a11y/critical-gate Python smoke с same-build TS parity и sentinel evidence | qa-auto2 | qa-auto-ver | `testframework/python/e2e/{auth,platform}/**`, `.planning/audits/test-rebase/inventory.json`, `.planning/evidence/test-rebase/137/**` | TASK-194 | 1д |

## Волны

- Wave 1: TASK-194.
- Wave 2: TASK-217 и TS/Python parity comparison.

## Риски

| Риск | Mitigation |
|---|---|
| SessionStorage/dev auth shape drift | Contract fixture из live response; никаких copied secrets |
| A11y library differs от TS | Одинаковый impact scope и accepted violation mapping |
| Green через skip | Fail-closed preflight + exact collected/executed/skip gate |

## Verification

- `qa-auto-ver`: два одинаковых smoke runs, TS/Python assertion map,
  3-spec collection, unexpected skip=0, deliberate failure red.
- `security-auditor`: trace/screenshot/log redaction.
