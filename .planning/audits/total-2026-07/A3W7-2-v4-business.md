# A3 W7 агент A2 — intranet+crm+hr+accounting+calendar (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline: [A2W7-2-v4-business.md](A2W7-2-v4-business.md) · A1: [A7-2-v4-business.md](A7-2-v4-business.md)  
> Assignment: [W7-ASSIGNMENTS.md](W7-ASSIGNMENTS.md) § A2 · Периметр: **72** · Прочитано: **72/72** · Пропущено: **0**  
> Фокус: IDOR/owner scope · soft-delete trash · optimistic lock · coffee/intranet polls  
> Remediation: TASK-024/057/063/067/089/094/095/098/106/118 (+ sibling `service_polls.go` / `repo_polls.go`)

## Сверка с A2 (W7-A2-* / A2W7-2-*)

| A2 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| W7-A2-001 | FIXED-OK | закрыт | TASK-098: `DeleteNews` author\|\|admin (`service.go:181–200`); polls+vote+update routes (`handler.go:45–47`); soft-delete hide poll (`repo_polls`); audit RestoreNewsPost |
| W7-A2-002 | FIXED-OK | закрыт | Odd-N → triple, `UnpairedCount=0`; `buildCoffeeMatches` + IT `coffee_odd_n_triple_task024` |
| W7-A2-003 | FIXED-OK | закрыт (news) | TASK-095: `req.Version < 1` → 400; `VersionEQ` в `UpdateNewsPost` (`repository.go:106–111`) |
| W7-A2-004 | FIXED-OK | закрыт | `CreateCoffeeRoundWithPairs` tx (`coffee_repository.go:65–100`); service edges+tx |
| W7-A2-005 | FIXED-OK | закрыт | TASK-089: `validateDealRefs` → `authorizeContactWrite` (`crm/service.go:285–294`) |
| W7-A2-006 | FIXED-OK | закрыт | TASK-118: audit trash `crm_contact`/`crm_deal` list/restore/purge (вне слота, подтверждено) |
| W7-A2-007 | FIXED-OK | закрыт | MAX(suffix)+1 + retry на unique (`invoice_repository.go:50–107`); IT concurrent |
| W7-A2-008 | **REGRESS** | открыт | ADR-U05/TASK-106: канон notify-only ✅ (`CoffeeMatchDeliveryMode`); **но** ошибки notify снова `_ =` без `slog.WarnContext` (`coffee_service.go:175–199`) — см. A3W7-2-001 |
| W7-A2-009 | FIXED-OK | закрыт | `ListApprovedLeaveInRange` viewer+project (`hr/repository.go:79–92`) |
| W7-A2-010 | FIXED-OK | закрыт | `MaxCalendarList=100` + `WithProject` (`calendar/repository.go:18–102`) |
| W7-A2-011 | FIXED-OK | закрыт | Cancel → `ErrAccessDenied` / «leave access denied» (`hr/handler.go:151–153`) |
| W7-A2-012 | FIXED-OK | закрыт | Unique SQLite DSN uuid (`handler_test.go:23`) |
| W7-A2-013 | FIXED-OK | закрыт | `MaxCRMList=100` + Limit |
| W7-A2-014 | FIXED-OK | закрыт | `CoffeePartnerResponse` без email |
| W7-A2-015 | FIXED-OK | закрыт | HR notifier/export → `slog.WarnContext` |
| W7-A2-016 | FIXED-OK | закрыт | Soft-delete news + AuditLog в tx |
| A2W7-2-N01 | FIXED-OK | закрыт | TASK-094: `stripTripleEdgesFromRoundResponse`; `PairCount==len(Pairs)`; `Triples[]` |

