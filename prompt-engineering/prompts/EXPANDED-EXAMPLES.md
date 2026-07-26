# Собранные recipe-эталоны (assembly из fragments)

> По одному полностью собранному recipe на каждую из 12 категорий. Каждый — `#1` соответствующего файла, собранный по [`../agents/templates/_MASTER-PROMPT-TEMPLATE.md`](../agents/templates/_MASTER-PROMPT-TEMPLATE.md): outcome → constraints → context profile → output schema → evidence → completion bar → stop/approval.
>
> **Как пользоваться:** `prompts/*.md` — фрагменты-слоты `[ЧТО]`. Bare fragment недостаточен: добавь `context/projects/<profile>/`, approved cards из `patterns/`, собери по профилю категории (блок «Recipe assembly» в шапке). Ниже — 12 эталонов результата assembly.
>
> `{{слоты}}` заполняются из `context/projects/<profile>/` или вручную. Профили категорий — таблица в конце.

---

## 01 — Дизайн тест-кейсов (разворот `01 #1`)

```markdown
[OUTCOME]
Полный список атомарных требований R1..Rn из {{спека}} (разделы {{разделы}}) —
фундамент для последующей генерации TC; сами кейсы на этом шаге НЕ пишем.

[CONSTRAINTS]
- Только извлечение; кейсы и код вне scope.
- Неоднозначности → «Вопросы к документации», НЕ в требования.
- Запрет invented-expectations (rules/40 этап 6).

[CONTEXT PROFILE]
- Profile: context/projects/<profile>/ (manifest, traceability map).
- Pattern cards: patterns/ — TC-design, requirement extraction.
- Вход: спецификация {{спека}}; active profile `TRACEABILITY.md` для ID.

[OUTPUT SCHEMA]
1. Резюме: сколько требований, какие разделы покрыты.
2. Список R1..Rn: ID | цитата «...» | раздел-источник.
3. Неявные контракты (лимиты/форматы/роли/дефолты) — отдельный блок.
4. Риски: неоднозначные места спеки.
5. Следующий шаг: передать список в дизайн кейсов (01 #23/#40).

[EVIDENCE]
- Каждая цитата сверена с источником (rules/40 этап 3).
- Каждое «должен/если-то» стало строкой (этап 2).
- Компактный блок Verification в ответе.

[COMPLETION BAR]
- Все разделы {{разделы}} покрыты или явно помечены gap.
- Нет строк без источника-цитаты.

[STOP/APPROVAL]
- STOP: спека противоречива → эскалация, не догадки.
- Approval: список R принят → handoff на 01 #23.
```

---

## 02 — Юнит-тесты (разворот `02 #1`)

```markdown
[OUTCOME]
Юнит-тесты для {{функция}} на {{стек}} по паттернам проекта; сьют зелёный, лог приложен.

[CONSTRAINTS]
- Только тест-файл {{путь_теста}}; прод-код не трогаем.
- Каждый тест обязан мочь упасть (rules/20 #1).
- Happy + каждый error-путь + границы.

[CONTEXT PROFILE]
- Profile: context/projects/<profile>/.
- Pattern cards: linked approved unit cards из `PATTERN-INDEX.md`.
- Команды: active profile `COMMANDS.md`.

[OUTPUT SCHEMA]
1. Резюме: сколько кейсов, классы (happy/error/границы).
2. Код тестов (AAA, канонические фабрики, нейминг).
3. Таблица покрытия: кейс → что проверяет → класс.
4. Риски: непокрытые ветки с причиной.
5. Команда прогона + ожидаемый результат.

[EVIDENCE]
- Few-shot эталон: human-approved card ID/path.
- Реальный вывод прогона (не «должно работать»).
- Verification rules/40: этапы 4, 6, 7.

[COMPLETION BAR]
- Сьют green; лог приложен.
- Нет фейковых тестов.

[STOP/APPROVAL]
- STOP: спека поведения {{ссылка}} неоднозначна → уточнение до кода.
- Approval: test-code-reviewer + CI merge-гейт.
```

---

## 03 — Интеграционные тесты (разворот `03 #1`)

```markdown
[OUTCOME]
Сетап интеграционных тестов {{модуль}}: окружение, изоляция, healthcheck — готов к написанию тестов.

[CONSTRAINTS]
- Сетап и стратегия изоляции; сами тесты — следующий шаг.
- Команды только из context (запрет выдумывания).
- Изоляция держит параллельный прогон.

[CONTEXT PROFILE]
- Profile: context/projects/<profile>/.
- Pattern cards: patterns/ — integration setup.
- `TEST-ARCHITECTURE.md` + `TEST-DATA.md` (инфраструктура).

[OUTPUT SCHEMA]
1. Резюме: как поднимается окружение, стратегия изоляции.
2. Код сетапа (подключение, миграции, фикстуры).
3. Стратегия изоляции с обоснованием (транзакция/уникальные ключи/truncate).
4. Риски: параллельность, FK, скорость.
5. Команда прогона + healthcheck зависимостей.

[EVIDENCE]
- Сетап проверен запуском (rules/40 этап 3).
- Cleanup не оставляет следов (этап 7).

[COMPLETION BAR]
- Healthcheck всех {{БД/очереди}} green.
- Документирована стратегия cleanup.

[STOP/APPROVAL]
- STOP: зависимость недоступна → зафиксировать блокер, не mock без ADR.
- Approval: авторы интеграционных тестов модуля.
```

