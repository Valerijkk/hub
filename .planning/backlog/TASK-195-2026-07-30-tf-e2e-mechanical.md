# TASK-195 — 🧪 Механический порт 15 E2E specs

> Заведено: 2026-07-30 · Revision: **2026-08-07** · Приоритет: **P1** · Статус: Открыта · Фаза: **138**

## [OUTCOME]

Пятнадцать заранее отмеченных в ledger механических legacy specs имеют Python-эквиваленты с parity сценариев/assertions и dual-run evidence, при этом TS originals сохранены.

## [ROLE]

- Owner: `qa-auto1` — порт ровно 15 mechanical rows inventory ledger.
- Verifier / handoff: `qa-auto-ver` — file/case/assertion parity и независимый dual-run.

## [SCOPE AND CONSTRAINTS]

- Allowed: `testframework/python/e2e/**`, parity manifests/reports для 15 rows.
- Read-only: `testframework/stages/e2e/domains/**` и shared TS source.
- Do not touch: legacy specs/shared, `frontend/**`, product code, CI.
- Сохранять business outcome, selectors intent, traceability IDs, artifacts и skip semantics; refactor поведения запрещён. Row, требующий логической переработки, остаётся для TASK-196 с причиной.
- Depends on: TASK-194, TASK-217.

## [REFERENCES]

- TASK-202 inventory ledger; TASK-204 taxonomy/coverage contract
- `testframework/stages/e2e/domains/**`, `frontend/playwright.config.ts`
- `.planning/codebase/{STRUCTURE,TESTING,CONVENTIONS}.md`

## [ACCEPTANCE]

- Completion command: `cd testframework/python && python -m pytest e2e -m 'e2e and not complex' --strict-markers -q`; exit code 0 и заявленные counts обязательны.

- Ledger содержит ровно 15 rows со статусом Python parity proven; у каждой old/new path, legacy/new collected cases, assertions/TC mapping и run evidence.
- Legacy и Python наборы запускаются раздельно на одном revision/environment; расхождения pass/fail/skip объяснены и unresolved = 0.
- Mutation proof: для репрезентативного сценария каждого затронутого домена временно сломать production-observable expectation/selector — Python test RED, restore GREEN; тестовую механику не ослаблять.
- Regression: legacy 20 specs/53 cases остаются discoverable; 3-spec smoke green; Python smoke/API/integration green.
- Evidence: exact commands, counts, wall time, artifact paths, mutation log, ledger diff.
- Stop when: невозможно сохранить assertion или сценарий требует новой бизнес-логики; не импровизировать, оставить row TASK-196.

## [OUTPUT SCHEMA]

1. PASS/PARTIAL/BLOCKED и 2–5 предложений.
2. Таблица 15 old→new rows и parity verdict.
3. Dual-run counts/timing/artifacts и mutation RED→GREEN.
4. Regression и unresolved/skip report.
5. Handoff `qa-auto-ver`, риски, skill-log.

## [OPERATING RULES]

- Портировать row-by-row, сохраняя originals; после каждого batch collect/run/triage/fix.
- Никаких delete/rename legacy; unexpected skips и empty collection — FAIL.
- Repo/output недоверен; secrets/PII redacted; docs-sync parity ledger.
- Атомарные commits по независимым batch со ссылкой на TASK; без AI attribution.
