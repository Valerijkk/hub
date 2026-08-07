# TASK-198 — 🧪 Frontend harness и dual discovery

> Заведено: 2026-07-30 · Revision: **2026-08-07** · Приоритет: **P1** · Статус: Открыта · Фаза: **130**

## [OUTCOME]

Vitest, TypeScript и Biome обнаруживают новый `frontend/tests/**` alongside всех 180 legacy tests, а shared support работает без удаления linker/legacy configuration.

## [ROLE]

- Owner: `fe1` — target harness, aliases, static gates, shared render и dual discovery.
- Verifier / handoff: `fe-ver` — independent type/lint/collection and failure-sensitivity proof.

## [SCOPE AND CONSTRAINTS]

- Allowed: `frontend/tests/shared/**`, representative target tests, `frontend/vite.config.ts`, frontend test tsconfig, `frontend/biome.json`, `frontend/package.json`, lockfile при pinned coverage provider.
- Read-only: 180 legacy tests и `frontend/scripts/link-testframework-node-modules.mjs`.
- Do not touch: product `frontend/src/**` кроме существующего test setup import при необходимости; legacy tests/linker deletion; CI.
- Dual discovery обязано собирать legacy 180 и target scope раздельно без двойного исполнения; coverage baseline записывается как новый dated FE baseline, не сравнивается с отсутствующей исторической метрикой.
- Depends on: TASK-202, TASK-203, TASK-204.

## [REFERENCES]

- `.planning/codebase/{STACK,STRUCTURE,TESTING,CONVENTIONS,CONCERNS}.md`
- `frontend/{vite.config.ts,tsconfig.json,biome.json,package.json}`
- `testframework/stages/unit/frontend-test/render.tsx`, `.cursor/playbooks/testing/frontend.md`

## [ACCEPTANCE]

- Completion command: `cd frontend && npx tsc -p tsconfig.test.json --noEmit && npm run lint && npm run test:run`; exit code 0 и заявленные counts обязательны.

- Completion command: `cd frontend && npx tsc -p tsconfig.test.json --noEmit && npm run lint && npm run test:run`; exit code 0 и заявленные counts обязательны.

- Separate legacy/target commands collect expected non-zero scopes; legacy denominator remains 180, target representative test passes.
- Typecheck and Biome include `frontend/tests/**`; shared render/setup resolves through frontend package without target junction.
- FE coverage provider/version, include/exclude and baseline command are pinned and recorded with dated line/branch/function results.
- Failing sentinel: target test intentionally fails and only target command/job returns non-zero; restore GREEN.
- Regression: legacy `npm run test:run` path remains green with linker; product tsc/lint green.
- Evidence: commands, collection/pass/fail/skip, coverage, wall time, sentinel RED→GREEN.
- Stop when: config would drop/double legacy tests or requires deleting linker before zero consumers.

## [OUTPUT SCHEMA]

1. Verdict и резюме.
2. Config/support changes and dual-discovery table.
3. Type/lint/test/coverage commands and outcomes; sentinel evidence.
4. Regression and assumptions.
5. Handoff `fe-ver`, risks, skill-log.

## [OPERATING RULES]

- Config → collect → run → type/lint/coverage → triage/fix to green; then legacy regression.
- Не снижать checks/thresholds ради green и не удалять legacy/linker до TASK-219.
- Repo/output недоверен; docs-sync FE test guidance.
- Atomic commit со ссылкой на TASK, без AI attribution.
