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
