# Hub — Стратегия тестового фреймворка

> **Статус:** принятая стратегия; implementation rebase начат 2026-08-07.
> **Реализация целевого состояния:** фазы **128–141**,
> [канон](../.planning/active/august-2026/06-TEST-ARCHITECTURE-REBASE.md).
> Микросервисный split больше не является prerequisite: сначала тестовый
> фундамент и evidence-based re-audit.
> **Связанные документы:** `.cursor/playbooks/testing.md` (рабочие практики), `docs/test_plan.md` (тест-план), `docs/TESTING.md` (TC-матрица), `testframework/PIPELINE.md` (карта CI), `.planning/COVERAGE.md` (coverage + тайминги).

---

## 1. Цели и принципы

Стратегия отвечает на вопрос «как устроено тестирование Hub сейчас и куда оно движется». Четыре несущих принципа:

### 1.1. Детерминизм

Тест либо всегда зелёный, либо указывает на реальный дефект. Флейки разрушают доверие к CI — команда начинает игнорировать красное. Правила:

- никакого реального времени (`time.Sleep`, `waitForTimeout`) — инъекция clock, web-first assertions;
- никакой реальной сети в unit-слое — только моки (MSW, `mockRepo`);
- изоляция состояния: свой sqlite/QueryClient/фикстура на каждый тест, уникальные данные (uuid/суффикс);
- `-race` обязателен для Go, `-shuffle=on` при подозрении на зависимость от порядка;
- unit-тесты не ретраятся никогда; ретраи допустимы только в E2E и только против инфраструктурной нестабильности (см. §6.2).

### 1.2. Трассируемость: требование → TC → автотест

Каждая цепочка прослеживается в обе стороны:

```
US/AC (vision_and_scope) / FR (system_specification)
   → TC-<MODULE>-NN (docs/TESTING.md, владелец qa-manual)
      → автотест (колонка «Автотест» в TC-таблице + Allure tms-link/AS_ID)
         → результат прогона в Hub TMS (Allure webhook)
```

Матрица «требование → код → тест» ведётся в `.planning/COVERAGE.md`. Автотест без привязки к TC не принимается верификатором (`qa-auto-ver`), TC без покрытия — это явный `TODO`/«ручной» в колонке «Автотест», а не пустота.

### 1.3. Пирамида

Много быстрых юнитов, меньше интеграционных, минимум медленных E2E:

```
        ┌─────────────┐
        │     E2E     │  критические user paths, smoke на PR / full на master
        ├─────────────┤
        │ Integration │  API/contract/компонентные — по верифицированным TC
        ├─────────────┤
        │    Unit     │  пишут разработчики вместе с кодом (be — Go, fe — TS)
        └─────────────┘
```

Антипаттерны: «мороженое» (всё в E2E), дублирование unit-проверок в E2E, тесты «ради процента» без ассертов поведения.

### 1.4. Замер времени

Длительность прогонов — метрика первого класса: CI-бюджет ограничен (~3000 мин/мес GitHub Actions), медленный тест умножается на каждый PR. Wall-time значимых прогонов фиксируется в `.planning/COVERAGE.md`; деградация >20% — заметка в `STATE.md` + задача на оптимизацию. Подробно — §8.

---

## 2. Текущее состояние (до cutover Phase 141)

### 2.1. Слои и стек

| Слой | Стек | Где живёт | CI job |
|------|------|-----------|--------|
| Unit backend | Go + testify, `//go:build hubtest` | `backend/testframework/{domain}/` | `backend-test` |
| Unit frontend | Vitest + RTL + MSW | `testframework/stages/unit/frontend-test/domains/{domain}/` | `frontend-test` |
| API / handler | Go httptest (в составе unit-слоя) | `backend/testframework/` | `backend-test` |
| E2E | Playwright (TypeScript) + POM | `testframework/stages/e2e/domains/` | `e2e-smoke` (3 spec), `e2e-full` (20 spec) |
| Coverage gate | `go tool cover` | `backend/testframework/` | `coverage-gate` |

Go-тесты физически живут в `backend/testframework/` (не в `testframework/stages/`), потому что Go запрещает импорт `backend/internal/*` из другого модуля; логически это job `backend-test`.

### 2.2. Метрики (inventory 2026-08-07)

