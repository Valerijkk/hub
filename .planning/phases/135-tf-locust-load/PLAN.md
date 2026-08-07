# PLAN: Phase 135 — backend business и integrations waves

**Статус:** PLANNED · **Оценка:** 2–3 дня

Канон: [`06-TEST-ARCHITECTURE-REBASE.md`](../../active/august-2026/06-TEST-ARCHITECTURE-REBASE.md).

## Цель (goal)

Завершить domain-by-domain disposition backend inventory: genuine units
получают `backend/tests/**`, а business/integration components — выполненный
Python replacement в корневом framework.

## Depends on

- Phase 134 PASS.

## Success criteria

- [ ] Calendar/intranet/accounting/crm/hr/bases/dashboard scope принят.
- [ ] AI/search/storage/backup/imex/github/gitlab/onec/marketplace/mail scope принят.
- [ ] Backend inventory `unmapped = 0`; component rows не находятся в
  `backend/tests/**` и имеют Python parity этой фазы.
- [ ] Legacy и native Go jobs green; общий Go coverage остаётся ≥82% по
  принятому TASK-203/204 contract; hooks ещё live.

## Задачи (атомарные)

| Task ID | Outcome | Owner | Verifier | Files | Depends | Estimate |
|---|---|---|---|---|---|---:|
| TASK-214 | Business wave: calendar/intranet/accounting/crm/hr/bases/dashboard | be1 | be-ver | `backend/tests/{calendar,intranet,accounting,crm,hr,bases,dashboard}/**`, `.planning/audits/test-rebase/inventory.json`, `.planning/evidence/test-rebase/135/business/**` | Phase 134 | 1–1.5д |
| TASK-215 | AI/search/storage/backup/imex/github/gitlab/onec/marketplace/mail wave | be2 | be-ver | `backend/tests/{ai,search,storage,backup,imex,github,gitlab,integrations/onec,marketplace,mail}/**`, `.planning/audits/test-rebase/hooks.json`, `.planning/evidence/test-rebase/135/integrations/**` | Phase 134 | 1–1.5д |
| TASK-229 | Business/integrations API/integration/contract component parity в Python | qa-auto1 | qa-auto-ver | `testframework/{api,integration,contract}/{calendar,intranet,accounting,crm,hr,bases,dashboard,ai,search,storage,backup,imex,github,gitlab,onec,marketplace,mail}/**` | TASK-193/205/214/215 | 1–2д |

## Волны

- Wave 1: TASK-214 и TASK-215 unit batches.
- Wave 2: TASK-229 получает frozen classification/assertion handoff.
- Wave 3: полный backend legacy/native/Python inventory и coverage reconciliation.

## Риски

| Риск | Mitigation |
|---|---|
| Provider tests зависят от сети/secrets | Deterministic local doubles для unit; real boundary → Phase 136 |
| Business DB semantics потеряны | PostgreSQL-sensitive rows → TASK-229 Python integration с same invariant |
| Mechanical volume >3 days | Никакого false-complete: stop, evidence и replan до Phase 140 |

## Verification

- `be-ver`: все затронутые Go packages, race, coverage, consumer ledger.
- `qa-auto-ver`: полный backend `unmapped = 0`; каждый non-unit row имеет
  выполненный Python replacement либо явный BLOCKED evidence.
