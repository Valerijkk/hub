# Онбординг — QA-автоматизатор

> Роль: автоматизация тестов Hub **по верифицированным тест-кейсам**.
> Legacy Go component/API/integration, static-contract Vitest и Playwright-TS — исторические и не портируются. Целевой greenfield стек: **Python pytest + allure-pytest + httpx + playwright-python + Locust**; package/launcher пока `NOT_READY`, runtime `NOT_RUN`.
> Главные документы: `.cursor/playbooks/testing.md` + `docs/testframework_strategy.md` · Индекс: [README.md](README.md)

---

## Кто ты и за что отвечаешь

- Автоматизируешь **только верифицированные TC** (прошедшие `qa-manual-ver`) — слой integration/API/contract/smoke/regression/E2E. Автоматизация неверифицированной доки = переделка.
- **Юниты не пишешь** — это обязанность разработчиков (`be*` — Go, `fe*` — Vitest); твой слой начинается выше пирамиды.
- Поддерживаешь трассируемость: каждый автотест привязан к TC (колонка «Автотест» в `docs/TESTING.md` + Allure label `AS_ID`/tms-link).
- Отвечаешь за детерминизм своего слоя: ноль флейков, изоляция данных, никаких sleep.
- Заливаешь Allure-результаты в Hub TMS (дог-фудинг: фреймворк отчитывается в наш собственный TMS-модуль).

Соседи по роли: агенты `qa-auto1` (backend/API + часть E2E) и `qa-auto2` (frontend-компонентные + часть E2E), верификатор `qa-auto-ver`.

## День 1

**Доступы и окружение** — общий чек-лист в [README.md](README.md). Дополнительно:

- [ ] Docker CLI/Compose доступен; Go/Node/Python test tooling на Windows host не запускается.
- [ ] Локальный стек поднят, dev-login работает, seed прогнан.
- [ ] Знаешь fixed `/qa-gate` targets: `python-contract`, `python-api`,
  `python-integration`, `python-e2e-smoke`, `python-e2e-full`, `python-load`.
  Они должны исполняться только sole versioned outer launcher в task-owned
  limited Docker. Launcher `NOT_READY`, runtime `NOT_RUN`; direct Go/Vitest/
  Playwright/pytest/Locust host commands и legacy runners не fallback.

**Что прочитать (в этом порядке):**

- [ ] **`.cursor/playbooks/testing.md`** — целиком; это твоя настольная книга: конвейер (§1), техники (§3), Go/Vitest/Playwright-паттерны (§4–6), трассируемость (§7), Allure (§8), тайминги (§9), регрессия (§11), triage (§12), флейки (§13).
- [ ] **`docs/testframework_strategy.md`** — текущее и целевое состояние фреймворка, план TF-PY.
- [ ] `testframework/PIPELINE.md` — карта CI: стадия → job → домен.
- [ ] `docs/TESTING.md` — TC-матрица (твой вход) и колонка «Автотест» (твой выход).
- [ ] `docs/test_plan.md`, `.planning/COVERAGE.md` — стратегия, coverage и тайминги.

## Неделя 1

- [ ] Изучи целевой package contract `testframework/PACKAGE-CONTRACT.md`, структуру `testframework/src/hub_testkit/**` и suites `testframework/tests/{contract,api,integration,e2e}/**`; legacy roots не копируй.
- [ ] Разбери, как один verified TC превращается в автотест: найди пару «TC-строка в TESTING.md ↔ Go/Playwright-тест» и проследи привязку (имя, `AS_ID`).
- [ ] Изучи контракт Allure/TMS (`POST /projects/:id/tms/allure/import`); runtime-цикл не заявляй выполненным до появления canonical launcher evidence.
- [ ] Пойми целевое ownership по Python suites и `hub_testkit`; один файл — один writer, legacy Go/Vitest/TS Playwright не портируются.
- [ ] **Первая задача (good-first):** автоматизируй 1–2 verified TC в целевой Python suite. Полный цикл требует canonical target evidence; если launcher всё ещё `NOT_READY`, handoff остаётся `NOT_READY`/`NOT_RUN`, а не green.

