# TASK-228 — 🐍 Work/realtime component parity

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: Открыта · Фаза: **134**

## [OUTCOME]

Component/integration rows tracker, project, sprint, tms, messenger и
notification заменены Python API/integration/contract tests; Redis/NATS/Mongo/
PostgreSQL semantics не маскируются под Go units.

## [ROLE]

- Owner: `qa-auto2`; verifier: `qa-auto-ver`.
- Handoff: `be1`/`be2` TASK-212/213.
- Depends on: TASK-193, TASK-205, TASK-212, TASK-213.

## [ACCEPTANCE]

- Ledger scope: unmapped=0, component-under-backend=0, parity-required=0.
- Isolated DB/broker namespaces, parallel-safe cleanup, required health fail-closed.
- Same assertions/negative paths/skip budget; sentinel RED→GREEN.
- Legacy retained до Phase 141; Go coverage separate; no external network.
- Atomic domain commits; `pro-qa-ai/**` вне scope; no push.
