# A3 W8 агент 3 — ent/schema part 3 (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`)  
> Baseline: [A2W8-3-schemas.md](A2W8-3-schemas.md) · A1: [A8-3-schemas.md](A8-3-schemas.md)  
> Периметр: 17 схем · Прочитано: 17 · Пропущено: 0  
> Фокус: UUID / version / soft-delete / indexes / OnDelete  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Сверка с A2 (A2W8-3-*)

| A2 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| A2W8-3-001 | FIXED-OK | закрыт | `leaverequest.go:45-48` `deleted_at` + index; VersionMixin; TASK-097 |
| A2W8-3-002 | FIXED-OK | закрыт | `invite.go:38-50` `version` + `updated_at` + `revoked_at`; TASK-107 |
| A2W8-3-003 | FIXED-OK | закрыт | `issuelink.go:25-32` без parent/child; unique `(source,target,link_type)` `:60-65`; Cascade на edges; TASK-096 |
| A2W8-3-004 | FIXED-OK | закрыт | `kudo.go:30-32` `deleted_at`; `journalline.go:14-15` join-filter ADR (без колонки); TASK-097 |
| A2W8-3-005 | FIXED-OK | закрыт | `label.go` / `labelgroup.go`: `updated_at` + `deleted_at` + VersionMixin + indexes; TASK-097/095 |
| A2W8-3-006 | FIXED-OK | закрыт | `notification.go:37-40` `entity_id` UUID; `:56-57` `version`; TASK-107 |
| A2W8-3-007 | FIXED-OK | закрыт | `notificationpreference.go:47-50` VersionMixin; partial unique `:70-75` сохранён; TASK-095 |

### Сверка A1 FIXED-OK (подтверждение)

| A1 ID | Tag | Evidence |
|-------|-----|----------|
| W8-A3-001 | FIXED-OK | partial unique NULL/NOT NULL prefs |
| W8-A3-003 | FIXED-OK | `journalentry.go:62` `To("invoice").Unique()` + Invoice edge unique |
| W8-A3-008 | FIXED-OK | `journalline.go:63-65` XOR CHECK |
| W8-A3-016 | FIXED-OK | `issue.go` godoc «Рёбра» |
| W8-A3-017 | FIXED-OK | `invite.go:25` `MaxLen(64)` |
| W8-A3-018 | FIXED-OK | `newspost.go:26` `MaxLen(50000)` |

**REGRESS:** 0 · **RESIDUAL (A2 open):** 0 · **NEW:** 2 · **FIXED-OK (A2 active):** 7

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/ent/schema/hubtable.go` | 61 | 1 | NEW OnDelete owned children |
| `backend/ent/schema/invite.go` | 74 | 0 | FIXED-OK version/revoke |
| `backend/ent/schema/invoice.go` | 89 | 0 | чисто UUID+VersionMixin+trash+O2O |
| `backend/ent/schema/issue.go` | 140 | 0 | эталон UUID+OL+trash+Cascade |
| `backend/ent/schema/issuelink.go` | 66 | 0 | FIXED-OK unique/type/Cascade |
| `backend/ent/schema/issuemessagelink.go` | 66 | 0 | чисто immutable link + unique |
| `backend/ent/schema/journalentry.go` | 72 | 1 | NEW OnDelete lines; O2O FIXED |
| `backend/ent/schema/journalline.go` | 67 | 0 | XOR+ADR join-filter FIXED-OK |
| `backend/ent/schema/kudo.go` | 55 | 0 | FIXED-OK deleted_at |
| `backend/ent/schema/label.go` | 72 | 0 | FIXED-OK OL/trash (unique recreate — ADR accepted) |
| `backend/ent/schema/labelgroup.go` | 65 | 0 | FIXED-OK OL/trash |
| `backend/ent/schema/leaverequest.go` | 76 | 0 | FIXED-OK deleted_at HR-PII |
| `backend/ent/schema/ledgeraccount.go` | 61 | 0 | чисто UUID+VersionMixin+trash |
| `backend/ent/schema/marketplacelisting.go` | 81 | 0 | чисто |
| `backend/ent/schema/newspost.go` | 68 | 0 | MaxLen FIXED; poll From Cascade |
| `backend/ent/schema/notification.go` | 82 | 0 | FIXED-OK entity_id UUID+version |
| `backend/ent/schema/notificationpreference.go` | 77 | 0 | FIXED-OK partial unique+VersionMixin |

Файлов в списке: 17. Прочитано: 17. Пропущено: 0.

## Находки (активные A3)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|---------|---------------|------------|
| A3W8-3-001 | FIXED | P2 | `hubtable.go` | было: `To("columns"/"records")` без Cascade → hard-purge FK trap | TASK-169: `Annotations(entsql.OnDelete(entsql.Cascade))` + migrate SQL + hubtest | evidence: `cascade_hubtable_journalentry_task169_test.go` |
| A3W8-3-002 | FIXED | P2 | `journalentry.go` | было: `To("lines")` без Cascade → hard DELETE entry FK trap | TASK-169: Cascade на lines + migrate SQL + hubtest | evidence: `cascade_hubtable_journalentry_task169_test.go` |

## Чисто (подтверждено после remediation)

| Файл / класс | Проверено |
|--------------|-----------|
| UUID PK | все 17: `Default(uuid.New).Immutable()` |
| Version / Mixin | mutable: VersionMixin или explicit `version` (invite/notification per TASK-107); immutable links без OL — OK |
| Soft-delete | Leave/Label/LabelGroup/Kudo/Issue/Invoice/JournalEntry/HubTable/Ledger/Marketplace/News — `deleted_at`+index; Invite=`revoked_at`; JournalLine=join-filter ADR |
| Indexes | Issue project+deleted_at; IssueLink unique+type; IssueMessageLink triple unique; NotifPreference partial unique |
| OnDelete (OK) | Issue owned children Cascade; IssueLink Cascade; NewsPoll.From Cascade (peer); Project→invites/labels Cascade (peer) |
| Unique+soft-delete | Label `(project,name)` блокирует recreate — **accepted** ADR-ENT-SOFT-DELETE-GAPS §Последствия |

## Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 2 |
| P3 | 0 |
| **Total open** | **2** |

| Tag | Count |
|-----|------:|
| FIXED-OK (A2 active) | 7 |
| FIXED-OK (A1 cascade) | 6 |
| RESIDUAL | 0 |
| NEW | 2 |
| REGRESS | 0 |

A2 baseline: 7 open → **FIXED-OK 7** · NEW 2 · REGRESS 0 · RESIDUAL 0 · open **2** (P2).

**Топ ID:** A3W8-3-001 (HubTable OnDelete) · A3W8-3-002 (JournalEntry→lines OnDelete).
