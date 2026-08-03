# Проверки: настоящие и фиктивные

Тест — утверждение о поведении, а не факт запуска кода. Проверка, которая
не может не пройти, даёт покрытие и не даёт ничего больше.

## 1. Тавтология

Плохо:

```python
def test_user_created(api):
    user = api.create_user(name="Иван")
    assert user.name == user.name
```

Хорошо:

```python
def test_user_created_with_given_name(api):
    user = api.create_user(name="Иван")
    assert user.name == "Иван"
```

**Что сломается:** ничего. Плохой вариант зелёный при любом имени, включая
пустое, включая затёртое сервером. Ошибка выглядит безобидно, потому что
слева и справа стоят похожие выражения.

Тот же класс — сравнение с только что присвоенным значением:
`expected = resp["id"]; assert resp["id"] == expected`.

## 2. Проверка существования как единственная

Плохо:

```python
def test_search_returns_results(api):
    results = api.search("телефон")
    assert results is not None
    assert len(results) > 0
```

Хорошо:

```python
def test_search_returns_matching_items_only(api, seeded_catalog):
    results = api.search("телефон")

    assert [r["id"] for r in results] == [seeded_catalog.phone.id]
    assert seeded_catalog.laptop.id not in [r["id"] for r in results]
```

**Что сломается:** поиск, возвращающий весь каталог, — плохой вариант
зелёный. Поиск, возвращающий чужие товары, — тоже зелёный. Проверено, что
что-то вернулось, а не что вернулось верное.

## 3. Только код ответа

Плохо:

```python
def test_get_profile(api, user):
    response = api.get("/profile", user=user)
    assert response.status_code == 200
```

Хорошо:

```python
def test_get_profile_returns_own_data(api, user, other_user):
    response = api.get("/profile", user=user)

    assert response.status_code == 200
    body = response.json()
    assert body["id"] == user.id
    assert body["email"] == user.email
    assert "password_hash" not in body
```

**Что сломается:** ручка, вернувшая чужой профиль, — 200. Ручка, вернувшая
пустой объект, — 200. Ручка, отдающая хеш пароля наружу, — тоже 200.
Последняя строка проверяет, чего быть **не должно**: утечка не ловится
проверкой «нужное есть».

## 4. Копия реализации в тесте

Плохо:

```python
def test_discount_applied(api):
    order = api.create_order(amount=1000, promo="SALE20")
    expected = 1000 - (1000 * 20 / 100)
    assert order.total == expected
```

Хорошо:

```python
def test_discount_20_percent_gives_800_from_1000(api):
    order = api.create_order(amount=1000, promo="SALE20")
    assert order.total == 800
```

**Что сломается:** если в продукте перепутали порядок операций и скидка
считается неверно, тест повторит **ту же** ошибку и останется зелёным.
Тест, вычисляющий ожидаемое тем же алгоритмом, проверяет сам себя.

Ожидаемое значение берётся из требования и пишется константой.

## 5. Проглоченное исключение

Плохо:

```python
def test_invalid_payload_rejected(api):
    try:
        api.post("/orders", json={"amount": "не число"})
    except ValidationError:
        pass
```

Хорошо:

```python
def test_invalid_amount_type_returns_400_with_field(api):
    response = api.post("/orders", json={"amount": "не число"})

    assert response.status_code == 400
    assert response.json()["errors"][0]["field"] == "amount"
```

**Что сломается:** плохой вариант зелёный и когда исключение брошено, и
когда его не было вовсе. Пустой обработчик превращает падение в успех, а
отсутствие ожидаемой ошибки — тоже в успех.

## 6. Избыточная проверка

Плохо:

```python
def test_order_created(api):
    order = api.create_order(amount=100)

    assert order.id is not None
    assert order.amount == 100
    assert order.created_at is not None
    assert order.updated_at is not None
    assert order.status == "new"
    assert order.currency == "RUB"
    assert order.version == 1
    assert order.customer_id is not None
```

Хорошо:

```python
def test_new_order_starts_in_new_status(api):
    order = api.create_order(amount=100)
    assert order.status == "new"
```

**Что сломается:** ничего не сломается — сломается тест. Добавили в модель
поле со значением по умолчанию, изменили формат `updated_at` — падает
проверка, не имеющая отношения к проверяемому требованию. Такой тест
начинают чинить механически, и однажды починят вместе с настоящим багом.

Одна причина падения на тест. Остальное — отдельными тестами, если это
отдельные требования.

## Быстрая проверка себя

Перед сдачей теста ответь одним предложением: **какое изменение поведения
системы обязано его уронить?**

Не находится ответа — проверки нет, сколько бы строк ни было. Ответ есть —
подтверди его прогоном по
[`rules/40`](../rules/40-verification-protocol.md) §Доказательство
осмысленности утверждения.