---

## 04 — API-тесты (разворот `04 #1`)

```markdown
[OUTCOME]
Полный тест-сьют {{метод}} {{эндпоинт}} по чек-набору A9; 401 и denial 403 покрыты.

[CONSTRAINTS]
- Только тесты этого эндпоинта.
- Ассерты СХЕМЫ ответа, не полный снапшот JSON.
- 401 и denial 403 обязательны (rules/20 #8, A11).

[CONTEXT PROFILE]
- Profile: context/projects/<profile>/.
- Pattern cards: linked approved API cards; baseline `rules/20`.
- OpenAPI {{путь}}; роли {{роли}}.

[OUTPUT SCHEMA]
1. Резюме: классы (happy/валидация/401/403/404/конфликт).
2. Код через тест-клиент проекта.
3. Матрица: класс → статус → что ассертит.
4. Риски: N+1, непокрытые error-пути.
5. Команда прогона.

[EVIDENCE]
- Прогон зелёный; denial 403 не пропущен.
- Verification rules/40: этапы 4, 6, 7.

[COMPLETION BAR]
- Все классы A9 для эндпоинта покрыты или явный gap с причиной.

[STOP/APPROVAL]
- STOP: спека эндпоинта {{ссылка}} противоречит OpenAPI → эскалация.
- Approval: merge-гейт + CI.
```

---

## 05 — E2E / UI (разворот `05 #1`)

```markdown
[OUTCOME]
Компонентные тесты {{компонент}}: loading/empty/error/success + a11y-минимум; прогон green.

[CONSTRAINTS]
- Только тест-файл компонента.
- Селекторы role/label > testid > ⛔ CSS/XPath (A10).
- Ассерты на видимый результат, не пропсы/внутренний стейт.

[CONTEXT PROFILE]
- Profile: context/projects/<profile>/.
- Pattern cards: patterns/ — UI A10.
- Active profile `TEST-ARCHITECTURE.md` (mock API, renderer).

[OUTPUT SCHEMA]
1. Резюме: покрытые состояния и a11y.
2. Код: accessible queries, действия пользователя, ассерты видимого.
3. Таблица: состояние/действие → видимый результат.
4. Риски: асинхронность, зависимость от вёрстки.
5. Команда прогона (+ prerequisites из active profile `COMMANDS.md`).

[EVIDENCE]
- Все состояния + a11y (rules/40 этап 4).
- Прогон зелёный (этап 7).

[COMPLETION BAR]
- Нет CSS/XPath селекторов без обоснования.

[STOP/APPROVAL]
- STOP: компонент без спеки поведения → запрос спеки до тестов.
- Approval: reviewer + CI.
```

---

## 06 — Ревью тестов (разворот `06 #1`)

```markdown
[OUTCOME]
Вердикт ревью тестов в {{diff}}: approve | request changes; находки с severity и исправлениями.

[CONSTRAINTS]
- Readonly — фиксы не пишу.
- Severity по rules/30; сомнение HIGH/MEDIUM → HIGH.
- Каждая находка: готовое исправление + ссылка на канон.

[CONTEXT PROFILE]
- Profile: context/projects/<profile>/.
- Pattern cards: patterns/ — review checklist.
- rules/20, rules/30; diff {{ссылка}}, TC {{список}}, лог {{лог}}.

[OUTPUT SCHEMA]
Строго test-code-reviewer §4:
- Находки [severity] file:line — проблема → исправление.
- Секция Naming (файлы + функции).
- Секция Прогон (лог приложен / отсутствует).
- Вердикт: approve | request changes.

[EVIDENCE]
- Инверсия условия для каждого нового теста — упадёт ли?
- Целевой проход на фейковые тесты и имена-вруны.
- CONFIRMED/PLAUSIBLE (rules/40 этапы 5, 6, 9).

[COMPLETION BAR]
- Все CRITICAL/HIGH адресованы или явно приняты риском с владельцем.

[STOP/APPROVAL]
- STOP: лог прогона отсутствует → request changes без approve.
- Approval: merge-гейт при approve.
```

---

## 07 — Дебаг падений (разворот `07 #1`)

