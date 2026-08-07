# TASK-229 — 🐍 Business/integration component parity

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: Открыта · Фаза: **135**

## [OUTCOME]

DB/API/provider/storage component rows business-доменов и integrations имеют
Python replacements; в `backend/tests/**` остаются только genuine units.

## [ROLE]

- Owner: `qa-auto1`; verifier: `qa-auto-ver`.
- Handoff: `be1`/`be2` TASK-214/215.
- Depends on: TASK-193, TASK-205, TASK-214, TASK-215.

## [ACCEPTANCE]

- Calendar/intranet/accounting/crm/hr/bases/dashboard/ai/search/storage/backup/
  imex/github/gitlab/onec/marketplace/mail scope полностью reconciled.
- Engine/provider invariants сохранены локальными doubles или task-owned infra;
  live external providers и secrets запрещены.
- Ledger unmapped=0, component-under-backend=0, parity-required=0; sentinel green.
- Legacy retained до Phase 141; Go coverage separate; atomic commits; no push.
