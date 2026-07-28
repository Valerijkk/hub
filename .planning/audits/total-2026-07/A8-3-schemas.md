# W8 агент A3 — schemas part 3

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: UUID+`version` · soft-delete (`deleted_at`) · PII · edges.

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/ent/schema/hubtable.go` | 56 | 0 | чисто |
| `backend/ent/schema/invite.go` | 61 | 3 | замечания |
| `backend/ent/schema/invoice.go` | 82 | 0 | чисто |
| `backend/ent/schema/issue.go` | 130 | 1 | замечания |
| `backend/ent/schema/issuelink.go` | 59 | 2 | замечания |
| `backend/ent/schema/issuemessagelink.go` | 66 | 0 | чисто |
| `backend/ent/schema/journalentry.go` | 66 | 1 | замечания |
| `backend/ent/schema/journalline.go` | 53 | 2 | замечания |
| `backend/ent/schema/kudo.go` | 50 | 1 | замечания |
| `backend/ent/schema/label.go` | 56 | 1 | замечания |
| `backend/ent/schema/labelgroup.go` | 49 | 1 | замечания |
| `backend/ent/schema/leaverequest.go` | 66 | 1 | замечания |
| `backend/ent/schema/ledgeraccount.go` | 56 | 0 | чисто |
| `backend/ent/schema/marketplacelisting.go` | 76 | 0 | чисто |
| `backend/ent/schema/newspost.go` | 58 | 1 | замечания |
| `backend/ent/schema/notification.go` | 77 | 2 | замечания |
| `backend/ent/schema/notificationpreference.go` | 63 | 2 | замечания |

Файлов в списке: 17. Прочитано: 17. Пропущено: 0.

## Counts

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 3 |
| P2 | 12 |
| P3 | 3 |
| **Σ** | **18** |

### Топ-3

1. **W8-A3-001 (P1)** — `NotificationPreference`: UNIQUE `(user_id, notification_type, channel, project_id)` с nullable `project_id` — в PostgreSQL NULL≠NULL, дубликаты глобальных prefs возможны при гонке Create.
2. **W8-A3-002 (P1)** — `LeaveRequest`: HR-PII (`reason`, sick/personal) без `deleted_at` — против TZ soft-delete/корзина и паттерна `CrmContact`.
3. **W8-A3-003 (P1)** — `JournalEntry.To("invoice")` без `.Unique()` → O2M вместо O2O; несколько счетов могут ссылаться на одну проводку.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W8-A3-001 | P1 | `backend/ent/schema/notificationpreference.go:60` | UNIQUE-индекс включает nullable `project_id` | Postgres допускает много строк с `project_id IS NULL` для одного `(user, type, channel)`; гонка двух `Create` обходит app-level upsert | Partial unique: `(user_id, notification_type, channel) WHERE project_id IS NULL` + отдельный unique с `project_id`; или sentinel UUID вместо NULL |
| W8-A3-002 | P1 | `backend/ent/schema/leaverequest.go:16` | Нет `deleted_at` при наличии version/updated_at и HR-PII (`reason`, тип sick) | TZ: soft-delete/корзина 30 дней; hard path только Admin/GDPR. Соседний `CrmContact` с PII имеет `deleted_at` | Добавить `deleted_at` Optional+Nillable + индекс; сервисный soft-delete / GDPR purge |
| W8-A3-003 | P1 | `backend/ent/schema/journalentry.go:56` | `edge.To("invoice", Invoice.Type)` без `.Unique()` | Invoice.From Unique даёт M2O: много Invoice → один JournalEntry; ломает 1:1 оплату/проводку | `.Unique()` на To → O2O; уникальный индекс на FK `journal_entry_id` у Invoice |
| W8-A3-004 | P2 | `backend/ent/schema/invite.go:17` | Нет `version` (и `updated_at`) при мутабельном `use_count` | TZ: `version` в каждой изменяемой Ent-схеме; revoke = hard `Delete` | Добавить `version`/`updated_at`; revoke → `deleted_at`/`revoked_at` вместо hard delete |
| W8-A3-005 | P2 | `backend/ent/schema/invite.go:17` | Нет `deleted_at`/`revoked_at`; revoke в сервисе = hard delete | Потеря аудита invite-кода; схема не поддерживает корзину/отзыв без удаления строки | Поле `deleted_at` или `revoked_at` + фильтр в List/Validate |
| W8-A3-006 | P2 | `backend/ent/schema/issuelink.go:56` | UNIQUE только `(source, target)` без `link_type` | Нельзя иметь одновременно `blocks` и `relates_to` между парой задач | Unique на `(source, target, link_type)` |
| W8-A3-007 | P2 | `backend/ent/schema/issuelink.go:29` | `link_type` включает `parent`/`child` при отдельном edge `Issue.parent`/`children` | Двойная модель иерархии → рассинхрон Gantt/дерева и links | Убрать parent/child из enum link_type **или** отказаться от self-edge parent |
| W8-A3-008 | P2 | `backend/ent/schema/journalline.go:21` | Нет DB-инварианта XOR debit/credit (оба ≥0 допустимы) | Инвариант только в service; прямой write в Ent/SQL может создать невалидную проводку | `entsql.Check` / Postgres CHECK: ровно одно из (debit>0, credit>0) |
| W8-A3-009 | P2 | `backend/ent/schema/journalline.go:16` | Нет `deleted_at`; soft-delete родителя `JournalEntry` не каскадирует на строки | Прямой query `JournalLine` видит «живые» строки удалённой проводки | `deleted_at` на line **или** обязательный join-фильтр + документированный cascade в сервисе |
| W8-A3-010 | P2 | `backend/ent/schema/kudo.go:16` | Нет `deleted_at`; `message` + `is_public=false` — персональный контент | Нет корзины/GDPR soft-path; только hard delete | Добавить `deleted_at` (+ опционально version при появлении edit) |
| W8-A3-011 | P2 | `backend/ent/schema/label.go:17` | Нет `version`/`updated_at`/`deleted_at`; `DeleteLabel` — hard delete | Имя/цвет не Immutable; hard delete против soft-delete TZ для пользовательского каталога | `version`+`updated_at`+`deleted_at`; soft-delete вместо `Delete()` |
| W8-A3-012 | P2 | `backend/ent/schema/labelgroup.go:17` | То же, что Label: нет version/updated_at/deleted_at | Группа меток — изменяемый каталог проекта без optimistic lock и корзины | Как W8-A3-011 |
| W8-A3-013 | P2 | `backend/ent/schema/notification.go:36` | `entity_id` — `String`; `actor_id`/`project_id` — UUID без edges | Полиморфная ссылка без FK; actor/project не защищены referential integrity | Edge на Project/Actor где применимо; для UUID-сущностей — `field.UUID` + comment для Mongo ID |
| W8-A3-014 | P2 | `backend/ent/schema/notification.go:47` | Мутабельный `is_read` без `version` | TZ optimistic lock на изменяемых сущностях; гонки mark-read/bulk-read | Добавить `version` (и при необходимости `updated_at`) |
| W8-A3-015 | P2 | `backend/ent/schema/notificationpreference.go:30` | Мутабельный `enabled`/`updated_at` без `version` | Lost update при параллельном toggle prefs | `field.Int("version").Default(1)` + Where version в Update |
| W8-A3-016 | P3 | `backend/ent/schema/issue.go:80` | Опечатка в комментарии: «ибра (edges)» | Косметика / шум в схеме | «Рёбра (edges)» |
| W8-A3-017 | P3 | `backend/ent/schema/invite.go:22` | `code` без `MaxLen` | Неограниченная строка в UNIQUE-колонке | `.MaxLen(64)` (или длина фактического генератора) |
| W8-A3-018 | P3 | `backend/ent/schema/newspost.go:24` | `body` Text без `MaxLen` | Потенциальный memory/DoS на уровне хранения | `.MaxLen(...)` по продуктовому лимиту intranet |

## Чисто

Проверялось по фокусу: UUID PK `Default(uuid.New).Immutable()`, наличие `version` у мутабельных, `deleted_at`+индекс, PII-поля, целостность edges (Required/Unique/O2O).

| Файл | Что проверено |
|------|----------------|
| `hubtable.go` | UUID+version+`deleted_at`+индекс; edges base Required Unique, columns/records O2M — ок |
| `invoice.go` | UUID+version+`deleted_at`; `customer_name` PII при наличии soft-delete; edges created_by Required, crm_deal/journal_entry optional Unique From — ок (дефект O2O на стороне JournalEntry) |
| `issuemessagelink.go` | UUID PK; Mongo `message_id` String обоснован; FK-поля+edges issue/channel/created_by; unique `(issue_id,channel_id,message_id)` — ок (immutable link, soft-delete N/A) |
| `ledgeraccount.go` | UUID+version+`deleted_at`; edge journal_lines O2M — ок |
| `marketplacelisting.go` | UUID+version+`deleted_at`; author/reviewer edges — ок |

## Замечания вне дефекта (контекст)

- `Issue`: UUID+version+`deleted_at`+составной индекс `project`+`deleted_at` — эталон soft-delete трекера.
- `JournalLine` / `IssueLink` / `Kudo` без `version` допустимы **только** если сущность реально immutable (create/delete); при появлении Update — обязателен `version`.
- `Notification.entity_id` как string частично оправдан Mongo/полиморфизмом (см. `notification/service.go`); находка — про отсутствие FK на UUID-ссылки, не про запрет string PK.