```markdown
[OUTCOME]
Triage падения {{тест}}: ROOT=code|test|infra|FLAKY|STOP — до любого фикса.

[CONSTRAINTS]
- Triage ДО фикса; запрет подгонки ассерта.
- STOP при неоднозначной спеке → эскалация.
- Воспроизведение реальной командой.

[CONTEXT PROFILE]
- Profile: context/projects/<profile>/.
- Pattern cards: patterns/ — debug triage A8.
- Active profile `TEST-ARCHITECTURE.md` risks + `COMMANDS.md`; лог {{лог}},
  спека {{спека}}, git {{диапазон}}.

[OUTPUT SCHEMA]
Debug Report (test-debugger §5):
- Triage: ROOT=… — обоснование с цитатой спеки.
- Воспроизведение: команда → результат.
- Гипотезы (если triage требует дебага).

[EVIDENCE]
- Сверка: что тест ДОЛЖЕН по спеке vs что ФАКТИЧЕСКИ ассертит.
- Команда воспроизведения → интерпретация вывода.
- Verification rules/40: этапы 3, 6.

[COMPLETION BAR]
- ROOT= зафиксирован с цитатой.
- Воспроизведение подтверждено.

[STOP/APPROVAL]
- STOP: спека неоднозначна → эскалация, не толкование.
- Approval: triage принят → следующий шаг фикса (если не STOP).
```

---

## 08 — Тестовые данные (разворот `08 #1`)

```markdown
[OUTCOME]
Фабрика {{сущность}} по канону: валидные дефолты, override, уникальность для параллельности.

[CONSTRAINTS]
- Только фабрика + именованные состояния при необходимости.
- Объект без override проходит валидацию системы (A4).
- Уникальные поля с префиксом теста.

[CONTEXT PROFILE]
- Profile: context/projects/<profile>/.
- Pattern cards: patterns/ — factories A4/B3.
- Схема {{сущность}} ({{путь}}).

[OUTPUT SCHEMA]
1. Резюме: поля с дефолтами, override.
2. Код фабрики.
3. Именованные состояния (active/expired/...) при необходимости.
4. Риски: параллельность, синхрон со схемой.
5. Пример использования в тесте.

[EVIDENCE]
- Фабрика прогнана через валидацию (rules/40 этап 3).
- Дефолты не создают битых данных (этап 6).

[COMPLETION BAR]
- Пример использования компилируется и проходит.

[STOP/APPROVAL]
- STOP: схема {{путь}} не найдена или устарела → sync перед фабрикой.
- Approval: авторы тестов модуля.
```

---

## 09 — Покрытие (разворот `09 #1`)

```markdown
[OUTCOME]
Отчёт покрытия {{модуль}}: проценты по пакетам + приоритетные дыры по риску (не гонка за %).

[CONSTRAINTS]
- Анализ + приоритизация; тесты не пишу.
- Проценты посчитаны командой; артефакты профиля удалить после снятия.
- Качество важнее % (rules/40 этап 5).

[CONTEXT PROFILE]
- Profile: context/projects/<profile>/.
- Pattern cards: patterns/ — coverage gap analysis.
- Active profile `COMMANDS.md`/`CI-PIPELINE.md`; пороги {{пороги}};
  критичные зоны из `PROJECT.md`.

[OUTPUT SCHEMA]
1. Резюме: % по пакетам, разрыв line vs branch.
2. Дыры: логика/риск vs тривиальное vs мёртвый код.
3. Приоритет: критичные зоны (риск × частота изменений).
4. Риски: иллюзия покрытия.
5. Следующий шаг: кейсы на топ-дыры (09 #12).

[EVIDENCE]
- Команда покрытия → чтение отчёта → классификация.
- Verification rules/40: этапы 3, 5.

[COMPLETION BAR]
- Топ-5 дыр ранжированы с обоснованием.

[STOP/APPROVAL]
- STOP: команда покрытия недоступна → зафиксировать, не оценивать «на глаз».
- Approval: планирование дозакрытия покрытия.
```

---

## 10 — CI (разворот `10 #1`)

```markdown
[OUTCOME]
Дизайн тестового пайплайна {{продукт}}: стадии, fail-fast, триггеры — готов к реализации конфига.

[CONSTRAINTS]
- Дизайн; реализация конфига — 10 #27.
- Команды из context (запрет выдумывания).
- Дешёвые гейты первыми, дорогие последними.

[CONTEXT PROFILE]
- Profile: context/projects/<profile>/.
- Pattern cards: patterns/ — CI pipeline.
- Существующий CI {{ссылка}}; раннер {{CI}}.

[OUTPUT SCHEMA]
1. Резюме: стадии, fail-fast логика.
2. Стадии: что гоняется, что блокирует мерж.
3. Триггеры: PR / merge / nightly.
4. Риски: время, флаки-зоны.
5. Следующий шаг: генерация конфига.

[EVIDENCE]
- Команды реальны (rules/40 этап 3).
- Формат конфига валиден для {{CI}} (этап 8).

[COMPLETION BAR]
- Каждая стадия привязана к команде из context.

[STOP/APPROVAL]
- STOP: {{CI}} неизвестен → уточнить до дизайна стадий.
- Approval: команда + CI-система.
```

