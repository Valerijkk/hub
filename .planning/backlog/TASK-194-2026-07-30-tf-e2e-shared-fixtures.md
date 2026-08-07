# TASK-194 — 🧪 Shared Python E2E fixtures и POM

> Заведено: 2026-07-30 · Revision: **2026-08-07** · Приоритет: **P1** · Статус: Открыта · Фаза: **137**

## [OUTCOME]

Python E2E имеет единый безопасный слой auth/data/browser fixtures и Page Objects, пригодный для портирования всех 20 legacy specs без копирования setup и shared state.

## [ROLE]

- Owner: `qa-auto1` — fixture/POM contracts, isolation, cleanup и redaction.
- Verifier / handoff: `qa-auto-ver` — два последовательных прогона, parallel proof и negative checks.

## [SCOPE AND CONSTRAINTS]

- Allowed: `testframework/python/e2e/conftest.py`, `testframework/python/e2e/pages/**`, `testframework/python/e2e/steps/**`, `testframework/python/fixtures/**`, fixture contract tests и профиль test env.
- Read-only source: `testframework/stages/e2e/shared/**` и 20 specs из `testframework/stages/e2e/domains/**`.
- Do not touch: legacy TS E2E, `frontend/playwright.config.ts`, product code, CI, production credentials.
- Auth — через разрешённый test/dev interceptor Magic Link или существующий dev-session contract; реальные mailbox/secrets запрещены. Mutable data function-scoped, namespace/run-id worker-safe, cleanup deterministic.
- Depends on: TASK-193, TASK-205, TASK-216.

## [REFERENCES]

- `testframework/stages/e2e/shared/{fixtures.ts,helpers,pages,steps}/`
- `.planning/codebase/{ARCHITECTURE,STRUCTURE,TESTING,INTEGRATIONS}.md`
- `docs/security_guidelines.md`, `.cursor/playbooks/testing.md`

## [ACCEPTANCE]

- Completion command: `cd testframework/python && python -m pytest e2e -m smoke --strict-markers -q`; exit code 0 и заявленные counts обязательны.

- Реализованы browser context, auth, user/project/channel factories, API client, cleanup и redaction; один contract test использует fixture через публичный Hub contract.
- Два последовательных прогона дают одинаковые pass/fail/skip counts и не оставляют конфликтующих записей; parallel smoke (`-n 2` либо документированный эквивалент) не сталкивается по данным.
- Mutation proof: временно отключить уникальный namespace или cleanup — isolation contract краснеет; восстановить — green. Не мутировать production code для доказательства fixture-инварианта.
- Regression: Python API/integration pilot green; legacy 3-spec smoke запускается прежней командой и не меняется.
- Evidence: команды, counts, wall time, sanitized artifacts, mutation RED→GREEN, список fixture scopes.
- Stop when: auth требует production secret/mailbox либо cleanup нельзя доказать без destructive broad delete; эскалировать точный blocker.

## [OUTPUT SCHEMA]

1. PASS/BLOCKED и резюме.
2. Таблица fixture → scope → creates → cleanup → parallel policy.
3. Команды/counts/timing и mutation evidence.
4. Regression legacy/Python, security/redaction checks.
5. Риски и handoff `qa-auto-ver`; skill-log PASS/FAIL.

## [OPERATING RULES]

- Implement → два run → parallel run → triage/fix до green; затем legacy regression.
- Не копировать secrets/PII в logs, traces, screenshots или Allure; skips только зарегистрированные и allowlisted.
- Не менять/удалять legacy до TASK-219; неожиданный skip считается FAIL.
- Docs-sync в testframework docs; атомарный коммит со ссылкой на TASK, без AI attribution.
