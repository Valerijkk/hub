---
name: pro-qa-ai
product: pro-qa-ai
version: 0.4.0-beta.1
status: foundation-ready
release_date: не выпущено — beta, GA после paired eval
owner: automation-quality
validation_status: STRUCTURAL_PASS
last_validation: 2026-08-03
release_eligibility: foundation-ready-eval-blocked
provenance: MANIFEST.md
supersedes: 0.3.0-beta.1
---

# Версия — pro-qa-ai

Текущая semantic version: **`0.4.0-beta.1`**.

Это **foundation-ready** издание линейки pro-qa-ai — AI-слой **для тестового
фреймворка**. Корень поставки: эта папка. Продуктовая карта:
[TOUCHSTONE.md](TOUCHSTONE.md).

## Ломающее изменение в `0.4.0-beta.1`

Папка переименована `prompt-engineering` → `pro-qa-ai`: имя описывает
объект работы, а не способ.

Убраны `evals/` (протокол замера, ни разу не прогнанный) и `adapters/`
вместе со всей машинерией task-packet, существовавшей только ради
адаптера generic-chat. `model-routing` и `models/` перенесены в издание
Pro: выбор модели — вопрос оркестрации.

Структура верхнего уровня сокращена с пятнадцати папок до девяти. Семь
каталогов с разнородным «знанием» сведены в `standards/` и `workflows/`,
записи работы — в `work/`. Прежние пути `rules/`, `practices/`,
`patterns/`, `review-feedback/`, `playbooks/`, `recipes/`, `prompts/`,
`planning/` на верхнем уровне больше не существуют.

Из валидатора снята проверка `cursor_adapter` — проверять нечего.

## Ломающее изменение в `0.3.0-beta.1`

Из пакет вынесены 26 файлов слоя разработки: 12 ролей продуктовой
команды, три скила жизненного цикла фазы, пять плейбуков поставки,
фрагменты категории 13, gsd-port и portable hub kit. Все они переехали
в Pro.

Причина: пакет раздавал корпоративный ростер автоматизаторам, которым
он не нужен, и одновременно делал вид, что покрывает разработку. Теперь
граница проходит по объекту работы — пакет про тестовый фреймворк,
Pro про продукт.

Кто ставил `0.2.0-rc.1` ради ростера — после обновления не найдёт его в
пакет; нужно поставить Pro рядом. Package: 174 → 148 файлов.

Почему `beta`, а не `rc`: release candidate обещает «релизим, если не
всплывёт багов», а по собственным воротам издание к релизу не допущено,
пока `paired eval` не прогнан.

## Что означает статус

| Термин | Значение |
|--------|----------|
| `foundation-ready` | Идентичность, вход, package, validator, governance готовы к переносу |
| `STRUCTURAL_PASS` | Дерево и валидатор прошли структурные проверки |
| `eval-blocked` | Paired effectiveness / ROI ещё `NOT_RUN` — GA `1.0.0` нельзя |

Это **не** заявление о completion effectiveness, conformance или ROI.

## Evidence (structural)

Замер 2026-08-03, перезамерено после реорганизации структуры. Команды воспроизводимы из корня
`pro-qa-ai/`.

| Проверка | Команда | Результат |
|----------|---------|-----------|
| Structural validation | `python scripts/validate_toolkit.py` | PASS · 0 errors / 0 warnings · exit 0 |
| Script tests | `python -m pytest scripts/tests -q` | 41 passed / 1 skipped (42 total); SKIP — symlink недоступен |
| Portable package | `python scripts/package_portable.py --root . --output <dst> --dry-run` | 162 files · exit 0 |

Package исключает generated Hub profile, smoke и raw feedback. Целостность
состава защищена `required_paths` в `scripts/package_manifest.json`: удаление
любого обязательного файла роняет пакер с exit 2 (проверено чтением утверждения на
`COMMANDS.md`, `standards/` и `work/` 2026-08-03).

## Продвижение к GA `1.0.0`

1. Зелёные validator unit tests на intended distribution
2. Структурная валидация exact package
3. Approved source + governance review
4. Representative paired eval evidence
5. Sign-off accountable owner

Пока п.4–5 открыты: `validation_status = STRUCTURAL_PASS`,
`release_eligibility = foundation-ready-eval-blocked`.

История: [CHANGELOG.md](CHANGELOG.md) · владение: [MANIFEST.md](MANIFEST.md).
