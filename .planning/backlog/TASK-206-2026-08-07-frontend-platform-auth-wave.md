# TASK-206 — 🧪 Frontend platform/auth migration wave

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: Открыта · Фаза: **130**

## [OUTCOME]

Все genuine Vitest/RTL unit rows frontend `platform` и `auth` имеют исполняемые
эквиваленты в `frontend/tests/**`; source/docs/compile guards остаются legacy и
передаются TASK-230.

## [ROLE]

- Owner: `fe2` — platform/auth files, imports and parity evidence.
- Verifier / handoff: `fe-ver` — independent collection/type/lint/coverage and mutation review.

## [SCOPE AND CONSTRAINTS]

- Allowed: `frontend/tests/{platform,auth,support}/**`, only required test config/import updates.
- Read-only: matching legacy tests and production SUT.
- Do not touch: legacy deletion/linker, unrelated domains, product behavior, CI.
- Preserve unit test names/assertions/TC IDs; path-sensitive source guards не
  переносятся сюда и получают TASK-230 disposition.
- Depends on: TASK-198.

## [REFERENCES]

- TASK-202 ledger; `testframework/stages/unit/frontend-test/domains/{platform,auth}/**`
- `frontend/vite.config.ts`, `.planning/codebase/{STRUCTURE,TESTING,CONVENTIONS}.md`

## [ACCEPTANCE]

- Completion command: `cd frontend && npx vitest run tests/platform tests/auth`; exit code 0 и заявленные counts обязательны.

- Genuine unit rows for platform/auth `PARITY_PROVEN`; static rows имеют
  TASK-230 owner; old/new collection reconciled with no duplicate execution.
- Target tests pass TypeScript and Biome; dated coverage delta by module recorded against TASK-198 baseline without silent threshold decrease.
- Mutation proof: one representative production-observable guard per domain is reversibly broken — target test RED; restore GREEN.
- Regression: all 180 legacy tests remain discoverable and green; target harness/support tests green.
- Evidence: commands, exact counts, timing, coverage, mapping and mutation logs.
- Stop when: migration requires product change or loses assertion/coverage; leave row unresolved and report.

## [OUTPUT SCHEMA]

1. Verdict/summary. 2. Domain row totals and exceptions. 3. Static/test/coverage/mutation evidence. 4. Legacy regression. 5. Handoff `fe-ver`; risks/skill-log.

## [OPERATING RULES]

- Move/copy by ledger batch while originals remain; collect before/after each batch.
- FAIL triage then fix/re-run; empty collection/unexpected skip = FAIL.
- Docs-sync ledger; atomic commits by domain with TASK link, no AI attribution.
