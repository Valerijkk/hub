# agents/templates/ — шаблоны запуска агентов

> Готовые user-промпты для запуска агентов тулкита. Схема: агент (файл из `agents/`) = system prompt, шаблон отсюда = первое сообщение. Заполни `{{слоты}}` — запускай. Каждый шаблон фиксирует вход, scope и ожидаемый формат выхода, чтобы агент не выпрашивал уточнения.
>
> **Все шаблоны — специализации [`_MASTER-PROMPT-TEMPLATE.md`](_MASTER-PROMPT-TEMPLATE.md)** — единого скелета «кто / что / где + best practices» (role prompting, контекст-инъекция, формат вывода, few-shot, chain-of-thought, step-back/ReAct, параметры генерации, встроенная верификация). Сценарий не покрыт готовым шаблоном — бери мастер и заполняй с нуля.

| Шаблон | Агент | Сценарий |
|--------|-------|----------|
| [_MASTER-PROMPT-TEMPLATE.md](_MASTER-PROMPT-TEMPLATE.md) | любой / делегирование | Единый скелет промпта со всеми техниками — база для остальных |
| [review-request.md](review-request.md) | test-code-reviewer | Ревью diff/PR с тестами |
| [naming-audit.md](naming-audit.md) | test-code-reviewer | Точечный аудит нейминга директории/домена |
| [test-case-generation.md](test-case-generation.md) | test-case-generator | TC из документации фичи |
| [test-case-verification.md](test-case-verification.md) | любой + скил test-verification | Независимый гейт TC-набора |
| [debug-session.md](debug-session.md) | test-debugger | Разбор упавшего теста/сьюта |
| [flaky-hunt.md](flaky-hunt.md) | test-debugger | Охота на flaky по истории CI |
| [docs-sync.md](docs-sync.md) | test-documenter | Синхрон доков после изменений |
| [coverage-gap.md](coverage-gap.md) | test-case-generator | Кейсы на непокрытые зоны по coverage-отчёту |
| [context-refresh.md](context-refresh.md) | любой + скил context-intake | Обновление context/ после реорганизации |

## Правила шаблонов

1. Слоты `{{...}}` заполняются ВСЕ; «не применимо» — пишется явно, не оставляется пустым.
2. Scope всегда ограничен (файлы/домен/diff) — «посмотри вообще всё» не задача, а желание.
3. Выход всегда сверяется с форматом из файла агента — шаблон напоминает об этом агенту.
