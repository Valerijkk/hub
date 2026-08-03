# Layer 01 — Core Quality Layer (= Standard)

Pro **не заменяет** Standard. Этот слой = обязательная база.

## Канон

Корень: [`../../pro-qa-ai/`](../../pro-qa-ai/)

| Нужно | Путь в Standard |
|-------|-----------------|
| Продукт | `TOUCHSTONE.md` |
| Вход | `START-HERE.md` |
| Установка | `INSTALL.md` |
| Агенты | `agents/` |
| Bootstrap | `skills/bootstrap-test-context/` |
| Recipes | `recipes/` |
| Validator | `scripts/validate_toolkit.py` |
| Evals | `evals/` |

## Pro-добавка к ядру

- Требовать generated profile + `PATTERN-INDEX` до любой генерации тестов.
- Любой claim PASS без пути команды/артефакта → FAIL (см. layer 05).
- Corporate delivery (волны) — только через layer 04, не ad-hoc чат.

## DoD слоя

Агент может: inventory → bootstrap → recipe → structural validate, не читая Pro.
