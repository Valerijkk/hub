# PLAN: Phase 135 — backend business и integrations waves

**Статус:** PLANNED · **Оценка:** 2–3 дня

Канон: [`06-TEST-ARCHITECTURE-REBASE.md`](../../active/august-2026/06-TEST-ARCHITECTURE-REBASE.md).

## Цель (goal)

Завершить domain-by-domain disposition backend inventory: business domains и
оставшиеся AI/search/storage/integration packages получают Go-native target или
явный Python parity dependency.

## Depends on

- Phase 134 PASS.

## Success criteria

- [ ] Calendar/intranet/accounting/crm/hr/bases/dashboard scope принят.
- [ ] AI/search/storage/backup/imex/github/gitlab/onec/marketplace/mail scope принят.
- [ ] Backend inventory `unmapped = 0` и `PARITY_REQUIRED = 0` для Go-wave scope;
  Python candidates блокируют deletion до 136.
- [ ] Legacy и native Go jobs green; общий Go coverage остаётся ≥82% по
  принятому TASK-203/204 contract; hooks ещё live.

## Задачи (атомарные)

| Task ID | Outcome | Owner | Verifier | Files | Depends | Estimate |
|---|---|---|---|---|---|---:|
| TASK-214 | Business wave: calendar/intranet/accounting/crm/hr/bases/dashboard | be1 | be-ver | `backend/tests/{calendar,intranet,accounting,crm,hr,bases,dashboard}/**`, `.planning/audits/test-rebase/inventory.json`, `.planning/evidence/test-rebase/135/business/**` | Phase 134 | 1–1.5д |
| TASK-215 | AI/search/storage/backup/imex/github/gitlab/onec/marketplace/mail wave | be2 | be-ver | `backend/tests/{ai,search,storage,backup,imex,github,gitlab,integrations/onec,marketplace,mail}/**`, `.planning/audits/test-rebase/hooks.json`, `.planning/evidence/test-rebase/135/integrations/**` | Phase 134 | 1–1.5д |

## Волны

- Wave 1: TASK-214 и TASK-215 параллельно.
- Wave 2: полный backend legacy/native inventory и coverage reconciliation.

## Риски

| Риск | Mitigation |
|---|---|
| Provider tests зависят от сети/secrets | Deterministic local doubles для unit; real boundary → Phase 136 |
| Business DB semantics потеряны | PostgreSQL-sensitive rows → Python integration с same invariant |
| Mechanical volume >3 days | Никакого false-complete: stop, evidence и replan до Phase 140 |

## Verification

- `be-ver`: все затронутые Go packages, race, coverage, consumer ledger.
- `qa-auto-ver`: полный backend `unmapped = 0`; каждый non-Go row связан с 136.
