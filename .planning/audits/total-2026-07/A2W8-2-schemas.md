# A2 W8 агент 2 — ent/schema part 2 (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> HEAD: `8d9c0950` · Baseline: [A8-2-schemas.md](A8-2-schemas.md)  
> Периметр: 17 схем · Прочитано: 17 · Пропущено: 0  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Сверка с Stage A (W8-A2-*)

| A1 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| W8-A2-001 | RESIDUAL | открыт | `comment.go` — `updated_at`/`deleted_at` есть, `version` нет |
| W8-A2-002 | RESIDUAL | открыт | `dashboardlayout.go` — mutable `items`+`updated_at`, нет `version` |
| W8-A2-003 | FIXED-OK | закрыт | `component.go:27-30` `deleted_at` + индексы; TASK-039 soft-delete |
| W8-A2-004 | RESIDUAL | частично | `deleted_at` FIXED (TASK-039); `version` всё ещё нет |
| W8-A2-005 | RESIDUAL | открыт | `customfielddefinition.go` — нет version/updated_at/deleted_at |
| W8-A2-006 | FIXED-OK | закрыт | `component.go:53` unique `(project, name)` (P2-BE-22 / TASK-046) |
| W8-A2-007 | FIXED-OK | закрыт | `environment.go:59` unique `(project, name)` |
| W8-A2-008 | RESIDUAL | открыт | `crmcontact.go` — нет `index.Fields("email")` |
| W8-A2-009 | RESIDUAL | открыт | `crmdeal.go` — нет `Edges("owner")+deleted_at` |
| W8-A2-010 | RESIDUAL | открыт | `filemetadata.go:34` — `entity_id` String, не UUID |
| W8-A2-011 | RESIDUAL | открыт | FileMetadata: soft-delete без `version`; `size` без Min; uploader не Required |
| W8-A2-012 | FIXED-OK | закрыт | `githubintegration.go` — `webhook_secret` Sensitive+NotEmpty (TASK-039) |
| W8-A2-013 | RESIDUAL | частично | GitLab `webhook_secret` Sensitive FIXED; всё ещё `Optional()` |
| W8-A2-014 | RESIDUAL | открыт | GitHubLink: `updated_at` без version; нет unique integration+external |
| W8-A2-015 | RESIDUAL | открыт | GitLabLink: симметрично W8-A2-014 |
| W8-A2-016 | RESIDUAL | открыт | HubBase: только `deleted_at`, нет `Edges("owner")+deleted_at` |
| W8-A2-017 | RESIDUAL | открыт | HubColumn: нет unique `(table, name)` / `table+deleted_at` |
| W8-A2-018 | RESIDUAL | открыт | HubRecord: нет `Edges("table")+deleted_at` |
| W8-A2-019 | FIXED-OK | закрыт | `comment.go:29` `mentions` → `[]uuid.UUID` (TASK-068) |
| W8-A2-020 | FIXED-OK | закрыт | Mojibake comments → «Рёбра»/UTF-8; grep `ибра`/`Р` = 0 |

