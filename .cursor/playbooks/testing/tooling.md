# Инструментарий QA: пакеты, артефакты и доказательства

> Часть плейбука тестирования. Канон выбора уровней и качества тестов —
> [`../testing.md`](../testing.md). Этот документ фиксирует только целевую
> структуру инструментария и правила доказательств; он не является инструкцией
> для запуска проверок.

## 1. Каноническая топология

Non-unit QA Hub — единый будущий Python-проект в `testframework/`.
Переиспользуемый код находится исключительно в устанавливаемом пакете
`testframework/src/hub_testkit/**`; исполняемые pytest-наборы, проверяющие SUT,
находятся исключительно в `testframework/tests/{contract,api,integration,e2e}/**`.
Тесты импортируют `hub_testkit` как установленный пакет, а не получают доступ к
исходникам через правки `sys.path`, относительные обходы дерева или копирование
fixture-кода между suites.

Отдельные корни сохраняют разные обязанности и не смешиваются с пакетом или
исполняемыми SUT suites:

| Путь | Назначение |
| --- | --- |
| `testframework/selftests/` | проверки самого QA-инструментария |
| `testframework/cases/` | человекочитаемые кейсы и каталоги |
| `testframework/schemas/` | схемы контрактов |
| `testframework/policies/` | политики и правила исполнения |
| `testframework/load/` | профили Locust |
| `testframework/reports/` | task-owned результаты запусков |

`smoke` — pytest marker, а `full` — профиль будущего launcher; это не
взаимозаменяемые названия каталогов или старых раннеров.

## 2. Статус исполнения

Метаданные пакета (`testframework/pyproject.toml`, `uv.lock` и
`src/hub_testkit/py.typed`) и versioned outer launcher ещё не реализованы.
До отдельной реализации и независимой проверки launcher имеет статус
`NOT_READY`, а runtime — `NOT_RUN`.

Единственным будущим владельцем outer-launcher API является
`testframework/runner/**`. Он будет отвечать за task-owned имена, лимиты
CPU/RAM/PID, wall-clock timeout, изолированные артефакты
`testframework/reports/<task-run>/` с owner marker, сохранение exit code и
безусловный cleanup exact resources в `finally`.

`testframework/docker/**` содержит только immutable execution assets и их
контракты. В нём не появляется launcher-код, альтернативная точка входа или
замена будущему outer launcher. До готовности launcher не допускаются fallback
на прямые команды среды, host execution, direct Docker/Compose или legacy
runner.

## 3. Allure и TMS

Allure — формат переносимых результатов и метаданных; его использование не
заменяет поведенческий oracle, traceability или доказательство чувствительности
проверки. Когда будущий suite формирует Allure-результат для TMS, запись должна
содержать стабильную связь с тест-кейсом Hub: `AS_ID` label или tms-link с
ключом `HUB-TC-*`.

Минимально полезная структура результата включает имя и статус проверки,
duration, вложенные steps, labels, links и безопасные attachments. Связь по
явному идентификатору приоритетнее эвристик по имени класса, методу или
`fullName`, поскольку эти имена меняются при рефакторинге.

Метаданные Allure применяются там, где они помогают расследованию:

- epic, feature и story описывают навигацию по предметной области;
- severity отражает риск, а не декоративную важность;
- steps показывают существенные действия сценария;
- attachment содержит только нужную для воспроизведения диагностику;
- tms-link или `AS_ID` связывает результат с конкретным TC.

Статусы результата различают проверочное падение (`failed`), проблему среды или
исполнения (`broken`) и разрешённый skip (`skipped`). Неизвестный статус не
трактуется как успешный. Любая будущая интеграция с импортом TMS обязана
сохранять требования к аутентификации, RBAC и отсутствию секретов в логах.

## 4. Артефакты и конфиденциальность

Артефакты изолируются по task/run/worker и не являются исходным кодом. В них
могут входить sanitised request/response diagnostics, идентификаторы fixture,
trace, screenshot, video, логи, результаты Allure и сводка времени. Все пути
артефактов принадлежат `testframework/reports/<task-run>/`; shared mutable
каталоги, порты, записи и аккаунты запрещены.

Никогда не помещайте в attachment, trace или лог JWT, пароль, magic-link code,
API key, webhook secret, PII или неотредактированный payload. Отладочный вывод
должен содержать только данные, достаточные для воспроизведения: TC ID, request
ID, sanitised endpoint, identity fixture и timing. Политика redaction относится
к отчётам, скриншотам, видео и трассам одинаково.

Coverage-профили, временные файлы и диагностические результаты не коммитятся.
Cleanup артефактов — часть будущего launcher-контракта, а не ручной постскриптум
после запуска.

## 5. Время и производительность

Будущая запись доказательства фиксирует wall-clock duration и достаточно данных
для поиска медленных сценариев без раскрытия секретов. Наблюдаемыми метриками
остаются длительность suite и теста, результат collection, число passed/failed/
skipped, выделенные ресурсы, сервисный профиль, версия образа или digest,
расположение артефактов и итог cleanup.

Сравнение времени полезно только при сопоставимых версии SUT, образе,
ресурсах, профиле сервисов и нагрузке. Рост длительности сам по себе не
доказывает регрессию: его исследуют вместе с изменением сценария, количеством
данных, конкуренцией и состоянием зависимостей.

Нагрузочные профили принадлежат `testframework/load/` и отделены от pytest
suites. Они получают собственную traceability, изоляцию, лимиты ресурсов и
отчётные артефакты.

## 6. Минимальное runnable evidence

После появления и независимой проверки launcher доказательство каждого
значимого запуска должно быть воспроизводимым и включать не менее следующего:

```yaml
job: <target job>
source_sha: <source revision>
image: <immutable tag or digest>
launcher: <versioned outer-launcher API>
inner_command: <recorded command>
resources: { cpus: <limit>, memory: <limit>, pids: <limit> }
service_profile: <none or profile>
result: { collected: <N>, passed: <N>, failed: <N>, skipped: <N> }
duration: <seconds>
artifacts: testframework/reports/<task-run>/
cleanup: <recorded result>
```

Такое evidence не заменяется наличием файлов, raw grep, покрытием, зелёным
mock-only assert, историческим результатом или self-report. Для применимого
высокого риска требуется отрицательный sentinel/mutation, который показывает,
что защищаемое поведение действительно чувствительно к ошибке.

## 7. Чек-лист автора

- [ ] Переиспользуемый Python-код находится под `src/hub_testkit/**`, а suite —
  под соответствующим `tests/{contract,api,integration,e2e}/**`.
- [ ] `selftests`, `cases`, `schemas`, `policies`, `load` и `reports` не
  используются как замена package или SUT suite.
- [ ] TC, requirement и observable oracle однозначно связаны с проверкой.
- [ ] Диагностика не содержит секретов или PII и имеет task/run/worker namespace.
- [ ] Для runtime нет заявлений до состояния launcher `NOT_READY` → отдельной
  реализации → независимого evidence; текущий runtime остаётся `NOT_RUN`.
- [ ] Артефакты, ресурсы и cleanup фиксируются только через будущий
  `testframework/runner/**`; альтернативный раннер не вводится.

Связанные документы: [`../testing.md`](../testing.md),
[`../../skills/qa-gate/SKILL.md`](../../skills/qa-gate/SKILL.md),
[`../../../testframework/PACKAGE-CONTRACT.md`](../../../testframework/PACKAGE-CONTRACT.md).
