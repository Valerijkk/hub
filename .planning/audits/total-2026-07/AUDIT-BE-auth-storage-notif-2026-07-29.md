# File-auditor — auth / storage / notification

> TASK-008 · File-auditor · READ-ONLY · 2026-07-29  
> Модель: `cursor-grok-4.5` · ветка `cursor-grok-4.5`  
> HEAD: `617af22372be61b98e7ea63d825b35075f495588` (`617af223`)  
> Периметр: **все** production `.go` (не `*_test.go`) в  
> `backend/internal/auth/` · `backend/internal/storage/` · `backend/internal/notification/`  
> Фокус: timing compare · rate limit · SSRF · soft-delete · secrets · layering (`*ent.Client`)  
> Выход: **только этот файл** · коммит **не** делать · **TASK-008 остаётся открытым**

Связанные A3 (не дублировать закрытие TASK-008): [A3W6-1-auth-core.md](A3W6-1-auth-core.md) · [A3W7-1-notification-mail.md](A3W7-1-notification-mail.md) · [A3W7-4-storage-dash-backup.md](A3W7-4-storage-dash-backup.md)

## Прочитанные файлы (32/32)

### auth (15)

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/internal/auth/dto.go` | 92 | 0 | чисто |
| `backend/internal/auth/events.go` | 21 | 0 | чисто — `magic_link only` |
| `backend/internal/auth/export_hooks.go` | 63 | 0 | hubtest-only (`//go:build hubtest`) |
| `backend/internal/auth/handler.go` | 524 | 0 | Magic Link only + ws-ticket; rune-safe code |
| `backend/internal/auth/magiclink.go` | 155 | 0 | CTC + rate 3/10m + HI-04 fails |
| `backend/internal/auth/middleware.go` | 130 | 0 | fail-closed status; Redis cache |
| `backend/internal/auth/repository.go` | 410 | 0 | `*ent.Client` только в EntRepository |
| `backend/internal/auth/resend.go` | 75 | 0 | timeout client; LimitReader 10MiB |
| `backend/internal/auth/service.go` | 108 | 0 | orchestrator → repo interface |
| `backend/internal/auth/service_login.go` | 96 | 0 | Magic Link only (ADR-AUTH-01) |
| `backend/internal/auth/service_organization.go` | 96 | 0 | pending 7d; invite token |
| `backend/internal/auth/service_profile.go` | 96 | 0 | avatar `ValidateOutboundURL` |
| `backend/internal/auth/service_registration.go` | 87 | 0 | fail-closed invite_only |
| `backend/internal/auth/service_session.go` | 299 | 0 | JWT/refresh Redis; CTC hash |
| `backend/internal/auth/ws_ticket.go` | 68 | 0 | one-shot GetDel; 32B entropy |

### storage (6)

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/internal/storage/export_hooks.go` | 65 | 0 | hubtest-only |
| `backend/internal/storage/filename.go` | 32 | 0 | path traversal sanitize |
| `backend/internal/storage/handler.go` | 253 | 0 | CR-04 RBAC; SoftDelete→404 (A3W7-4-004 closed) |
| `backend/internal/storage/orphan_cleanup.go` | 75 | 0 | drain + stop-on-zero |
| `backend/internal/storage/repository.go` | 191 | 0 | SoftDelete `DeletedAtIsNil`; scope helpers |
| `backend/internal/storage/service.go` | 324 | 0 | aws-sdk-go-v2; NF-09; entity scope |

### notification (11)

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/internal/notification/consumer.go` | 655 | 0 | repoLookupErr Nak/Ack; без `*ent.Client` |
| `backend/internal/notification/dto.go` | 62 | 0 | чисто |
| `backend/internal/notification/events.go` | 92 | 0 | TypeNewLogin = notify, не login |
| `backend/internal/notification/export_hooks.go` | 178 | 0 | hubtest-only |
| `backend/internal/notification/handler.go` | 348 | 0 | owner-scope; NF-07→400 |
| `backend/internal/notification/handler_telegram_bind.go` | 121 | 0 | bind/confirm/unlink/test/status only |
| `backend/internal/notification/push.go` | 265 | 1 | NF-07 OK; **`*ent.Client` в PushSender** |
| `backend/internal/notification/repository.go` | 312 | 1 | `*ent.Client` OK (repo); GetIssue без soft-delete |
| `backend/internal/notification/service.go` | 357 | 0 | prefs/quiet/dispatch; repo interface |
| `backend/internal/notification/telegram.go` | 120 | 0 | rate ~28/s; ctx cancel |
| `backend/internal/notification/telegram_bind.go` | 313 | 1 | Telegram ≠ login; **нет rate limit на link-code** |

