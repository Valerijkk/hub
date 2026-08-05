# Шаблоны context profile

Эти файлы — канонические, project-neutral шаблоны, которые `bootstrap-test-context`
копирует в `context/projects/<project-slug>/`.

Правила:

- Держи placeholders generic; никогда не добавляй proprietary project examples.
- Сгенерированные утверждения требуют source IDs из `SOURCE-MAP.md`.
- Используй литерал `UNKNOWN`, когда evidence отсутствует.
- Сохраняй блоки `BEGIN MANUAL` / `END MANUAL` при refresh.
- Обнаруженная команда остаётся `not-run`, пока человек явно не одобрит
  execution.
- Этот каталог — единственный канонический источник шаблонов; не зеркаль
  contracts под `context/projects/`.

## Состав

Восемь артефактов профиля. `tf-init` заполняет их фактами из репозитория,
`tf-refresh` обновляет затронутые.

| Шаблон | Что описывает | Кто грузит |
|--------|---------------|------------|
| [PROJECT.md](PROJECT.md) | Стек, назначение, границы | `test-debugger`, `test-documenter`, `tf-triage`, `tf-verify`, `tf-antiai` |
| [COMMANDS.md](COMMANDS.md) | Дословные команды прогона | `test-sdet`, `test-code-reviewer`, `test-debugger`, `test-documenter`, `tf-gate`, `tf-status`, `tf-triage`, `tf-review` |
| [TEST-ARCHITECTURE.md](TEST-ARCHITECTURE.md) | Слои, уровни, инфраструктура | все пять ролей, `tf-review`, `tf-triage`, `tf-naming`, `tf-verify` |
| [PATTERN-INDEX.md](PATTERN-INDEX.md) | Принятые паттерны с примерами из кода | `test-case-generator`, `test-sdet`, `tf-review`, `tf-naming` |
| [TEST-DATA.md](TEST-DATA.md) | Фикстуры, фабрики, окружение данных | `test-documenter`, `tf-triage`, `tf-verify` |
| [CI-PIPELINE.md](CI-PIPELINE.md) | Стадии, гейты, артефакты | `test-code-reviewer`, `test-documenter`, `tf-review`, `tf-triage` |
| [SOURCE-MAP.md](SOURCE-MAP.md) | Откуда взят каждый факт | все: [rules/00](../../standards/rules/00-global-agent-rules.md) §1.1 берёт отсюда пути |
| [TRACEABILITY.md](TRACEABILITY.md) | Связь требований, кейсов и тестов | `test-case-generator`, `test-code-reviewer`, `test-documenter`, `tf-verify`, `tf-naming` |

Шаблон допускает незаполненные слоты только при `status: template`.
Заполненный профиль с оставшимся слотом — дефект: он выглядит фактом.

## У каждого поля должен быть потребитель

Профиль грузится при каждом запуске команды. Поле, которое не названо ни в
одной секции «Что гружу» роли и ни в одном шаге скила, стоит токенов на
каждой загрузке и не влияет ни на один вердикт.

Поэтому при добавлении поля в шаблон называется тот, кто его читает: роль,
скил или рецепт. Не называется — поле не добавляется. Проверка обратная и
такая же механическая: грепни имя секции по `agents/` и `skills/`; ноль
попаданий — кандидат на удаление, а не «пусть лежит, вдруг пригодится».
