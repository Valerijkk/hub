# CRM / HR / Accounting / Calendar

**Дата:** 2026-07-26  
**Ветка:** `cursor-grok-4.5`  
**Зона A6:** `backend/internal/{crm,hr,accounting,calendar}`  
**Режим:** READ-ONLY (кроме этого отчёта)

## Что проверено

| Пакет | Файлы (production `.go`) | Фокус |
|-------|--------------------------|--------|
| `crm/` | `handler.go`, `service.go`, `repository.go`, `dto.go` | IDOR owner/created_by, RBAC routes, optimistic lock WHERE, won→invoice side-effect, pagination |
| `hr/` | `handler.go`, `service.go`, `repository.go`, `dto.go` | Admin gate approve/reject, cancel requester scope, version=0 bypass, ErrVersionConflict dead path, `_ =` notify/export |
| `accounting/` | `handler.go`, `service.go`, `service_invoice.go`, `repository.go`, `invoice_repository.go`, `dto.go` | Admin gate, MarkInvoicePaid atomicity/race, invoice number race, deal-invoice uniqueness, journal tx rollback |
| `calendar/` | `handler.go`, `service.go`, `repository.go`, `ical.go`, `dto.go` | Project membership scope, leave privacy, iCal export SSRF (inbound fetch), RRULE, range limits |

**Метод:** полное чтение handler/service/repository + Grep по `_ =`, `go func`, `Tx(`/`Rollback`, `http.DefaultClient`, `ValidateOutbound`, `Version`, `authorize`, `Limit(`, `io.ReadAll`. Сверка со схемой Ent (`crmdeal`, `crmcontact`, `invoice`, `leaverequest`, `ledgeraccount`) и prior scan `W3-v4-bases-crm-hr.md`.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это баг | Как чинить |
|----|----------|-------------|----------|----------------|------------|
| A6-001 | P0 | `crm/service.go:80-123`, `172-214`; `crm/handler.go:59-84`, `110-138`; `crm/repository.go:81-112`, `163-200` | Update/Delete contact/deal без owner/admin: `UpdateContact`/`DeleteContact`/`DeleteDeal` не принимают `userID`; `UpdateDeal` не сверяет `owner` | Edge `owner`/`created_by` пишется при Create, но никогда не enforced. Любой JWT → правка PII контакта / суммы / soft-delete любой сделки / `stage=won` → draft invoice по чужой сделке (`tryCreateInvoice` 200-201). Классический IDOR (как bases до `authorizeBase`). | `authorizeContact` / `authorizeDeal`: actor == created_by/owner **или** system admin/founder; Delete/Update только после; denial-suite как `testframework/bases/rbac_test.go`. |
| A6-002 | P0 | `accounting/service_invoice.go:116-168` | `MarkInvoicePaid` неатомарен: create journal → post → mark paid вне одной tx; status check TOCTOU | Два concurrent `POST .../pay`: оба видят `issued`, оба постят журналы; второй `MarkInvoicePaid` падает на status — **осиротевший posted journal** (двойной revenue/банк в trial balance / P&L). При fail после `PostJournalEntry` — та же рассинхронизация книг и счёта. | Одна Ent-tx: `SELECT … FOR UPDATE` invoice `status=issued` → create+post lines → set paid+journal_id; rollback при любой ошибке. Idempotency key / unique partial index на payment journal per invoice. |
| A6-003 | P1 | `crm/service.go:88-90`, `180-181`; `crm/repository.go:82-84`, `164-166` | Optimistic lock CRM: `if req.Version > 0` + Update без `VersionEQ` в WHERE | При `version: 0`/omit проверка пропускается; при concurrent update оба читают v=N, оба `AddVersion(1)` — lost update (TZ: version optimistic lock). | Требовать `version >= 1`; `Where(…, Version(req.Version))` + map NotFound→409 Conflict. |
| A6-004 | P1 | `accounting/service_invoice.go:85-93`; `invoice_repository.go:71-98`, `144-151`; schema `invoice.go:67-69` | `InvoiceExistsForDeal` + Create — TOCTOU; FK `crm_deal` **не** unique (edge `To("invoices")` без Unique на deal) | Два concurrent create (HTTP + won-hook или два admin) → два invoice на одну сделку. App-check не атомарна. | Unique index на `invoice.crm_deal_id` WHERE deleted_at IS NULL (или status≠cancelled); create в tx с conflict→`ErrInvoiceExists`. |
| A6-005 | P1 | `calendar/service.go:77-83`; `calendar/repository.go:96-104` | `ListApprovedLeaveInRange` без фильтра по проектам/команде — **все** approved leave org любому JWT (и в `export.ics`) | Project-scoped calls/sprints/issues фильтруются `projectIDs`; leave — нет. Утечка HR (sick/personal + имя/email requester) всем authenticated, включая iCal attachment. | Ограничить leave: своё + members общих проектов / admin; или org-policy flag. Не класть email в title (`leaveEvent` 216). |
| A6-006 | P2 | `hr/service.go:110-120`, `149-159`, `178-191`; `hr/handler.go:85-87`, `127` | `version == 0` подставляет текущую версию; Approve/Cancel игнорируют bind error (`_ =` / empty req) | Optimistic lock фактически optional; conflict при stale version мапится в `ErrNotFound` (repo Where Version), а `ErrVersionConflict` **нигде не возвращается** из service — клиент получает 404 вместо 409. | Require version≥1; `ent.IsNotFound` после versioned update → `ErrVersionConflict`; не глотать Bind error. |
| A6-007 | P2 | `crm/service.go:150-169` | `CreateDeal` со `stage=won` не вызывает `tryCreateInvoice` | Invoice hook только на transition в `UpdateDeal` (200-201). Create сразу в won → сделка без draft invoice, расхождение с V4-03.5 automation. | После Create: если stage won — `tryCreateInvoice`; или запретить create с terminal stage. |
| A6-008 | P2 | `accounting/invoice_repository.go:43-56` | `nextInvoiceNumber`: Count+1 без сериализации | Concurrent create → одинаковый `INV-YYYY-NNN`; unique на `number` даёт 500 одному из запросов (не silent corrupt, но fragile UX/retry). | Sequence/advisory lock в tx; или `INSERT … ON CONFLICT` retry. |
| A6-009 | P2 | `crm/service.go:43-53`, `125-147`; `repository.go:66-71`, `140-151` | List contacts / pipeline — полный org dump без limit/cursor | Неbounded list; blast radius A6-001 + PII (email/phone/notes) всем JWT. Accounting хотя бы `Limit(50)`. | Зафиксировать ADR «org-shared read»; иначе owner/role filter + pagination (limit/offset или cursor). |
| A6-010 | P2 | `hr/handler.go:148-149` vs `service.go:187-188` | `ErrForbidden` всегда → `"admin access required"` | Non-owner `CancelLeave` с version=0 тоже Forbidden → ложный сигнал «нужен admin». | Разделить `ErrNotOwner` / message «not your leave request». |
| A6-011 | P3 | `hr/service.go:100`, `134`, `137-139`, `173` | `_ =` на Notify* / export err без лога (`_ = err`) | Сбой уведомления/1C export молчаливый; approval уже committed — ops не видит. | `slog.WarnContext` с leave_id/err; метрика export_fail. |
| A6-012 | P3 | `calendar/repository.go:58-63` | `ListScheduledCalls` без `WithProject()` | `expandCallEvents` читает `sc.Edges.Project` → `projectKey` всегда пустой в JSON/iCal. | `.WithProject()` в query. |
| A6-013 | P3 | `accounting/service.go:57-58`; `invoice_repository.go:90-93` | `AccountType` не валидируется до Ent; invalid `due_date` silently dropped | Невалидный enum → 500 вместо 400; битая due date молча игнорируется. | Allowlist AccountType → `ErrInvalidInput`; parse due fail → 400. |
| A6-014 | P3 | `hr/service.go:235-241` | `isValidLeaveType("") == true` | Пустой type проходит service-валидацию (handler подставляет vacation, но прямой вызов service — дыра). | `default: return false`. |

