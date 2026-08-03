# Негативные классы: что покрыто только успешным путём

Покрыт только успешный путь — задача не закрыта. Ошибки живут в отказах,
а не в успехах: успешный путь проверяют вручную при разработке, отказы —
никогда.

## Обязательные классы

| Класс | Что проверяется | Типичный пропуск |
|---|---|---|
| Авторизация | без токена, с чужим, с истёкшим, с недостаточными правами | проверяют только «свой токен работает» |
| Валидация | нет обязательного поля, неверный тип, выход за границы | проверяют один невалидный случай из десяти |
| Состояние | несуществующий id, удалённый объект, конфликт версий | забывают удалённый — он ведёт себя иначе, чем несуществующий |
| Идемпотентность | повторный одинаковый запрос | не проверяют вовсе |
| Ограничения | превышение лимита, слишком большое тело, частота | вспоминают после инцидента |
| Границы | пусто, максимум, максимум плюс один, минимум минус один | проверяют середину диапазона |

Неприменимый класс помечается `N/A: причина`, а не пропускается молча.
Пропущенный класс неотличим от неприменимого, и через месяц никто не
вспомнит, было это решением или забывчивостью.

## 1. Авторизация: «чужой» важнее «своего»

Плохо:

```python
def test_get_document(api, user, document):
    response = api.get(f"/documents/{document.id}", user=user)
    assert response.status_code == 200
```

Хорошо:

```python
def test_owner_can_read_document(api, user, document):
    assert api.get(f"/documents/{document.id}", user=user).status_code == 200

def test_stranger_cannot_read_document(api, other_user, document):
    assert api.get(f"/documents/{document.id}", user=other_user).status_code == 403

def test_anonymous_cannot_read_document(api, document):
    assert api.get(f"/documents/{document.id}", user=None).status_code == 401
```

**Что сломается:** ручка, отдающая документ кому угодно, проходит первый
тест. Утечка доступа не ловится проверкой «владелец может» — только
проверкой «чужой не может».

Разница между 401 и 403 существенна: первое означает «не знаем, кто ты»,
второе — «знаем, но нельзя». Смешивать их в одном тесте нельзя.

## 2. Состояние: удалённое не равно несуществующему

Плохо:

```python
def test_get_missing_order_returns_404(api):
    assert api.get("/orders/999999").status_code == 404
```

Хорошо:

```python
def test_never_existed_order_returns_404(api):
    assert api.get("/orders/999999").status_code == 404

def test_deleted_order_returns_410(api, deleted_order):
    assert api.get(f"/orders/{deleted_order.id}").status_code == 410
```

**Что сломается:** удалённый объект часто ведёт себя иначе — возвращает
410, отдаёт заголовок с датой удаления, попадает в другой обработчик.
Этот путь не покрывается вообще, потому что «404 же проверили».

Конкретные коды берутся **из контракта**. Нет в контракте 410 — не
выдумывать, а вынести в вопросы.

## 3. Границы: середина ничего не доказывает

Плохо:

```python
def test_username_length(api):
    assert api.create_user(name="Иван").status_code == 201
```

Хорошо:

```python
@pytest.mark.parametrize("length,expected", [
    (1, 201),    # минимум
    (2, 201),    # минимум плюс один
    (49, 201),   # максимум минус один
    (50, 201),   # максимум
    (51, 400),   # максимум плюс один
    (0, 400),    # пусто
])
def test_username_length_boundaries(api, length, expected):
    assert api.create_user(name="и" * length).status_code == expected
```

**Что сломается:** ошибка на единицу — самая частая в проверках длины.
Значение из середины диапазона проходит при любом варианте сравнения:
и при строгом, и при нестрогом, и при сдвинутом на единицу.

Проверяются **обе стороны каждой границы**, а не диапазон целиком.

## 4. Идемпотентность: повтор — не то же самое

Плохо:

```python
def test_payment_created(api, order):
    response = api.pay(order, idempotency_key="abc")
    assert response.status_code == 201
```

Хорошо:

```python
def test_repeated_payment_does_not_charge_twice(api, order):
    first = api.pay(order, idempotency_key="abc")
    second = api.pay(order, idempotency_key="abc")

    assert first.status_code == 201
    assert second.status_code == 200
    assert second.json()["payment_id"] == first.json()["payment_id"]
    assert api.charges_count(order) == 1
```

**Что сломается:** повторный запрос при обрыве связи — обычное дело.
Двойное списание при этом обнаруживают клиенты, а не тесты. Последняя
строка проверяет главное: списание одно, сколько бы запросов ни пришло.

## 5. Валидация: сообщение об ошибке — тоже контракт

Плохо:

```python
def test_invalid_email_rejected(api):
    assert api.register(email="не-почта").status_code == 400
```

Хорошо:

```python
def test_invalid_email_rejected_with_field_name(api):
    response = api.register(email="не-почта")

    assert response.status_code == 400
    error = response.json()["errors"][0]
    assert error["field"] == "email"
    assert error["code"] == "invalid_format"
```

**Что сломается:** 400 вернётся и при неверной почте, и при отсутствии
обязательного поля, и при слишком длинном имени. Тест не отличает
причину отказа. Клиент, показывающий пользователю подсветку неверного
поля, при этом сломан, а тест зелёный.

Проверять поле и код ошибки — если они документированы. Не
документированы — вопрос к контракту, а не догадка.

## Проверка себя

Открой матрицу покрытия. Если в строке требования заполнена только
колонка `happy`, а остальные пусты — требование **не покрыто**,
независимо от того, что показывает отчёт покрытия по строкам кода.
