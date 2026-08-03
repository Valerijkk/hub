# Детерминизм: откуда берутся флейки

Мигающий тест хуже отсутствующего. Отсутствующий честно молчит, мигающий
приучает игнорировать красный.

## 1. Пауза вместо ожидания условия

Плохо:

```python
def test_order_processed(api, order):
    api.submit(order)
    time.sleep(3)
    assert api.get(order.id).status == "processed"
```

Хорошо:

```python
def test_order_processed(api, order, env):
    api.submit(order)
    wait_until(lambda: api.get(order.id).status == "processed",
               timeout=env.processing_timeout)
```

**Что сломается:** машина под нагрузкой обработает за четыре секунды —
тест упадёт, хотя система работает. Обработает за 50 мс — тест всё равно
простоит три секунды, и на наборе из двухсот тестов это десять минут
впустую.

Пауза одновременно и ненадёжна, и медленна. Ожидание условия быстрее на
зелёном пути и стабильнее на медленном.

## 2. Зависимость от текущего времени

Плохо:

```python
def test_report_covers_current_month(api):
    report = api.monthly_report()
    assert report.period == datetime.now().strftime("%Y-%m")
```

Хорошо:

```python
def test_report_covers_requested_month(api, freeze_time):
    freeze_time("2026-03-15")
    report = api.monthly_report()
    assert report.period == "2026-03"
```

**Что сломается:** прогон в 23:59:59 последнего числа месяца. Тест падал
раз в месяц, разбирали каждый раз заново. Отдельный подкласс — тесты,
падающие только в конце года, в високосный день или при переводе часов.

## 3. Случайные данные без зафиксированного зерна

Плохо:

```python
def test_search_handles_any_query(api):
    query = "".join(random.choices(string.ascii_letters, k=10))
    assert api.search(query).status_code == 200
```

Хорошо:

```python
@pytest.mark.parametrize("query", ["abc", "с пробелом", "UNICODE", "кавычка-и-точка-с-запятой"])
def test_search_handles_special_queries(api, query):
    assert api.search(query).status_code == 200
```

**Что сломается:** случайная строка однажды попадёт в неудачный класс —
тест упадёт один раз и больше не воспроизведётся. Отладка невозможна:
входные данные не сохранились.

Случайность допустима только с записанным зерном, которое печатается при
падении. Но чаще нужна не случайность, а **явный список интересных
случаев**.

## 4. Порядок неупорядоченного

Плохо:

```python
def test_list_users(api, seeded_users):
    result = api.list_users()
    assert result == [seeded_users.ivan, seeded_users.petr]
```

Хорошо:

```python
def test_list_users_returns_all_seeded(api, seeded_users):
    result = api.list_users()
    assert {u.id for u in result} == {seeded_users.ivan.id, seeded_users.petr.id}
```

**Что сломается:** база вернёт записи в другом порядке — а она вправе,
если сортировка не задана явно. Тест начнёт мигать после смены версии
СУБД или при другом плане запроса.

Порядок проверяется, только если он **является требованием**. Тогда и
сортировка в запросе должна быть явной.

## 5. Ретрай как маскировка

Плохо:

```python
@pytest.mark.flaky(reruns=3)
def test_notification_sent(api, user):
    api.trigger_event(user)
    assert api.notifications(user)
```

Хорошо:

```python
def test_notification_sent(api, user, env):
    api.trigger_event(user)
    wait_until(lambda: api.notifications(user), timeout=env.async_timeout)
```

**Что сломается:** ретрай прячет настоящую гонку в продукте. Уведомление
теряется в одном случае из четырёх — с тремя перезапусками тест зелёный,
а у пользователей проблема есть.

Ретрай допустим только против доказанно внешней нестабильности, и с
записанной причиной, а не «чтобы стало зелёным».

## 6. Зависимость от локали и часового пояса

Плохо:

```python
def test_price_formatted(api):
    assert api.get_price(1234.5) == "1 234,50 руб."
```

Хорошо:

```python
def test_price_formatted_for_ru_locale(api):
    assert api.get_price(1234.5, locale="ru_RU") == "1 234,50 руб."
```

**Что сломается:** локально проходит, в контейнере CI с локалью по
умолчанию — падает. Причина неочевидна, потому что код не менялся.

Всё, что зависит от окружения — локаль, часовой пояс, кодировка,
разделитель путей — задаётся в тесте явно.

## Проверка себя

Прогони новый тест **десять раз подряд**. Все десять зелёные — базовая
стабильность есть. Хотя бы одно падение — это не «бывает», а найденный
дефект, и он не станет реже от того, что его проигнорировали.