**REGRESS:** 0 · **RESIDUAL (open classes):** 12 · **NEW:** 0 · **FIXED-OK:** 6 (+2 partial→open)

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/ent/schema/coffeeround.go` | 45 | 0 | чисто |
| `backend/ent/schema/comment.go` | 67 | 1 | RESIDUAL version |
| `backend/ent/schema/component.go` | 57 | 1 | soft FIXED; RESIDUAL version |
| `backend/ent/schema/crmcontact.go` | 67 | 1 | RESIDUAL email index |
| `backend/ent/schema/crmdeal.go` | 74 | 1 | RESIDUAL owner+deleted_at |
| `backend/ent/schema/customfielddefinition.go` | 57 | 1 | RESIDUAL OL/trash |
| `backend/ent/schema/dashboardlayout.go` | 59 | 1 | RESIDUAL version |
| `backend/ent/schema/environment.go` | 63 | 1 | soft FIXED; RESIDUAL version |
| `backend/ent/schema/filemetadata.go` | 69 | 1 | RESIDUAL entity_id/version/size |
| `backend/ent/schema/githubintegration.go` | 65 | 0 | Sensitive FIXED |
| `backend/ent/schema/githublink.go` | 72 | 1 | RESIDUAL version/unique |
| `backend/ent/schema/gitlabintegration.go` | 61 | 1 | Sensitive FIXED; Optional residual |
| `backend/ent/schema/gitlablink.go` | 75 | 1 | RESIDUAL version/unique |
| `backend/ent/schema/helpers.go` | 14 | 0 | чисто |
| `backend/ent/schema/hubbase.go` | 54 | 1 | RESIDUAL composite index |
| `backend/ent/schema/hubcolumn.go` | 60 | 1 | RESIDUAL unique/composite |
| `backend/ent/schema/hubrecord.go` | 53 | 1 | RESIDUAL composite index |

Файлов в списке: 17. Прочитано: 17. Пропущено: 0.

## Находки (открытые A2)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|---------|---------------|------------|
| A2W8-2-001 | RESIDUAL (W8-A2-001) | P1 | `comment.go:17-42` | Нет `version` при `updated_at` + soft-delete | TZ OL; concurrent edit body → LWW | VersionMixin + VersionEQ в tracker UpdateComment |
| A2W8-2-002 | RESIDUAL (W8-A2-002) | P1 | `dashboardlayout.go:17-39` | Нет `version` при mutable `items` JSON | Concurrent SaveLayout LWW | `version` + CAS в dashboard Save |
| A2W8-2-003 | RESIDUAL (W8-A2-004∩version) | P1 | `component.go`; `environment.go` | Soft-delete есть, `version` нет | Mutable TMS catalog без CAS | `version` + VersionEQ на Update |
| A2W8-2-004 | **FIXED** (TASK-097) | P2 | `customfielddefinition.go` | soft-delete + list filter (version → TASK-095) | tracker/soft_delete_task097_test.go |
| A2W8-2-005 | RESIDUAL (W8-A2-008/009) | P2 | `crmcontact.go:62-66`; `crmdeal.go:69-73` | Нет email index; нет owner+deleted_at | Lookup/list hot-path full scan | `index.Fields("email")`; `Edges("owner").Fields("deleted_at")` |
| A2W8-2-006 | RESIDUAL (W8-A2-010/011) | P2 | `filemetadata.go:17-46` | `entity_id` String; нет version; size/uploader | TZ UUID; soft без OL; отрицательный size | UUID entity_id; version; Min(0); uploader Required |
| A2W8-2-007 | RESIDUAL (W8-A2-013) | P2 | `gitlabintegration.go:33-36` | `webhook_secret` Optional после Sensitive | Схема разрешает интеграцию без секрета | Required/NotEmpty (как GitHub) |
| A2W8-2-008 | RESIDUAL (W8-A2-014/015) | P2 | `githublink.go`; `gitlablink.go` | `updated_at` без version; нет unique external | Гонки status; дубликаты links | version; unique Edges(integration)+external+type |
| A2W8-2-009 | RESIDUAL (W8-A2-016…018) | P2 | `hubbase.go`; `hubcolumn.go`; `hubrecord.go` | Нет composite owner/table+deleted_at; column name unique | ListBases/ListRecords hot-path; дубли колонок | A8-001 pattern + unique `(table, name)` |

## Чисто

| Файл | Проверено |
|------|-----------|
| `coffeeround.go` | UUID PK; append-only; indexes created_at |
| `helpers.go` | `nowUTC()` UTC canon; keyRegexp |
| `githubintegration.go` | Sensitive secrets FIXED (TASK-039) |
| `component` / `environment` soft-delete | TASK-039 + unique name FIXED |

## Сводка counts (открытые)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 3 |
| P2 | 6 |
| P3 | 0 |
| **Total open** | **9** |

FIXED-OK: 6 · NEW: 0 · REGRESS: 0
