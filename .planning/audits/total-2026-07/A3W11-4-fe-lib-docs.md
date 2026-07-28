# A3 W11 агент 4 — fe lib+docs part 4 (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline A2: [A2W11-4-fe-lib-docs.md](A2W11-4-fe-lib-docs.md)  
> Stage A: [A11-4-fe-lib-docs.md](A11-4-fe-lib-docs.md)  
> Периметр: 19 файлов (W11-ASSIGNMENTS A4) · Прочитано: 19 · Пропущено: 0

## Сверка с A2 (A2W11-4-*)

| A2 ID | Tag (A2) | Статус в HEAD | Доказательство |
|-------|----------|---------------|----------------|
| A2W11-4-001 | FIXED → FIXED-OK | закрыт | `docs/BUSINESS.md:189` голос+webcam shipped; `:443` Out of Scope = polish; `:462` Phase 16+ без webcam-as-blocker; `CallControls.tsx:143,152` `setCameraEnabled` |
| A2W11-4-002 | FIXED → FIXED-OK | закрыт | `use-websocket.ts` — только `useWebSocket` / `useWsEventListener`; `rg useRooms` по `frontend/` = 0; Vitest `i18n-push-rooms-p2` |

### Cascade FIXED-OK (Stage A W11-A4-* / B–C)

| Stage A ID | Evidence в HEAD |
|------------|-----------------|
| W11-A4-001 | `use-unread.ts:23-28` → `setUnreadCounts(query.data)`; store `setUnreadCounts: (counts) => set({ unreadCounts: { ...counts } })` |
| W11-A4-002 | `architecture.md:56` thin `main.go` (~34) + `wiring_*.go` |
| W11-A4-003 | `architecture.md:127-137` порядок = `wiring_fiber.go:49-76` (RequestID→Recover→Trace→…→Maintenance→Audit) |
| W11-A4-004 | `use-unread.ts:44-54` rollback `updateUnread(channelId, previous)` |
| W11-A4-005 | `format-time.ts` — `Intl.RelativeTimeFormat` + locale ru\|en |
| W11-A4-006 | future: `past = diffMs >= 0` → «через N…» / `in N…` |
| W11-A4-007 | dead `useRooms` удалён (≡ A2W11-4-002) |
| W11-A4-008 | `PUSH_DISMISS_TTL_MS` = 7d; `shouldShowPushBanner` TTL-check |
| W11-A4-009 | TD-01 DONE; evidence на `wiring_*` |
| W11-A4-010 | webcam shipped wording (≡ A2W11-4-001) |
| W11-A4-011 | SSO/LDAP vs social OAuth разделены (`BUSINESS.md:438-439`) |
| W11-A4-012 | `code_style_guide.md:17` — комментарии RU |
| W11-A4-013 | деревья тестов → `backend/testframework/` + `testframework/stages/…` |
| W11-A4-014 | Magic Link + Telegram notify bind (`competitive_analysis_2026.md:150`) |
| W11-A4-015 | SSO after ADR (`competitive_audit.md:127`) |
| W11-A4-016 | `extractApiError`: `code` number \|\| `request_id` string; Vitest `type-casts-p3` |
| W11-A4-017 | routes metric **77** `.tsx` (`architecture.md:105`, `code_style_guide.md:154`; факт `(Get-ChildItem routes -Filter *.tsx).Count` = **77**) |

