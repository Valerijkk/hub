# Singleton и глобальное состояние

Default: **mutable singleton в тестовом фреймворке запрещён**. Он прячет
dependency, связывает тесты порядком и ломает параллельный запуск.

## Когда применять

Только для immutable-after-build process metadata или registry, если:

- создание детерминировано;
- зависимость можно инжектировать;
- тест получает отдельный instance или проверяемый reset;
- нет user/session/token/browser/test-data state;
- thread/process ownership и lifetime документированы.

## Когда НЕ применять

Никогда не хранить в singleton: HTTP session с auth, текущего пользователя,
browser context/page, mutable config, test data, transaction, clock/random,
cache, который влияет на oracle.

## BAD

```python
class Session:
    token = None

def login(api, user):
    Session.token = api.login(user)
```

Параллельный тест перезаписывает token; результат зависит от порядка, а
dependency не видна в signature.

## GOOD

```python
@dataclass(frozen=True)
class ApiSession:
    token: str

def create_order(api, session: ApiSession, request: OrderRequest):
    return api.post("/orders", token=session.token, json=request.to_json())
```

State принадлежит сценарию и может безопасно существовать параллельно.

## Почему легче изменить

Lifetime и owner видны из аргументов. Новая auth-схема меняет создание session,
не все вызовы через неявную глобальную переменную.

## Trade-offs

Immutable registry может убрать дорогую повторную загрузку справочника, но
должен строиться до запуска workers и передаваться read-only. Module constant
проще Singleton class. Если нужен reset между тестами, это уже injected
fixture lifetime, а не настоящий singleton.

## Review checks

- Есть module/class mutable state?
- Может ли один тест изменить поведение другого?
- Dependency видна в constructor/function signature?
- Worker/process isolation определён?
- Reset детерминирован и выполняется даже при падении?
- Immutable registry действительно immutable после build?

## Primary sources

- `SRC-GOOGLE-GLOBAL-STATE-001`: скрытые зависимости и тестовые риски global
  state.
- `SRC-PYTEST-FIXTURES-001`: explicit fixture scopes/dependencies и safe
  teardown.
