# Touchstone Standard — Test Automation AI Quality Layer

**Продукт:** [Touchstone](TOUCHSTONE.md) · **Издание:** Standard `0.4.0-beta.1`
**Pro (максималка):** sibling `prompt-engineering-pro/` (рядом в исходном репо)

**Политика языка:** Все инструкции тулкита — на русском. Идентификаторы, пути, CLI и машинные enum остаются как в схемах.

**Первый вход:** [START-HERE.md](START-HERE.md). Не читай дерево целиком.

> Переносимая AI-прослойка для автоматизаторов: она изучает существующий
> тестовый фреймворк, превращает его факты и проверенные паттерны в контекст для
> моделей, направляет специализированных агентов и измеряет качество результата.
> Это **не** каркас тестового фреймворка и не каталог «магических промптов».
> Эта папка — **фундамент Touchstone** для тестовых фреймворков.

## Ценность

Без слоя модель получает разовый промпт и вынуждена заново угадывать устройство
проекта. Со слоем она получает:

- верифицированную карту конкретного тестового фреймворка;
- реальные project patterns с provenance;
- вручную добавленные автоматизатором пары «плохо → комментарий коллеги → хорошо»;
- task recipes, которые ссылаются на нужный контекст, а не копируют его целиком;
- GSD-подобный цикл plan → execute → verify → handoff;

Эффект измеряется на пилоте: first-pass acceptance, время ревью, дефекты,
hallucinated paths/commands, scope escapes, flaky-risk и стоимость. Заявления
«ускоряет в N раз» без такого замера запрещены.

## Как устроен пакет

Три вопроса — три места, куда смотреть:

```
Что запустить?          COMMANDS.md → skills/
Чему следовать?         standards/
Как выполнять работу?   workflows/
```

Факты о конкретном проекте собирает `tf-init` в `context/projects/<profile>/`.
Записи работы — задачи, дефекты, обсуждения — в `work/`.

## Быстрый старт

```
1. Поставь пакет в репозиторий тестового фреймворка (INSTALL.md)
2. Запусти tf-init — он изучит репозиторий и заполнит контекст
3. Запусти tf-status — он скажет, что делать дальше
4. Дальше работай командами: tf-cover, tf-triage, tf-review, tf-gate
```

Замечания коллег с ревью переносит человек в
`standards/team-review/threads/inbox/`, дальше `tf-curate` превращает их
в проверяемые правила.

## Структура

| Папка | Вопрос, на который отвечает |
|-------|------------------------------|
| [`skills/`](skills/README.md) | Что запустить |
| [`agents/`](agents/README.md) | Кто исполняет |
| [`context/`](context/README.md) | Что известно о моём проекте |
| [`standards/`](standards/README.md) | Чему следовать: правила, практики, паттерны, решения команды |
| [`workflows/`](workflows/README.md) | Как выполнять работу: плейбуки, рецепты, фрагменты промптов |
| [`work/`](work/README.md) | Где записывается: задачи, дефекты, обсуждения, размышления |
| `scripts/` | Валидатор и упаковщик |
| `docs/`, `sources/` | Управленческие документы и реестр источников |

**Знание команды сильнее общих практик.** Треды из реальных ревью лежат в
[`review-feedback/threads/`](standards/team-review/threads/README.md) и читаются
ревьюером **до** общих практик из [`practices/`](standards/practices/README.md).
Порядок источников — в [`rules/20`](standards/rules/20-test-quality.md).

**Пакет командный.** Полная поверхность из десяти команд и типовые цепочки —
[COMMANDS.md](COMMANDS.md). Первая команда в новом репозитории — `tf-init`,
в уже настроенном — `tf-status`.

**Слой Corporate/GSD переехал в издание Pro.** Роли разработки, волны,
GSD-порт и portable hub kit живут в sibling-папке `prompt-engineering-pro/`.
Standard остаётся чисто про тестовый фреймворк.

## Роли и skills

Ростер Standard: [agents/README.md](agents/README.md). Здесь **только
тестовый фреймворк** — четыре роли.

| Роль | Объект работы |
|------|---------------|
| test-case-generator | Тест-кейсы из документации |
| test-code-reviewer | Код тестового фреймворка |
| test-debugger | Упавшие и мигающие тесты |
| test-documenter | Документация тестового фреймворка |

Роли разработки продукта — orchestrator, product-owner, architect,
backend-dev, frontend-dev, qa-manual, qa-auto, code-reviewer, debugger,
security-reviewer, devops, docs-writer — в издании **Pro**, sibling-папка
`prompt-engineering-pro/`. Ссылкой отсюда не достаются намеренно: Standard
переносится самостоятельным пакетом и не вправе зависеть от наличия Pro.

Skills — явные процедуры, а не личности: bootstrap/refresh context,
curation, planning, generation/review checks, goal verification и handoff.

## Конвенции тулкита

- `{{placeholder}}` — слот для подстановки (значение из active profile или руками).
- `MUST/MUST NOT` — жёсткое требование; `PASS/FAIL` — результат проверки.
- Severity: `CRITICAL > HIGH > MEDIUM > LOW` (критерии: [rules/30-severity-policy.md](standards/rules/30-severity-policy.md)).
- Entry points самодостаточны; детальные patterns/references загружаются только
  когда на них ссылается задача (progressive disclosure).
