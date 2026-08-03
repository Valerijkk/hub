# 10 — Канон именования тестов по умолчанию

> Действует, пока project profile не заполнит свой naming canon. **Приоритет
> ниже**, чем у project context и approved pattern cards (см. [Приоритет](#приоритет)).

## Приоритет

| Источник | Роль |
|----------|------|
| [`context/projects/<profile>/`](../../context/projects/README.md) — `TEST-ARCHITECTURE.md`, `TRACEABILITY.md`, `PATTERN-INDEX.md`, [`SOURCE-MAP.md`](../../context/templates/SOURCE-MAP.md) | Конвенции проекта и evidence |
| [`patterns/approved/`](../patterns/approved/README.md) | Human-approved паттерны именования с provenance |
| Этот файл (`rules/10`) | Переносимые defaults, когда project canon отсутствует или `template` |

Когда project canon и эти defaults расходятся, **побеждает project canon**.
Отклонение только от `rules/10` само по себе не дефект, пока оно не нарушает
конвенцию проекта или approved card.

Общие стилевые предпочтения (язык, разделитель, describe/it vs плоские имена) —
это **рекомендации**, а не findings ревью, пока проект явно не зафиксировал их
в profile docs или approved card.

## Когда именование — finding

Публиковать naming finding только когда выполнены **все** условия:

1. **Указана конвенция** — ссылка на правило проекта, approved card или
   повторяющийся in-repo паттерн, зафиксированный в `SOURCE-MAP` /
   `PATTERN-INDEX` (по правилам provenance в [`rules/50`](50-data-handling.md)).
2. **Сформулирован impact** — как имя мешает triage, трассируемости, поиску
   дубликатов или искажает поведение (например, имя обещает `_ReturnsError`, а
   ассерт проверяет успех).
3. **Severity** — по [`rules/30`](30-severity-policy.md), не по вкусу.

Эвристический grep по `test1`, `misc_`, или `temp*` даёт **candidate**, пока
семантически не подтверждены конвенция проекта и impact
([`rules/40`](40-verification-protocol.md), этап 3).

## 1. Имена файлов (defaults по стеку)

| Стек | Unit / integration | E2E | Helpers / fixtures |
|------|-------------------|-----|-------------------|
| Go | `<unit>_test.go` рядом или в test module | — | `helpers_test.go`, `fixtures_test.go`, `factory/` |
| TS/JS (vitest/jest) | `<unit>.test.ts(x)` | `<flow>.spec.ts` (Playwright/Cypress) | `test-utils.ts`, `__fixtures__/` |
| Python (pytest) | `test_<unit>.py` | `test_<flow>_e2e.py` | `conftest.py`, `factories.py` |
| Java (JUnit) | `<Unit>Test.java` | `<Flow>IT.java` | `<X>TestFixtures.java` |

Общее:

- Один тестовый файл соответствует одному модулю или компоненту (сюит сценариев
  может использовать описательное групповое имя).
- Разделители и регистр — как в основном кодебазе этого стека.
- Избегать непрозрачных имён: `test2.go`, `new_test_final.spec.ts`,
  `misc_test.py`, `temp*`.

## 2. Имена тестовых функций / кейсов

Структура по умолчанию — **три части**: субъект / условие / ожидаемый исход.

| Runner | Шаблон | Пример |
|--------|--------|--------|
| Go | `Test<Unit>_<Condition>_<Expected>` | `TestCreateOrder_EmptyCart_ReturnsValidationError` |
| vitest/jest | `describe('<Unit>')` + `it('<condition> → <expected>')` | `it('rejects expired token with 401')` |
| pytest | `test_<unit>__<condition>__<expected>` | `test_create_order__empty_cart__validation_error` |
| Playwright | `test('<actor> <flow> <outcome>')` | `test('user completes checkout with saved card')` |

Правила:

- Имя должно отвечать «что сломалось?» в выводе CI без открытия файла.
- Язык по умолчанию — English; project profile может переопределить.
- Избегать пустых имён: `TestOK`, `TestFix`, `Test1`, `it('works')`,
  `test('should work correctly')`, голый `testHappyPath`.
- Подкейсы таблиц: имя каждой строки по той же трёхчастной схеме, коротко.

## 3. TC-ID и трассируемость

- Формат ID по умолчанию: `TC-<DOMAIN>-<NNN>` (пример: `TC-AUTH-042`).
- Связывать автотесты с формальными TC через comment, annotation или встроенный
  ID — точная форма выбирается в
  [`TRACEABILITY.md`](../../context/templates/TRACEABILITY.md) активного profile.
- ID требований (`R<number>`) следуют формату трекера проекта в том же profile.

## 4. Идентификаторы внутри теста

| Сущность | Default |
|----------|---------|
| System under test | `sut` или доменное имя (`svc`, `handler`) |
| Expected / actual | `want` / `got` (Go), `expected` / `actual` (остальные) |
| Table cases | `tests` / `tt` (Go), `cases` (TS) |
| Поле метки кейса | `name` |
| Factories | `make<Entity>(overrides)` / `new_test_<entity>` |
| Тестовые константы | `test<What>` / `TEST_<WHAT>` |

## 5. Анти-каталог (smell-имена)

Использовать как подсказки ревью после семантического чтения — не как
автоматические findings:

- Цепочки `And`: `TestLoginAndCreateOrderAndLogout` — вероятно несколько
  поведений в одном тесте.
- Имя реализации: `TestCallsRepositorySaveOnce` — тестирует internals, не
  контракт.
- Вводящее в заблуждение имя: обещает `ReturnsError`, а проверяет только
  отсутствие panic.
- Дубликат соседа с суффиксом `2` / `New` / `Copy`.
- Имя «длиной в твит» со всеми полями — детали переносить в table row `name`.

## 6. Правило выводится из кода, а не объявляется

Конвенция именования этого репозитория — **наблюдаемый факт**, а не
предпочтение. Порядок установления:

```
1. Собрать фактические имена тестов и файлов по всему дереву.
2. Вывести доминирующую форму.
3. Посчитать долю соответствия и записать её.
4. Перечислить зоны-исключения с причиной.
```

Запись выглядит так:

```
Правило:       test_<объект>_<условие>_<ожидание>
Соответствие:  84% (211 из 251)
Исключения:    legacy/ — старый стиль, миграция не планируется
               contract/ — имена генерируются из спеки
```

**Честная доля важнее красивого правила.** Конвенция, заявленная как
стопроцентная при фактических 60%, приводит к тому, что каждое второе имя
становится находкой — и правило отключают целиком вместо того, чтобы
довести долю.

Доля пересчитывается при обновлении контекста. Её падение — сигнал, что
конвенция размывается либо что появилась вторая, неучтённая.
