# A3 W6 агент A1 — auth core (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`)  
> Baseline: [A2W6-1-auth-core.md](A2W6-1-auth-core.md) · TASK-102 / TASK-116  
> Периметр: `W6-ASSIGNMENTS.md` §A1 (30 путей) · теги: NEW | REGRESS | RESIDUAL | FIXED-OK

## Сверка с A2 (A2W6-A1-*)

| A2 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| A2W6-A1-001 | FIXED-OK | закрыт | `telegram_bot.go` удалён (TASK-102); denial `TestAuth_NoTelegramLoginSurface_Task102` |
| A2W6-A1-002 | FIXED-OK | закрыт | `telegram.go` / `VerifyTelegramAuth` удалены (TASK-102) |
| A2W6-A1-003 | FIXED-OK | закрыт | `validateProfileAvatarURL` — `service_profile.go:57-80` + whitelist-тест |
| A2W6-A1-004 | FIXED-OK | закрыт | `ListSessions_MarksCurrent` decode + ровно один `is_current` = JWT sid |
| A2W6-A1-005 | FIXED-OK | закрыт | `events.go:23` — `Method` comment: `magic_link only (ADR-AUTH-01)` |
| A2W6-A1-006 | FIXED-OK | закрыт | `GetUserByTelegramID` отсутствует в `Repository` / `EntRepository` |

Stage A closes (W6-A1-003…010 из A2-таблицы): registration fail-closed, DTO validate, avatar, sessions assert, org invite token, onboarding tx — **без регрессии**.

**REGRESS:** 0 · **RESIDUAL (открытые):** 0 · **NEW:** 0

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| backend/internal/auth/dto.go | 108 | 0 | FIXED-OK (validate tags) |
| backend/internal/auth/events.go | 25 | 0 | FIXED-OK (TASK-102 Method) |
| backend/internal/auth/export_hooks.go | 93 | 0 | чисто (hubtest exports) |
| backend/internal/auth/handler.go | 590 | 0 | чисто (Magic Link only routes + ws-ticket) |
| backend/internal/auth/magiclink.go | 175 | 0 | чисто (rate 3/10m, CTC, fails HI-04) |
| backend/internal/auth/middleware.go | 150 | 0 | чисто (fail-closed status) |
| backend/internal/auth/repository.go | 463 | 0 | FIXED-OK (нет GetUserByTelegramID) |
| backend/internal/auth/resend.go | 87 | 0 | чисто (LimitReader 10MB) |
| backend/internal/auth/service.go | 130 | 0 | чисто (AUTH-01 comment) |
| backend/internal/auth/service_login.go | 109 | 0 | чисто (Magic Link only) |
| backend/internal/auth/service_organization.go | 113 | 0 | чисто (pending 7d) |
| backend/internal/auth/service_profile.go | 106 | 0 | FIXED-OK (avatar whitelist) |
| backend/internal/auth/service_registration.go | 100 | 0 | FIXED-OK (fail-closed invite) |
| backend/internal/auth/service_session.go | 323 | 0 | чисто (JWT/refresh Redis, CTC hash) |
| backend/internal/auth/telegram.go | — | 0 | **ABSENT** FIXED-OK (TASK-102 delete) |
| backend/internal/auth/telegram_bot.go | — | 0 | **ABSENT** FIXED-OK (TASK-102 delete) |
| backend/testframework/auth/handler_admin_extra_test.go | 45 | 0 | чисто |
| backend/testframework/auth/handler_admin_test.go | 162 | 0 | чисто |
| backend/testframework/auth/handler_extra_test.go | 154 | 0 | FIXED-OK (`is_current`) |
| backend/testframework/auth/handler_integration_test.go | 169 | 0 | чисто |
| backend/testframework/auth/handler_invite_coverage_test.go | 62 | 0 | чисто |
| backend/testframework/auth/handler_invite_test.go | 120 | 0 | чисто |
| backend/testframework/auth/handler_magiclink_runes_test.go | 81 | 0 | чисто (rune-safe) |
| backend/testframework/auth/handler_organization_test.go | 125 | 0 | чисто |
| backend/testframework/auth/handler_profile_test.go | 160 | 0 | чисто |
| backend/testframework/auth/handler_test.go | 147 | 0 | чисто |
| backend/testframework/auth/handler_wave115_test.go | 79 | 0 | чисто |
| backend/testframework/auth/magiclink_test.go | 365 | 0 | чисто |
| backend/testframework/auth/middleware_failclosed_test.go | 92 | 0 | чисто |
| backend/testframework/auth/middleware_test.go | 196 | 0 | чисто |

Путей в assignment: **30**. Учтено: **30** (28 на диске + 2 удалённых). Пропущено: **0**.

> Примечание: в пакете появился `ws_ticket.go` (вне списка A1; ADR-WS-JWT-HANDSHAKE) — handler монтирует `POST /auth/ws-ticket`; глубокий аудит ticket/redeem — смежный слот/WS.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|----------|-------------------|------------|
| — | — | — | — | Нет активных находок | A2 open (4 RESIDUAL) закрыты TASK-102; TZ-инварианты держатся | — |

## Фокус TZ

| Инвариант | Статус | Evidence |
|-----------|--------|----------|
| Magic Link only (HTTP) | OK | `handler.go:37-50` — только magic-link / refresh / protected; нет Telegram login routes |
| JWT 15m / refresh Redis 30d | OK | defaults `config.go:186-187`; session Redis `service_session.go`; тесты `15*time.Minute` / `30*24*time.Hour` |
| Rate 3/10min | OK | `magicLinkRateMax=3`, `magicLinkRateTTL=10m`; wiring `AuthRateLimit` Max=3 / 10m |
| ConstantTimeCompare | OK | code: `magiclink.go:141`; refresh hash: `service_session.go:233` |
| Rune-safe codes | OK | `handler.go:126` `utf8.RuneCountInString`; `handler_magiclink_runes_test.go` |
| Telegram NOT login | OK | файлы login удалены; `service.go` / `events.go` ADR-AUTH-01; denial tests TASK-102 (соседний A2-слот) |

## Чисто

- Routes: Magic Link send/verify, refresh, logout/me/sessions/ws-ticket; admin registration/pending/org invite — без Telegram login.
- Layering: handler → AuthService / субсервисы → Repository; `*ent.Client` не в handler.
- Security: masked email в логах; Resend fail → 500; one-shot code Del fail-closed; middleware DB status fail → 503.
- Org: invite_only default, pending 7d purge hook в wiring (вне слота), founder/admin admin-gate.
- Avatar: `ValidateOutboundURL` + http(s) / relative `/…`.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 0 |
| P3 | 0 |
| **Total open** | **0** |

| Tag | n |
|-----|--:|
| NEW | 0 |
| REGRESS | 0 |
| RESIDUAL (open) | 0 |
| FIXED-OK (A2 closes + closes в этом проходе) | 6 (A2W6-A1-001…006) + Stage A closes |

A2 baseline open Σ **4** → все **FIXED-OK** · REGRESS **0** · NEW **0**.