| Метрика | Значение |
|---------|----------|
| Ledger | **1,047** source rows; SHA-256 `01dae4e2…`; `unmapped = 0` |
| Backend test files | **830** total / **3,667** declared: legacy `backend/testframework/` 829/3,665 + server 1/2; taxonomy 594 genuine unit + 236 component (31 API + 205 integration); 0 в target `backend/tests/` |
| Backend `export_hooks.go` | **39** canonical + 4 fragments; deletion only TASK-219 |
| Frontend Vitest | **180** files / **519** literal cases: 156 frontend unit + 24 static contract |
| Frontend support | **3** source rows → `frontend/tests/support/**` |
| Playwright E2E | 20 specs / 53 declared cases (smoke: 3 specs) + 14 support rows |
| Target Python tests | **0** до Phase 129 |
| Statement coverage `internal/` | **UNKNOWN pending TASK-203**; gate ≥82% не ослаблен |
| Security-пакеты (auth/ws/rbac/audit/calls + platform) | 6/6 ≥80% |

### 2.3. Команды прогона (локально)

```bash
# Backend (как CI)
cd backend && go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...

# Один домен backend
cd backend && go test -tags hubtest ./testframework/auth/...

# Frontend unit
cd frontend && node scripts/link-testframework-node-modules.mjs && npm run test:run

# E2E
cd frontend && npm run test:e2e:smoke   # 3 spec, без backend
cd frontend && npm run test:e2e         # 20 spec, нужен backend
```

### 2.4. Паттерны текущего фреймворка

| Паттерн | Где |
|---------|-----|
| Page Object | `testframework/stages/e2e/shared/pages/` |
| Steps | `testframework/stages/e2e/shared/steps/` |
| Env Builder | `backend/testframework/env/` |
| Factory | `backend/testframework/factory/` |
| TestRenderer | `testframework/stages/unit/frontend-test/render.tsx` |

---

## 3. Целевое состояние (уточнено rebase 2026-08-07)

### 3.1. Матрица слоёв

Юниты остаются на языке кода; **весь не-юнит слой переезжает на Python**; нагрузочные — **Locust**.

| Слой | Язык/стек | Где живёт | Кто пишет |
|------|-----------|-----------|-----------|
| Unit backend | **Go** | `backend/tests/{domain}/` | be1/be2 |
| Unit frontend | **TypeScript** (Vitest + RTL) | `frontend/tests/{domain}/`; support в `frontend/tests/support/` | fe1/fe2 |
| Integration / API / contract / smoke / regression / E2E | **Python**: pytest + allure-pytest + httpx (API) + playwright-python (E2E) + pytest-xdist (параллелизм) | `testframework/{api,integration,contract,e2e}/` в едином корневом Python-проекте | qa-auto1/qa-auto2 **по верифицированным TC** |
| Нагрузочные | **Python: Locust** (общий `core` с pytest-слоем) | `testframework/load/` в том же Python-проекте | qa-auto1/qa-auto2 |

`backend/tests/**` принимает только genuine Go unit tests. Текущие
**236** `go-component` rows (31 API + 205 integration) нельзя механически
переносить туда: каждая строка
либо переезжает в Python `api/integration/contract`, либо сначала
рефакторится в настоящий isolated unit с verifier-approved evidence.

Обоснование выбора Python: стандарт индустрии QA-автоматизации (нанимаемость ручников и автоматизаторов), allure-pytest — самая зрелая Allure-интеграция (стыкуется с Hub TMS напрямую), отделение QA-слоя от языков продакшена соответствует модели «юниты пишут разработчики, остальное — QA-конвейер».

### 3.2. Структура каталогов Python-фреймворка

Это **один Python-проект прямо в корне `testframework/`**, без
`testframework/python/` и без вложенных самостоятельных проектов.
Раскладка по слоям и доменам:

```
testframework/
├── pyproject.toml              # единые dependencies/tooling и pytest config
├── pytest.ini                  # опциональный config вместо [tool.pytest.ini_options]
├── conftest.py                 # корневой env/base_url; один runner
├── core/                       # client, settings, factories, shared fixtures, Allure helpers
├── api/                        # HTTP/API tests по доменам
├── integration/                # PostgreSQL/Mongo/Redis/NATS/storage/provider boundaries
├── contract/                   # schemas/events и fixture contracts
├── e2e/                        # playwright-python, POM/steps/domains
├── load/                       # Locust scenarios; исполняемый тоже только Python
├── docker/                     # изолированные test-service profiles/orchestration
└── reports/                    # sanitized generated reports/artifact metadata
```

