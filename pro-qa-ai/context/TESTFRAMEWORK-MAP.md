---
status: template            # template | filled
compatibility: deprecated
read_only: true
replacement: "context/projects/<profile>/{TEST-ARCHITECTURE,TEST-DATA,COMMANDS,CI-PIPELINE}.md"
filled_at: {{YYYY-MM-DD}}
---

# Test Framework Map

> **DEPRECATED / READ-ONLY COMPATIBILITY.** Новые workflows используют active
> profile `TEST-ARCHITECTURE.md`, `TEST-DATA.md`, `COMMANDS.md` и
> `CI-PIPELINE.md`.

## 1. Пирамида тестов проекта (фактическая, не желаемая)

| Уровень | Есть? | Раннер | Где живут | Примерное кол-во |
|---------|-------|--------|-----------|------------------|
| Unit | {{yes/no}} | {{go test / vitest / jest / pytest}} | {{path}} | {{count}} |
| Integration | {{yes/no}} | {{runner + как поднимаются зависимости}} | {{path}} | {{count}} |
| Contract/API | {{yes/no}} | {{...}} | {{path}} | {{count}} |
| E2E | {{yes/no}} | {{playwright / cypress / ...}} | {{path}} | {{count}} |
| Performance | {{yes/no}} | {{...}} | {{path}} | {{count}} |

## 2. Организация тестов

- **Принцип группировки:** {{по доменам | по слоям | рядом с кодом | отдельный модуль}}
- **Список доменов/групп:** {{domains_list}}
- **Почему так** (если есть архитектурная причина): {{reason_or_link}}

```
{{test_tree — дерево тестовых директорий до 3 уровней}}
```

## 3. Инфраструктура тестов

| Компонент | Реализация | Путь |
|-----------|-----------|------|
| Фикстуры / фабрики | {{factory pattern / fixtures files / builders}} | {{path}} |
| Тестовая БД | {{testcontainers / compose / in-memory / mock}} | {{path/config}} |
| Моки внешних API | {{msw / gomock / ручные стабы / нет}} | {{path}} |
| Page Objects (E2E) | {{yes/no}} | {{path}} |
| Общие хелперы | {{...}} | {{path}} |
| Тестовые env-переменные | {{names}} | {{где задаются}} |

## 4. Запуск (детали сверх PROJECT-CONTEXT §4)

- Обязательные шаги перед прогоном: {{link scripts / поднять compose / миграции}}
- Как гоняется один домен/файл/тест: `{{cmd}}`
- Флаги обязательные всегда: {{-race / -count=1 / --runInBand / ...}} — почему: {{reason}}
- Таймауты: {{defaults}}
- Артефакты прогона, которые нельзя коммитить: {{coverage dumps, traces, screenshots}}

## 5. Соглашения фреймворка

- Разделение unit vs integration: {{build tags / suffix / папки}}
- Параллельность: {{t.Parallel() policy / workers}}
- Скip-политика: {{когда допустим skip и как помечается}}
- Флаки-политика: {{retry / quarantine / запрещены}}
- Snapshot-тесты: {{policy}}

## 6. Coverage

- Как считается: `{{coverage_cmd}}`
- Пороги: {{thresholds по пакетам/модулям}}
- Где фиксируются результаты: {{coverage doc path}}

## 7. Известные болевые точки

> Из истории проекта: flaky-зоны, медленные сьюты, хрупкие места. Дебаггер читает это ПЕРВЫМ.

- {{pain_point_1: сьют/файл — симптом — известная причина}}
- {{pain_point_2}}
