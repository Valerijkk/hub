# Изоляция: независимость, состояние, очистка

Тест обязан проходить в одиночку, в любом порядке и параллельно с
другими. Зависимость от порядка — дефект теста, а не особенность набора.

## 1. Общее изменяемое состояние

Плохо:

```python
@pytest.fixture(scope="module")
def cart(api, user):
    return api.create_cart(user)

def test_add_item(api, cart):
    api.cart_add(cart, item_id=1)
    assert len(api.cart_get(cart)["items"]) == 1

def test_remove_item(api, cart):
    api.cart_remove(cart, item_id=1)
    assert api.cart_get(cart)["items"] == []
```

Хорошо:

```python
@pytest.fixture
def cart(api, user):
    created = api.create_cart(user)
    yield created
    api.cart_delete(created)
```

**Что сломается:** второй тест зависит от того, что первый уже положил
товар. Запусти его отдельно — упадёт. Запусти параллельно — упадёт
случайным образом. Порядок сейчас удачен случайно, и это самая дорогая
разновидность зелёного.

Правило границы: **дорогое создание переиспользовать можно, изменяемое
состояние — нет.** Нужна общая фикстура ради скорости — чисти состояние
после каждого теста.

## 2. Очистка, которая не выполнится

Плохо:

```python
def test_upload_and_process(api, storage):
    file_id = api.upload("report.csv")

    result = api.process(file_id)
    assert result.rows == 100

    storage.delete(file_id)
```

Хорошо:

```python
@pytest.fixture
def uploaded_file(api, storage):
    file_id = api.upload("report.csv")
    yield file_id
    storage.delete(file_id)

def test_processing_counts_all_rows(api, uploaded_file):
    result = api.process(uploaded_file)
    assert result.rows == 100
```

**Что сломается:** упадёт проверка — очистка не выполнится, файл
останется. Через сто прогонов хранилище забито, а причина неочевидна.
Очистка последней строкой тела работает только на зелёном пути, то есть
тогда, когда она нужна меньше всего.

Как это устроено в коде фикстуры — [`../code/fixtures.md`](../code/fixtures.md) §4.

## 3. Порядок как скрытая зависимость

Плохо:

```python
def test_1_create_user(api):
    global created_user
    created_user = api.create_user()
    assert created_user.id

def test_2_update_user(api):
    api.update_user(created_user.id, name="Новое")
    assert api.get_user(created_user.id).name == "Новое"
```

Хорошо:

```python
def test_user_name_updated(api, user):
    api.update_user(user.id, name="Новое")
    assert api.get_user(user.id).name == "Новое"
```

**Что сломается:** первый тест отфильтровали по маркеру — второй падает с
ошибкой неопределённого имени, и она не имеет отношения к продукту.
Нумерация в именах (`test_1_`, `test_2_`) — признак того, что зависимость
от порядка осознавали и закрепили вместо того, чтобы убрать.

Если сценарий действительно последовательный — это **один** тест с
несколькими шагами, а не несколько тестов.

## 4. Общий внешний ресурс

Плохо:

```python
def test_export_creates_file(api):
    api.export_report(path="/tmp/report.csv")
    assert os.path.exists("/tmp/report.csv")
```

Хорошо:

```python
def test_export_creates_file(api, tmp_path):
    target = tmp_path / "report.csv"
    api.export_report(path=str(target))
    assert target.exists()
```

**Что сломается:** два теста, пишущих в один путь, при параллельном
прогоне затирают файл друг друга. Хуже: тест проходит, потому что файл
остался от **предыдущего** прогона, а текущий экспорт сломан.

Тот же класс — фиксированные порты, общие имена очередей, общие ключи
кэша, записи в базе с захардкоженным идентификатором.

## 5. Данные, конфликтующие по уникальному полю

Плохо:

```python
def test_register(api):
    user = api.register(email="test@example.com")
    assert user.id
```

Хорошо:

```python
def test_register_returns_created_user(api, unique_email):
    user = api.register(email=unique_email)
    assert user.email == unique_email
```

**Что сломается:** второй прогон падает на уникальности почты, потому что
первый её занял. Тест «работает один раз» — на локальной машине это
замечают быстро, в CI с чистой базой не замечают никогда.

Генератор уникальных значений — [`../code/fixtures.md`](../code/fixtures.md) §5.

## Проверка себя

Три вопроса перед сдачей:

1. Пройдёт ли тест, запущенный **в одиночку**?
2. Пройдёт ли он **дважды подряд** на том же окружении?
3. Пройдёт ли он **параллельно** с соседями по файлу?

Хотя бы одно «не уверен» — изоляция не сделана. Первые два проверяются
прогоном за секунды, третий — прогоном набора с параллелизмом.
