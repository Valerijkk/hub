# Онбординг — QA-автоматизатор

> Роль: автоматизация тестов Hub **по верифицированным тест-кейсам**.
> Стек сейчас: Go integration/API + Vitest компонентные + Playwright-TS E2E. Целевой стек (фаза TF-PY): **Python pytest + allure-pytest + httpx + playwright-python + Locust**.
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

- [ ] Go 1.25+, Node 22, Docker Compose v2 (Python 3.12+ понадобится с фазы TF-PY).
- [ ] Локальный стек поднят, dev-login работает, seed прогнан.
- [ ] Все слои гоняются локально:

```bash
# Backend integration/API
cd backend && go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...

# Frontend (перед Vitest/E2E — линковка обязательна)
cd frontend && node scripts/link-testframework-node-modules.mjs && npm run test:run

# E2E
cd frontend && npm run test:e2e:smoke   # 3 spec, без backend
cd frontend && npm run test:e2e         # 20 spec, нужен backend
```

**Что прочитать (в этом порядке):**

- [ ] **`.cursor/playbooks/testing.md`** — целиком; это твоя настольная книга: конвейер (§1), техники (§3), Go/Vitest/Playwright-паттерны (§4–6), трассируемость (§7), Allure (§8), тайминги (§9), регрессия (§11), triage (§12), флейки (§13).
- [ ] **`docs/testframework_strategy.md`** — текущее и целевое состояние фреймворка, план TF-PY.
- [ ] `testframework/PIPELINE.md` — карта CI: стадия → job → домен.
- [ ] `docs/TESTING.md` — TC-матрица (твой вход) и колонка «Автотест» (твой выход).
- [ ] `docs/test_plan.md`, `.planning/COVERAGE.md` — стратегия, coverage и тайминги.

## Неделя 1

- [ ] Изучи паттерны фреймворка: env `Builder` и `factory` (`backend/testframework/env|factory/`), POM и Steps (`testframework/stages/e2e/shared/`), `render.tsx` для компонентных.
- [ ] Разбери, как один verified TC превращается в автотест: найди пару «TC-строка в TESTING.md ↔ Go/Playwright-тест» и проследи привязку (имя, `AS_ID`).
- [ ] Прогони Allure-цикл локально: сгенерируй результаты, посмотри на матчинг в Hub TMS (`POST /projects/:id/tms/allure/import`).
- [ ] Пойми разделение доменов между потоками (playbook §1.4): qa-auto1 — Go + E2E auth/navigation/tracker/tms/messenger; qa-auto2 — Vitest + E2E admin/ai/platform/v4. Один файл — один владелец.
- [ ] **Первая задача (good-first):** автоматизируй 1–2 verified TC, у которых в колонке «Автотест» стоит TODO. Полный цикл: тест → привязка TC → зелёная регрессия домена → обновление колонки → handoff на `qa-auto-ver`.

## Инструменты и команды

```bash
# Один домен backend (быстрый цикл)
cd backend && go test -tags hubtest -race ./testframework/tracker/...

# Один домен frontend
cd frontend && npx vitest run ../testframework/stages/unit/frontend-test/domains/auth

# Один E2E-spec, детект флейка
npx playwright test path/to.spec.ts --repeat-each=10
npx playwright show-trace test-results/<...>/trace.zip

# Флейк-детект Go
go test -tags hubtest -race -run TestSuspect -count=20 ./testframework/<pkg>/...
```

**Целевой стек (фаза TF-PY — после микросервисного split M2–M4):** pytest + allure-pytest + httpx (API), playwright-python (E2E, POM), pytest-xdist (параллелизм), **Locust** (`testframework/load/`) для нагрузки с SLO-целями (API p95<200ms, WS<100ms, 500+ WS). Locust-baseline заводится раньше остальной фазы. До завершения TF-PY работаем на текущем стеке — детали и структура каталогов в `docs/testframework_strategy.md` §3, §7.

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
- **Кросс-ревью:** твой PR смотрит второй auto-поток, затем `qa-auto-ver` (детерминизм, привязка TC↔тест, отсутствие flaky/sleep, green regression + лог).
- **Упал тест — triage строго до фикса** (`.cursor/skills/test-triage/SKILL.md`): воспроизведи функционал по спеке → ROOT=code|test → запись `TRIAGE/FIX/VERIFIED` в лог. Менять assertion «чтобы прошло» запрещено; ROOT=code эскалируется разработчику.
- **Общие фикстуры не форкать:** env Builder, factory, E2E shared — изменения через согласованный handoff (влияют на оба потока).
- **Docs-sync:** обновляй колонку «Автотест» в `docs/TESTING.md` в той же сессии; session note — `.planning/STATE.md`.
- **Git:** фича-ветки + Conventional Commits (`test(tracker): automate TC-TRACKER-12..15`); **мерж в master — только владелец**.

## Чек-лист готовности

- [ ] Все три слоя (Go/Vitest/Playwright) гоняются локально, зелёные.
- [ ] Прочитаны testing.md (целиком) и testframework_strategy.md.
- [ ] Понимаю конвейер: работаю только по verified TC, юниты — не моя зона.
- [ ] Понимаю привязку TC↔автотест и Allure-матчинг (AS_ID/tms-link).
- [ ] Знаю правила флейков, quarantine SLA и запрет sleep.
- [ ] Первые TC автоматизированы, колонка «Автотест» обновлена, qa-auto-ver дал verified.
- [ ] Понимаю целевой Python-стек и план TF-PY (что и когда переезжает).
