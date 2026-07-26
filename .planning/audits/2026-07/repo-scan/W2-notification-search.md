# W2 — Notification + Search (readonly)

> **Provenance class:** PRE_EXISTING · **Confidence:** MIXED · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> Агент: `security-auditor` (Grok 4.5) · Дата: **2026-07-23** · Scope: `backend/internal/notification/`, `backend/internal/search/`
> Фокус: **NF-03** user-scope search · **NF-07** WebPush `ValidateOutboundURL` · **Telegram notify-only** (не login, ADR-AUTH-01)
> Код **не менялся**.

## Чек-лист (результат)

| Контроль | Статус | Evidence |
|----------|--------|----------|
| NF-03: issues / test_cases / projects / files | PASS | `search/service.go:221-499` — фильтр `projectIDs` / `id = ANY($2)` |
| NF-03: messages (Mongo `$text`) | PASS | `search/service.go:565-582` — `channel_id ∈ getUserChannelIDs` (SEC-H2) |
| NF-03: users | PASS | `search/service.go:329-351` — shared-project scope; admin/founder bypass `searchUsersUnscoped` |
| NF-03: channels | **PARTIAL** | `search/service.go:502-555` — только `project_id`, **не** `channel_members` → private name leak |
| NF-03: saved filters CRUD | **PARTIAL** | owner check на delete/share OK; create/list **без** project membership |
| NF-07: WebPush `ValidateOutboundURL` | PASS | `notification/push.go:157-165` (send), `:218-220` (save) |
| NF-07: SafeHTTPClient / DNS rebinding | PASS | `push.go:25`, `:175-176` — `pushHTTPClient` + `SafeDialContext` |
| Telegram в notification = notify bind/delivery | PASS | `handler_telegram_bind.go`, `telegram_bind.go` — bind/confirm/unlink/test/status; нет JWT/session |
| Telegram ≠ login (HTTP surface) | PASS (модуль) | `/api/notifications/telegram/*` за `authMw`; login-роутов в notification нет |
| Auth Magic Link only (смежный) | PASS (HTTP) | `auth/handler.go:35-36` — только magic-link; `POST /auth/telegram` **не** зарегистрирован |
| Auth dead code Telegram login | INFO | `auth/service_login.go:94+`, `auth/service.go:106-113` — back-compat методы без route |
| Layering notification PushSender | FAIL (LOW) | `push.go:45` — `*ent.Client` напрямую в sender |

## Findings

| Sev | file:line | Уязвимость / наблюдение | Рекомендация | OWASP/ASVS | Статус |
|-----|-----------|-------------------------|--------------|------------|--------|
| **MEDIUM** | `search/service.go:502-555` | `searchChannels` фильтрует каналы по `project_id = ANY(projectIDs)`, **без** членства в `channel_members`. Типы `private` / `group` (schema `channel.go:22-23`) раскрывают **имя + UUID + URL** любому member проекта, даже без invite/membership. Сообщения при этом корректно scoped (ниже) — асимметрия NF-03. | Для `private`/`group`/`dm`: фильтр `id = ANY(getUserChannelIDs)`; для `public` — project scope допустим. Эталон: `searchMessages` + messenger `IsMember`. Тест: non-member private channel → пусто в `/api/search?type=channel`. | A01 Broken Access Control / ASVS V4.1 | open |
| **MEDIUM** | `search/service.go:628-660`; `search/repository.go:65-84` | `CreateSavedFilter` / `ListSavedFilters` не проверяют membership в `projectID`. Auth user с известным UUID чужого проекта может: (1) создать фильтр в чужом проекте; (2) читать **shared** фильтры (`filter_config`) чужого проекта. Delete/Share — owner OK (`:662-688`). | Перед create/list: `project membership` (или RBAC project view). List — 404/403 если не member. Тест denial как `bases/rbac_test.go`. | A01 IDOR / ASVS V4.2 | open |
| **MEDIUM** | `notification/push.go:252-261`; `handler.go:340-364` | `RemoveSubscription` / `PushUnsubscribe` удаляют по `endpoint` **без** `userID`. Любой аутентифицированный пользователь, зная/угадав endpoint чужой подписки, снимает push у жертвы (DoS уведомлений). Handler берёт `userID` и **игнорирует** его (`handler.go:341`). | `Delete().Where(EndpointEQ, UserIDEQ(userID))`; 404 если чужой. | A01 Broken Access Control / ASVS V4.2 | open |
| **MEDIUM** | `notification/push.go:230-238` | `SaveSubscription`: при существующем `endpoint` делается `SetUserID(caller)` — **перехват** чужой push-подписки. Эффект: DoS push жертвы + доставка уведомлений нового владельца на устройство старого endpoint (контент чужих нотификаций на чужой браузер). | Не менять `UserID`, если `existing.UserID != caller` → 409/403; либо удалять только свою подписку. | A01 IDOR / A04 Insecure Design | open |
| **LOW** | `notification/push.go:45`; `NewPushSender` | `PushSender` держит `*ent.Client` (нарушение layering TZ). Сопутствует IDOR выше. | Вынести PushSubscription CRUD в `Repository` interface. | A04 / Hub layering | open |
| **LOW** | `notification/telegram_bind.go:322-327`, `:108-124` | Link-код привязки — 6 цифр (`10⁶`), TTL 5 мин. Confirm за `authMw` (атакующий уже в Hub). Аналог NF-10 (accepted для magic-link); отдельного rate-limit на `/telegram/confirm` нет (только global API 100/s). | Rate-limit per-user на confirm (как auth 3/10м) или удлинить код; мониторить brute. | A07 Auth Failures / ASVS V2.2 | open (hardening; близко к accepted NF-10) |
| **INFO** | `auth/service_login.go:94+`; `auth/dto.go:20` | Остаточный `LoginWithTelegram` / `TelegramAuthRequest` без HTTP route. Риск регрессии, если кто-то снова повесит route. | Удалить dead code или явно `// deprecated ADR-AUTH-01` + тест «нет /auth/telegram». Вне scope notification/, но релевантно инварианту. | A07 / ADR-AUTH-01 | open (вне scope файлов; для W1-auth) |
| — | — | CRITICAL / HIGH по трём целевым инвариантам **ядра** (messages scope, WebPush SSRF, Telegram login в notification) — **не найдены** | — | — | — |