**REGRESS:** 0 · **RESIDUAL (A2 actives):** 0 · **NEW (этот слот):** 3

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `frontend/src/hooks/use-unread.ts` | 58 | 0 | FIXED-OK replace map + rollback |
| `frontend/src/hooks/use-websocket.ts` | 71 | 0 | FIXED-OK dead `useRooms` gone |
| `frontend/src/lib/onboarding/profile-schema.ts` | 57 | 0 | чисто |
| `frontend/src/lib/onboarding/project-schema.ts` | 59 | 0 | чисто |
| `frontend/src/lib/format-api-error.ts` | 106 | 0 | FIXED-OK code/request_id guard |
| `frontend/src/lib/format-time.ts` | 48 | 0 | FIXED-OK i18n + future |
| `frontend/src/lib/is-system-admin.ts` | 4 | 0 | чисто (admin\|founder) |
| `frontend/src/lib/keyboard.ts` | 11 | 0 | чисто |
| `frontend/src/lib/permissions.ts` | 56 | 0 | чисто — 26 бит = `rbac/permissions.go` iota 0…25 |
| `frontend/src/lib/push-notifications.ts` | 177 | 0 | FIXED-OK TTL 7d |
| `frontend/src/lib/sanitize-color.ts` | 8 | 0 | чисто `#RRGGBB` |
| `frontend/src/lib/utils.ts` | 21 | 0 | чисто |
| `docs/architecture.md` | 239 | 0 | FIXED-OK thin main + middleware + **77** routes |
| `docs/BUSINESS.md` | 465 | 0 | FIXED-OK webcam/SSO (TASK-123) |
| `docs/code_style_guide.md` | 290 | 0 | FIXED-OK RU + TF paths + **77** routes / **12** stores |
| `docs/competitive_analysis_2026.md` | 198 | 0 | FIXED-OK Telegram notify + webcams ✅ |
| `docs/competitive_audit.md` | 175 | 2 | NEW — gap matrix webcam + stale HEAD |
| `docs/GLOSSARY.md` | 159 | 0 | чисто |
| `docs/README.md` | 97 | 1 | NEW — inventory `618` vs факт |

Файлов в списке: **19**. Прочитано: **19**. Пропущено: **0**.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A3W11-4-001 | NEW | P2 | `docs/README.md:84` | Команда/комментарий: «все **618** тестов в `backend/testframework/`, measured 2026-07-28» | Факт HEAD: **759** `*_test.go` под `backend/testframework/`. Соседние каноны A5 (`TROUBLESHOOTING.md:48`, `test_plan.md:52`) уже **652** (тоже устарели vs 759) — входной `docs/README` вводит агентов в ложный inventory | Пересчитать `Get-ChildItem … *_test.go` → одна цифра во `README` (+ синхрон с `test_plan`/`TROUBLESHOOTING` в A5) |
| A3W11-4-002 | NEW | P3 | `docs/competitive_audit.md:36` | Gap Calls: «webcams — Phase 16 IN» | Противоречит тому же файлу `:132` (webcam UI shipped) + `BUSINESS`/TASK-123 + `CallControls.setCameraEnabled` | Заменить на «webcams ✅ shipped; polish — backlog» (как competitive_analysis §3.4) |
| A3W11-4-003 | NEW | P3 | `docs/competitive_audit.md:22` | Таблица позиции: ветка `master` · HEAD `bbdaf16` | SHA/ветка не канон аудита (`ee353287` / `cursor-grok-4.5`); устаревший якорь для onboarding/аудитов | Убрать конкретный SHA **или** обновить на tip + дату; не претендовать на live HEAD в competitive doc |

## FIXED-OK (сверка фокуса слота)

| Зона | Проверено в HEAD |
|------|------------------|
| Unread sync | Полная замена map + markRead rollback; mutation-тест `reply-thread-unread.test.ts` |
| WS hook | Нет dead export; connect on auth / disconnect cleanup |
| format-api-error | Guard + rate-limit countdown; `type-casts-p3` |
| format-time / push | i18n RTF + future; TTL 7d — `i18n-push-rooms-p2` |
| permissions | FE bigint 0…25 ≡ backend `1 << iota` (26) |
| Zod onboarding | min/max/trim; project key `^[A-Z0-9]+$` |
| architecture | thin main, middleware order, 77 routes, Ent **82** |
| BUSINESS / competitive_analysis | Magic Link only; Telegram bind; webcam shipped; SSO after ADR |

## Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 1 |
| P3 | 2 |
| **Total** | **3** |

| Метрика | Значение |
|---------|----------|
| Файлов | **19 / 19 / 0** |
| Активных | **3** (NEW 3 · REGRESS 0 · RESIDUAL 0) |
| FIXED-OK (A2 actives) | **2/2** |
| FIXED-OK (Stage A via A2) | **17/17** без регресса |

### Топ-3

1. **A3W11-4-001 (P2, NEW)** — `docs/README.md` inventory **618** vs факт **759** BE `*_test.go`.
2. **A3W11-4-002 (P3, NEW)** — `competitive_audit` gap «webcams Phase 16 IN» при shipped UI.
3. **A3W11-4-003 (P3, NEW)** — stale HEAD `bbdaf16` в competitive_audit.

---

Файлов в списке: 19. Прочитано: 19. Пропущено: 0.  
Counts: P0=0 · P1=0 · P2=1 · P3=2 · Total=3.  
Топ: A3W11-4-001.
