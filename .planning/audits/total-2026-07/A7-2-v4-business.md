# W7 агент A2 — intranet+crm+hr+accounting+calendar

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: trash/soft-delete (FR-INTR-08, FR-CAL-05), IDOR owner-scope, RBAC.  
> Сверка: TASK-001/002 (restore `newspost`/`scheduled_call` в `audit`) **закрыты**; calendar фильтрует `DeletedAtIsNil`; CRM RT-003/A6-001 **держат**.

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/intranet/coffee_dto.go` | 54 | 1 | замечания |
| `backend/internal/intranet/coffee_repository.go` | 175 | 1 | замечания |
| `backend/internal/intranet/coffee_service.go` | 238 | 3 | замечания |
| `backend/internal/intranet/dto.go` | 176 | 1 | замечания |
| `backend/internal/intranet/export_hooks.go` | 29 | 0 | чисто |
| `backend/internal/intranet/handler.go` | 438 | 1 | замечания |
| `backend/internal/intranet/repository.go` | 229 | 2 | замечания |
| `backend/internal/intranet/service.go` | 326 | 3 | замечания |
| `backend/testframework/intranet/coffee_repository_test.go` | 76 | 0 | чисто |
| `backend/testframework/intranet/coffee_service_test.go` | 56 | 0 | чисто |
| `backend/testframework/intranet/dto_test.go` | 38 | 0 | чисто |
| `backend/testframework/intranet/fixtures_test.go` | 54 | 0 | чисто |
| `backend/testframework/intranet/handler_test.go` | 297 | 1 | замечания |
| `backend/testframework/intranet/handler_wave91_test.go` | 221 | 0 | чисто |
| `backend/testframework/intranet/service_test.go` | 144 | 0 | чисто |
| `backend/testframework/intranet/service_wave114_test.go` | 30 | 0 | чисто |
| `backend/internal/crm/dto.go` | 171 | 0 | чисто |
| `backend/internal/crm/export_hooks.go` | 17 | 0 | чисто |
| `backend/internal/crm/handler.go` | 195 | 0 | чисто |
| `backend/internal/crm/repository.go` | 243 | 2 | замечания |
| `backend/internal/crm/service.go` | 327 | 1 | замечания |
| `backend/testframework/crm/fixtures_test.go` | 45 | 0 | чисто |
| `backend/testframework/crm/handler_test.go` | 185 | 0 | чисто |
| `backend/testframework/crm/handler_wave101_test.go` | 54 | 0 | чисто |
| `backend/testframework/crm/handler_wave107_test.go` | 108 | 0 | чисто |
| `backend/testframework/crm/handler_wave111_test.go` | 89 | 0 | чисто |
| `backend/testframework/crm/list_scope_rt003_test.go` | 90 | 0 | чисто |
| `backend/testframework/crm/optimistic_lock_a6003_test.go` | 95 | 0 | чисто |
| `backend/testframework/crm/rbac_a6001_test.go` | 86 | 0 | чисто |
| `backend/testframework/crm/service_invoice_test.go` | 117 | 0 | чисто |
| `backend/testframework/crm/service_test.go` | 185 | 0 | чисто |
| `backend/testframework/crm/service_wave101_test.go` | 97 | 0 | чисто |
| `backend/testframework/crm/service_wave111_test.go` | 185 | 0 | чисто |
| `backend/internal/hr/dto.go` | 101 | 0 | чисто |
| `backend/internal/hr/export_hooks.go` | 34 | 0 | чисто |
| `backend/internal/hr/handler.go` | 163 | 1 | замечания |
| `backend/internal/hr/repository.go` | 174 | 1 | замечания |
| `backend/internal/hr/service.go` | 240 | 1 | замечания |
| `backend/testframework/hr/fixtures_test.go` | 58 | 0 | чисто |
| `backend/testframework/hr/handler_a6006_test.go` | 78 | 0 | чисто |
| `backend/testframework/hr/handler_test.go` | 186 | 0 | чисто |
| `backend/testframework/hr/handler_wave101_test.go` | 72 | 0 | чисто |
| `backend/testframework/hr/handler_wave111_test.go` | 95 | 0 | чисто |
| `backend/testframework/hr/repository_ent_test.go` | 94 | 1 | замечания |
| `backend/testframework/hr/service_extended_test.go` | 149 | 0 | чисто |
| `backend/testframework/hr/service_wave111_test.go` | 105 | 0 | чисто |
| `backend/internal/accounting/dto.go` | 137 | 0 | чисто |
| `backend/internal/accounting/export_hooks.go` | 25 | 0 | чисто |
| `backend/internal/accounting/handler.go` | 212 | 0 | чисто |
| `backend/internal/accounting/invoice_repository.go` | 286 | 1 | замечания |
| `backend/internal/accounting/repository.go` | 226 | 0 | чисто |
| `backend/internal/accounting/service.go` | 247 | 0 | чисто |
| `backend/internal/accounting/service_invoice.go` | 170 | 0 | чисто |
| `backend/testframework/accounting/create_invoice_concurrent_a6004_test.go` | 93 | 0 | чисто |
| `backend/testframework/accounting/fixtures_test.go` | 64 | 0 | чисто |
| `backend/testframework/accounting/handler_test.go` | 165 | 0 | чисто |
| `backend/testframework/accounting/handler_wave108_test.go` | 131 | 0 | чисто |
| `backend/testframework/accounting/handler_wave98_test.go` | 151 | 0 | чисто |
| `backend/testframework/accounting/pay_concurrent_a6002_test.go` | 154 | 0 | чисто |
| `backend/testframework/accounting/service_invoice_test.go` | 61 | 0 | чисто |
| `backend/testframework/accounting/service_wave108_test.go` | 96 | 0 | чисто |
| `backend/testframework/accounting/service_wave98_test.go` | 69 | 0 | чисто |
| `backend/internal/calendar/dto.go` | 39 | 0 | чисто |
| `backend/internal/calendar/export_hooks.go` | 31 | 0 | чисто |
| `backend/internal/calendar/handler.go` | 109 | 0 | чисто |
| `backend/internal/calendar/ical.go` | 59 | 0 | чисто |
| `backend/internal/calendar/repository.go` | 118 | 1 | замечания |
| `backend/internal/calendar/service.go` | 239 | 0 | чисто |
| `backend/testframework/calendar/handler_test.go` | 134 | 0 | чисто |
| `backend/testframework/calendar/leave_scope_a6005_test.go` | 90 | 0 | чисто |
| `backend/testframework/calendar/repository_series_trash_test.go` | 103 | 0 | чисто |
| `backend/testframework/calendar/service_wave98_test.go` | 111 | 0 | чисто |

Файлов в списке: 72. Прочитано: 72. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W7-A2-001 | P1 | `intranet/service.go:167-178`; `intranet/handler.go:201-220`; `intranet/dto.go` (нет poll); `intranet/repository.go:124-130` | **FR-INTR-07/08 не закрыты в модуле:** `DeleteNews` только `requireAdmin` (автор-member не может удалить «свой» пост); нет `news_poll*` API/схем (есть только messenger `poll`); soft-delete поста не каскадирует опрос/голоса; restore опроса (буквальный FR-INTR-08) нечем возвращать | Спека: Admin любой + автор свой → корзина + audit; restore → пост **с опросом и счётчиками**. Сейчас CMS admin-only без опросов; комментарий в `audit.RestoreNewsPost` сам признаёт «когда появятся news_poll*» | Реализовать `news_poll*` + author-or-admin delete; каскад soft-hide poll/votes; тест restore через `/admin/trash/newspost/:id/restore` с голосами; мутация — снять author-ветку |
| W7-A2-002 | P1 | `intranet/coffee_service.go:201-229` | **FR-INTR-10:** при нечётном N остаётся `unpaired`, а не «одна тройка» | Спека: нечётный N → одна тройка; код оставляет одного без пары (`UnpairedCount`) | `buildCoffeePairs`: при `len%2==1` собрать тройку (или отдельная сущность); тест N=3 → 0 unpaired / 1 triple |
| W7-A2-003 | P2 | `intranet/repository.go:105-121`; `intranet/service.go:136-138` | Optimistic lock news: service сравнивает `version` только если `req.Version > 0`; repo `AddVersion(1)` **без** `VersionEQ` | TOCTOU / omit version → last-write-wins (как A6-003 в CRM уже починен) | Repo: `Where(VersionEQ(req.Version))`; требовать `version≥1`; тест concurrent PUT → 409 |
| W7-A2-004 | P2 | `intranet/coffee_service.go:137-164` | `CreateCoffeeRound` + цикл `CreateCoffeePair` без одной tx | Ошибка на N-й паре → orphan round с частичными парами | `WithTx`: round+все pairs атомарно; rollback при любой ошибке |
| W7-A2-005 | P2 | `crm/service.go:285-292` | `validateDealRefs`: `GetContact(id)` без owner-scope | Member знает UUID чужого контакта → привязывает к своей сделке (обход RT-003 list-scope) | Для non-admin: контакт только `created_by=actor` (или shared policy); тест stranger contact_id → 400/403 |
| W7-A2-006 | P2 | `crm/repository.go:117-123`; `crm/repository.go:212-218` | Soft-delete contact/deal есть; в admin trash (`audit`) типов `crm_contact`/`crm_deal` **нет** | ТЗ: soft delete через корзину 30д + restore; сущности «исчезают» без restore/purge API | Зарегистрировать в trash list/restore/purge **или** явный ADR «CRM-lite без корзины» |
| W7-A2-007 | P2 | `accounting/invoice_repository.go:44-57` | `nextInvoiceNumber` = `Count(prefix)+1` вне уникального constraint / serial | Два concurrent create (без deal-lock A6-004) → одинаковый `INV-YEAR-00N` / unique fail → 500 | Sequence/advisory lock/`MAX(number)+1` в tx; unique на `number` |
| W7-A2-008 | P2 | `intranet/coffee_service.go` | **FR-INTR-11 FIXED (TASK-106):** ADR-INTRANET-U05 notify-only; `notifyCoffeeMatch` + slog; без auto-DM | Спека синхронизирована с notify-only; IT `coffee_notify_task106_test.go` | Auto-DM — только после нового ADR |
| W7-A2-009 | P2 | `hr/repository.go:68-77`; `testframework/hr/repository_ent_test.go:65-80` | `ListApprovedLeaveInRange(from,to)` — **org-wide** без viewer-scope | Calendar уже закрыл A6-005 своим scoped API; HR-метод — footgun IDOR leave PII при будущем reuse | Удалить/пометить deprecated; либо требовать `viewerID`+project scope как в calendar; не тестировать org-wide как «норму» |
| W7-A2-010 | P2 | `calendar/repository.go:55-65`; `calendar/repository.go:82-95` | `ListScheduledCalls` / `ListIssuesDueInRange` без `Limit`; calls без `WithProject()` | Рост проектов → unbounded response; `ProjectKey` у call-событий всегда пуст (edge не грузится) | Cap/пагинация; `.WithProject()` на calls |
| W7-A2-011 | P3 | `hr/handler.go:151-152` | `CancelLeave` → `ErrForbidden` мапится в `"admin access required"` | Non-owner cancel — не admin-кейс; вводит в заблуждение клиент/аудит | Отдельный sentinel / сообщение `"leave access denied"` |
| W7-A2-012 | P3 | `testframework/intranet/handler_test.go:23` | SQLite DSN `file:intranet?mode=memory&cache=shared` фиксирован | Parallel `go test` шарит БД → flaky (CRM/calendar уже на uuid DSN) | `file:intranet_`+uuid |
| W7-A2-013 | P3 | `crm/repository.go:69-77`; `crm/repository.go:151-165` | `ListContacts`/`ListDeals` без limit | Рост CRM-lite → unbounded JSON | Offset/limit или cap (как accounting invoices 50) |
| W7-A2-014 | P3 | `intranet/coffee_dto.go:20`; `intranet/coffee_repository.go:132-148` | `CoffeePartnerResponse.Email` отдаётся в match/admin rounds | PII email всем участникам раунда/admin list без необходимости (достаточно display_name) | Убрать email из DTO или gate admin-only |
| W7-A2-015 | P3 | `hr/service.go:100`; `hr/service.go:127`; `hr/service.go:131-132` | `_ = notifier` / export error проглочен | Чек-лист TASK-008: необработанные `_ =`; silent drop notify/1C export | slog.Warn + metric; export fail → outbox retry |
| W7-A2-016 | P3 | `intranet/service.go:167-178` | Soft-delete news без явной записи domain audit `(actor, post_id)` в модуле | FR-INTR-07 требует audit log; полагается только на HTTP-audit middleware (если включён) | Вызов audit SoftDelete/Log с actor+entity_id в той же операции |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| CRM `authorizeContactWrite` / `authorizeDealWrite` | Owner или system admin/founder; stranger → `ErrForbidden` (A6-001) |
| CRM `listOwnerScope` + repo filters | RT-003: member list/pipeline scoped; admin org-wide; тесты `list_scope_rt003`, `rbac_a6001` на **real Ent** |
| CRM `UpdateContact`/`UpdateDeal` | `VersionEQ` в WHERE (A6-003) — эталон для intranet news |
| CRM soft-delete filters | `DeletedAtIsNil` на get/list/update; hard delete HTTP нет |
| HR approve/reject/cancel | Admin на pending; cancel только requester + `Version` в WHERE; A6-006 HTTP version=0 → 400 |
| Accounting `requireAdmin` | Все HTTP paths admin/founder; member → 403 |
| Accounting pay/create invoice | `PayInvoiceAtomic` tx (A6-002); deal invoice lock (A6-004); `DeletedAtIsNil` на invoice/account/journal |
| Calendar project scope | `ListUserProjectIDs` membership; calls/sprints/issues только своих проектов |
| Calendar leave A6-005 | Scoped leave; PII email не в title (`leaveEvent`); тест `leave_scope_a6005` |
| Calendar FR-CAL-05 (read path) | `ListScheduledCalls` + `DeletedAtIsNil`; restore series снова в выборке — `repository_series_trash_test` (мутация-ready) |
| Calendar DELETE events | Вне модуля (planned в API / calls+audit trash) — в scope A2 нет дыры фильтра |
| Intranet news soft-delete basics | `DeletedAtIsNil` на get/list/update; member delete → Forbidden (`service_wave114`); admin delete → 404 get |
| Intranet coffee RBAC | Stats/rounds/run — admin; enrollment/match — self |
| Intranet kudos | Self-kudo reject; recipient must be active; pagination capped 100 |
| `export_hooks` (все 5) | Только hubtest re-export, без prod-логики |
| Большинство `*_test.go` | Есть assert/require; RBAC/scope/concurrency тесты бьют в real Ent, не только мок |

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | 72 / 72 / 0 |
| Находок | **16** |
| P0 | 0 |
| P1 | 2 |
| P2 | 8 |
| P3 | 6 |

### Топ-3

1. **W7-A2-001** — FR-INTR-07/08: нет author-delete и `news_poll*`; restore опроса из корзины формально невыполним.  
2. **W7-A2-002** — FR-INTR-10: нечётный N → unpaired, а не тройка.  
3. **W7-A2-005** — CRM: привязка чужого `contact_id` к своей сделке обходит RT-003 list-scope.

### Вердикт по фокусу аудита

| Инвариант | Статус |
|-----------|--------|
| FR-INTR-08 restore newspost (audit path) | **TASK-001 закрыт** вне модуля; в intranet — soft-delete ок, poll-часть **gap** (W7-A2-001) |
| FR-INTR-07 trash + author delete | Soft-delete admin-only; author/member + poll cascade **не держат** |
| FR-CAL-05 series trash в calendar read | **держит** (`DeletedAtIsNil` + regression test) |
| CRM IDOR owner-scope (update/delete/list) | **держит** (A6-001, RT-003); дыра на contact ref (W7-A2-005) |
| HR/Accounting RBAC | Admin gates **держат**; cancel message P3; HR org-wide leave helper — footgun |
| Soft delete → корзина | News/calls — через audit; CRM entities — **без trash** (W7-A2-006) |
