# A2W7-2 — intranet+crm+hr+accounting+calendar (Stage A2)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Baseline: `A7-2-v4-business.md` · Assignments: `W7-ASSIGNMENTS.md` §A2

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/intranet/coffee_dto.go` | 62 | 1 | замечания |
| `backend/internal/intranet/coffee_repository.go` | 174 | 1 | замечания |
| `backend/internal/intranet/coffee_service.go` | 266 | 1 | замечания |
| `backend/internal/intranet/dto.go` | 175 | 1 | замечания |
| `backend/internal/intranet/export_hooks.go` | 33 | 0 | чисто |
| `backend/internal/intranet/handler.go` | 437 | 1 | замечания |
| `backend/internal/intranet/repository.go` | 255 | 1 | замечания |
| `backend/internal/intranet/service.go` | 325 | 1 | замечания |
| `backend/testframework/intranet/coffee_repository_test.go` | 75 | 0 | чисто |
| `backend/testframework/intranet/coffee_service_test.go` | 55 | 0 | чисто |
| `backend/testframework/intranet/dto_test.go` | 37 | 0 | чисто |
| `backend/testframework/intranet/fixtures_test.go` | 53 | 0 | чисто |
| `backend/testframework/intranet/handler_test.go` | 298 | 1 | замечания |
| `backend/testframework/intranet/handler_wave91_test.go` | 220 | 0 | чисто |
| `backend/testframework/intranet/service_test.go` | 143 | 0 | чисто |
| `backend/testframework/intranet/service_wave114_test.go` | 29 | 0 | чисто |
| `backend/internal/crm/dto.go` | 170 | 0 | чисто |
| `backend/internal/crm/export_hooks.go` | 16 | 0 | чисто |
| `backend/internal/crm/handler.go` | 194 | 0 | чисто |
| `backend/internal/crm/repository.go` | 248 | 1 | замечания |
| `backend/internal/crm/service.go` | 326 | 1 | замечания |
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
| `backend/internal/hr/handler.go` | 165 | 1 | замечания |
| `backend/internal/hr/repository.go` | 173 | 1 | замечания |
| `backend/internal/hr/service.go` | 251 | 1 | замечания |
| `backend/testframework/hr/fixtures_test.go` | 57 | 0 | чисто |
| `backend/testframework/hr/handler_a6006_test.go` | 77 | 0 | чисто |
| `backend/testframework/hr/handler_test.go` | 185 | 0 | чисто |
| `backend/testframework/hr/handler_wave101_test.go` | 71 | 0 | чисто |
| `backend/testframework/hr/handler_wave111_test.go` | 94 | 0 | чисто |
| `backend/testframework/hr/repository_ent_test.go` | 93 | 1 | замечания |
| `backend/testframework/hr/service_extended_test.go` | 148 | 0 | чисто |
| `backend/testframework/hr/service_wave111_test.go` | 104 | 0 | чисто |
| `backend/internal/accounting/dto.go` | 136 | 0 | чисто |
| `backend/internal/accounting/export_hooks.go` | 24 | 0 | чисто |
| `backend/internal/accounting/handler.go` | 211 | 0 | чисто |
| `backend/internal/accounting/invoice_repository.go` | 285 | 1 | замечания |
| `backend/internal/accounting/repository.go` | 225 | 0 | чисто |
| `backend/internal/accounting/service.go` | 246 | 0 | чисто |
| `backend/internal/accounting/service_invoice.go` | 169 | 0 | чисто |
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
| `backend/internal/calendar/handler.go` | 108 | 0 | чисто |
| `backend/internal/calendar/ical.go` | 58 | 0 | чисто |
| `backend/internal/calendar/repository.go` | 117 | 1 | замечания |
| `backend/internal/calendar/service.go` | 238 | 0 | чисто |
| `backend/testframework/calendar/handler_test.go` | 133 | 0 | чисто |
| `backend/testframework/calendar/leave_scope_a6005_test.go` | 89 | 0 | чисто |
| `backend/testframework/calendar/repository_series_trash_test.go` | 102 | 0 | чисто |
| `backend/testframework/calendar/service_wave98_test.go` | 110 | 0 | чисто |

Файлов в списке: **72**. Прочитано: **72**. Пропущено: **0**.

## Находки

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект / статус |
|----|-----|-----|-------------|---------|------------------------|
| W7-A2-001 | **FIXED** | P1 | `intranet/` + `news_poll*` | FR-INTR-07/08: author-delete + polls + restore counters (TASK-098 / ADR-INTRANET-U03) |
| W7-A2-002 | FIXED-OK | P1 | `coffee_service.go:230-258` | Нечётный N → unpaired | Закрыто TASK-024: `buildCoffeeMatches` → тройка, `UnpairedCount=0` |
| W7-A2-003 | RESIDUAL | P2 | `intranet/repository.go:107-123`; `service.go:136-138` | Update news: нет `VersionEQ`; service сравнивает version только если `>0` | TOCTOU / omit version → LWW; TASK-038 вне scope |
| W7-A2-004 | RESIDUAL | P2 | `coffee_service.go:140-175` | `CreateCoffeeRound` + пары/тройки без одной tx | Mid-fail → orphan round с частичными парами |
| W7-A2-005 | RESIDUAL | P2 | `crm/service.go:285-292` | `validateDealRefs`: `GetContact` без owner-scope | Member привязывает чужой contact UUID к своей сделке |
| W7-A2-006 | RESIDUAL | P2 | `crm/repository.go` soft-delete vs audit trash | Нет типов `crm_contact`/`crm_deal` в admin trash | Soft-delete без restore/purge API |
| W7-A2-007 | RESIDUAL | P2 | `accounting/invoice_repository.go:44-57` | `nextInvoiceNumber` = Count+1 без serial/unique | Concurrent create → duplicate INV / 500 |
| W7-A2-008 | RESIDUAL | P2 | `coffee_service.go:160-194` | FR-INTR-11: только `NotifyCoffeeMatch`, `_=`; нет auto-DM | W12 U05 — greenfield DM |
| W7-A2-009 | RESIDUAL | P2 | `hr/repository.go:68-77` | `ListApprovedLeaveInRange` org-wide без viewer-scope | Footgun IDOR leave PII при reuse |
| W7-A2-010 | RESIDUAL | P2 | `calendar/repository.go:55-65`, `:82-95` | Lists без `Limit`; `ListScheduledCalls` без `WithProject()` | Unbounded; `ProjectKey` у calls пуст (issues — WithProject есть) |
| W7-A2-011 | FIXED-OK | P3 | `hr/handler.go:151-153` | Cancel → «admin access required» | Закрыто TASK-063: `ErrAccessDenied` → «leave access denied» |
| W7-A2-012 | FIXED-OK | P3 | `testframework/intranet/handler_test.go:23` | Shared SQLite DSN | Закрыто TASK-067: uuid DSN |
| W7-A2-013 | RESIDUAL | P3 | `crm/repository.go` ListContacts/Deals | Без limit | Unbounded CRM-lite JSON |
| W7-A2-014 | FIXED-OK | P3 | `coffee_dto.go:16-22` | Email в partner DTO | Закрыто TASK-067: email убран |
| W7-A2-015 | FIXED-OK | P3 | `hr/service.go` notifier/export | `_ = notifier/export` | Закрыто TASK-057: slog на err |
| W7-A2-016 | FIXED-OK | P3 | `intranet/repository.go:126-149` | Soft-delete без domain audit | Закрыто TASK-067: AuditLog в tx |
| A2W7-2-N01 | NEW | P3 | `coffee_service.go:150-204` | После фикса троек: `Pairs[]` содержит edge-пары троек, а `PairCount` — только бинарные | Клиенты `len(pairs)` / PairCount расходятся; побочный эффект TASK-024 |

## Чисто (инварианты держат)

| Зона | Статус |
|------|--------|
| CRM owner-scope write/list (A6-001, RT-003) | Держит |
| CRM VersionEQ update | Держит |
| Accounting admin + pay/create atomic | Держит |
| Calendar leave A6-005 + series DeletedAtIsNil | Держит |
| Coffee FR-INTR-10 odd-N triple | FIXED-OK |
| News soft-delete + audit row | FIXED-OK (author-delete/poll — gap) |

## Counts

| Метрика | n |
|---------|--:|
| Файлов | 72 |
| Open (RESIDUAL+NEW) | **11** (10 RESIDUAL + 1 NEW) |
| FIXED-OK | 6 |
| REGRESS | 0 |
| Open by sev | P0:0 · P1:1 · P2:8 · P3:2 |
