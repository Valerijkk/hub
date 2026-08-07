# TASK-227 — 🐍 Security/platform component parity

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: Открыта · Фаза: **133**

## [OUTCOME]

Все API/DB/broker/source-contract rows auth, rbac, audit, ws, calls, platform,
apigateway, dev, ent и status имеют Python replacements в корневом
`testframework/{api,integration,contract}/**`; ни один не попал в
`backend/tests/**`.

## [ROLE]

- Owner: `qa-auto1`.
- Verifier: `qa-auto-ver` + `security-auditor` для negative paths.
- Handoff: `be1`/`be2` дают ledger/assertion contract TASK-210/211.
- Depends on: TASK-193, TASK-205, TASK-210, TASK-211.

## [ACCEPTANCE]

- Ledger scope: unmapped=0, component-under-backend=0, parity-required=0.
- Required services fail closed; no live providers/secrets; data namespaced.
- Exact behavior/negative-path/skip parity и deliberate sentinel RED→GREEN.
- Legacy остаётся до Phase 141; Go coverage ≥82% измеряется отдельно.
- Atomic Python-domain commits; `pro-qa-ai/**` вне scope; no push.