Корневой `conftest.py` содержит только общие fail-closed настройки;
browser и доменные fixtures живут в своём слое. Общие
фабрики/fixtures хранятся в `core/`, а не в отдельном корневом
`fixtures/`. Smoke/full — pytest markers и CI selection, а не отдельные
проекты или каталоги. После Phase 141 в `testframework/**` не остаётся
исполняемого Go/Vitest/TypeScript E2E-кода.

### 3.3. Нагрузочные: Locust и SLO

Locust-сценарии живут в `testframework/load/`, переиспользуют фикстуры/фабрики pytest-слоя. Целевые SLO (бюджеты продукта):

| Метрика | Цель |
|---------|------|
| API response time (p95) | < 200 ms |
| WebSocket message latency | < 100 ms |
| Одновременных WS-подключений | **500+** |
| Kanban board load (100 задач) | < 500 ms |
| Messenger: канал на 1000 сообщений | < 300 ms |
| Глобальный поиск | < 500 ms |
| Allure import (1000 результатов) | < 5 s |

Сценарии: WS connect + сообщения, auth-flow (magic link/refresh), board API, messenger throughput, Allure-import 1000 результатов. Режим прогонов: nightly / pre-release, **не в основном CI** (дорого); результаты и тайминги — в `.planning/COVERAGE.md`.

**Важно:** Locust-baseline заводится **раньше** остальной фазы TF-PY — до старта выноса realtime-трио (Track P M4). Locust самостоятелен и не ждёт pytest-каркаса; baseline 500 WS нужен как опорная точка «до/после» микросервисного split (заменяет ранее рассматривавшийся k6).

---

## 4. QA-конвейер (порядок обязателен)

Поток «бизнес-дока → тест-кейс → верификация доки → автотест → верификация автотестов» строго последователен. **Юнит-тесты в конвейер не входят — их пишут разработчики (`be*`/`fe*`) вместе с кодом (rule 56).**

```
BA/SA (vision_and_scope, system_specification — US/AC/FR)
      │
      ▼
qa-manual      ── анализирует бизнес-доку → пишет TC-<MODULE>-NN в docs/TESTING.md
      │            (шаги, ожидаемый результат; колонка «Автотест» = TODO/«ручной»)
      ▼
qa-manual-ver  ── верифицирует тестовую доку (полнота, техники, трассируемость);
      │            без verified:true автотесты НЕ стартуют
      ▼
qa-auto1/2     ── автоматизирует ВЕРИФИЦИРОВАННЫЕ TC; обновляет колонку «Автотест»
      ▼
qa-auto-ver    ── верифицирует автотесты (детерминизм, привязка TC↔тест, green regression)
```

Ключевые правила:

1. **Порядок жёсткий.** Автоматизация неверифицированной доки = переделка.
2. **Юниты — зона разработчиков.** `qa-auto*` не пишут unit-тесты за `be`/`fe`.
3. **Пара TC ↔ автотест.** `qa-manual` владеет «что и зачем», `qa-auto` — «как»; точка встречи — колонка «Автотест» в `docs/TESTING.md`.
4. **Разделение потоков без конфликтов:** один файл — один владелец в рамках задачи; общие фикстуры (env Builder, factory, shared POM) не форкаются — изменения через согласованный handoff.

---

## 5. Allure и дог-фудинг в Hub TMS

Hub **сам парсит** Allure-отчёты (`backend/internal/tms`: `allure_parser`, `allure_matcher`) — команда заливает результаты собственных прогонов в свой же TMS-модуль.

### 5.1. Аннотации в автотестах

Обязательный минимум: иерархия `epic/feature/story`, `severity`, привязка к TC через **tms-link или label `AS_ID`** (высший приоритет матчинга). В целевом Python-стеке — allure-pytest:

```python
import allure

@allure.epic("Tracker")
@allure.feature("Project lifecycle")
@allure.story("Create and restore project")
@allure.severity(allure.severity_level.CRITICAL)
@allure.label("AS_ID", "HUB-TC-PROJ-01")          # матчинг by annotation
@allure.link("https://hub.local/tms/HUB-TC-PROJ-01", link_type="tms", name="HUB-TC-PROJ-01")
def test_create_project_lifecycle(api_client, project_factory):
    with allure.step("Create project"):
        ...
```

### 5.2. Стратегии матчинга (в порядке приоритета)

| Приоритет | Стратегия | Источник в Allure |
|-----------|-----------|-------------------|
| 1 | По аннотации/tms-link | `links[].type == "tms"` либо label `AS_ID` |
| 2 | По automation ID | automation-идентификатор |
| 3 | По case key | ключ кейса |
| 4 | По полному имени | `fullName` |
| 5 | По классу/методу | `suite` label + `name` |

Полагаться на fullName/className хрупко (переименование рвёт связь) — всегда ставить явный `AS_ID`/tms-link.

### 5.3. Доставка результатов

| Путь | Endpoint | Auth |
|------|----------|------|
| Webhook из CI (основной) | `POST /projects/:id/tms/webhook/allure` | API Key |
| Ручной импорт | `POST /projects/:id/tms/allure/import` (multipart) | JWT + RBAC |

Целевой цикл (TF-PY-ALLURE): pytest-прогон → allure-результаты → CI-шаг заливает webhook'ом в Hub TMS → тест-ран с результатами и привязкой к TC виден в продукте. Фреймворк сам отчитывается в наш TMS — это и проверка TMS-модуля боем.

Правила attachments: только релевантное для диагностики, **никогда** секреты (JWT, пароли, ключи, PII), корректный MIME для бинарей.

---

## 6. Гейты качества

### 6.1. Coverage per-service

| Область | Порог |
|---------|-------|
| Core-монолит `internal/` (statement) | **≥ 82%** (текущий gate; факт 83.0%, 2026-06-22 Docker baseline) |
| Выделенные сервисы (после split: hub-notification, hub-ai, realtime-трио) | **≥ 80%** каждый |
| Security-критичные пакеты (auth, ws, rbac, audit, calls; после split — соответствующие сервисы) | **≥ 80%** |

Gate живёт в CI (`coverage-gate`); падение ниже порога блокирует. Профили — только в TMP (rule 45), `-covermode=atomic` вместе с `-race`. Coverage — не самоцель: тест без ассертов поведения не принимается, приоритет — error paths, границы, security-инварианты.

### 6.2. Флейки

- **Unit (Go/Vitest): ретраев нет.** Флейк = дефект, чинится причина.
- **E2E: `retries: 2`** — только против инфраструктурной нестабильности CI; трейс собирается `on-first-retry` и обязателен для диагностики.
- Детект: `-count=20` / `--repeat-each=10` для подозрительных; `-shuffle=on`.
- **Quarantine:** флейк, который нельзя починить немедленно, помечается skip с причиной + тикетом + owner и заносится в `.planning/STATE.md`. **SLA карантина — 1 волна:** к воротам следующей волны тест либо починен, либо удалён с обоснованием. Верификатор проверяет отсутствие «голых» skip.

### 6.3. Contract-тесты

Все NATS-события версионируются JSON-схемами в `contracts/events/*.json` (реестр — план M0-02). Contract-слой (`testframework/contract/`) проверяет **обе стороны**: publisher сериализует по схеме, consumer принимает все версионные вариации. Ломающее изменение схемы без версии = красный CI. Это ворота микросервисного split: сервисы общаются только через проверенные контракты.

### 6.4. Топология E2E

| Триггер | Что гоняем |
|---------|-----------|
| PR | `e2e-smoke` (3 spec) + spec'и затронутого домена (по path-filter) |
| Push/merge в master | `e2e-full` (все домены, 20 spec) |
| Nightly / pre-release | full + Locust-профиль |

### 6.5. CI-бюджет

Лимит ~3000 мин/мес (студенческий пакет GitHub Actions). Меры экономии:

- **path-filtered CI** (dorny/paths-filter, план M0-05): PR в один домен запускает только его jobs;
- smoke на каждом PR, full — на master/по необходимости;
- Playwright: шардинг + blob/merge-reports, `storageState` (логин один раз);
- pytest-xdist для параллелизма Python-слоя;
- нагрузочные и бенчмарки — вне основного CI;
- замер минут фиксируется при изменении pipeline (см. §8).

