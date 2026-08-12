# Аналитика: архитектурный слой

> Часть плейбука. Канон и оглавление — [`../analysis.md`](../analysis.md).

Контекст, варианты, решения, ADR.

---

## 5. Архитектура (ARCH)

Владелец раздела: `arch`. Основные артефакты: `docs/architecture.md`, `docs/stack_and_tools.md`,
`docs/code_style_guide.md` (при изменении стандартов), ADR в `.planning/decisions/`.
Верификатор: `arch-ver`. Закон границ: `docs/tz_constraints.md` (mandatory/forbidden).

### 5.1 Что делает ARCH и чего не делает

**Делаю:** C4-модель (context/container/component), выбор библиотек с обоснованием, ADR,
deployment-топология, границы модулей, атрибуты качества (безопасность, производительность,
масштабируемость, надёжность), соответствие TZ, паттерны (Clean Architecture, Repository,
Event-Driven/Outbox).

**НЕ делаю:** реализация кода (→ BE/FE), бизнес-приоритеты (→ BA), детальные API-контракты
(→ SA; ARCH только согласует их непротиворечивость).

### 5.2 C4-модель

Hub документирует три уровня C4. Уровень 1 (Context) — Hub и внешние акторы/системы; уровень 2
(Container) — процессы и хранилища; уровень 3 (Component) — модули backend/frontend.

Образец стиля (Container, из architecture.md) — псевдографика:

```
Browser (React SPA) ⇄ Caddy :80/443 ⇄ ┌ Go Backend (Fiber v3) :3000
                                        ├ WS Handler (Fiber) :3000
                                        └ static (Vite build)
Backend ⇄ PostgreSQL 17 (Ent) · MongoDB 8 (messages) · Redis 7 (sessions/limits)
        · NATS JetStream (pub/sub) · SeaweedFS (S3) · LiveKit (WebRTC)
```

Шаблон C4 Context:

```
Акторы:
  - <Персона/роль> → взаимодействует с Hub через <канал>
Внешние системы:
  - GitHub/GitLab (webhooks, PR/MR)
  - Telegram (notify bind)
  - LiveKit (WebRTC media)
  - SMTP/Resend/hub-mail (email)
Hub System:
  - принимает <входы>, отдаёт <выходы>, хранит <данные>
```

Шаблон C4 Component (модуль):

```
Модуль: tms
Слои: handler (handler_cases.go, handler_runs.go)
      → service (service.go)
      → repository (repository.go interface, ent_repository.go)
      → Ent (PostgreSQL) + Mongo (allure_reports)
Зависимости: tracker (issue link), notification (bug alert), storage (attachments)
События: publishes hub.tms.run.finished (NATS/outbox)
```

При изменении C4 ARCH держит диаграммы **синхронными с кодом**: реальный дефект из лога
`arch-ver` — docs заявляли «18 модулей», код содержал 21; «Go 1.24» при фактической 1.25.
Правило: каждое число/факт в architecture.md сверяется с кодом (grep, go.mod, routeTree).

### 5.3 Выбор библиотек с обоснованием

Любой выбор/замена библиотеки — решение с обоснованием и записью в `stack_and_tools.md` +
ADR. Критерии выбора:

| Критерий | Вопрос |
|----------|--------|
| Соответствие TZ | Не нарушает ли mandatory/forbidden? (Fiber, NATS, SeaweedFS, TanStack Router — фиксированы) |
| Зрелость/поддержка | Активно ли развивается, есть ли релизы, лицензия совместима? |
| Безопасность | Нет ли известных CVE, безопасны ли дефолты (таймауты, лимиты)? |
| Совместимость | Работает ли с Go 1.25 / React 19 / текущим стеком? |
| Стоимость владения | Размер зависимостей, влияние на CI-время (бюджет ~3000 мин/мес) |
| Альтернативы | Рассмотрены ≥2 варианта, зафиксирован проигравший и почему |

Шаблон записи выбора:

