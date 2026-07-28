# A2 W8 агент 1 — ent/schema part 1 (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> HEAD: `0d6a659f` · Baseline: [A8-1-schemas.md](A8-1-schemas.md)  
> Периметр: 17 схем · Прочитано: 17 · Пропущено: 0  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Сверка с Stage A (W8-A1-*)

| A1 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| W8-A1-001 | FIXED-OK | закрыт | `coffeeenrollment.go:45` `index.Edges("user").Unique()` + `user.go:135` `.Unique()`; TASK-025 / W18-C3 PASS |
| W8-A1-002 | RESIDUAL | открыт | `channel.go` — `updated_at`/`deleted_at` есть, `version` нет (P1-BE-U06 ADR) |
| W8-A1-003 | RESIDUAL | открыт | Mutable без `version`: AIProvider, AutomationRule, AIUserPreference, Board, ChannelMember, ChannelInvite, Call, APIKey, CoffeeEnrollment |
| W8-A1-004 | RESIDUAL | открыт | `OnDelete`/Cascade — **0** вхождений во всём `ent/schema/` (P1-BE-U05 / P2-BE-U03) |
| W8-A1-005 | RESIDUAL | открыт | `board.go` — нет `Indexes()`, нет `updated_at` |
| W8-A1-006 | RESIDUAL | открыт | `aiuserpreference.go:22-24` — `default_provider_id` UUID без edge на `AIProvider` |
| W8-A1-007 | FIXED-OK | закрыт | `bookmark.go:44-48` — `edge.From("channel", …).Field("channel_id")` |
| W8-A1-008 | RESIDUAL | открыт | `board.go` — нет `deleted_at` |
| W8-A1-009 | FIXED-OK | закрыт | `apikey.go:62-68` — дубль unique на `key_hash` убран (P3-BE-17) |
| W8-A1-010 | FIXED-OK | закрыт | Mojibake «ибра» → «Рёбра»; grep `ибра`/`Р` = 0 (TASK-068) |

**REGRESS:** 0 · **RESIDUAL (open):** 6 · **NEW:** 0 · **FIXED-OK:** 4

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/ent/schema/activitylog.go` | 62 | 0 | чисто |
| `backend/ent/schema/aiprovider.go` | 74 | 1 | RESIDUAL version |
| `backend/ent/schema/aiusagedaily.go` | 51 | 0 | чисто |
| `backend/ent/schema/aiuserpreference.go` | 54 | 2 | RESIDUAL version + FK |
| `backend/ent/schema/apikey.go` | 70 | 1 | RESIDUAL version; soft-delete FIXED |
| `backend/ent/schema/auditlog.go` | 63 | 0 | чисто |
| `backend/ent/schema/automationrule.go` | 69 | 1 | RESIDUAL version |
| `backend/ent/schema/backuprecord.go` | 53 | 0 | чисто (mojibake FIXED) |
| `backend/ent/schema/board.go` | 48 | 1 | RESIDUAL indexes/OL/trash |
| `backend/ent/schema/bookmark.go` | 59 | 0 | FIXED-OK channel FK |
| `backend/ent/schema/call.go` | 88 | 1 | RESIDUAL version + OnDelete children |
| `backend/ent/schema/callparticipant.go` | 60 | 0 | чисто (junction) |
| `backend/ent/schema/channel.go` | 91 | 1 | RESIDUAL version |
| `backend/ent/schema/channelinvite.go` | 75 | 1 | RESIDUAL version |
| `backend/ent/schema/channelmember.go` | 69 | 1 | RESIDUAL version |
| `backend/ent/schema/coffeeenrollment.go` | 48 | 1 | FIXED unique; RESIDUAL version |
| `backend/ent/schema/coffeepair.go` | 49 | 0 | чисто |

Файлов в списке: 17. Прочитано: 17. Пропущено: 0.

## Находки (открытые A2)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|---------|---------------|------------|
| A2W8-1-001 | RESIDUAL (W8-A1-002) | P1 | `channel.go:17-57` | Нет `field.Int("version")` при `updated_at` + soft-delete | TZ optimistic lock; concurrent rename/mute/settings → LWW | VersionMixin + `VersionEQ` в messenger repo |
| A2W8-1-002 | RESIDUAL (W8-A1-003) | P1 | `aiprovider.go`; `automationrule.go`; `aiuserpreference.go`; `board.go`; `channelmember.go`; `channelinvite.go`; `call.go`; `apikey.go`; `coffeeenrollment.go` | Mutable-схемы A1 без `version` | Системный TZ-gap P1-BE-U06 | ADR VersionMixin + exclusion list |
| A2W8-1-003 | RESIDUAL (W8-A1-004) | P2 | `channel.go:63-78`; `call.go:76-77` (+ пакет) | Нет `entsql.OnDelete` на parent→child | FK trap / orphan при hard-purge | ADR Cascade vs SetNull vs app purge |
| A2W8-1-004 | FIXED (TASK-107) | P2 | `board.go` | `updated_at`/`deleted_at` + project+deleted_at index | — | TASK-107 |
| A2W8-1-005 | FIXED (TASK-107) | P2 | `aiuserpreference.go` | FK edge `default_provider` → AIProvider | — | TASK-107 |
| A2W8-1-006 | RESIDUAL (W8-A1-003∩APIKey) | P2 | `apikey.go:17-48` | Soft-delete есть, `version` нет | Concurrent revoke/update без CAS | `version` + VersionEQ на revoke |

## Чисто

| Файл | Проверено |
|------|-----------|
| `activitylog.go` | UUID PK; append-only; edges Required; index issue+created_at |
| `aiusagedaily.go` | UUID; unique (user, usage_date); счётчики |
| `auditlog.go` | UUID; append-only; индексы action/entity/user |
| `callparticipant.go` | UUID; unique (call, user); junction |
| `coffeepair.go` | UUID; unique (round, A, B) |
| `bookmark.go` | UUID; channel+user FK; unique (user, message) |
| `backuprecord.go` | UUID; godoc «Рёбра» OK |

## Сводка counts (открытые)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 2 |
| P2 | 4 |
| P3 | 0 |
| **Total open** | **6** |

FIXED-OK: 4 · NEW: 0 · REGRESS: 0