---

## 7. План миграции — фаза TF-PY

**Предусловие (Depends on):** Phase 128 приняла baseline,
inventory и coverage contract. Микросервисный split не является prerequisite.

| План | Содержание | Ворота |
|------|------------|--------|
| TF-PY-01 | Спека структуры Python-фреймворка (sa + qa-auto): раскладка по доменам/сервисам, conftest-иерархия, фикстуры, фабрики, POM для playwright-python, Allure-аннотации с TC-ID | sa-ver + qa-auto-ver |
| TF-PY-02 | Каркас прямо в `testframework/`: корневой pyproject/pytest config, базовые fixtures в `core/`, allure-pytest config | Чистая установка + ненулевой smoke; CI добавляется в dual-run фазе |
| TF-PY-03..NN | Миграция по доменам: API-тесты каждого сервиса/домена (notification, ai, ws/messenger/calls, core-домены) — **один атомарный план на домен** | Паритет сценариев с Go-integration слоем домена |
| TF-PY-E2E | Перенос Playwright-спеков TS → playwright-python (POM/steps), сохранение smoke/full-разбиения | Паритет 20 spec; smoke стабилен на PR |
| TF-PY-LOAD | Locust-сценарии в `testframework/load/` с целями SLO (§3.3); nightly/pre-release; **baseline — раньше фазы, до выноса realtime** | Baseline 500 WS зафиксирован в COVERAGE.md |
| TF-PY-ALLURE | Сквозная привязка TC↔автотест (tms-link/AS_ID) + автозаливка в Hub TMS webhook'ом из CI | Результаты прогона видны в TMS, unmapped → 0 |

**Ворота фазы целиком:** паритет покрытия сценариев со старым слоем **до**
удаления legacy Go/Vitest/TS E2E; тайминги зафиксированы; CI-бюджет
соблюдён. После успешного dual-run Phase 141 обязана удалить эти
исполняемые legacy слои и оставить `testframework/**` единым Python-проектом.

---

## 8. Замер времени прогонов и бюджеты

### 8.1. Что и куда фиксируем

- **Wall-time** каждого значимого прогона (полная регрессия, coverage-прогон, e2e-full, Locust) — в `.planning/COVERAGE.md` рядом с процентами.
- **Топ медленных тестов** — `go test -json` / Playwright `--reporter=json` / `pytest --durations=15`; медленные ревьюятся и либо оправданы, либо чинятся (мок вместо БД, web-first вместо sleep).
- **Деградация >20%** — заметка в `.planning/STATE.md` + задача на оптимизацию.

### 8.2. Ориентировочные бюджеты прогонов

| Прогон | Бюджет |
|--------|--------|
| Один домен backend (локально) | < 10 s |
| Полный backend-test (CI) | ≤ 5 мин |
| frontend-test (tsc + lint + vitest) | ≤ 5 мин |
| e2e-smoke | ≤ 5 мин |
| e2e-full | ≤ 20 мин |
| Python API-слой (после TF-PY, с xdist) | ≤ 10 мин |
| Locust nightly-профиль | вне CI-бюджета, отдельное окружение |

Бюджеты пересматриваются при изменении состава тестов; факт против бюджета — часть отчёта волны.

---

## 9. Ответственность (сводка)

| Артефакт | Владелец | Верификатор |
|----------|----------|-------------|
| Unit-тесты backend (Go) | be1/be2 | be-ver |
| Unit-тесты frontend (Vitest) | fe1/fe2 | fe-ver |
| TC-матрица (`docs/TESTING.md`) | qa-manual | qa-manual-ver |
| Integration/API/contract/E2E (по TC) | qa-auto1/qa-auto2 | qa-auto-ver |
| Locust-профили | qa-auto1/qa-auto2 | qa-auto-ver |
| CI jobs / бюджет минут | devops-engineer | qa-auto-ver / security-auditor |
| Эта стратегия и её drift | docs-writer | product-ver |

*Обновлено: 2026-08-07 — единый root `testframework/` Python contract,
Go/Vitest unit-only targets и mandatory legacy cutover cleanup.*
