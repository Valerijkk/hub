# A2 W11 агент 4 — fe lib+docs part 4 (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Зеркало: `W11-ASSIGNMENTS.md` A4 · Stage A: `A11-4-fe-lib-docs.md`  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `frontend/src/hooks/use-unread.ts` | 57 | 0 | FIXED-OK replace map + rollback |
| `frontend/src/hooks/use-websocket.ts` | 113 | 0 | FIXED TASK-123 — dead `useRooms` удалён |
| `frontend/src/lib/onboarding/profile-schema.ts` | 53 | 0 | чисто |
| `frontend/src/lib/onboarding/project-schema.ts` | 59 | 0 | чисто |
| `frontend/src/lib/format-api-error.ts` | 92 | 0 | FIXED-OK code/request_id guard |
| `frontend/src/lib/format-time.ts` | 39 | 0 | FIXED-OK i18n + future |
| `frontend/src/lib/is-system-admin.ts` | 4 | 0 | чисто |
| `frontend/src/lib/keyboard.ts` | 11 | 0 | чисто |
| `frontend/src/lib/permissions.ts` | 53 | 0 | чисто (rbac iota mirror) |
| `frontend/src/lib/push-notifications.ts` | 158 | 0 | FIXED-OK TTL 7d |
| `frontend/src/lib/sanitize-color.ts` | 8 | 0 | чисто |
| `frontend/src/lib/utils.ts` | 21 | 0 | чисто |
| `docs/architecture.md` | 239 | 0 | FIXED-OK thin main + wiring |
| `docs/BUSINESS.md` | 465 | 0 | FIXED TASK-123 — webcam/Phase16 consistency |
| `docs/code_style_guide.md` | 290 | 0 | FIXED-OK RU comments + TF paths |
| `docs/competitive_analysis_2026.md` | 198 | 0 | FIXED-OK Telegram notify bind |
| `docs/competitive_audit.md` | 175 | 0 | FIXED-OK SSO after ADR |
| `docs/GLOSSARY.md` | 159 | 0 | чисто |
| `docs/README.md` | 97 | 0 | чисто |

Файлов в списке: **19**. Прочитано: **19**. Пропущено: **0**.

## Находки (открытые)

| ID | Tag | Severity | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|----------|-------------|----------|---------------|------------|
| — | — | — | — | Нет открытых (TASK-123) | — | — |

## FIXED (TASK-123 · 2026-07-28)

| ID | Evidence |
|----|----------|
| A2W11-4-001 | `docs/BUSINESS.md`: голос + Out of Scope + Phase 16+ — webcam UI shipped; residual = JQL/hub-mail + polish |
| A2W11-4-002 | удалён dead export `useRooms` из `frontend/src/hooks/use-websocket.ts`; Vitest `i18n-push-rooms-p2` |

## FIXED-OK (Stage A → закрыто после B/C)

| Stage A ID | Evidence |
|------------|----------|
| W11-A4-001 | `setUnreadCounts(query.data)` — полная замена map |
| W11-A4-002 | architecture: thin main ~34 + `wiring_*.go` |
| W11-A4-003 | middleware порядок → `wiring_fiber.go` |
| W11-A4-004 | markRead fail → `updateUnread(..., previous)` |
| W11-A4-005 | `Intl.RelativeTimeFormat` + i18n locale |
| W11-A4-006 | future `diffMs < 0` → «через N…» |
| W11-A4-007 | roomsKey join (остаток dead — RESIDUAL) |
| W11-A4-008 | `PUSH_DISMISS_TTL_MS` 7 дней |
| W11-A4-009 | TD/evidence на wiring_*; TD-01 DONE |
| W11-A4-010 | webcam out-of-scope переписан (частично; NEW drift) |
| W11-A4-011 | SSO/LDAP vs social OAuth разделены |
| W11-A4-012 | комментарии — русский канон |
| W11-A4-013 | деревья тестов → testframework |
| W11-A4-014 | Magic Link + Telegram notify bind |
| W11-A4-015 | competitive_audit SSO after ADR |
| W11-A4-016 | extractApiError требует code \|\| request_id |
| W11-A4-017 | routes metric обновлён в architecture/vision |

## Чисто

Zod onboarding schemas; `permissions` iota 0…25; sanitize `#RRGGBB`; GLOSSARY/README индекс.

## Сводка агента

| Метрика | Значение |
|---------|----------|
| Файлов | 19/19/0 |
| Открытых находок | **0** (TASK-123) |
| NEW | 0 |
| REGRESS | 0 |
| RESIDUAL | 0 |
| FIXED-OK | 17 + TASK-123 (2) |
