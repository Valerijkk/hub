# Тестовые данные, фикстуры, фабрики

Данные — половина теста. Плохие данные делают зелёным тест, который
ничего не проверяет, и красным тест, с которым всё в порядке.

## 1. Данные, не относящиеся к проверке

Плохо:

```python
def test_discount_for_premium(api):
    user = api.create_user(
        name="Иван Петров",
        email="ivan@example.com",
        phone="+70000000000",
        address="Москва, ул. Ленина, 1",
        birthday="1990-01-01",
        tier="premium",
    )
    assert api.calc_discount(user, amount=1000) == 100
```

Хорошо:

```python
def test_discount_for_premium(api, user_factory):
    user = user_factory(tier="premium")
    assert api.calc_discount(user, amount=1000) == 100
```

**Что сломается:** ничего, но читатель теста не понимает, что важно.
Через полгода кто-то поменяет город и будет гадать, влияет ли это на
скидку. В тесте видно **только то, что относится к проверке**, остальное
даёт фабрика значениями по умолчанию.

## 2. Магические значения без смысла

Плохо:

```python
def test_bulk_limit(api):
    items = [{"id": i} for i in range(101)]
    assert api.bulk_create(items).status_code == 400
```

Хорошо:

```python
MAX_BULK_ITEMS = 100

def test_bulk_above_limit_rejected(api):
    items = [{"id": i} for i in range(MAX_BULK_ITEMS + 1)]
    assert api.bulk_create(items).status_code == 400
```

**Что сломается:** лимит поменяли на 200 — тест падает, и непонятно,
это баг или устаревшее ожидание. Число 101 не говорит, что оно означает
«на единицу больше предела».

Ещё лучше — брать предел из того же источника, что и продукт, если он
доступен: тогда изменение лимита не требует правки теста.

## 3. Фикстура, делающая слишком много

Плохо:

```python
@pytest.fixture
def prepared_env(api):
    user = api.create_user()
    api.verify_email(user)
    api.add_payment_method(user)
    order = api.create_order(user)
    api.pay(order)
    api.ship(order)
    return order
```

Хорошо:

```python
@pytest.fixture
def paid_order(api, verified_user):
    order = api.create_order(verified_user)
    api.pay(order)
    return order

@pytest.fixture
def shipped_order(api, paid_order):
    api.ship(paid_order)
    return paid_order
```

**Что сломается:** тесту на оплату не нужна отправка, но он её ждёт —
прогон медленнее, а падение отправки роняет тест про оплату. Диагноз
получается неверный: причина в шаге, к проверке не относящемся.

Фикстуры строятся лесенкой, каждая доводит состояние ровно до нужного.

## 4. Общий набор данных на весь прогон

Плохо:

```python
# conftest.py
@pytest.fixture(scope="session", autouse=True)
def seed_database(db):
    db.execute(open("fixtures/full_dump.sql").read())
```

Хорошо:

```python
@pytest.fixture
def catalog_with_two_items(api):
    phone = api.create_item(name="Телефон", category="electronics")
    book = api.create_item(name="Книга", category="media")
    yield SimpleNamespace(phone=phone, book=book)
    api.delete_items([phone.id, book.id])
```

**Что сломается:** большой общий дамп создаёт невидимые связи. Тест
проходит, потому что в дампе случайно есть подходящая запись; удалили её
— упало пять тестов, и связь неочевидна. Плюс любой тест, пишущий в
общие данные, ломает соседей.

Данные создаются под проверку и удаляются после.

## 5. Реальные данные в фикстурах

Плохо:

```python
@pytest.fixture
def customer(api):
    return api.create_user(
        email="a.ivanova@realcompany.ru",
        phone="+79161234567",
        passport="4510 123456",
    )
```

Хорошо:

```python
@pytest.fixture
def customer(api, unique_email):
    return api.create_user(
        email=unique_email,
        phone="+7000" + str(random.randint(1000000, 9999999)),
        passport="0000 000000",
    )
```

**Что сломается:** тестовое окружение однажды отправит письмо или смс на
реальный адрес. Персональные данные в репозитории — отдельная проблема,
не техническая.

Правило: в фикстурах только синтетические данные из заведомо
недействительных диапазонов.

## 6. Фикстура, зависящая от внешнего состояния

Плохо:

```python
@pytest.fixture
def existing_user(api):
    return api.get_user_by_email("test@example.com")
```

Хорошо:

```python
@pytest.fixture
def existing_user(api, unique_email):
    return api.create_user(email=unique_email)
```

**Что сломается:** пользователь должен уже быть в базе — значит, кто-то
его создал раньше, и это неявная зависимость от порядка либо от ручной
подготовки окружения. На чистой базе всё падает, и причина не в коде.

Фикстура создаёт то, что ей нужно, а не рассчитывает найти.

## Проверка себя

Прочитай тест и назови, **какие данные влияют на результат**. Если в теле
есть значения, про которые нельзя сказать «вот это здесь потому-то» —
они лишние и мешают понимать проверку.
