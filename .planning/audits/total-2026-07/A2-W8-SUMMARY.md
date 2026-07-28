# A2-W8 SUMMARY — ent/schema (Stage A2 re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> HEAD: `0d6a659f` · Периметр: `W8-ASSIGNMENTS.md` · **83/83** файлов  
> Отчёты: `A2W8-1` … `A2W8-5`

## Severity (открытые: NEW + REGRESS + RESIDUAL)

| Агент | Зона | Файлов | P0 | P1 | P2 | P3 | Σ open |
|-------|------|-------:|---:|---:|---:|---:|-------:|
| A2W8-1 | schemas part 1 | 17 | 0 | 2 | 4 | 0 | **6** |
| A2W8-2 | schemas part 2 | 17 | 0 | 3 | 9 | 0 | **12** |
| A2W8-3 | schemas part 3 | 17 | 0 | 1 | 9 | 0 | **10** |
| A2W8-4 | schemas part 4 | 17 | 0 | 3 | 4 | 1 | **8** |
| A2W8-5 | schemas part 5 | 15 | 0 | 5 | 3 | 0 | **8** |
| **Σ** | | **83** | **0** | **14** | **29** | **1** | **44** |

## По тегам

| Tag | Count |
|-----|------:|
| NEW | **4** |
| REGRESS | **0** |
| RESIDUAL | **40** |
| FIXED-OK (справочно, не в Σ open) | **~23** Stage A findings закрыты (полные); +partial soft-delete |

## NEW (4)

| ID | Sev | Суть |
|----|-----|------|
| A2W8-2-004 | P2 | Component/Environment: unique `(project, name)` без partial `WHERE deleted_at IS NULL` — trash блокирует recreate |
| A2W8-2-005 | P2 | DashboardLayout unique с nullable `project_id` (Postgres NULL≠NULL) — класс P1-BE-41 |
| A2W8-4-007 | P3 | Organization: дубль Unique на `slug`/`invite_token` (field + Indexes) |
| A2W8-5-007 | P2 | TestSuite soft-delete без unique `(project, name)` (асимметрия vs Component/Environment) |

## REGRESS

Нет. CONSTRAINT-фиксы TASK-025 (coffee/notif/invoice O2O), Sensitive secrets, mojibake/MaxLen (TASK-068), bookmark→channel FK, Role/Component/Environment/TestCaseVersion uniques — на месте.

## Топ RESIDUAL

1. **P1-BE-U06 VersionMixin** — Channel/Project/Comment/User/Team/TestRun/TestSuite/WebhookEndpoint/Org/Poll/Role/… без `version` (~14 схем с version из 82 entity)
2. **Soft-delete gaps** — LeaveRequest, Sprint, Label*, Kudo, Recording, Poll*, Board, Invite revoke
3. **OnDelete policy** — 0 `OnDelete` в пакете (P1-BE-U05 / P2-BE-U03)
4. **Composite indexes** — Hub*/CrmDeal owner+deleted_at; Git*Link unique external; IssueLink без link_type
5. **GitLab webhook_secret Optional** — Sensitive FIXED, Required — нет

## Stage A → сейчас

| Stage A Σ | Было | A2 open |
|-----------|-----:|--------:|
| P0 | 0 | **0** |
| P1 | 19 | **14** (constraints FIXED; VersionMixin/trash residual) |
| P2 | 40 | **29** (+3 NEW P2; −partial BE-22/secrets) |
| P3 | 12 | **1** (mojibake/MaxLen FIXED; +1 NEW dup index) |
| **Σ** | **71** | **44** |

## FIXED-OK highlights

| Class | Evidence |
|-------|----------|
| P1-BE-40..42 | Coffee unique, notif partial unique, journal↔invoice O2O — TASK-025 W18-C3 PASS |
| P2-BE-09 secrets | GitHub/GitLab/Org invite/Push Sensitive |
| P2-BE-22 (часть) | Component/Environment/Role/TestCaseVersion unique; JournalLine CHECK; Bookmark channel FK |
| P2-BE-08 (часть schema) | Component/Environment/TestSuite/WebhookEndpoint `deleted_at` |
| P3-BE-17/18 | mentions UUID; MaxLen invite/news; mojibake «Рёбра» — TASK-068 |

**Caveat:** TASK-046 mutation W18-C3 **FAIL** (source-grep, не Ent-runtime) — schema guards присутствуют, evidence layer слабый.

## Пути артефактов

```
.planning/audits/total-2026-07/A2W8-1-schemas.md
.planning/audits/total-2026-07/A2W8-2-schemas.md
.planning/audits/total-2026-07/A2W8-3-schemas.md
.planning/audits/total-2026-07/A2W8-4-schemas.md
.planning/audits/total-2026-07/A2W8-5-schemas.md
.planning/audits/total-2026-07/A2-W8-SUMMARY.md
```

→ Следующая волна A2: W9 frontend/components (если оркестратор идёт по порядку).
