# A3 W7 агент A5 — marketplace+status+dev (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline: [A2W7-5-mp-status-dev.md](A2W7-5-mp-status-dev.md) · A1: [A7-5-mp-status-dev.md](A7-5-mp-status-dev.md)  
> Assignment: [W7-ASSIGNMENTS.md](W7-ASSIGNMENTS.md) § A5 · Периметр: **27** · Прочитано: **27/27** · Пропущено: **0**  
> Remediation: TASK-121 (W7-A5-001…006) · TASK-057/059/060/061/063/067 (W7-A5-007…016)

## Сверка с A2 (W7-A5-*)

| A2 ID | Tag | Статус | Комментарий |
|-------|-----|--------|-------------|
| W7-A5-001 | FIXED-OK | закрыт | `/api/docs` + `/docs/doc.json` → `requireAuth` + `RequireAdmin` (`wiring_core.go:47–57`); IT `apidocs_auth_gate_task121_test.go` |
| W7-A5-002 | FIXED-OK | закрыт | публичный `/api/status` — агрегат `hub`; `/status/details` за `authMw` (`status/handler.go:19–22`, `service.go:29–49`); IT handler/service W7A5002 |
| W7-A5-003 | FIXED-OK | закрыт | `ListMine`/`ListPending`/`ListApproved` — `normalizePage` cap 100 + `has_more` (`repository.go:39–47,101–148`); IT `task121_exposure_test.go` |
| W7-A5-004 | FIXED-OK | закрыт | `UpdateApprovedListingToPending` один UPDATE fields+pending+version (`repository.go:162–175`); IT atomic remoderation |
| W7-A5-005 | FIXED-OK | закрыт | `displayName(…, allowEmail=false)` → «Пользователь» (`dto.go:129–141`); IT no-email catalog |
| W7-A5-006 | FIXED-OK | закрыт | `priceCents≥0` + currency allowlist (`service.go:369–375`, `dto.go:152–160`); IT validate price/currency |
| W7-A5-007 | FIXED-OK | закрыт | `IsSystemAdmin` err → wrap 500 (`service.go:277–280`) |
| W7-A5-008 | FIXED-OK | закрыт | stranger-delete → `ErrNotOwner` (`service.go:283–285`); HTTP «not listing owner» |
| W7-A5-009 | FIXED-OK | закрыт | title/desc `utf8.RuneCountInString` (`service.go:362–364`) |
| W7-A5-010 | FIXED-OK | закрыт | invalid category → `ErrInvalidInput` до Ent (`service.go:48–51`) |
| W7-A5-011 | FIXED-OK | закрыт | `SoftDeleteListing` `VersionEQ` + `DeletedAtIsNil` (`repository.go:214–225`) |
| W7-A5-012 | FIXED-OK | закрыт | `/api/dev/login` `StatusActive` (`handler.go:110–113`); IT task067 |
| W7-A5-013 | REGRESS | живо | outer Seed `partial`→500 есть (`seed.go:194–198`), но helpers глотают err → 200 (см. A3W7-5-002) |
| W7-A5-014 | FIXED-OK | закрыт | демо Telegram = notification bind (`seed_data.go:168–172`); IT task067 |
| W7-A5-015 | FIXED-OK | закрыт | `BuildAPIDocsJSON` marshal err → error (`apidocs.go:199–206`) |
| W7-A5-016 | FIXED-OK | закрыт | DELETE stranger HTTP denial suite (TASK-059) |

