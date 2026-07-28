# A2 W8 агент 3 — ent/schema part 3 (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> HEAD: `8d9c0950` · Baseline: [A8-3-schemas.md](A8-3-schemas.md)  
> Периметр: 17 схем · Прочитано: 17 · Пропущено: 0  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Сверка с Stage A (W8-A3-*)

| A1 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| W8-A3-001 | FIXED-OK | закрыт | `notificationpreference.go:63-68` partial unique NULL/NOT NULL (TASK-025 / P1-BE-41) |
| W8-A3-002 | RESIDUAL | открыт | `leaverequest.go` — HR-PII без `deleted_at` |
| W8-A3-003 | FIXED-OK | закрыт | `journalentry.go:57` `To("invoice").Unique()` O2O (TASK-025 / P1-BE-42) |
| W8-A3-004 | RESIDUAL | открыт | Invite: mutable `use_count` без `version`/`updated_at` |
| W8-A3-005 | RESIDUAL | открыт | Invite: нет `deleted_at`/`revoked_at`; revoke = hard delete |
| W8-A3-006 | RESIDUAL | открыт | IssueLink unique только `(source, target)` без `link_type` |
| W8-A3-007 | RESIDUAL | открыт | `link_type` parent/child + Issue.self-edge parent — dual model |
| W8-A3-008 | FIXED-OK | закрыт | `journalline.go:61-63` XOR CHECK (P2-BE-22 / TASK-046) |
| W8-A3-009 | RESIDUAL | открыт | JournalLine без `deleted_at` при soft-delete родителя |
| W8-A3-010 | RESIDUAL | открыт | Kudo без `deleted_at` |
| W8-A3-011 | RESIDUAL | открыт | Label без version/updated_at/deleted_at |
| W8-A3-012 | RESIDUAL | открыт | LabelGroup — то же |
| W8-A3-013 | RESIDUAL | открыт | Notification: `entity_id` String; actor/project без edges |
| W8-A3-014 | RESIDUAL | открыт | Notification: mutable `is_read` без `version` |
| W8-A3-015 | RESIDUAL | открыт | NotificationPreference: `enabled`/`updated_at` без `version` |
| W8-A3-016 | FIXED-OK | закрыт | `issue.go` godoc «Рёбра» (TASK-068) |
| W8-A3-017 | FIXED-OK | закрыт | `invite.go:25` `MaxLen(64)` |
| W8-A3-018 | FIXED-OK | закрыт | `newspost.go:26` `MaxLen(50000)` |

**REGRESS:** 0 · **RESIDUAL (open):** 11 · **NEW:** 0 · **FIXED-OK:** 7

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/ent/schema/hubtable.go` | 56 | 0 | чисто |
| `backend/ent/schema/invite.go` | 63 | 1 | RESIDUAL version/revoke |
| `backend/ent/schema/invoice.go` | 84 | 0 | чисто (O2O peer FIXED) |
| `backend/ent/schema/issue.go` | 130 | 0 | эталон UUID+version+trash |
| `backend/ent/schema/issuelink.go` | 59 | 1 | RESIDUAL unique/dual hierarchy |
| `backend/ent/schema/issuemessagelink.go` | 66 | 0 | чисто |
| `backend/ent/schema/journalentry.go` | 67 | 0 | O2O FIXED |
| `backend/ent/schema/journalline.go` | 65 | 1 | CHECK FIXED; RESIDUAL trash |
| `backend/ent/schema/kudo.go` | 50 | 1 | RESIDUAL deleted_at |
| `backend/ent/schema/label.go` | 56 | 1 | RESIDUAL OL/trash |
| `backend/ent/schema/labelgroup.go` | 49 | 1 | RESIDUAL OL/trash |
| `backend/ent/schema/leaverequest.go` | 66 | 1 | RESIDUAL deleted_at |
| `backend/ent/schema/ledgeraccount.go` | 56 | 0 | чисто |
| `backend/ent/schema/marketplacelisting.go` | 76 | 0 | чисто |
| `backend/ent/schema/newspost.go` | 60 | 0 | MaxLen FIXED |
| `backend/ent/schema/notification.go` | 77 | 1 | RESIDUAL entity_id/version |
| `backend/ent/schema/notificationpreference.go` | 70 | 1 | partial unique FIXED; RESIDUAL version |

Файлов в списке: 17. Прочитано: 17. Пропущено: 0.

## Находки (открытые A2)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|---------|---------------|------------|
| A2W8-3-001 | RESIDUAL (W8-A3-002) | P1 | `leaverequest.go:16-46` | HR-PII (`reason`, sick) без `deleted_at` | TZ корзина 30д; CrmContact с PII имеет soft-delete | `deleted_at` + index; soft-delete / GDPR purge |
| A2W8-3-002 | RESIDUAL (W8-A3-004/005) | P2 | `invite.go:17-41` | Нет version; revoke = hard delete | Mutable use_count; потеря аудита кода | version + `revoked_at`/`deleted_at` |
| A2W8-3-003 | RESIDUAL (W8-A3-006/007) | P2 | `issuelink.go:22-57` | Unique без `link_type`; parent/child dual | Нельзя blocks+relates; рассинхрон дерева | Unique (source,target,type); убрать parent/child **или** self-edge |
| A2W8-3-004 | RESIDUAL (W8-A3-009/010) | P2 | `journalline.go`; `kudo.go` | Нет soft-delete на line/kudo | Orphan lines / нет GDPR soft-path | deleted_at **или** join-filter + ADR |
| A2W8-3-005 | RESIDUAL (W8-A3-011/012) | P2 | `label.go`; `labelgroup.go` | Нет version/updated_at/deleted_at | Hard DeleteLabel; catalog без OL | version + soft-delete |
| A2W8-3-006 | RESIDUAL (W8-A3-013/014) | P2 | `notification.go:36-48` | entity_id String; is_read без version | Полиморф без FK; mark-read race | UUID where applicable; version |
| A2W8-3-007 | RESIDUAL (W8-A3-015) | P2 | `notificationpreference.go:31-43` | Mutable enabled без version | Lost update parallel toggle | `version` + VersionEQ |

## Чисто

| Файл | Проверено |
|------|-----------|
| `hubtable.go` | UUID+version+deleted_at |
| `invoice.go` | UUID+version+deleted_at; O2O peer FIXED |
| `issue.go` | эталон soft-delete + composite project+deleted_at |
| `issuemessagelink.go` | unique (issue,channel,message); Mongo message_id OK |
| `journalentry.go` | O2O invoice FIXED |
| `ledgeraccount.go` / `marketplacelisting.go` / `newspost.go` | UUID+version+trash OK |

## Сводка counts (открытые)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 1 |
| P2 | 6 |
| P3 | 0 |
| **Total open** | **7** |

FIXED-OK: 7 · NEW: 0 · REGRESS: 0