## Чисто

- **Accounting HTTP authz:** все list/mutate/invoice/P&L — `requireAdmin` (founder\|\|admin) в service; маршруты только `authMw` + service gate — OK для admin-only модуля.
- **HR approve/reject:** admin gate; `CancelLeave` в repo scoped `HasRequesterWith` + `status=pending` + version WHERE — IDOR cancel закрыт на уровне Update.
- **HR overlap:** `HasOverlappingLeave` по requester + pending/approved.
- **Journal CreateJournalEntry:** Ent tx + Rollback на ошибках строк (`repository.go:128-158`).
- **PostJournalEntry / IssueInvoice / MarkInvoicePaid (status):** WHERE status draft/issued — защита от повторного post/issue на уровне update.
- **Calendar project scope:** calls/sprints/issues фильтруются `ListUserProjectIDs`; range validate + `maxRangeDays=366`.
- **iCal SSRF:** inbound fetch / subscriptions **не реализованы** в зоне (только `RenderICalendar` export + auth Bearer). `http.DefaultClient` / user-controlled URL fetch в `calendar/` отсутствуют. (Spec FR-CAL subscriptions — future; при добавлении — обязателен `ValidateOutboundURL`.)
- **CRM project_id:** `IsProjectMember` при привязке проекта; invalid stage allowlist.
- **Secrets in logs:** CRM invoice skip логирует `deal_id` + err, не credentials.
- **Goroutines / races в shared maps:** в зоне нет `go func` / незащищённых maps.

## Оценка зоны: 4.5/10

**Обоснование:** HR approve/cancel и accounting admin-gate выглядят осознанно (~7). Сильно тянут вниз: полный CRM write-IDOR с фин. побочным эффектом invoice (P0), неатомарный/race `MarkInvoicePaid` с риском двойной проводки (P0), слабый optimistic lock CRM, race duplicate invoice-per-deal, и org-wide leak approved leave в календарь/iCal (P1). После A6-001…A6-005 зона поднялась бы к ~7.5/10.
