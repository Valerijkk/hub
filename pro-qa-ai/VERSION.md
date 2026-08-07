---
name: pro-qa-ai
product: pro-qa-ai
version: 0.4.0-beta.2
status: foundation-ready
release_date: не выпущено — beta, GA после paired eval
owner: automation-quality
validation_status: STRUCTURAL_PASS
last_validation: 2026-08-07
release_eligibility: foundation-ready-eval-blocked
provenance: MANIFEST.md
supersedes: 0.4.0-beta.1
---

# Версия — pro-qa-ai

Текущая semantic version: **`0.4.0-beta.2`**.

**foundation-ready** — AI-слой для тестового фреймворка. Корень поставки:
эта папка. Продуктовая карта: [TOUCHSTONE.md](TOUCHSTONE.md).

## Усиление portable release в `0.4.0-beta.2`

Упаковка стала транзакционной: отдельный sibling staging, exact inventory с
размерами и SHA-256, повторная проверка destination перед atomic publish и
fail-closed cleanup. Добавлены safe-I/O, secret scan, provenance gate и
14-шаговый clean-room protocol. Правила размера функций, SOLID, DRY, mocks и
Page Objects явно классифицированы как review guidance, а не универсальные
hard gates.

## Ломающее изменение в `0.4.0-beta.1`

Папка переименована `prompt-engineering` → `pro-qa-ai`: имя описывает
объект работы, а не способ.

Убраны `evals/` (протокол замера, ни разу не прогнанный) и `adapters/`
вместе со всей машинерией task-packet, существовавшей только ради
адаптера generic-chat. Убраны `model-routing` и `models/`: выбор модели
задаёт среда запуска, а не текст промпта.

Структура верхнего уровня сокращена с пятнадцати папок до девяти. Семь
каталогов с разнородным «знанием» сведены в `standards/` и `workflows/`,
записи работы — в `work/`. Прежние пути `rules/`, `practices/`,
`patterns/`, `review-feedback/`, `playbooks/`, `recipes/`, `prompts/`,
`planning/` на верхнем уровне больше не существуют.

Из валидатора снята проверка `cursor_adapter` — проверять нечего.
Добавлена `inline_paths`: ссылки в обратных кавычках раньше не
проверялись, и три пути протухли незамеченными.

Роли и скилы приведены к единой форме: девять фиксированных секций у
роли, семь у скила. Файлы, написанные под старую форму, валидатор теперь
отклоняет (`agent_contract`, `skill_contract`).

Удалён слой совместимости в корне `context/`: четыре файла —
`PROJECT-CONTEXT`, `TESTFRAMEWORK-MAP`, `NAMING-CONVENTIONS`,
`PATTERNS-LIBRARY`. Их ничто не заполняло, а четыре скила из-за них
носили колонку fallback.

Два параллельных словаря слоя ревью схлопнуты в один. Каталоги
`standards/team-review/{inbox,approved}` и `TEMPLATE.md` со схемой `RF-*`
удалены; остались `threads/` и `checks/` со схемой `THREAD-*`/`CHECK-*`.
Кто заполнял inbox по старому шаблону — переносит записи в
`threads/inbox/`.

## Ломающее изменение в `0.3.0-beta.1`

Из пакета вынесены 26 файлов слоя разработки продукта: 12 ролей
продуктовой команды, три скила жизненного цикла фазы, пять плейбуков
поставки, фрагменты категории 13, gsd-port и portable hub kit.

Причина: пакет раздавал корпоративный ростер автоматизаторам, которым
он не нужен, и одновременно делал вид, что покрывает разработку. Граница
теперь проходит по объекту работы — пакет про тестовый фреймворк.

Кто ставил `0.2.0-rc.1` ради ростера — после обновления его не найдёт.
Package: 174 → 148 файлов.

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

Актуальная приёмка выполняется целиком по
[`docs/PORTABLE-RELEASE.md`](docs/PORTABLE-RELEASE.md) в resource-limited
Docker: source validator/tests, dry-run, transactional package, exact artifact
validator/tests и digest. Число файлов, тестов и digest являются свойствами
конкретного candidate и сохраняются во внешнем release evidence, а не как
вечные константы этого документа.

Package исключает generated project profiles, локальные прогоны и private
review material. Целостность состава защищена `required_paths`, exact staged
inventory и негативными sentinels для race, verification failure, unexpected
file/directory и cleanup.

## Продвижение к GA `1.0.0`

1. Зелёные validator unit tests на intended distribution
2. Структурная валидация exact package
3. Approved source + governance review
4. Representative paired eval evidence
5. Sign-off accountable owner

Пока п.4–5 открыты: `validation_status = STRUCTURAL_PASS`,
`release_eligibility = foundation-ready-eval-blocked`.

История: [CHANGELOG.md](CHANGELOG.md) · владение: [MANIFEST.md](MANIFEST.md).
