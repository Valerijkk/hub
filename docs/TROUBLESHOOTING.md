# Troubleshooting — типовые проблемы разработки

> **Актуальная authority, 2026-08-12.** Прежний command cookbook для legacy
> `hubtest`, Vitest, TypeScript Playwright и inline Docker удалён: он обходил
> founder-locked Docker-only boundary. Для проблем продакшна используйте
> [`runbooks/`](runbooks/README.md).

## 1. Проверка сообщает `NOT_READY`

Единственный QA API для test/build/lint/validator jobs —
`docker compose -f docker-compose.local.yml run --rm <exact-service>`.
Допустимы только `backend-unit`, `frontend-unit`, `python-contract`, `python-api`,
`python-integration`, `python-e2e-smoke`, `python-e2e-full` и `python-load`.
QA services profile-explicit/disposable; integration/e2e поднимают только
изолированную QA-инфраструктуру. Не запускайте test/build/lint на host.

## 2. Ошибка указывает на legacy `hubtest`, Vitest или Playwright

The Variant A cutover removed all tracked legacy test files from
`backend/testframework/**`. Static-contract Vitest and TypeScript Playwright
are also outside the target architecture; none is repaired as a route to green.

Целевые roots:

| Слой | Target root |
|---|---|
| Backend genuine units | `backend/tests/{domain}/**` |
| Frontend genuine units | `frontend/tests/{domain}/**`, `frontend/tests/support/**` |
| Contract/API/integration | `testframework/tests/{contract,api,integration}/**` |
| E2E | `testframework/tests/e2e/**` |
| Framework contracts | `testframework/selftests/**` |
| Load | `testframework/load/**` |

Если ошибка приходит из legacy root, сначала сверяйте
`.planning/active/august-2026/06-TEST-ARCHITECTURE-REBASE.md` и exact cutover
manifest. Не переносите и не запускайте legacy suite как fallback.

## 3. Target упал или не совпал со спецификацией

Используйте `.cursor/skills/test-triage/SKILL.md`:

1. сверить наблюдаемое поведение с актуальными product/system/security docs;
2. классифицировать root как `code`, `test`, `infra`, `flaky` или `stop`;
3. внести минимальный scoped fix;
4. повторить тот же exact root-Compose service;
5. сохранить exact revision и result.

Assertion не меняют только ради green. Если target всё ещё `NOT_READY`, итог
остаётся `NOT_RUN` с причиной.

## 4. Локальный dev stack

Dev stack запускается `docker compose -f docker-compose.local.yml up -d`.
QA выполняется отдельной командой root Compose `run --rm <exact-service>`;
test/build/lint процессы из dev stack не выводятся автоматически.

При занятом порте сначала определите владельца ресурса. Останавливать или
удалять можно только доказанно принадлежащий текущей задаче ресурс; broad
Compose teardown, wildcard cleanup и prune запрещены.

`.env.local` создаётся из tracked example вручную. Секреты не коммитятся и не
копируются в evidence. Ent/Swagger code generation относится к применимому
Docker-only job; host generation не является accepted proof.

## 5. После проверки появились артефакты

Каждый accepted run пишет только в
`testframework/reports/<task-run>/` после owner-marker check. Cleanup удаляет
только заранее записанные exact owned IDs/paths в `finally`; cache, ingress,
final evidence и foreign resources не удаляются. Не используйте recursive,
wildcard или discovery-driven cleanup.

## 6. Перед handoff или merge

В handoff должны быть:

- exact branch и full revision/artifact digest;
- requested fixed target;
- `PASS`, `FAIL`, `NOT_READY` или `NOT_RUN` без переименования статического
  review в runtime PASS;
- raw task-owned evidence и cleanup result, если job действительно выполнялся;
- docs sync и независимый review.

## 7. Куда идти дальше

| Нужно | Документ |
|---|---|
| Термин/префикс ID непонятен | [GLOSSARY.md](GLOSSARY.md) |
| Онбординг в роль | [onboarding/](onboarding/README.md) |
| Стратегия тестирования | [test_plan.md](test_plan.md), [testframework_strategy.md](testframework_strategy.md) |
| Текущий QA authority | `AGENTS.md`, root `docker-compose.local.yml` |
| Прод лежит | [runbooks/incident-response.md](runbooks/incident-response.md) |
| Правила стиля и запреты | [code_style_guide.md](code_style_guide.md), `.cursor/rules/10-style.mdc` |