---

## 11 — Нагрузка / Security (разворот `11 #1`)

```markdown
[OUTCOME]
Нагрузочный сценарий {{эндпоинт/флоу}}: профиль, SLO, критерии успеха/провала — готов к исполнению.

[CONSTRAINTS]
- Дизайн; исполнение — 11 #6.
- Профиль из реальной аналитики {{источник}}, не выдуман.
- Security — только авторизованное тестирование своего продукта.

[CONTEXT PROFILE]
- Profile: context/projects/<profile>/.
- Pattern cards: patterns/ — load/security.
- SLO {{цели}} (RPS, p95/p99, error rate).

[OUTPUT SCHEMA]
1. Резюме: профиль, метрики, критерии успеха/провала.
2. Сценарий: микс, ramp-up/плато/ramp-down, данные.
3. Что мерить: ресурсы, очереди, деградация зависимостей.
4. Риски: нереалистичность окружения.
5. Следующий шаг: реализация в {{инструмент}}.

[EVIDENCE]
- SLO как критерий (rules/40 этап 1).
- Реалистичность профиля (этап 5).

[COMPLETION BAR]
- Критерии pass/fail измеримы до запуска.

[STOP/APPROVAL]
- STOP: нет SLO {{цели}} → зафиксировать до сценария.
- Approval: запуск в тестовом контуре.
```

---

## 12 — Документация (разворот `12 #1`)

```markdown
[OUTCOME]
Тест-план фичи {{фича}} на одну страницу: scope, пирамида, риски, матрица трассируемости.

[CONSTRAINTS]
- Только тест-план; тесты не пишу.
- Факты из context; без «скоро/планируется» без тикета.
- Одна страница, без воды.

[CONTEXT PROFILE]
- Profile: context/projects/<profile>/.
- Pattern cards: patterns/ — test plan template.
- Спека {{ссылка}}; риски {{контекст}}.

[OUTPUT SCHEMA]
1. Scope: что тестируем / что нет.
2. Уровни пирамиды и что на каждом.
3. Риски и митигация.
4. Критерии входа/выхода, окружения, ответственные.
5. Матрица: требование → TC.

[EVIDENCE]
- Команды из context проверены (rules/40 этап 3).
- Нет вранья/устаревшего (этап 6).

[COMPLETION BAR]
- Каждый Must-требование имеет строку в матрице или явный gap.

[STOP/APPROVAL]
- STOP: спека {{ссылка}} недоступна → запрос до плана.
- Approval: команда QA + разработчики фичи.
```

---

## Профили категорий (сводка для assembly)

> Тот же профиль вынесен в блок «Recipe assembly» каждого файла-категории.

| # | Категория | [РОЛЬ] | [OUTPUT SCHEMA] | Rationale | Evidence |
|---|-----------|--------|-----------------|-----------|----------|
| 01 | Дизайн кейсов | test-case-generator | список R + матрица | step-back стратегии | цитаты из спеки |
| 02 | Unit | инженер-автотестер | код + таблица покрытия | few-shot B2 | лог прогона |
| 03 | Integration | инженер-автотестер | сетап + изоляция | проверка окружения | healthcheck + cleanup |
| 04 | API | инженер-автотестер | код + матрица классов | чек-набор A9 | 401/403 + схема |
| 05 | E2E/UI | инженер-автотестер UI | код + таблица состояний | видимый результат | role/label, состояния |
| 06 | Ревью | test-code-reviewer | вердикт §4 | инверсия условия | CONFIRMED/PLAUSIBLE |
| 07 | Дебаг | test-debugger | Debug Report §5 | triage до фикса | ROOT= + воспроизведение |
| 08 | Данные/моки | инженер-автотестер | код фабрики + пример | валидные дефолты | валидация объекта |
| 09 | Покрытие | инженер по качеству | отчёт + приоритеты | риск × изменения | команда покрытия |
| 10 | CI | DevOps/CI-инженер | стадии + rationale | дешёвые гейты первыми | команды из context |
| 11 | Perf/Security | perf/security-тестер | сценарий/отчёт | SLO как критерий | аналитика нагрузки |
| 12 | Документация | test-documenter | док по формату | факты, не планы | сверка с кодом/CI |

**Общее для всех:** `[CONTEXT PROFILE]` включает `context/projects/<profile>/` + `patterns/`; `[УСЛОВИЯ]` — верификация `rules/40` + компактный блок Verification; каждая находка/рекомендация — с готовым исправлением и ссылкой на канон.