```
Библиотека: <name@version>
Задача: <что решает>
Альтернативы: <A vs B vs C>
Выбор: <name>, потому что <обоснование по критериям>
Риски: <что может пойти не так>, митигация: <как>
TZ: не нарушает (или ADR-XXX если нужно исключение)
```

Запрещённые паттерны (из `code_style_guide.md`, TZ) — ARCH следит, чтобы выбор их не вводил:
`http.DefaultClient` для user-URL (SSRF), `io.ReadAll` без `LimitReader`, `!=` для секретов,
`*ent.Client` в Service/Handler, `panic()` в проде, `seaweedfs:latest` без pin.

### 5.4 ADR (Architecture Decision Record)

ADR фиксирует значимое архитектурное решение: контекст, варианты, решение, последствия. ADR
неизменяем после принятия (новое решение = новый ADR, старый помечается Superseded).

Полный шаблон ADR:

```markdown
# ADR-<NNN>: <краткое название решения>

- Статус: Proposed | Accepted | Superseded by ADR-XXX | Deprecated
- Дата: YYYY-MM-DD
- Автор: arch
- Связанные: US-XX, FR-XX, NFR-XX, TZ-раздел

## Контекст
<Какая проблема/сила требует решения. Ограничения TZ, NFR, текущее состояние кода.>

## Рассмотренные варианты
1. <Вариант A> — плюсы / минусы
2. <Вариант B> — плюсы / минусы
3. <Вариант C> — плюсы / минусы

## Решение
<Что выбрано и почему. Явная привязка к критериям 5.3 и к TZ.>

## Последствия
- Положительные: <...>
- Отрицательные / долг: <...> (регистрируется как TD-XX при необходимости)
- Влияние на NFR: <...>
- Влияние на deployment: <...>

## Проверка соответствия
<Как arch-ver убедится, что решение реализовано: файл:строка, команда, тест.>
```

Пример (в стиле проекта — ADR-AUTH-01 упоминается в TZ):

```markdown
# ADR-AUTH-01: Telegram — только notify bind, не login

- Статус: Accepted
- Дата: 2026-06-21
- Связанные: US-M01, NFR-05, TZ Authentication (mandatory)

## Контекст
Требуется единый метод входа. TZ: Magic Link (email) — единственный метод; SSO/OAuth forbidden.
Telegram обсуждался как второй способ входа.

## Рассмотренные варианты
1. Telegram login — противоречит TZ (единственный метод — Magic Link), forbidden.
2. Telegram только для notify bind/delivery — соответствует TZ.

## Решение
Telegram используется исключительно для привязки уведомлений, не для аутентификации.

## Последствия
- Положительные: соответствие TZ, меньше поверхность атаки на auth.
- Отрицательные: пользователи не могут войти через TG (осознанно).

## Проверка соответствия
Нет endpoint auth через TG; `backend/internal/auth/` не содержит TG-login; notify bind в notification.
```

### 5.5 Deployment

ARCH описывает топологию dev (Docker Compose) и prod (K3s), reverse proxy (Caddy), не меняя их
без решения devops. Образец: dev — `docker-compose.local.yml` в корне; prod —
`deployments/docker-compose.prod.yml` / K3s manifests (планируются). CI — GitHub Actions,
бюджет ~3000 мин/мес.

Шаблон deployment-описания:

```
Среда: dev | prod
Оркестрация: Docker Compose (dev) / K3s (prod)
Reverse proxy: Caddy (:80/443), TLS
Сервисы: backend (Fiber :3000), frontend (Vite static), PG17, Mongo8, Redis7, NATS JS,
         SeaweedFS, LiveKit, Meilisearch, hub-mail
Секреты: через env/secret store; в репо запрещены (TZ forbidden: hardcoded credentials)
Пины образов: обязательны (запрещён seaweedfs:latest)
Observability: Loki/Prometheus/Grafana — планируется (TD-05), не в prod compose сейчас
```

### 5.6 Ограничения TZ (mandatory / forbidden)

`docs/tz_constraints.md` (зеркало `.cursor/rules/60-tz-constraints.mdc`) — закон. ARCH — его
хранитель. Ключевые mandatory и forbidden, которые ARCH проверяет в каждом решении:

| Категория | MANDATORY | FORBIDDEN |
|-----------|-----------|-----------|
| Auth | Magic Link only; JWT 15 мин; refresh 30 дней Redis; rate 3/10 мин | SSO/LDAP/OAuth; Telegram login |
| Data | UTC в БД; version optimistic lock; UUID v4 PK; soft delete 30д | string PK; hard delete без запроса Admin/GDPR |
| API | ошибки `{"error","request_id"}`; пагинация offset/cursor; rate limits | — |
| Real-time | WebSocket через Fiber; NATS JetStream only | net/http/gorilla для WS; Redis pub/sub |
| Storage | SeaweedFS через aws-sdk-go-v2; presigned URLs; max 25MB | minio-go; seaweedfs:latest |
| Frontend | TanStack Router; Tailwind 4 + Shadcn/ui | замена роутера/UI-фреймворка; `!important`; index как key |
| Infra | Docker Compose (dev), K3s (prod); Caddy; GitHub Actions | nginx; DEV_MODE=true в prod CI |
| Security | webhook signature check; WS RBAC; search user-scope; ValidateOutboundURL | skip подписи; WS без membership; search без scope |

Любое отклонение от mandatory или ввод forbidden — только через ADR с явным согласованием.
Без ADR `arch-ver` ставит FAIL.

### 5.7 Атрибуты качества (нефункциональные)

ARCH обеспечивает, что архитектура покрывает NFR из `system_specification.md`. Ключевые
атрибуты и их реализация в Hub:

| Атрибут | Как обеспечивается |
|---------|--------------------|
| Безопасность | Auth middleware до сервиса; RBAC на room/endpoint; webhook signature; SafeHTTPClient + ValidateOutboundURL; security headers (CSP/HSTS); секреты вне репо |
| Производительность | p95 < 500ms (NFR-01); индексы под фильтры; пагинация; кеш Redis; presigned upload (без backend buffer) |
| Масштабируемость | Модульный монолит; NATS JetStream для async; горизонтальное масштабирование stateless backend; WS 50+ conn/instance |
| Надёжность | Outbox pattern (гарантия доставки событий); optimistic lock; бэкапы PG+Mongo+S3; graceful degradation |
| Сопровождаемость | Clean Architecture (handler→service→repository); Repository pattern; атомарные изменения; docs sync |
| Наблюдаемость | request_id в ошибках; slog с контекстом; метрики (планово Prometheus) |

Шаблон анализа атрибута:

```
Атрибут: производительность (NFR-01 p95<500ms)
Сценарий качества: 1000 конкурентных листингов задач
Тактика: индекс (project_id, deleted_at); пагинация limit≤100; Redis-кеш прав RBAC
Проверка: нагрузочный тест / профиль; фиксация в COVERAGE/test_plan
Риск: N+1 при загрузке связей → eager load в repository
```

### 5.8 Tech debt registry

ARCH ведёт реестр технического долга (TD-XX) в architecture.md: каждый компромисс, отход от
идеала фиксируется с evidence (file:line). Образец из проекта:

```
| ID | Summary | Key evidence |
|----|---------|--------------|
| TD-02 | 7 packages без Repository | auth, dashboard, backup, imex, storage, apigateway, rbac |
| TD-03 | GetDB() escape hatch | tracker/repository.go:153 |
| TD-06 | security headers | (закрыт: main.go:190) |
```

TD связывается с ADR (если долг — следствие сознательного решения) и с ROADMAP (когда
планируется погашение).

### 5.9 Handoff ARCH

После обновления architecture.md/ADR ARCH:
1. Описание решений простым языком (2–5 предложений).
2. Обновляет `stack_and_tools.md` (если стек), `.planning/STATE.md`, лог в `ARCH_skill.md`.
3. Сверяет каждое число/факт с кодом (grep/go.mod/routeTree) — во избежание stale-claims.
4. `verified: false` → `arch-ver`.
5. Handoff BE/FE с указанием паттернов, границ модулей, ADR-ограничений.

---