## Инструменты и команды

| Слой | Fixed target | Текущий статус |
|---|---|---|
| Contract | `python-contract` | launcher `NOT_READY`, runtime `NOT_RUN` |
| API | `python-api` | launcher `NOT_READY`, runtime `NOT_RUN` |
| Integration | `python-integration` | launcher `NOT_READY`, runtime `NOT_RUN` |
| E2E | `python-e2e-smoke`, `python-e2e-full` | launcher `NOT_READY`, runtime `NOT_RUN` |
| Load | `python-load` | launcher `NOT_READY`, runtime `NOT_RUN` |

Прямые host-команды и прямой Docker/Compose запрещены как fallback. Принимается
только task-owned limited Docker evidence sole `/qa-gate` outer launcher.

**Целевой стек:** pytest + allure-pytest + httpx (API), playwright-python (E2E, POM), pytest-xdist (параллелизм), **Locust** (`testframework/load/`). До готовности package/launcher нет допустимого legacy fallback; детали и структура — в `testframework/PACKAGE-CONTRACT.md` и плане rebase.

**Железные правила слоя:**

- unit-тесты не ретраятся; E2E — `retries: 2` только против инфры, трейс `on-first-retry`;
- никаких `time.Sleep`/`waitForTimeout` — web-first assertions, fake clock;
- изоляция данных: uuid-суффиксы, свой ресурс на тест;
- quarantine только с причиной + тикетом + owner, SLA — 1 волна;
- coverage-профили — в TMP (rule 45), артефакты (`test-results/`, `allure-results/`) не коммитятся;
- wall-time значимых прогонов — в `.planning/COVERAGE.md`.

## Как работает AI-команда и как с ней взаимодействовать

Ростер: `.cursor/playbooks/README.md`. Твоё место в конвейере:

```
qa-manual (TC) → qa-manual-ver → ТЫ (автотесты по verified TC) → qa-auto-ver
```

- **Не стартуй без verified TC.** Нет TC или он не верифицирован — верни задачу в конвейер (через тимлида), а не пиши «по коду».
- **Кросс-ревью:** твой PR смотрит второй auto-поток, затем `qa-auto-ver` (детерминизм, привязка TC↔тест, отсутствие flaky/sleep). `verified: true` требует `PASS` canonical Docker target; `NOT_READY`/`NOT_RUN` не green.
- **Упал тест — triage строго до фикса** (`.cursor/skills/test-triage/SKILL.md`): воспроизведи функционал по спеке → ROOT=code|test → запись `TRIAGE/FIX/VERIFIED` в лог. Менять assertion «чтобы прошло» запрещено; ROOT=code эскалируется разработчику.
- **Общие фикстуры не форкать:** env Builder, factory, E2E shared — изменения через согласованный handoff (влияют на оба потока).
- **Docs-sync:** обновляй колонку «Автотест» в `docs/TESTING.md` в той же сессии; session note — `.planning/STATE.md`.
- **Git:** назначенная `codex/<task>` от refreshed master + Conventional Commits (`test(tracker): automate TC-TRACKER-12..15`); direct master implementation запрещён. После Docker-only evidence и независимого review routine merge/push/cleanup выполняет `team-lead`; security/auth, DB migrations, deploy/CI и ADR требуют founder review до integration.

## Чек-лист готовности

- [ ] Понимаю status boundary: package/launcher `NOT_READY`, runtime `NOT_RUN`; host/legacy fallback запрещён, green без canonical evidence не заявляется.
- [ ] Прочитаны testing.md (целиком) и testframework_strategy.md.
- [ ] Понимаю конвейер: работаю только по verified TC, юниты — не моя зона.
- [ ] Понимаю привязку TC↔автотест и Allure-матчинг (AS_ID/tms-link).
- [ ] Знаю правила флейков, quarantine SLA и запрет sleep.
- [ ] Первые TC автоматизированы, колонка «Автотест» обновлена, qa-auto-ver дал verified.
- [ ] Понимаю целевой Python-стек и план TF-PY (что и когда переезжает).
