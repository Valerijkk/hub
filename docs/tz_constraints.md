# Hub — Технические ограничения (tracked mirror)

> **Текущая product-truth (founder, 2026-08-12):** Hub остаётся
> незавершённым MVP. Ни один продуктовый модуль не принят как `complete`,
> `stable` или `production-ready`. Метки `SHIPPED`/`COMPLETE`, milestone/audit
> `PASS` и строки «реализовано» ниже фиксируют только историческое
> planning/technical evidence. PG-03/FA-06 — отдельный manual ops-блокер
> production cutover, а не индикатор готовности MVP.
>
> **Канон для агентов (локально):** `.cursor/rules/60-tz-constraints.mdc`
> **Обновлено:** 2026-08-12 · **ADR:** `.planning/decisions/ADR-TZ-REWRITE-2026.md` · пагинация задач: `ADR-PAGINATION-CURSOR-ISSUES-2026.md`

Документ фиксирует то, что **нельзя менять** без явного согласования с PO/архитектором, и то, что **строго запрещено** в коде и инфраструктуре.

---

## ОБЯЗАТЕЛЬНО (MANDATORY)

### Аутентификация

| Правило | Детали |
|---------|--------|
| Единственный вход | **Magic Link (email)** — без паролей (ADR-AUTH-01). SSO (OIDC/LDAP) — дополнительный корпоративный метод после `ADR-AUTH-SSO` (2026-07-03); Magic Link остаётся дефолтом и break-glass |
| Telegram | **Только** привязка и доставка уведомлений, **не** login |
| JWT access | TTL **15 минут** |
| Refresh token | Redis, TTL **30 дней**, мгновенный отзыв |
| Rate limit auth | **3 попытки / 10 минут** (по IP) |

### Организация (v2.5+)

| Правило | Детали |
|---------|--------|
| Модель | Таблица `organization`; **1 org на инстанс** (multi-org → v5) |
| Регистрация | По **invite**; очередь **Pending** у admin; **7 дней** без approve → удаление |
| Founder | `system_role founder` — **один** на org |

### Слой данных

| Правило | Детали |
|---------|--------|
| Время | UTC в БД всегда; фронтенд конвертирует в TZ пользователя |
| Optimistic locking | Поле `version` в каждой изменяемой Ent-схеме |
| Primary keys | UUID v4; **строковые ID запрещены** |
| Удаление | Soft delete через корзину (**30 дней**); hard delete — только Admin/GDPR |

### API

| Правило | Детали |
|---------|--------|
| Ошибки | `{"error": "message", "request_id": "uuid"}` |
| Пагинация | Задачи (list-issues) — **cursor** + limit (max 100); чат — **cursor** (infinite scroll). Offset/limit — для admin/intranet/wiki списков, не для list-issues (`ADR-PAGINATION-CURSOR-ISSUES-2026`) |
| Rate limiting | API **100/сек**, файлы **10/мин**, сообщения **60/мин** |

### Real-time

| Правило | Детали |
|---------|--------|
| WebSocket | Через **Fiber (fasthttp)** — не net/http, не gorilla |
| Pub/sub | **NATS JetStream** — единственный; Redis pub/sub **не использовать** |

### Файловое хранилище

| Правило | Детали |
|---------|--------|
| Backend | **SeaweedFS** (S3-compatible) через `aws-sdk-go-v2` — не minio-go |
| Загрузка | Presigned URLs (не через буфер backend) |
| Лимит | Max **25 MB**: Fiber BodyLimit + presigned PUT через обязательный `size` и `ContentLength` bind (ADR-STORAGE-NF09-SIZE-2026; NF-09 closed) |
| Версия образа | `chrislusf/seaweedfs:3.97` (pinned), не `latest` |

### Frontend

| Правило | Детали |
|---------|--------|
| Роутер | **TanStack Router** — единственный |
| Стили | **Tailwind CSS 4** + **Shadcn/ui** — не заменять на другой UI framework |
| Shadcn | Только изменения `className`; поведение Radix **не трогать** |
| CSS | **Без `!important`** (Phase 8 очистил 56 вхождений — не добавлять новые) |

### Инфраструктура

