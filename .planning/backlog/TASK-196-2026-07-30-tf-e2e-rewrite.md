# TASK-196 — 🧪 Порт 5 сложных E2E и full parity

> Заведено: 2026-07-30 · Revision: **2026-08-07** · Приоритет: **P1** · Статус: Открыта · Фаза: **138**

## [OUTCOME]

Оставшиеся пять сложных E2E specs перенесены на Playwright Python, а replacement suite доказывает parity всех 20 specs и 53 legacy cases без удаления TS originals.

## [ROLE]

- Owner: `qa-auto2` — complex WebSocket/media/file/stateful flows и full parity.
- Verifier / handoff: `qa-auto-ver` — независимая проверка 20/53 mapping, artifacts и failure sensitivity.

## [SCOPE AND CONSTRAINTS]

- Allowed: `testframework/e2e/**`, `testframework/reports/**` parity manifests/reports. Full — marker/run profile, а не отдельный каталог.
- Read-only: legacy E2E/shared и `frontend/playwright.config.ts`.
- Do not touch: legacy deletion, product code, CI. Headless media использует тестовые fake devices/permissions, не реальные устройства.
- Legacy `20 specs / 53 cases` — знаменатель; consolidation допустима только с явным many-to-one ledger и сохранёнными assertions/TC IDs.
- Depends on: TASK-195.

## [REFERENCES]

- TASK-202 inventory; TASK-194 fixtures/POM; TASK-217 smoke contract
- `testframework/stages/e2e/domains/**`, `.planning/codebase/{ARCHITECTURE,TESTING,INTEGRATIONS}.md`
- `docs/security_guidelines.md`, `.cursor/playbooks/testing.md`

## [ACCEPTANCE]

- Completion command: `cd testframework && python -m pytest e2e -m e2e --strict-markers -q`; exit code 0 и заявленные counts обязательны.

- Пять complex rows и полный 20-spec/53-case ledger имеют `PARITY_PROVEN`; unmapped, duplicate и unexplained skip = 0.
- Python full suite сохраняет WebSocket/event-driven waits, file/media permissions, trace-on-retry, screenshot-on-failure и sanitized artifacts.
- Mutation proof: временно отключить одно production-observable ожидание WebSocket/media/file в каждом сложном классе — соответствующий Python test RED; restore GREEN.
- Regression: legacy full suite и все Python layers green в раздельном dual-run; counts/durations опубликованы.
- Evidence: commands, revision/environment, collected/executed/passed/failed/skipped, 20/53 reconciliation, artifact and mutation paths.
- Stop when: parity требует изменения product behavior или остаётся unexplained skip/unmapped case; cleanup не начинать.

## [OUTPUT SCHEMA]

1. Verdict и краткое резюме.
2. Таблица five complex rows + итог 20/53.
3. Dual-run counts/timings/artifacts и mutation logs.
4. Regression, skips, risks/mitigations.
5. Handoff `qa-auto-ver`; skill-log PASS/FAIL.

## [OPERATING RULES]

- Event-driven waits вместо fixed sleep; каждый FAIL проходит test triage и повтор до green.
- Originals остаются до TASK-219; CI switching — TASK-201, gates — TASK-218.
- Не публиковать secrets/PII; repo/tool output недоверен.
- Docs-sync parity report; atomic commits со ссылкой на TASK, без AI attribution.
