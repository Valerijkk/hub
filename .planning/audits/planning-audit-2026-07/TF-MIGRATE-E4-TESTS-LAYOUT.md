# E4 — `/tests` Python project layout (PLAN ONLY)

**Дата:** 2026-07-27  
**Источник:** [`02-TEST-FRAMEWORK-MIGRATION.md`](02-TEST-FRAMEWORK-MIGRATION.md) §Э1, целевая структура  
**Статус:** design — **файлы `/tests/**` не создавать в этой задаче**

---

## Цель

Спроектировать корневой Python-проект для интеграционных, E2E и нагрузочных тестов так, чтобы Э1 (scaffold) реализовывался по чеклисту без архитектурных развилок.

---

## Целевое дерево (план)

```
tests/                              # новый корневой Python-проект
├── pyproject.toml                  # pytest, playwright, locust, httpx, allure-pytest
├── README.md                       # как поднять Docker + pytest + locust
├── conftest.py                     # корневые фикстуры: env, DB, auth clients
├── pytest.ini                      # optional; можно всё в pyproject [tool.pytest.ini_options]
│
├── fixtures/                       # данные и фабрики (не pytest plugin alone)
│   ├── __init__.py
│   ├── auth.py                     # DEV_EMAIL, ADMIN_EMAIL, seed_dev, login, session payload
│   ├── env.py                      # API_URL, FRONTEND_URL, reachability
│   ├── db.py                       # Postgres/Mongo/Redis handles (integration only)
│   ├── factories.py                # project/issue/channel builders via API
│   └── seed.py                     # POST /api/dev/seed wrappers
│
├── integration/                    # pytest + httpx (и при нужде async)
│   ├── conftest.py                 # markers: integration, requires_docker
│   ├── auth/
│   ├── tracker/
│   ├── messenger/
│   ├── tms/
│   ├── search/
│   ├── notification/
│   ├── bases/
│   └── platform/                   # health, metrics, webhooks
│
├── e2e/                            # Playwright Python
│   ├── conftest.py                 # page, browser, skip_* fixtures
│   ├── pages/                      # POM (port from stages/e2e/shared/pages)
│   ├── steps/
│   ├── helpers/                    # a11y, ai
│   └── domains/
│       ├── auth/
│       ├── platform/               # smoke: login/a11y/critical-gate живут здесь
│       ├── admin/
│       ├── ai/
│       ├── messenger/
│       ├── navigation/
│       ├── tms/
│       ├── tracker/
│       └── v4/
│
└── load/                           # Locust
    ├── locustfile.py               # entry (или пакет)
    ├── users.py                    # HttpUser subclasses
    ├── scenarios/                  # см. TF-MIGRATE-E5-LOCUST.md
    └── data/                       # credentials / project ids for load env
```

**Не входит в `/tests`:** Go/TS unit-тесты (живут рядом с кодом после Э3/Э4).

---

## `pyproject.toml` (контракт зависимостей)

| Пакет | Зачем |
|-------|--------|
| `pytest` ≥8 | runner |
| `pytest-playwright` / `playwright` | E2E |
| `httpx` | API integration + auth seed |
| `locust` | load |
| `allure-pytest` | отчёты (паритет с Testing Guide) |
| `psycopg[binary]` | optional DB asserts в integration |
| `redis` / `pymongo` | optional integration probes |
| `axe-playwright-python` или inject axe | a11y port |

Минимальные tool-настройки:

```toml
[project]
name = "hub-tests"
version = "0.1.0"
requires-python = ">=3.12"

[project.optional-dependencies]
dev = ["pytest", "playwright", "pytest-playwright", "httpx", "locust", "allure-pytest"]

[tool.pytest.ini_options]
testpaths = ["integration", "e2e"]
markers = [
  "integration: API/DB integration",
  "e2e: browser",
  "smoke: CI smoke subset",
  "load: locust scenarios (not default pytest)",
]
addopts = "-ra --strict-markers"
```

Locust **не** гоняется через default `pytest` — отдельная команда `locust -f tests/load/...`.

---

## `conftest.py` — фикстуры (спека)

### Env / stack

| Fixture | Scope | Поведение |
|---------|-------|-----------|
| `api_url` | session | `HUB_API_URL` или `http://localhost:3000` |
| `frontend_url` | session | `HUB_FRONTEND_URL` или `http://localhost:5173` |
| `backend_available` | session | GET `/api/health` timeout 2s → bool |
| `frontend_available` | session | GET frontend root → bool |
| `dev_mode_available` | session | probe `/api/dev/login` or env `DEV_MODE` |
| `docker_stack` | session | документирует зависимость от `docker-compose.local.yml` / test compose |

### Auth

| Fixture | Scope | Поведение |
|---------|-------|-----------|
| `seed_dev` | function/session | `POST /api/dev/seed` (idempotent) |
| `dev_tokens` | function | login `dev1@hub.local` → tokens+user |
| `admin_tokens` | function | login `admin@hub.local` |
| `auth_headers` | function | `Authorization: Bearer …` |
| `authenticated_page` | function (e2e) | Playwright page + sessionStorage inject (порт `authenticatePage`) |

### DB (integration only)

| Fixture | Scope | Поведение |
|---------|-------|-----------|
| `pg` | session/function | DSN из `POSTGRES_TEST_URL` / local; skip если нет |
| `mongo` | session | `MONGODB_TEST_URL` |
| `redis_client` | session | optional; для session/revocation probes |

**Инвариант:** E2E по умолчанию **не** открывает сырой SQL — только API + UI. DB fixtures — для `integration/`.

---

## Слои и границы

| Слой | Что тестирует | Запрещено |
|------|---------------|-----------|
| `integration/` | HTTP API, side-effects в PG/Mongo/NATS через API | Browser; импорт `backend/internal` |
| `e2e/` | UI flows + критичные cross-module paths | Прямой SQL; хрупкие CSS без role/testid |
| `load/` | Throughput/latency hot paths | Assert UI; смешивание с pytest CI PR |

Маркеры CI (план):

- `pytest -m smoke` → e2e smoke (3 сценария)
- `pytest -m "e2e and not smoke"` → full e2e
- `pytest -m integration` → integration job
- Locust → scheduled / manual workflow (см. E5/E6)

---

## Связь со старым `testframework/`

| Сейчас | После миграции |
|--------|----------------|
| `testframework/stages/e2e/` | `tests/e2e/` |
| `testframework/stages/unit/frontend-test/` | `frontend/src/**/*.test.ts(x)` (не `/tests`) |
| `backend/testframework/` | `backend/internal/<pkg>/*_test.go` + часть сценариев → `tests/integration/` |
| `testframework/PIPELINE.md` | новый `tests/README.md` + CI jobs в Э7 |

Переходный период: оба дерева живут; паритет сценариев; удаление старого только в phase 136.

---

## Acceptance для Э1 (когда будут писать код)

1. `tests/pyproject.toml` существует; `pip install -e ".[dev]"` / `uv sync` документирован.  
2. `conftest` поднимает auth: один integration-тест `GET /api/health` + один authenticated `GET` (например `/api/projects`) **зелёный** против local Docker.  
3. Playwright browsers установлены (`playwright install chromium`).  
4. Нет удаления старого `testframework/` в Э1.

---

## Mapping → GSD

Реализация layout: **Phase 128** — [`../phases/128-tf-python-scaffold/PLAN.md`](../../phases/128-tf-python-scaffold/PLAN.md).  
Hard dep: **A3 coverage baseline** до старта Э1.