**REGRESS:** 1 · **RESIDUAL (A2 open, не закрыты кодом/ADR):** 0 · **NEW:** 2

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/intranet/coffee_dto.go` | 62 | 0 | чисто (TripleCount/Triples; без email) |
| `backend/internal/intranet/coffee_repository.go` | 219 | 0 | WithTx round+pairs FIXED-OK; Find First для троек |
| `backend/internal/intranet/coffee_service.go` | 331 | 2 | REGRESS `_=` notify; GetMyCoffeeMatch 1 partner |
| `backend/internal/intranet/dto.go` | 215 | 1 | poll Version в DTO; OL gap в service_polls |
| `backend/internal/intranet/export_hooks.go` | 38 | 0 | чисто |
| `backend/internal/intranet/handler.go` | 526 | 0 | poll routes + author-delete docs; layering OK |
| `backend/internal/intranet/repository.go` | 241 | 0 | news VersionEQ FIXED-OK; SoftDeleteKudo GDPR |
| `backend/internal/intranet/service.go` | 347 | 0 | author-delete + version&lt;1 FIXED-OK |
| `backend/testframework/intranet/coffee_repository_test.go` | 75 | 0 | чисто |
| `backend/testframework/intranet/coffee_service_test.go` | 55 | 0 | чисто |
| `backend/testframework/intranet/dto_test.go` | 37 | 0 | чисто |
| `backend/testframework/intranet/fixtures_test.go` | 53 | 0 | чисто |
| `backend/testframework/intranet/handler_test.go` | 298 | 0 | unique DSN FIXED-OK |
| `backend/testframework/intranet/handler_wave91_test.go` | 220 | 0 | чисто |
| `backend/testframework/intranet/service_test.go` | 143 | 0 | чисто |
| `backend/testframework/intranet/service_wave114_test.go` | 29 | 0 | чисто |
| `backend/internal/crm/dto.go` | 170 | 0 | чисто |
| `backend/internal/crm/export_hooks.go` | 16 | 0 | чисто |
| `backend/internal/crm/handler.go` | 194 | 0 | чисто |
| `backend/internal/crm/repository.go` | 248 | 0 | VersionEQ + MaxCRMList + soft-delete |
| `backend/internal/crm/service.go` | 328 | 0 | owner-scope write/list + validateDealRefs |
| `backend/testframework/crm/fixtures_test.go` | 44 | 0 | чисто |
| `backend/testframework/crm/handler_test.go` | 184 | 0 | чисто |
| `backend/testframework/crm/handler_wave101_test.go` | 53 | 0 | чисто |
| `backend/testframework/crm/handler_wave107_test.go` | 107 | 0 | чисто |
| `backend/testframework/crm/handler_wave111_test.go` | 88 | 0 | чисто |
| `backend/testframework/crm/list_scope_rt003_test.go` | 89 | 0 | чисто |
| `backend/testframework/crm/optimistic_lock_a6003_test.go` | 94 | 0 | чисто |
| `backend/testframework/crm/rbac_a6001_test.go` | 85 | 0 | чисто |
| `backend/testframework/crm/service_invoice_test.go` | 116 | 0 | чисто |
| `backend/testframework/crm/service_test.go` | 184 | 0 | чисто |
| `backend/testframework/crm/service_wave101_test.go` | 96 | 0 | чисто |
| `backend/testframework/crm/service_wave111_test.go` | 184 | 0 | чисто |
| `backend/internal/hr/dto.go` | 100 | 0 | чисто |
| `backend/internal/hr/export_hooks.go` | 33 | 0 | чисто |
| `backend/internal/hr/handler.go` | 165 | 0 | ErrAccessDenied mapping FIXED-OK |
| `backend/internal/hr/repository.go` | 214 | 0 | leave scope + Version WHERE + soft-delete |
| `backend/internal/hr/service.go` | 251 | 0 | slog notify/export FIXED-OK |
| `backend/testframework/hr/fixtures_test.go` | 57 | 0 | чисто |
| `backend/testframework/hr/handler_a6006_test.go` | 77 | 0 | чисто |
| `backend/testframework/hr/handler_test.go` | 185 | 0 | чисто |
| `backend/testframework/hr/handler_wave101_test.go` | 71 | 0 | чисто |
| `backend/testframework/hr/handler_wave111_test.go` | 94 | 0 | чисто |
| `backend/testframework/hr/repository_ent_test.go` | 150 | 0 | W7-A2-009 scope IT |
| `backend/testframework/hr/service_extended_test.go` | 148 | 0 | чисто |
| `backend/testframework/hr/service_wave111_test.go` | 104 | 0 | чисто |
| `backend/internal/accounting/dto.go` | 136 | 0 | чисто |
| `backend/internal/accounting/export_hooks.go` | 24 | 0 | чисто |
| `backend/internal/accounting/handler.go` | 211 | 0 | admin-only API |
| `backend/internal/accounting/invoice_repository.go` | 322 | 0 | nextInvoiceNumber FIXED-OK |
| `backend/internal/accounting/repository.go` | 242 | 0 | soft-delete journal paths |
| `backend/internal/accounting/service.go` | 246 | 0 | admin gate |
| `backend/internal/accounting/service_invoice.go` | 169 | 0 | pay atomic + deal invoice hook |
| `backend/testframework/accounting/create_invoice_concurrent_a6004_test.go` | 92 | 0 | чисто |
| `backend/testframework/accounting/fixtures_test.go` | 63 | 0 | чисто |
| `backend/testframework/accounting/handler_test.go` | 164 | 0 | чисто |
| `backend/testframework/accounting/handler_wave108_test.go` | 130 | 0 | чисто |
| `backend/testframework/accounting/handler_wave98_test.go` | 150 | 0 | чисто |
| `backend/testframework/accounting/pay_concurrent_a6002_test.go` | 153 | 0 | чисто |
| `backend/testframework/accounting/service_invoice_test.go` | 60 | 0 | чисто |
| `backend/testframework/accounting/service_wave108_test.go` | 95 | 0 | чисто |
| `backend/testframework/accounting/service_wave98_test.go` | 68 | 0 | чисто |
| `backend/internal/calendar/dto.go` | 38 | 0 | чисто |
| `backend/internal/calendar/export_hooks.go` | 30 | 0 | чисто |
| `backend/internal/calendar/handler.go` | 108 | 0 | events + ics scoped via service |
| `backend/internal/calendar/ical.go` | 58 | 0 | escapeICal OK |
| `backend/internal/calendar/repository.go` | 126 | 0 | Limit+WithProject+leave scope |
| `backend/internal/calendar/service.go` | 238 | 0 | projectIDs scope + range expand |
| `backend/testframework/calendar/handler_test.go` | 133 | 0 | чисто |
| `backend/testframework/calendar/leave_scope_a6005_test.go` | 89 | 0 | чисто |
| `backend/testframework/calendar/repository_series_trash_test.go` | 102 | 0 | DeletedAtIsNil series |
| `backend/testframework/calendar/service_wave98_test.go` | 110 | 0 | чисто |

Файлов в списке: **72**. Прочитано: **72**. Пропущено: **0**.

> Sibling (вне assignment, нужны для polls): `service_polls.go`, `repo_polls.go` — прочитаны для сверки W7-A2-001 / OL poll.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|----------|-------------------|------------|
| A3W7-2-001 | REGRESS (W7-A2-008 / TASK-106) | P2 | `coffee_service.go:175–199` | ADR-INTRANET-U05 §3: ошибки `NotifyCoffeeMatch` → `slog.WarnContext`; в HEAD снова `_ =` (silent swallow) | False-close TASK-106: notify-only + IT есть, observability-ветка ADR откатилась/не доехала; ops не видит failed match-notify | Вынести `notifyCoffeeMatch` helper: err → `slog.WarnContext(ctx, …, "user_id", …, "partner_id", …, "error", err)`; mutation IT на slog или export helper |
| A3W7-2-002 | NEW (хвост W7-A2-003) | P2 | `service_polls.go:83–85`; `repo_polls.go:221`; DTO `dto.go:44` | `UpdateNewsPoll`: version опционален (`>0`); repo `UpdateOneID` **без** `VersionEQ` | Тот же TOCTOU/LWW, что закрыли на news post (TASK-095). Спека PUT poll: `"version"` + 409. Concurrent close/edit → lost update; omit version → silent overwrite | Как news: `version < 1` → 400; `Where(newspoll.VersionEQ(in.Version))`; 409 на NotFound; IT concurrent + omit-version |
| A3W7-2-003 | NEW | P3 | `coffee_service.go:76–102`; `coffee_dto.go:51–57`; `coffee_repository.go:156–168` | `GetMyCoffeeMatch` возвращает один `Partner` через `FindUserPairInRound`=`First`; участник тройки видит только одного из двух напарников | Notify шлёт 2 партнёра (×6 calls); GET `/match` — нет. UI «текущий матч» неполон после odd-N | Собрать все рёбра user в раунде → `Partners[]` (или Partner+PartnerB); strip/hub logic как в triples; IT N=3 |

## Чисто (подтверждено после remediation)

- **Intranet author-delete + polls:** FR-INTR-07/08 path; poll hide via post `deleted_at`; vote unique path; percent after vote/author/admin.
- **News optimistic lock:** VersionEQ + required version (TASK-095).
- **Coffee:** odd-N triple; round+edges tx; PairCount/Triples strip (N01); notify-only канон ADR-U05 (кроме slog regress).
- **CRM:** owner write + RT-003 list-scope; contact-ref owner (TASK-089); VersionEQ contact/deal; MaxCRMList; trash types.
- **HR:** leave Version WHERE; cancel owner→403 access denied; approved leave viewer-scope; slog on notify/export fail.
- **Accounting:** admin-only; invoice MAX+retry; deal-lock create; pay atomic.
- **Calendar:** project membership scope; MaxCalendarList; leave A6-005; series `DeletedAtIsNil`; ics escape.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 2 |
| P3 | 1 |
| **Total** | **3** |

### Теги vs A2

| Tag | Count |
|-----|------:|
| FIXED-OK (A2 IDs) | 16 (W7-A2-001…007, 009…016, A2W7-2-N01) |
| REGRESS | 1 (`A3W7-2-001` ← W7-A2-008 slog) |
| RESIDUAL | 0 |
| NEW | 2 (`A3W7-2-002`, `A3W7-2-003`) |

A2 baseline: **11** open (10 RESIDUAL + 1 NEW) → **FIXED-OK 16** class closes · REGRESS **1** · NEW **2** · open **3** (P2=2 · P3=1).

**Топ ID:** A3W7-2-001 (coffee notify slog regress) · A3W7-2-002 (poll VersionEQ) · A3W7-2-003 (GetMyCoffeeMatch triple partners).
