# Слои фреймворка

Четыре роли, и правило одно: **нижняя роль не знает про то, что выше неё**.
Это не приказ создавать четыре каталога для одного прямого вызова.

```
тесты          что проверяем
  ↓
сценарии       бизнес-действия: «зарегистрировать», «оформить заказ»
  ↓
клиенты        как обратиться: API, UI, очередь, база
  ↓
инфраструктура конфиг, http, логи, ожидания
```

- action/page/client object скрывает mechanics;
- flow/scenario оркестрирует несколько бизнес-действий;
- reporting step только подписывает и прикладывает evidence;
- test владеет business oracle.

Если тест остаётся ясным с одним прямым вызовом клиента, отдельный scenario
layer не нужен. Новый слой обязан уменьшать повтор mechanics или давать
стабильный контракт, иначе это ceremony.

## 1. Утверждение живёт в тесте, а не под ним

Плохо:

```python
# helpers/order.py
def create_and_check_order(api, amount):
    order = api.post("/orders", json={"amount": amount})
    assert order.status_code == 201
    assert order.json()["status"] == "new"
    return order.json()

# test_orders.py
def test_order_created(api):
    create_and_check_order(api, 100)
```

Хорошо:

```python
# scenarios/order.py
def create_order(api, amount):
    response = api.post("/orders", json={"amount": amount})
    response.raise_for_status()
    return response.json()

# test_orders.py
def test_new_order_starts_in_new_status(api):
    order = create_order(api, amount=100)
    assert order["status"] == "new"
```

**Что сломается:** в плохом варианте тест не показывает, что проверяет.
При падении читаешь имя, идёшь в хелпер, находишь два утверждения и не
знаешь, какое из них сработало. Плюс хелпер нельзя переиспользовать там,
где ожидается отказ — он всегда требует успеха.

Разделение: `raise_for_status` — это **предусловие** сценария, а не
проверка теста. Оно говорит «дальше идти бессмысленно», а не «требование
выполнено».

## 2. Клиент не знает про сценарий

Плохо:

```python
class ApiClient:
    def create_order(self, amount):
        user = self.login_as_default_user()      # клиент сам логинится
        return self.post("/orders", json={"amount": amount}, user=user)
```

Хорошо:

```python
class ApiClient:
    def post(self, path, *, json=None, token=None):
        ...

# scenarios/order.py
def create_order(api, token, amount):
    return api.post("/orders", json={"amount": amount}, token=token)
```

**Что сломается:** тест на создание заказа от чужого пользователя
написать нельзя — клиент всегда подставляет своего. Клиент, принимающий
решения за тест, ограничивает то, что вообще можно проверить.

## 3. Инфраструктура не знает про домен

Плохо:

```python
# core/http.py
def request(method, path, **kw):
    if path.startswith("/orders"):
        kw.setdefault("timeout", 30)     # заказы медленные
    return session.request(method, path, **kw)
```

Хорошо:

```python
# core/http.py
def request(method, path, *, timeout, **kw):
    return session.request(method, path, timeout=timeout, **kw)

# clients/orders.py
class OrdersClient:
    TIMEOUT = env.orders_timeout
```

**Что сломается:** знание о домене, зашитое в транспорт, невозможно
найти. Тест на заказы стал медленнее — идёшь в клиент заказов, там
ничего нет, а причина в общем http-слое, куда никто не смотрит.

## 4. Зависимость только вниз

Плохо: `core/wait.py` импортирует `clients/orders.py`, чтобы дождаться
статуса заказа.

Хорошо: `core/wait.py` принимает предикат; знание о заказах остаётся в
слое заказов.

**Что сломается:** цикл импортов и невозможность переиспользовать
ожидание в другом домене. Признак нарушения — импорт «снизу вверх»: файл
нижнего слоя тянет что-то из верхнего.

## 5. Прямой доступ мимо слоя

Плохо:

```python
def test_order_visible_in_list(api, db):
    db.execute("INSERT INTO orders (id, amount) VALUES (1, 100)")
    assert api.get("/orders").json()["items"][0]["id"] == 1
```

Хорошо:

```python
def test_order_visible_in_list(api, token):
    created = create_order(api, token, amount=100)
    ids = [o["id"] for o in api.get("/orders", token=token).json()["items"]]
    assert created["id"] in ids
```

**Что сломается:** запись напрямую в базу минует валидацию, вычисляемые
поля, события и кеш. Тест зелёный, а через продуктовый путь такой заказ
не создался бы вовсе — или создался бы иначе.

Прямой доступ к хранилищу допустим для **подготовки того, что не
относится к проверке**, и для очистки. Для создания проверяемого объекта —
нет.

## Проверка себя

Открой файл и спроси: **какому слою он принадлежит и что он импортирует
из соседних**. Импорт снизу вверх, утверждение в хелпере, домен в
транспорте — три признака, что слой размыт.
