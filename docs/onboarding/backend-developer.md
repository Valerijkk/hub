# Онбординг — Backend-разработчик

> Роль: разработка backend Hub на **Go 1.24 + Fiber v3 + Ent (PostgreSQL 17)**.
> Главный playbook: `.cursor/playbooks/backend.md` · Общий индекс онбординга: [README.md](README.md)

---

## Кто ты и за что отвечаешь

- Реализуешь фичи backend по цепочке handler → service → repository (**layering строгий**: `*ent.Client` живёт только в repository; в service/handler — интерфейсы).
- **Unit-тесты — твоя обязанность**, а не QA: пишешь их вместе с кодом фичи в `backend/testframework/{domain}/` (rule 56). QA-конвейер покрывает integration/API/E2E — не юниты.
- Соблюдаешь security-инварианты (SSRF, timing attacks, RBAC, rate limits) и TZ-ограничения — они не обсуждаются в коде-ревью, они блокеры.
- Доводишь свою ветку до merge-ready; мерж в `master` делает только владелец.

Твои соседи по роли в AI-команде — агенты `be1`/`be2` (параллельные потоки) и верификатор `be-ver`.

## День 1

**Доступы и окружение** — общий чек-лист в [README.md](README.md). Дополнительно для BE:

- [ ] Go 1.25+ установлен (`go version`), Docker Compose v2.
- [ ] Поднят локальный стек: `docker compose --env-file .env.local -f docker-compose.local.yml up --build` → UI на `http://localhost:8080`, dev-login работает.
- [ ] Демо-данные: `curl -X POST http://localhost:8080/api/dev/seed`.
- [ ] Backend собирается и тесты зелёные:

```bash
cd backend
go generate ./ent
go build -tags release ./...
go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...
```

**Что прочитать (в этом порядке):**

- [ ] `docs/architecture.md` — модули, потоки, deployment (31+ модулей в `backend/internal/`).
- [ ] `docs/system_specification.md` — FR/NFR, ERD, API-контракты.
- [ ] `docs/code_style_guide.md` — стиль Go: ошибки через `fmt.Errorf("...: %w", err)`, `slog` с контекстом, UUID для PK.
- [ ] `docs/security_guidelines.md` + `docs/tz_constraints.md` — запреты (см. таблицу ниже).
- [ ] **`.cursor/playbooks/backend.md`** — главный рабочий документ: Fiber, Ent, NATS, Mongo, конкурентность.
- [ ] `.cursor/playbooks/observability.md`, `code-comments.md`, `git-workflow.md`.

## Неделя 1

- [ ] Пройдись по продукту как пользователь: создай проект, задачи на Kanban, спринт; напиши в мессенджер; заведи тест-кейс в TMS; посмотри дашборды.
- [ ] Пройди код одного домена насквозь: `backend/internal/tracker/` — handler → service → repository → ent-схема (`backend/ent/schema/`) — и его тесты в `backend/testframework/tracker/`.
- [ ] Разберись с платформенными пакетами: `internal/platform/` (errors, middleware, http-helpers, outbox, nats) — они используются везде.
- [ ] **Первая задача (good-first):** маленький фикс или тест в одном домене — например, добить error-path тестом существующий сервис или закрыть мелкий пункт из `.planning/AUDIT-*.md`. Согласуй с тимлидом; план на ≤200–300 строк.
- [ ] Первый PR: фича-ветка → регрессия домена → handoff.

## Инструменты и команды

```bash
# Полная регрессия backend (как CI)
cd backend && go build -tags release ./... && \
  go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...

# Один домен (быстрый цикл, < 10 сек)
go test -tags hubtest -race ./testframework/tracker/...

# Ent codegen после изменения схем
go generate ./ent

# Coverage (профиль ТОЛЬКО в TMP — rule 45)
go test -tags hubtest -race -covermode=atomic \
  -coverprofile="${TMPDIR:-/tmp}/hub-cov.out" ./testframework/...
go tool cover -func="${TMPDIR:-/tmp}/hub-cov.out" | tail -1
```

**Запрещённые паттерны (блокеры ревью):**

| Паттерн | Альтернатива |
|---------|--------------|
| `http.DefaultClient` для user-controlled URL | `safehttp.NewSafeHTTPClient(10*time.Second)` |
| `io.ReadAll` без лимита | `io.LimitReader(r, maxBytes)` |
| `!=` для секретов | `subtle.ConstantTimeCompare` |
| `*ent.Client` в service/handler | Repository-интерфейс |
| `panic()` в production paths | return error |

Гейты: coverage `internal/` ≥82% (security-пакеты ≥80%); `-race` всегда зелёный.

## Как работает AI-команда и как с ней взаимодействовать

Ростер и процессы: `.cursor/playbooks/README.md`, `team-workflow.md`. Что важно тебе:

- **Требования и спеки готовят агенты-аналитики** (`ba1`/`ba2` → US/AC, `sa1`/`sa2` → FR/API/ERD). Не начинай фичу без спеки — если её нет, запроси через тимлида.
- **Кросс-ревью:** твой PR ревьюит второй BE-поток (человек или агент `be1`/`be2`), затем верификатор `be-ver` (layering, security, ошибки, тесты, race). `verified: true` — только при зелёной регрессии с логом.
- **Упал тест — сначала triage** (`.cursor/skills/test-triage/SKILL.md`): проверь функционал по спеке, классифицируй ROOT=code|test, только потом чини. Подгонять assertion «чтобы прошло» запрещено.
- **Сложный баг** — можно поручить агенту `debugger` (root-cause по трейсам/логам); security-вопросы — `security-auditor`.
- **Docs-sync обязателен** в той же сессии: изменил API → `docs/system_specification.md` + `docs/wiki/API.md`; любой код → session note в `.planning/STATE.md` (rule 55).
- **Git:** коммиты в фича-ветки свободно (Conventional Commits: `feat(tracker): ...`), push ветки — ок; **мерж в master только владелец**. Работая параллельно с другим потоком, не трогай его файлы — границы модулей на волну назначает tech-lead.

## Чек-лист готовности

- [ ] Локальный стек поднят, dev-login работает, seed прошёл.
- [ ] Полная backend-регрессия зелёная локально.
- [ ] Прочитаны: architecture, system_specification, code_style_guide, security_guidelines, tz_constraints, playbook backend.md.
- [ ] Понимаю layering и таблицу запрещённых паттернов.
- [ ] Знаю QA-конвейер: юниты пишу я, integration/E2E — QA по верифицированным TC.
- [ ] Первый PR прошёл кросс-ревью + be-ver и доведён до merge-ready.
- [ ] Знаю git-политику: фича-ветки — можно, master — только владелец.
