# E2E — greenfield Python Playwright

> Target: `testframework/e2e/**`, CI `python-e2e-smoke/full`. TypeScript
> Playwright, legacy specs/config/shared helpers — delete-not-port.

## Когда нужен E2E

E2E защищает user-visible journey, который нельзя достаточно доказать unit,
API или integration test: auth transition, project lifecycle, messenger
delivery, TMS flow, critical admin/security denial, cross-domain recovery.
Низкоуровневые permutations остаются в более быстрых suites.

`smoke` и `full` — pytest markers, не разные каталоги. Smoke выбирается по
release risk и критичности, а не по историческому списку TS specs.

## UI abstraction

1. Начать с accessible semantic locator в test.
2. Вынести tiny action helper после повторения meaningful action.
3. Создать Component Object для повторяемого widget.
4. Создать Page Object только для стабильного page knowledge в крупной suite.

Giant `AppPage`, обязательные Steps/POM и assertions внутри generic helpers
скрывают scenario. Specific oracle обычно остаётся в test. POM — option, не
gate. `data-testid` — fallback, когда semantic locator невозможен.

## Determinism и failure model

- web-first/bounded eventual assertions вместо sleep;
- task/run/worker-specific tenant, user, records и artifact path;
- fresh state либо явная idempotent fixture;
- cleanup принадлежит fixture и проверяется;
- retry не маскирует product/test flake;
- unexpected skip и zero collection — FAIL.

Выбирать применимое: happy, invalid, authn, authz/tenant, empty/one/many,
duplicate/reorder, timeout/unavailable/partial/retry/recovery, cancellation,
concurrency, replay/idempotency, resource exhaustion, version compatibility,
a11y/security/privacy. Нет обязательного числа scenarios на journey.

## Python shape

```python
import pytest


@pytest.mark.e2e
@pytest.mark.smoke
@pytest.mark.tc("TC-AUTH-01")
@pytest.mark.requirement("FR-AUTH-01")
@pytest.mark.task("TASK-000")
def test_magic_link_opens_authorized_home(page, user):
    page.get_by_role("link", name="Войти").click()
    assert page.get_by_role("heading", name="Главная").is_visible()
```

Форма markers иллюстративна до scaffold; точную schema задаёт root
`testframework/pyproject.toml`. Case обязан дополнительно фиксировать risk,
observable oracle и cleanup через принятую metadata schema.

## Docker evidence

E2E запускается только через canonical `/qa-gate` outer launcher
и target Compose profile из `testframework/docker/**`. Этот playbook не
дублирует `docker run`/`docker compose` snippets: прямой inner command
не доказывает наличие required services, limits, isolation и cleanup.

Если target E2E image/profile/runner ещё не существует, verdict —
`NOT_READY`, а не legacy TypeScript Playwright fallback. Host
Playwright/pytest/backend/frontend processes запрещены.

Все E2E reports, traces, screenshots и videos пишутся только в
`testframework/reports/<task-run>/`. Перед запуском каталог
получает owner marker с `task`, `run`, `owner`, `source_sha`, `created_at`
и cleanup policy. Shared report root без task-run namespace нельзя
использовать как output.

## Review

- journey и marker выведены из current risk;
- semantic locator first; abstraction имеет evidence повторения;
- assertions доказывают user result и запрещённые side effects;
- state, wait, retry, artifacts и cleanup детерминированы;
- TC/task/requirement links разрешаются;
- temporary sentinel/mutation доказывает чувствительность high-risk oracle.

См. [test design](test-design.md), [tooling](tooling.md) и
[`NEG-03`, `NEG-07`, `NEG-08`](../../knowledge/code-practices/controls/negative-sentinels.md).