### Закрытые / подтверждённые инварианты (не findings)

1. **NF-07 / SEC-H3:** `ValidateOutboundURL` на регистрации и перед каждой отправкой; небезопасные записи удаляются (`push.go:157-164`).
2. **NF-07 DNS rebinding:** исходящий HTTP через `safehttp.NewSafeHTTPClient` (`push.go:25,175`), не `http.DefaultClient`.
3. **NF-03 messages:** `getUserChannelIDs` + Mongo filter `channel_id $in` (`service.go:178-196,578-582`); empty channels → early nil.
4. **NF-03 core shards:** issues/test_cases/files/projects — `projectIDs` fail-closed при `len==0`.
5. **NF-03 users:** HI-12 shared-project; `IsSystemAdmin` = founder\|admin (`repository.go:56-61`) — осознанный bypass.
6. **Search auth:** `RegisterRoutes` — Group + `authMw` (`handler.go:26-38`); userID из Locals.
7. **Notification auth:** все routes под `api.Group("/notifications", authMw)` (`handler.go:24-48`).
8. **Telegram notify-only (модуль):** polling обрабатывает только `/start bind_`, `/link`, `/unlink`, `/test` (`telegram_bind.go:202-219`); `linkTelegramID` пишет `telegram_id` профиля, **не** создаёт session. Delivery: `service.go:186-204` → `TelegramSender.Send`.
9. **ADR-AUTH-01 HTTP:** публичный auth — magic-link only (`auth/handler.go:35-36`); Telegram login route отсутствует.
10. **http.DefaultClient:** в notification только комментарий (`push.go:21`); prod client = SafeHTTPClient.

## Grep-лог (кратко)

```
ValidateOutboundURL / NewSafeHTTPClient  → push.go:25,157,218  PASS
http.DefaultClient (notification)        → comment only
searchMessages channel scope             → service.go:578-582  PASS
searchChannels membership                → project_id only     PARTIAL
CreateSavedFilter project membership     → отсутствует         PARTIAL
RemoveSubscription user scope            → endpoint only       FAIL
Telegram login в notification/           → 0 hits
Telegram bind/confirm/unlink             → handler_telegram_bind.go  PASS notify-only
LoginWithTelegram в auth                 → service only, no route  INFO
```

## Вердикт

| Метрика | Значение |
|---------|----------|
| CRITICAL | 0 |
| HIGH | 0 |
| MEDIUM | 4 (channels scope; saved-filter IDOR; push unsubscribe IDOR; push subscribe hijack) |
| LOW | 2 |
| INFO | 1 (auth dead code, вне file-scope) |

**Целевые инварианты:**
- NF-07 WebPush SSRF — **закрыт** (ядро PASS).
- NF-03 message/project search — **закрыт**; channels + saved filters — **дыры MEDIUM**.
- Telegram notify-only в `notification/` — **PASS**; HTTP login через Telegram — **не экспонирован**.

**verified: false** — 2026-07-23 01:00 UTC+3

Причина: 4× MEDIUM (A01) без mitigation/accepted. Блокеров CRITICAL/HIGH нет; для merge-ready волны желательно закрыть MEDIUM или зафиксировать accepted-risk с владельцем.

**Handoff:** `be1`/`be2` — fix searchChannels membership + saved-filter project authz + PushSubscribe/Unsubscribe user-scope; затем повторный проход `security-auditor`.