Файлов в периметре: **32**. Прочитано: **32**. Пропущено: **0**.

## Находки (активные)

| ID | Severity | Файл:строка | Фокус | Проблема | Почему дефект | Как чинить |
|----|----------|-------------|-------|----------|---------------|------------|
| ASN-001 | P2 | `notification/push.go:53–71`, `:130–267` | layering | `PushSender` держит `db *ent.Client` и делает Query/Update/Delete напрямую | TZ / `10-style` / `20-backend`: `*ent.Client` запрещён в Service (и аналогах service-слоя). Handler → Service → Repository нарушен; consumer уже вынесен через `Repository`, push — нет | Вынести push CRUD в `Repository` (или `PushSubscriptionRepository`); `PushSender` принимает интерфейс; mutation-тест: service без ent-импорта |
| ASN-002 | P2 | `notification/telegram_bind.go:107–131`; `handler_telegram_bind.go:46–64` | rate limit | `ConfirmLinkCode`: 6-значный код = Redis-ключ, **без** счётчика неудач / per-user rate | В отличие от magic-link (HI-04 + Redis verify RL), brute-force кода привязки идёт под JWT по `/api/notifications/…` (глобальный API 100/с). Окно TTL 5 мин ≈ десятки тысяч попыток → реалистичный захват чужого pending `/link` | Redis fails-counter (как `magiclink:fails:`) или RateLimitByUser на `POST …/telegram/confirm`; тест: >N fail → 429 |
| ASN-003 | P3 | `notification/repository.go:203–216` | soft-delete | `GetIssueAssignee` → `Issue.Get` **без** `DeletedAtIsNil` | Soft-deleted issue всё ещё резолвится → consumer может слать notify по «мёртвой» задаче (late NATS / race с корзиной). `ListIssuesWithUpcomingDeadline` фильтр уже есть (`:255`) — асимметрия | `Where(issue.IDEQ, issue.DeletedAtIsNil()).Only` / QueryAssignee; not-found → nil как сейчас |
| ASN-004 | P3 | `notification/repository.go:218–229` | soft-delete | `GetChannel` / `GetChannelMembers` без `channel.DeletedAtIsNil` | Soft-deleted канал (messenger trash) может продолжать порождать call/message notify, пока живёт FK | Фильтр `DeletedAtIsNil` + Exist membership с тем же predicate, что storage/messenger (NF-02) |
| ASN-005 | P3 | `storage/service.go:269–270` | soft-delete / RBAC | `voice` + `entityID=="temp"` обходит channel membership | Сознательный FE-путь (voice-recorder), но confirm без membership до send — окно для confirm «чужого» voice-файла в project при наличии `PermFileUpload` | Сузить: только uploader + pending TTL; или требовать membership сразу, temp только в FE state без confirm; regression denial |

**REGRESS относительно A3W6/A3W7 (фокус-инварианты):** 0  
**A3W7-4-004 (SoftDelete→500):** FIXED-OK на HEAD (`handler.go:278–282`).

## Фокус-чеклист (сводка)

