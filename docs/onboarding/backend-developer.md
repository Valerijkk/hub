# Онбординг — Backend-разработчик

> Роль: разработка backend Hub на **Go 1.25 + Fiber v3 + Ent (PostgreSQL 17)**.
> Главный playbook: `.cursor/playbooks/backend.md` · Общий индекс онбординга: [README.md](README.md)

---

## Кто ты и за что отвечаешь

- Реализуешь фичи backend по цепочке handler → service → repository (**layering строгий**: `*ent.Client` живёт только в repository; в service/handler — интерфейсы).
- **Genuine unit-тесты — твоя обязанность**, а не QA: пишешь их вместе с кодом фичи в `backend/tests/{domain}/` (rule 56). Python QA-конвейер покрывает contract/API/integration/E2E — не юниты.
- Соблюдаешь security-инварианты (SSRF, timing attacks, RBAC, rate limits) и TZ-ограничения — они не обсуждаются в коде-ревью, они блокеры.
- Работаешь в `codex/<task>`, созданной от refreshed `master`; не реализуешь напрямую в `master`. После Docker-only проверки и независимого review `team-lead` интегрирует routine accepted branch и удаляет её; security/auth, DB migrations, deploy/CI и ADR требуют founder review до integration.

Твои соседи по роли в AI-команде — агенты `be1`/`be2` (параллельные потоки) и верификатор `be-ver`.

## День 1

**Доступы и окружение** — общий чек-лист в [README.md](README.md). Дополнительно для BE:

- [ ] Docker CLI/Compose доступен; Go build/test/codegen на Windows host не запускается.
- [ ] Поднят локальный стек: `docker compose -f docker-compose.local.yml up -d` → UI на `http://localhost:8080`, dev-login работает.
- [ ] Демо-данные: `curl -X POST http://localhost:8080/api/dev/seed`.
- [ ] Знаешь текущую runtime-правду: единственный QA-вызов для `backend-unit` —
  `docker compose -f docker-compose.local.yml run --rm backend-unit`.
  До фактического запуска статус остаётся `NOT_RUN`; host-команды Go и
  удалённые legacy runners не являются fallback.

**Что прочитать (в этом порядке):**

- [ ] `docs/architecture.md` — модули, потоки, deployment (31+ модулей в `backend/internal/`).
- [ ] `docs/system_specification.md` — FR/NFR, ERD, API-контракты.
- [ ] `docs/code_style_guide.md` — стиль Go: ошибки через `fmt.Errorf("...: %w", err)`, `slog` с контекстом, UUID для PK.
- [ ] `docs/security_guidelines.md` + `docs/tz_constraints.md` — запреты (см. таблицу ниже).
- [ ] **`.cursor/playbooks/backend.md`** — главный рабочий документ: Fiber, Ent, NATS, Mongo, конкурентность.
- [ ] `.cursor/playbooks/observability.md`, `code-comments.md`, `git-workflow.md`.

## Неделя 1

- [ ] Пройдись по продукту как пользователь: создай проект, задачи на Kanban, спринт; напиши в мессенджер; заведи тест-кейс в TMS; посмотри дашборды.
- [ ] Пройди код одного домена насквозь: `backend/internal/tracker/` — handler → service → repository → ent-схема (`backend/ent/schema/`) — и его genuine units в `backend/tests/tracker/`.
- [ ] Разберись с платформенными пакетами: `internal/platform/` (errors, middleware, http-helpers, outbox, nats) — они используются везде.
- [ ] **Первая задача (good-first):** маленький фикс или тест в одном домене — например, добить error-path тестом существующий сервис или закрыть мелкий пункт из `.planning/AUDIT-*.md`. Согласуй с тимлидом; план на ≤200–300 строк.
- [ ] Первый handoff: task-ветка → evidence `backend-unit`; до фактического запуска честно передай `NOT_RUN`, не заявляя green.

## Инструменты и команды

Единственный принимаемый execution path —
`docker compose -f docker-compose.local.yml run --rm backend-unit`. До фактического
evidence результат равен `NOT_RUN`; host Go и удалённые legacy runners не допускаются
как fallback.

**Запрещённые паттерны (блокеры ревью):**

| Паттерн | Альтернатива |
|---------|--------------|
| `http.DefaultClient` для user-controlled URL | `safehttp.NewSafeHTTPClient(10*time.Second)` |
| `io.ReadAll` без лимита | `io.LimitReader(r, maxBytes)` |
| `!=` для секретов | `subtle.ConstantTimeCompare` |
| `*ent.Client` в service/handler | Repository-интерфейс |
| `panic()` в production paths | return error |

Acceptance targets: coverage нового кода и security-пакетов проверяются только
каноническим target evidence; без него нельзя писать, что coverage или race green.

## Как работает AI-команда и как с ней взаимодействовать

Ростер и процессы: `.cursor/playbooks/README.md`, `team-workflow.md`. Что важно тебе:

- **Требования и спеки готовят агенты-аналитики** (`ba1`/`ba2` → US/AC, `sa1`/`sa2` → FR/API/ERD). Не начинай фичу без спеки — если её нет, запроси через тимлида.
- **Кросс-ревью:** твой PR ревьюит второй BE-поток (человек или агент `be1`/`be2`), затем верификатор `be-ver` (layering, security, ошибки, тесты, race). `verified: true` допустим только при `PASS` canonical Docker target с логом; `NOT_READY`/`NOT_RUN` не green.
- **Упал тест — сначала triage** (`.cursor/skills/test-triage/SKILL.md`): проверь функционал по спеке, классифицируй ROOT=code|test, только потом чини. Подгонять assertion «чтобы прошло» запрещено.
- **Сложный баг** — можно поручить агенту `debugger` (root-cause по трейсам/логам); security-вопросы — `security-auditor`.
- **Docs-sync обязателен** в той же сессии: изменил API → `docs/system_specification.md` + `docs/wiki/API.md`; любой код → session note в `.planning/STATE.md` (rule 55).
- **Git:** коммиты делай только в назначенной `codex/<task>` от refreshed master (Conventional Commits); не реализуй напрямую в protected master. После Docker-only evidence и независимого review routine merge/push/cleanup выполняет `team-lead`; security/auth, DB migrations, deploy/CI и ADR требуют founder review до integration. Работая параллельно с другим потоком, не трогай его файлы — границы модулей на волну назначает tech-lead.

## Чек-лист готовности

- [ ] Локальный стек поднят, dev-login работает, seed прошёл.
- [ ] Понимаю status boundary: без actual evidence runtime `NOT_RUN`; host/legacy fallback запрещён, green без actual evidence не заявляется.
- [ ] Прочитаны: architecture, system_specification, code_style_guide, security_guidelines, tz_constraints, playbook backend.md.
- [ ] Понимаю layering и таблицу запрещённых паттернов.
- [ ] Знаю QA-конвейер: юниты пишу я, integration/E2E — QA по верифицированным TC.
- [ ] Первый PR прошёл кросс-ревью + be-ver и доведён до merge-ready.
- [ ] Знаю git-политику: `codex/<task>` от refreshed master; direct master implementation запрещён; routine integration выполняет team-lead после ворот, а security/auth, DB migrations, deploy/CI и ADR требуют founder review.
