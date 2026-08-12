# A3 W5 агент A5 — search remaining tests (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Mirror A2: [A2W5-5-search-tests.md](A2W5-5-search-tests.md) · A1: [A5-5-search-tests.md](A5-5-search-tests.md)  
> Assignment: [W5-ASSIGNMENTS.md](W5-ASSIGNMENTS.md) § A5 · Периметр: **14** · Прочитано: **14/14** · Пропущено: **0**  
> Remediation: TASK-113 (A2W5-A5-001…006) · Фокус: assert-free · wrong URL · mock vs guard · mutation gaps

## Сверка с A2 (A2W5-A5-*)

| A2 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| A2W5-A5-001 | FIXED-OK | закрыт | Hollow sqlite Empty убран; `TestSearch_UsesProjectScopeForIssues` → `t.Run` → `TestPostgresSearch_RejectsForeignProjectIssues_TASK037` (`service_scope_test.go:184-188`; proof: `authz_scope_ent_negative_task037_test.go:19-68`, Len==1 + foreign ID deny) |
| A2W5-A5-002 | FIXED-OK | закрыт | Runtime soft-delete: issues + files **+** test_cases + projects — `soft_delete_a4002_runtime_test.go:18-90` (`Len==1` на каждый шард) |
| A2W5-A5-003 | FIXED-OK | закрыт | Static `Contains("deleted_at IS NULL")` помечен secondary; primary = runtime (`soft_delete_a4002_test.go:14-17`) |
| A2W5-A5-004 | FIXED-OK | закрыт | Канон empty-scope: `TestSearchScopedEntities_EmptyProjectIDsReturnNothing` (`service_scope_test.go:137-165`); дубли wave100/108/109/114 сняты |
| A2W5-A5-005 | FIXED-OK | закрыт | Один `DeleteSavedFilter` NotFound + `ShareFilter_NotFound` sentinel — `service_wave109_test.go:14-25`; дубль shard_errors убран |
| A2W5-A5-006 | FIXED-OK | закрыт | Membership + missing entity tables → `ErrSearchUnavailable` (wave106/108/115, shard_errors, `FiltersByType` / admin user shard) |

**REGRESS:** 0 · **RESIDUAL (этот слот):** 0 · **NEW (этот слот):** 2

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/testframework/search/service_scope_test.go` | 188 | 0 | FIXED-OK TASK037 delegate; GetUser*IDs Len+UUID; empty-scope канон |
| `backend/testframework/search/service_shard_errors_test.go` | 53 | 0 | FIXED-OK total outage; closed DB; no-membership messages |
| `backend/testframework/search/service_test.go` | 374 | 1 | membership mock+Ent OK; **Limit Normalizes hollow** |
| `backend/testframework/search/service_wave100_test.go` | 77 | 0 | чисто — sqlite QueryFails + uuidSlice; empty-scope дубль снят |
| `backend/testframework/search/service_wave103_mongo_test.go` | 166 | 0 | FIXED-OK Len==1 foreign channel + soft-delete Mongo |
| `backend/testframework/search/service_wave106_test.go` | 115 | 1 | Unavailable OK; **LimitCapped misnamed**; дубль closedDB/nil-mongo |
| `backend/testframework/search/service_wave108_test.go` | 37 | 0 | чисто — DefaultLimit→Unavailable; ProjectIDsFailure |
| `backend/testframework/search/service_wave109_test.go` | 25 | 0 | FIXED-OK один NotFound Delete/Share |
| `backend/testframework/search/service_wave114_test.go` | 29 | 0 | чисто (дубль nil-mongo — см. NEW-002) |
| `backend/testframework/search/service_wave115_test.go` | 54 | 0 | FIXED-OK shard-only → Unavailable |
| `backend/testframework/search/service_wave116_postgres_test.go` | 319 | 0 | чисто — PG issues/TC/files/users/channels + P1BE32 |
| `backend/testframework/search/service_wave92_test.go` | 55 | 0 | чисто — Ent GetSavedFilter |
| `backend/testframework/search/soft_delete_a4002_runtime_test.go` | 90 | 0 | FIXED-OK 4 шарда mutation-ready |
| `backend/testframework/search/soft_delete_a4002_test.go` | 62 | 0 | FIXED-OK secondary static lint |

Файлов в списке: **14**. Прочитано: **14**. Пропущено: **0**.

## Находки (активные A3)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|---------|---------------|------------|
| A3W5-5-001 | **FIXED** (TASK-157) | P3 | `limit_clamp_a3w55_test.go` · was `service_wave106_test.go:40-54`; `service_test.go:233-242` | clamp по Len | — | Postgres multi-shard seed → `Len==20` for limit 100/0/-1; wave106 → `TestSearch_MissingIssuesTable_Unavailable`; hollow NormalizesLimit убран |
| A3W5-5-002 | **FIXED** (TASK-157) | P3 | каноны: `TestSearchMessages_NilMongoReturnsEmpty` · `TestGetUserChannelIDs_ClosedDB` | дубли сняты | — | удалены дубли scope/shard_errors/wave106/wave114 |

## Чисто (перепроверено)

| Чек-лист TASK-008 (тесты) | Статус |
|---------------------------|--------|
| assert-free `Test*` | **Не найдено** — у каждого есть `assert.*` / `require.*` |
| wrong URL | **N/A** — в 14 файлах нет HTTP/`fiber`/`httptest`; только service/repo/SQL/Mongo |
| mock vs guard | **OK** — `requireProjectMembership` в service → mock `memberOK=false` (`Create`/`List` NonMemberForbidden); Ent `IsProjectMember` / `IsSystemAdmin` на реальном sqlite Ent; Ctrl+K scope — SQL/Postgres, не мок membership |
| тест зелёный на сломанном guard | **Закрыто для NF-03/A4-002** (TASK037, wave103 Len, soft_delete runtime 4 шарда, P1BE32 channels). **Открыто** только clamp limit (A3W5-5-001) |
| soft_delete mutation | Runtime issues/files/test_cases/projects — mutation-ready; static — secondary lint |

### Дополнительно чисто

- Saved-filter owner/share/forbidden + Ent CRUD (`service_test.go`, wave92).
- wave116: outsider users excluded; channels soft-delete + private membership.
- wave103: foreign channel excluded (`Len==1` + NotContains `"hidden"`); Mongo soft-delete live vs trash.
- Empty `projectIDs` early-return — один канон.
- Missing-table total outage → `ErrSearchUnavailable` (не маскируется Empty).

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **14 / 14 / 0** |
| Находок open | **2** (NEW 2 · REGRESS 0 · RESIDUAL 0) |
| FIXED-OK (A2) | **6** |
| P0 | 0 |
| P1 | 0 |
| P2 | 0 |
| P3 | 2 |

### Теги vs A2

| Tag | Count |
|-----|------:|
| FIXED-OK (A2 checklist) | 6 |
| REGRESS | 0 |
| RESIDUAL | 0 |
| NEW | 2 |

### Топ

1. **A3W5-5-001** — clamp `limit` в `Search` не покрыт поведением (имя wave106 врёт; NormalizesLimit hollow).  
2. **A3W5-5-002** — дубли nil-mongo / closed-DB после чистки empty-scope.

P0/P1/P2 в remaining search tests **нет**. Remediation TASK-113 по NF-03 / soft-delete / Unavailable держится; остаток — hygiene лимита и дублей.
