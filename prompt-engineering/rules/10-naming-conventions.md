# 10 — Дефолтный канон нейминга тестов

> Действует, пока `context/NAMING-CONVENTIONS.md` не заполнен проектным каноном; после заполнения — проектный канон приоритетнее. Эти дефолты — индустриальный мейнстрим, чтобы тулкит работал «из коробки».

## 1. Имена файлов (дефолты по стекам)

| Стек | Unit/Integration | E2E | Хелперы/фикстуры |
|------|------------------|-----|-------------------|
| Go | `<unit>_test.go` рядом или в тест-модуле | — | `helpers_test.go`, `fixtures_test.go`, `factory/` |
| TS/JS (vitest/jest) | `<unit>.test.ts(x)` | `<flow>.spec.ts` (playwright/cypress) | `test-utils.ts`, `__fixtures__/` |
| Python (pytest) | `test_<unit>.py` | `test_<flow>_e2e.py` | `conftest.py`, `factories.py` |
| Java (JUnit) | `<Unit>Test.java` | `<Flow>IT.java` | `<X>TestFixtures.java` |

Общие правила:
- Один тестовый файл ↔ один модуль/компонент кода (best-effort; сценарные сьюты — исключение с говорящим именем).
- Разделители/регистр — как в основном коде стека, не изобретать.
- ⛔ `test2.go`, `new_test_final.spec.ts`, `misc_test.py`, `temp*`.

## 2. Имена тестовых функций/кейсов

Дефолтная структура — **три компонента**: что тестируем / при каких условиях / что ожидаем.

| Раннер | Шаблон | Пример |
|--------|--------|--------|
| Go | `Test<Unit>_<Condition>_<Expected>` | `TestCreateOrder_EmptyCart_ReturnsValidationError` |
| vitest/jest | `describe('<Unit>')` + `it('<condition> → <expected>')` | `it('rejects expired token with 401')` |
| pytest | `test_<unit>__<condition>__<expected>` | `test_create_order__empty_cart__validation_error` |
| Playwright | `test('<actor> <flow> <outcome>')` | `test('user completes checkout with saved card')` |

Правила:
- Имя читается в CI-отчёте без открытия файла и отвечает «что сломалось».
- Язык — английский (дефолт; проект может переопределить).
- ⛔ Информационно-пустые: `TestOK`, `TestFix`, `Test1`, `it('works')`, `test('should work correctly')`, `testHappyPath` без контекста.
- Табличные тесты: имя каждой строки — по той же трёхкомпонентной схеме, коротко.

## 3. TC-ID и привязка

- Дефолтный формат: `TC-<DOMAIN>-<NNN>` (`TC-AUTH-042`).
- Привязка автотеста к TC: комментарий/аннотация у теста (`// TC-AUTH-042`) либо включение ID в имя — канон выбирает проект, но привязка обязательна для тестов, автоматизирующих формализованные TC.
- Требования: `R<число>` внутри одного TC-набора; сквозные — формат трекера проекта.

## 4. Внутренности теста

| Сущность | Дефолт |
|----------|--------|
| System under test | `sut` (или доменное имя: `svc`, `handler`) |
| Ожидание/факт | `want` / `got` (Go), `expected` / `actual` (остальные) |
| Табличная переменная | `tests` / `tt` (Go), массив `cases` (TS) |
| Поле имени кейса | `name` |
| Фабрики | `make<Entity>(overrides)` / `new_test_<entity>` |
| Тестовые константы | `test<What>` / `TEST_<WHAT>` |

## 5. Анти-каталог (имена-запахи)

- Имя с `And`: `TestLoginAndCreateOrderAndLogout` — три теста в одном.
- Имя-реализация: `TestCallsRepositorySaveOnce` — тестирует внутренности, не контракт.
- Имя-враньё: обещает `ReturnsError`, ассертит отсутствие паники.
- Имя-дубль соседа с суффиксом `2`/`New`/`Copy`.
- Имя длиной в твит из-за перечисления всех полей — вынести детали в имя строки таблицы.
