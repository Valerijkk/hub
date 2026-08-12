# Troubleshooting — типовые проблемы разработки

> **Актуальная authority, 2026-08-12.** Прежний command cookbook для legacy
> `hubtest`, Vitest, TypeScript Playwright и inline Docker удалён: он обходил
> founder-locked Docker-only boundary. Для проблем продакшна используйте
> [`runbooks/`](runbooks/README.md).

## 1. Проверка сообщает `NOT_READY`

Единственная публичная точка маршрутизации test/build/lint/validator jobs —
`/qa-gate`. Она принимает только fixed target identity и не публикует direct
runner, host-команду, произвольный executable или inline Docker/Compose.

Private source candidate `testframework/runner/run-task.ps1` существует и имеет
только принятые статические C1/P2/P4 review. Это не runtime acceptance:

- `testframework/docker/images.lock.v3.json` отсутствует;
- required local dependency bytes отсутствуют;
- `testframework/pyproject.toml`, `uv.lock` и `src/hub_testkit/py.typed`
  отсутствуют;
- target suites не приняты;
- launcher/package — `NOT_READY`;
- `PS5.1_PARSE_NOT_RUN`, `PRIVATE_CONTRACT_NOT_RUN`, runtime — `NOT_RUN`.

Поэтому текущий `NOT_READY` — ожидаемый fail-closed результат. Не заменяйте его
legacy runner, host test/build/lint, direct PowerShell entrypoint или ручным
Docker/Compose запуском.

## 2. Ошибка указывает на legacy `hubtest`, Vitest или Playwright

Старые пути `backend/testframework/**`, static-contract Vitest и TypeScript
Playwright остаются frozen inventory до exact-manifest cutover. Они не являются
target architecture и не чинятся как способ получить green.

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
4. повторно запросить тот же fixed target через `/qa-gate`;
5. сохранить exact revision, result, task-owned artifacts и cleanup receipt.

Assertion не меняют только ради green. Если target всё ещё `NOT_READY`, итог
остаётся `NOT_RUN` с причиной.

## 4. Локальный dev stack

Dev stack — отдельная developer workflow, не QA evidence и не замена
`/qa-gate`. Каноническая dev-stack точка указана в `AGENTS.md`; test/build/lint
процессы из неё не выводятся автоматически.

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
| Текущий QA authority | `AGENTS.md`, `.cursor/skills/qa-gate/SKILL.md` |
| Прод лежит | [runbooks/incident-response.md](runbooks/incident-response.md) |
| Правила стиля и запреты | [code_style_guide.md](code_style_guide.md), `.cursor/rules/10-style.mdc` |
