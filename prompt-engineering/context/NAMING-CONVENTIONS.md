---
status: template            # template | filled
compatibility: deprecated
read_only: true
replacement: "context/projects/<profile>/{TEST-ARCHITECTURE,TRACEABILITY,PATTERN-INDEX}.md"
filled_at: {{YYYY-MM-DD}}
---

# Naming Conventions — тесты

> **DEPRECATED / READ-ONLY COMPATIBILITY.** Источник истины — active profile
> `TEST-ARCHITECTURE.md`, `TRACEABILITY.md`, `PATTERN-INDEX.md` и linked
> human-approved cards.

## 1. Файлы тестов

| Уровень | Шаблон имени | Пример из проекта | Источник |
|---------|--------------|-------------------|----------|
| Unit | `{{pattern, напр. <unit>_test.<ext>}}` | {{real_example_path}} | {{style guide / факт кода}} |
| Integration | `{{pattern}}` | {{...}} | {{...}} |
| E2E | `{{pattern, напр. <flow>.spec.<ext>}}` | {{...}} | {{...}} |
| Фикстуры/хелперы | `{{pattern}}` | {{...}} | {{...}} |

- Регистр и разделители: {{snake_case / kebab-case / camelCase}} — где какой.
- Соответствие «файл теста ↔ файл кода»: {{правило, напр. service.go → service_test.go}}
- Группирующие суффиксы/префиксы (волны, версии, теги): {{pattern или «не используются»}}

## 2. Тестовые функции / кейсы

| Раннер | Шаблон | Пример |
|--------|--------|--------|
| {{go test}} | `{{Test<Unit>_<Scenario>_<Expected>}}` | {{TestCreateIssue_EmptyTitle_ReturnsError}} |
| {{vitest/jest}} | `{{describe('<Unit>') + it('<behavior>')}}` | {{...}} |
| {{playwright}} | `{{test('<user flow>')}}` | {{...}} |

Правила формулировки:

- Имя отвечает на три вопроса: **что тестируем / при каких условиях / что ожидаем**.
- Язык имён: {{english / russian / mixed}} — канон: {{...}}.
- Запрещённые имена: `Test1`, `TestFix`, `TestNew`, `it('works')`, `test('should work')` — не несут информации.
- Длина: имя можно читать в отчёте CI без открытия файла.

## 3. Соответствие «тест ↔ требование/кейс»

- Формат ID тест-кейсов: {{TC-<domain>-<nnn> или другой}}
- Как тест ссылается на кейс/требование: {{комментарий / имя / аннотация / никак}}
- Обязательность привязки: {{для каких уровней обязательна}}

## 4. Переменные и хелперы внутри тестов

- Фикстуры: {{naming, напр. newTestUser(), makeIssue(), userFixture}}
- Sut-переменная (system under test): {{канон: sut / svc / handler}}
- Ожидания: {{want/got | expected/actual}} — канон: {{...}}
- Табличные тесты: имя поля кейса {{name/desc}}, имя переменной {{tt/tc/tests}}

## 5. Известные отступления (легаси)

> Файлы/зоны, где нейминг не соответствует канону и это принятый долг — naming-check их репортит как WARN, не ERROR.

- {{legacy_zone_1: путь — чем отличается — тикет/решение}}
- {{legacy_zone_2}}
