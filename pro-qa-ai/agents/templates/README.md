# agents/templates/ — шаблоны запуска агентов

> Готовые контракты делегирования. Человек может ставить задачу обычным языком; агент нормализует её по мастер-контракту и спрашивает только о материальных пробелах. Полный шаблон обязателен между агентами, в CI и при handoff, где принимающая сторона не видит исходный чат.
>
> Core prompts не являются готовыми конфигами конкретного runtime.
> Модель и параметры генерации задаёт среда запуска, а не текст промпта.

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
| [context-refresh.md](context-refresh.md) | любой + скил refresh-test-context | Обновление существующего profile после реорганизации |

## Правила шаблонов

1. Для делегирования заполнить outcome, scope, do-not-touch, references, done, evidence, stop, approval boundary и handoff.
2. Необязательный слот можно обозначить `N/A: причина`; не спрашивать человека о нем без влияния на результат.
3. В references указывать generated context/pattern cards, а не копировать их содержимое.
4. Scope ограничивать файлами, diff или доменом. Broad audit допустим только с явным outcome и stopping condition.
5. Выход сверять с role-specific schema; факты сопровождаются evidence и confidence.
6. Prompt body не назначает модель и runtime-параметры — это ответственность adapter/routing layer.
