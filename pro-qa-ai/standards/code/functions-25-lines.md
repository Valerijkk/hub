# Функция до 25 logical lines — review budget

Цель числа — вовремя заметить смешение обязанностей, а не выиграть метрику.
Обычная функция или метод имеет review budget **25 logical lines**. Превышение
само по себе не делает CI красным: оно просит проверить cohesion, nesting,
change reasons и side effects.

## Когда применять

К production-like коду тестового фреймворка: clients, fixtures, builders,
parsers, validators, packagers, Page/Component Objects и orchestration flows.

## Когда НЕ применять буквально

- не дробить последовательный DAMP-тест на helpers, скрывающие смысл;
- не выносить однократные выражения только ради строки 25;
- декларативная таблица случаев не равна алгоритмической сложности;
- generated/vendor code проверяется отдельной policy либо исключается.

## Что считается

Logical line — исполняемый statement в теле функции. Не считаются signature,
decorators, blank lines, docstring и comment-only lines. Продолжение одного
выражения считается одним statement. Вложенная функция считается частью
владельца и отдельно — это предотвращает обход правила.

## BAD

```python
def create_paid_order(api, user, amount):
    # внутри: login, fixture data, create, retry, payment, polling, assertions,
    # attachments и cleanup — всё одним блоком на 60 строк
    ...
```

Смена payment retry затрагивает подготовку и cleanup; повторно использовать
создание без оплаты нельзя; падение имеет много независимых причин.

## GOOD

```python
order = orders.create(user=user, amount=amount)
payment = payments.pay(order.id)
paid = orders.wait_for_status(order.id, "paid", timeout=timeouts.payment)
assert payment.accepted
assert paid.status == "paid"
```

Каждая механика имеет один контракт, orchestration читается как сценарий,
бизнес-проверка остаётся на виду.

## Что делать при превышении

Сначала собрать контекст: logical lines, максимальную вложенность, независимые
ветки, категории side effects и затронутые зависимости. Если функция остаётся
линейной и cohesive, её сохраняют целиком и при необходимости записывают
короткое rationale:

```yaml
path: scripts/example.py
symbol: Parser.parse_document
measured_logical_lines: 31
rationale: one grammar production; splitting duplicates state transitions
risk: parser change has wide blast radius
owner: automation-quality
review_by: 2026-10-01
```

Такая запись делает решение проверяемым, но не превращает число 25 в закон.
Ошибка допустима только для более сильного составного sentinel — например,
когда превышение сочетается с глубокой вложенностью и несколькими несвязанными
категориями side effects. До локальной калибровки даже он остаётся warning.

## Почему легче изменить

Бюджет направляет review к cohesion, nesting и explicit dependencies. Он не
заменяет complexity: функция в 12 строк с шестью ветками хуже линейной функции
в 27 строк.

## Trade-offs

25 — founder review budget. Внешние источники поддерживают малые focused units,
ясность и single behavior, но не доказывают универсальность именно числа 25.
Наблюдательная работа на Java/SLOC также не доказывает универсальный предел для
Python/Go/TypeScript. Поэтому одиночный line counter не является quality gate.

## Review checks

- Сколько независимых причин изменить функцию?
- Nesting можно заменить guard clauses?
- Side effect отделён от pure decision?
- Extraction имеет имя-контракт, а не `do_part_2`?
- Тест после extraction остаётся локально понятным?
- Для осознанного превышения записаны rationale, owner и срок пересмотра?

## Primary sources

- `SRC-MICROSOFT-UNIT-001`: один Act и один поведенческий фокус теста.
- `SRC-GOOGLE-GOOD-TEST-001`: clarity, completeness и conciseness.
- Число 25: `PROJECT-POLICY`, founder review budget 2026-08-07; warning, не
  универсальный научный закон и не самостоятельный CI blocker.