| Правило | Детали |
|---------|--------|
| Dev | **Docker Compose** (`docker-compose.local.yml`, `deployments/*.yml`) |
| Prod | **K3s** (planned) — не менять без решения devops |
| Reverse proxy | **Caddy** — не nginx |
| CI | **GitHub Actions** — лимит ~3000 мин/мес (студенческий пакет); экономить |

---

## ЗАПРЕЩЕНО (FORBIDDEN)

### Фичи вне scope (до отдельного ADR)

- ⚠️ SSO / LDAP — **одобрено в scope (2026-07-03)**; реализация только после принятого `ADR-AUTH-SSO` (OIDC + LDAP/AD bind; org-политика magic-link/sso/both с break-glass Magic Link для founder; JIT + LDAP-sync; без собственного MFA — делегируется IdP; маппинг групп только на system role). OAuth-провайдеры соцсетей — по-прежнему вне scope
- Multi-tenant multi-org (один инстанс = одна org; v5 — multi-org)
- Native App Store / Play Store (PWA достаточно)
- SaaS billing / подписки (milestone Later)

### Антипаттерны в коде

| Паттерн | Почему | Альтернатива |
|---------|--------|--------------|
| `http.DefaultClient` для user URLs | SSRF | `safehttp.NewSafeHTTPClient` |
| `io.ReadAll` без лимита | Memory DoS | `io.LimitReader` |
| `!=` для секретов | Timing attack | `subtle.ConstantTimeCompare` |
| `*ent.Client` в Service/Handler | Нарушение layering | Repository interface |
| `panic()` в production paths | Crash | return error |
| Hardcoded credentials в репо | Утечка | env / GitHub Secrets |
| `DEV_MODE=true` в prod CI/CD | Dev routes в prod | `-tags release` |
| `seaweedfs:latest` | Непредсказуемые обновления | pinned `3.97` |

### Security non-negotiables

- Webhook signature: **нельзя skip** даже при отсутствии claim (SEC-02)
- WS room subscription: **обязательна** проверка membership/RBAC (NF-02)
- Search: **обязателен** user-scope фильтр (NF-03)
- WebPush endpoint: **обязателен** `ValidateOutboundURL` (NF-07)

---

## IN SCOPE по фазам (PO 2026-06-21)

Ранее запрещённые фичи, разрешённые по roadmap:

| Фича | Фаза / историческое implementation evidence (2026-07-28) |
|------|------|
| Gantt | UI surface зафиксирован (`frontend/.../gantt.lazy.tsx`, `GanttChart`), но модуль не принят; остаток scope — backlog |
| Webcam в звонках | UI surface зафиксирован (`CallControls` / LiveKit `setCameraEnabled`), но модуль не принят; polish — backlog |
| Transactional email + marketing (hub-mail) | 15.2 / 16 |
| Full JQL automation | 17 |
| AI gateway BYOK + Ollama | historical AI Hub v3.0 capability slice; модуль не принят, BYOK/Ollama depth — Phase 17 остаток |
| Intranet, Calendar, CRM, HR, accounting, no-code DB | historical v4 code/UI surfaces; ни один модуль не принят |

---

## Email (политика)

| Тип | Назначение |
|-----|------------|
| Magic link | Аутентификация (обязательно) |
| Transactional alerts | Opt-in уведомления — ADR-NOTIF-01 |
| Marketing Mailchimp-lite | Admin-only, v2.6 — ADR-TZ-REWRITE-2026 |

Микросервис: `backend/cmd/hub-mail` — internal `POST /v1/send`.

---

## Верификация против репозитория

| Источник | Проверено |
|----------|-----------|
| Go | `backend/go.mod` → `go 1.25.0` |
| CI Go | `.github/workflows/ci.yml` → `go-version: '1.25'` |
| CI Node | `node-version: '22'` |
| Docker images | `deployments/docker-compose.prod.yml`, `docker-compose.local.yml` |
| Auth routes | `backend/internal/auth/handler.go` — только `/auth/magic-link*` |
| Message rate | `platform/middleware/ratelimit.go` → 60/min |

---
*Последнее обновление: 2026-06-21 — синхронизация с ADR-TZ-REWRITE-2026*