**REGRESS:** 1 (W7-A5-013 incomplete) · **A2 open residual закрыты TASK-121:** 6/6 FIXED-OK (кроме ложного close 013)

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/marketplace/dto.go` | 205 | 1 | review_note в общем DTO; PII email FIXED-OK |
| `backend/internal/marketplace/export_hooks.go` | 57 | 0 | чисто · hubtest re-exports |
| `backend/internal/marketplace/handler.go` | 247 | 0 | чисто · authMw group; mapServiceError |
| `backend/internal/marketplace/repository.go` | 254 | 0 | FIXED-OK pagination/atomic/OL soft-delete |
| `backend/internal/marketplace/service.go` | 377 | 2 | review_note leak; currency normalize gap |
| `backend/testframework/marketplace/dto_wave99_test.go` | 31 | 0 | чисто |
| `backend/testframework/marketplace/fixtures_test.go` | 74 | 0 | чисто |
| `backend/testframework/marketplace/handler_test.go` | 205 | 0 | чисто · 403 pending/member |
| `backend/testframework/marketplace/handler_wave105_test.go` | 156 | 0 | чисто |
| `backend/testframework/marketplace/handler_wave97_test.go` | 79 | 0 | чисто |
| `backend/testframework/marketplace/optimistic_lock_a6003_test.go` | 46 | 0 | чисто |
| `backend/testframework/marketplace/service_test.go` | 210 | 0 | чисто · owner/admin flows |
| `backend/testframework/marketplace/service_wave105_test.go` | 85 | 0 | чисто |
| `backend/internal/status/dto.go` | 20 | 0 | чисто |
| `backend/internal/status/handler.go` | 54 | 0 | FIXED-OK public vs details |
| `backend/internal/status/service.go` | 92 | 0 | FIXED-OK aggregate hub |
| `backend/testframework/status/handler_test.go` | 97 | 0 | чисто · W7A5002 |
| `backend/testframework/status/service_test.go` | 61 | 0 | чисто · W7A5002 |
| `backend/internal/dev/apidocs.go` | 207 | 0 | FIXED-OK marshal; gate в wiring |
| `backend/internal/dev/export_hooks.go` | 159 | 0 | чисто · hubtest |
| `backend/internal/dev/handler.go` | 129 | 0 | FIXED-OK DEV_MODE+StatusActive |
| `backend/internal/dev/handler_release.go` | 39 | 0 | чисто · ME-09 no-op |
| `backend/internal/dev/seed.go` | 788 | 1 | REGRESS swallow err в helpers |
| `backend/internal/dev/seed_data.go` | 271 | 0 | FIXED-OK ADR-AUTH-01 copy |
| `backend/testframework/dev/apidocs_wave102_test.go` | 103 | 0 | чисто |
| `backend/testframework/dev/handler_wave102_test.go` | 139 | 0 | чисто · DevMode guard |
| `backend/testframework/dev/seed_test.go` | 92 | 0 | чисто · fixtures |

Файлов в списке: **27**. Прочитано: **27**. Пропущено: **0**.

> Evidence вне периметра (не в таблице): `task121_exposure_test.go`, `apidocs_auth_gate_task121_test.go`, `login_status_task067_test.go`, `wiring_core.go` / `wiring_messenger.go` / `wiring_auth.go`.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A3W7-5-001 | NEW | P2 | `dto.go:100–126`; `service.go:57–59,203` | `toListingResponse` (витрина `ListApproved`) отдаёт `review_note` вместе с публичным DTO | Admin moderation note (может содержать PII/внутренние комментарии) виден всем auth-пользователям каталога; TASK-121 закрыл email, но не notes | В `toListingResponseOpts(allowEmail=false)` очищать `ReviewNote` (и при желании `ReviewerName`); private/admin — оставлять; regression: approve+note → public `review_note==""` |
| A3W7-5-002 | REGRESS | P3 | `seed.go:278–281,435–451,660–686` | `seedWorkflowTransitions` / `seedIssueLinks` / `seedChannels` при DB-err только `slog` + `continue`, return `(n, nil)` | W7-A5-013 помечен FIXED (outer `partial`→500), но эти шаги **не** поднимают err → Seed 200 при битых links/channels/transitions | На err: `return …, err` **или** локальный `failed` → err; mutation: force Create fail → HTTP 500 + `partial` |
| A3W7-5-003 | NEW | P3 | `service.go:369–375`; `repository.go:58–60` | `isValidCurrency` trim/upper только для check; в БД пишется сырой `req.Currency` | `"USD "` (пробел) проходит allowlist, `MaxLen(3)` Ent → **500** вместо 400; `"rub"` vs `"RUB"` — рассинхрон | Нормализовать `strings.ToUpper(TrimSpace(c))` до Save; invalid → `ErrInvalidInput`; IT |
| A3W7-5-004 | NEW | P3 | `dto.go:53`; `service.go:310`; schema `review_note` Text без MaxLen | `ReviewNote` без лимита длины на service | Admin может записать огромный text → раздувание ответа ListPending/ListMine и хранения | Cap rune (напр. 2000) → `ErrInvalidInput`; опционально MaxLen в schema+ADR |

## Чисто (подтверждено после remediation)

- **DEV_MODE / ME-09:** `handler.go` `!release` + runtime guard→404; `handler_release.go` no-op RegisterRoutes; bootstrap WARN; login StatusActive.
- **`/api/docs`:** не публичен — auth+admin в `wireCore`; apidocs без `!release` tag допустим при gate.
- **Status PII/recon:** anon — только `hub` aggregate + 503; per-component — `/status/details` + auth; тесты W7A5002.
- **Marketplace authz:** все routes под `authMw`; pending/approve/reject → `requireAdmin`; update/sold/withdraw → owner; delete → owner\|admin; stranger → `ErrNotOwner`.
- **Optimistic lock / soft-delete:** VersionEQ на update/status/soft-delete; atomic re-moderation approved→pending.
- **PII email на витрине:** публичный `author_name` без email fallback.
- **Price/currency allowlist** (с оговоркой A3W7-5-003 normalize).
- **Seed demo copy:** Telegram = notification bind (ADR-AUTH-01).

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 1 |
| P3 | 3 |
| **Total** | **4** |

### Теги vs A2

| Tag | Count |
|-----|------:|
| FIXED-OK (A2 cascade) | 15 (`W7-A5-001…012`, `014…016`) |
| REGRESS | 1 (`W7-A5-013` → A3W7-5-002) |
| NEW | 3 (`A3W7-5-001`, `003`, `004`) |
| RESIDUAL (open A2 besides 013) | 0 |

### Фокус-вердикт

DEV_MODE exposure и status recon **держат** после TASK-121. Marketplace authz/OL **solid**. Главный новый риск зоны — **`review_note` на публичной витрине** (A3W7-5-001); ложный close seed-partial (A3W7-5-002) — dev-only P3.
