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

| Шаблон | Что описывает |
|--------|---------------|
| [PROJECT.md](PROJECT.md) | Стек, назначение, границы |
| [COMMANDS.md](COMMANDS.md) | Дословные команды прогона |
| [TEST-ARCHITECTURE.md](TEST-ARCHITECTURE.md) | Слои, уровни, инфраструктура |
| [PATTERN-INDEX.md](PATTERN-INDEX.md) | Принятые паттерны с примерами из кода |
| [TEST-DATA.md](TEST-DATA.md) | Фикстуры, фабрики, окружение данных |
| [CI-PIPELINE.md](CI-PIPELINE.md) | Стадии, гейты, артефакты |
| [SOURCE-MAP.md](SOURCE-MAP.md) | Откуда взят каждый факт |
| [TRACEABILITY.md](TRACEABILITY.md) | Связь требований, кейсов и тестов |

Шаблон допускает незаполненные слоты только при `status: template`.
Заполненный профиль с оставшимся слотом — дефект: он выглядит фактом.
