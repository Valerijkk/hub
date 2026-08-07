# TASK-230 — 🐍 Frontend static/source contracts → Python

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: Открыта · Фаза: **131**

## [OUTCOME]

Legacy Vitest source/docs/compile guards не переезжают как фиктивные UI units:
они получают path-independent replacement в
`testframework/contract/frontend/**` или verifier-approved static CI gate.

## [ROLE]

- Owner: `qa-auto2`; verifier: `qa-auto-ver` + `fe-ver`.
- Handoff: `fe1`/`fe2` TASK-207/208 classification.
- Depends on: TASK-193, TASK-205, TASK-207, TASK-208.

## [ACCEPTANCE]

- Все fs/subprocess/hard-coded-root Vitest rows классифицированы; unmapped=0.
- `frontend/tests/**` содержит только Vitest/RTL behavior units.
- Replacement fail-closed, path-independent, с deliberate violation RED→GREEN.
- Legacy guard удаляется только Phase 141 после parity; no nested `domains/` in
  frontend target; `pro-qa-ai/**` вне scope; no push.