| Инвариант | auth | storage | notification |
|-----------|------|---------|--------------|
| Timing compare (`subtle.ConstantTimeCompare`) | **OK** — code (`magiclink.go:141`); refresh hash (`service_session.go:233`) | N/A (нет secret compare) | N/A для Redis-key tokens; CTC не требуется для high-entropy key lookup. Link-code — см. ASN-002 (rate, не `==`) |
| Rate limit | **OK** — Redis 3/10m email (`magicLinkRateMax`); HI-04 verify fails=5; wiring `AuthRateLimit` + verify email RL 10/10m (`wiring_auth.go`, вне пакета, подтверждено) | **OK** — `FileUploadRateLimit` 10/мин на `/files` (wiring) | Push 50 ms; Telegram 35 мс. **Gap:** telegram confirm — ASN-002 |
| SSRF | **OK** — avatar `ValidateOutboundURL` (`service_profile.go:69`); Resend = fixed `api.resend.com` + timeout (не user URL) | N/A outbound user URL; S3 endpoint из config | **OK** NF-07 — `ValidateOutboundURL` на save+send; `safehttp.NewSafeHTTPClient` вместо `DefaultClient` (`push.go`) |
| Soft-delete | N/A sessions/users (hard revoke / purge pending 7d — ORG-06) | **OK** — `DeletedAtIsNil` на active paths; SoftDelete idempotent→NotFound; orphan hard-delete pending | Inbox: hard delete по дизайну (схема без `deleted_at`). **Gap:** issue/channel lookup — ASN-003/004 |
| Secrets | **OK** — JWT/Resend из config; refresh SHA-256 в Redis+DB; OTP не в лог (maskEmail); export_hooks только hubtest | **OK** — S3 keys из `config.S3Config`, не хардкод | **OK** — VAPID/bot token из config; nil-safe disable |
| Layering (`*ent.Client`) | **OK** — Client только в `EntRepository`; Handler/Service на `Repository` | **OK** — Client только в `EntRepository`; Service на `Repository` | **Gap** — ASN-001 `PushSender`; Handler/Service/Consumer — OK |

## Чисто (держат TZ)

- Magic Link only HTTP routes; Telegram login surface отсутствует в auth.
- Telegram notification: только bind/confirm/unlink/test/status + bot `/start bind_*`, `/link`, `/unlink`, `/test`.
- JWT access через `ParseAccessToken` (HMAC); refresh в Redis с instant Del; single-session invalidate.
- WS ticket: 32 байта entropy, TTL 30s, atomic `GetDel`.
- Storage: SeaweedFS path-style `aws-sdk-go-v2`; presign PUT + обязательный `ContentLength` 1…25 MiB (NF-09); HTML/SVG/JS content-type reject; CR-04 project RBAC на upload/confirm/download/delete.
- Filename sanitize против path traversal.
- Нет `http.DefaultClient`, нет голого `io.ReadAll` без `LimitReader`, нет `panic()` в prod-paths периметра.
- Нет hardcoded credentials в этих пакетах.

## Сводка counts

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 2 |
| P3 | 3 |
| **Total open** | **5** |

| Tag | n |
|-----|--:|
| NEW | 5 (ASN-001…005) |
| REGRESS | 0 |
| FIXED-OK (ранее A3 в периметре) | A3W7-4-004 SoftDelete→404; CTC/rate/NF-07/NF-09 как в A3W6-1 / A3W7-1 / A3W7-4 |

## Рекомендации оркестратору

1. **Не закрывать TASK-008** по этому отчёту — file-auditor только evidence.
2. Remediation-кандидаты: ASN-001 (layering push) + ASN-002 (telegram confirm RL) — приоритет P2.
3. ASN-003/004 — мелкие soft-delete consistency fixes в `notification/repository.go`.
4. ASN-005 — продуктовое решение (оставить temp + сузить / убрать).

```
reason: file-auditor READ-ONLY; 32/32 go read; 5 findings (2×P2); TASK-008 stays open; no commit
```
