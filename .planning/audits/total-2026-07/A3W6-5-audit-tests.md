# A3 W6 агент A5 — audit remaining tests (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline: [A2W6-5-audit-tests.md](A2W6-5-audit-tests.md) · Assignment: [W6-ASSIGNMENTS.md](W6-ASSIGNMENTS.md) §A5  
> Периметр: **12** · Прочитано: **12/12** · Пропущено: **0**  
> Remediation: TASK-114 (A2W6-A5-001/003..009 + A4); evidence также в `task114_residual_test.go` (вне списка A5)  
> Фокус: assert-free · mutation gaps · purge/retention tests

## Сверка с A2 (A2W6-A5-*)

| A2 ID | Tag | Статус | Комментарий |
|-------|-----|--------|-------------|
| A2W6-A5-001 | FIXED-OK | закрыт (частично) | `TestAuditService_PurgeExpired_CapturesCutoffs_A2W6A5001` — callCounts×6 + cutoff±1m; **CRM×2 не в wantKeys** → A3W6-5-001 |
| A2W6-A5-002 | FIXED-OK | закрыт | `PurgeExpired_PropagatesError` — `require.Error` + contains `db unavailable` / `purge expired` (бывший cement fail-open удалён) |
| A2W6-A5-003 | FIXED-OK | закрыт | RetentionHides −40d: issue/project (`repository_trash_test`) + test_case/file (`repository_extra`); newspost/scheduled уже были |
| A2W6-A5-004 | FIXED-OK | закрыт (частично) | `SoftDelete/RestoreItem_RecordsAllTypes_A2W6A5004` — map×6; **CRM не в списке / mock не пишет** → A3W6-5-002 |
| A2W6-A5-005 | FIXED-OK | закрыт | Purge test_case/file/project + issue/newspost/scheduled: `ent.IsNotFound` после hard delete |
| A2W6-A5-006 | FIXED-OK | закрыт | `ListTrash_CapturesProjectID_A2W6A5006` (task114); hollow sibling `ListTrash_WithProjectFilter` жив → A3W6-5-005 |
| A2W6-A5-007 | FIXED-OK | закрыт | `DeletedAt==nil` после restore: issue/file/test_case (+ newspost/scheduled/project) |
| A2W6-A5-008 | FIXED-OK | закрыт | `ReassignTasks` — `reassignAssignedCalls` + count=3 |
| A2W6-A5-009 | FIXED-OK | закрыт | `DeleteUserGDPR_SuccessCascadeCalls_A2W6A5009`; hollow `DeleteUserGDPR_Success` → A3W6-5-007 |

**REGRESS:** 0 · **assert-free (0 asserts):** не найдено

### Stage A (W6-A5) cascade

| Старый ID | Tag | Комментарий |
|-----------|-----|-------------|
| W6-A5-001…006,008,010 | FIXED-OK / см. A2 | закрыты TASK-114 + repo IT |
| W6-A5-007 | FIXED-OK | ListTrash cutoff ≈ now−30d |
| W6-A5-009 | FIXED-OK | SkipsOnCreateError + createCalls |

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/testframework/audit/repository_extra_test.go` | 260 | 1 | Retention/purge IsNotFound OK; GDPR Success hollow |
| `backend/testframework/audit/repository_newspost_trash_test.go` | 113 | 0 | чисто · soft/restore/retention/purge + service ViaTrashAPI |
| `backend/testframework/audit/repository_scheduled_call_trash_test.go` | 117 | 0 | чисто · soft/restore/retention/purge + ViaTrashAPI |
| `backend/testframework/audit/repository_test.go` | 115 | 0 | чисто · ListAuditLogs pagination/filters/date; SlowOnly sqlite skip |
| `backend/testframework/audit/repository_trash_test.go` | 236 | 0 | FIXED-OK RetentionHides issue/project · purge IsNotFound · restore DeletedAt |
| `backend/testframework/audit/retention_test.go` | 175 | 0 | чисто · recordings purge + GetRetentionDays fallback · nil mongo |
| `backend/testframework/audit/service_deactivate_test.go` | 121 | 2 | cache-assert gap · hollow project_id filter sibling |
| `backend/testframework/audit/service_limits_test.go` | 281 | 0 | FIXED-OK Reassign counters · limits fail-closed · batch N+1 |
| `backend/testframework/audit/service_purge_test.go` | 222 | 2 | PropagatesError+ListTrash cutoff OK; Succeeds/AllTypes hollow; CRM purge assert gap |
| `backend/testframework/audit/service_security_test.go` | 537 | 1 | FIXED-OK SkipsOnCreateError; SoftDeleteCRM* без call-record |
| `backend/testframework/audit/service_test.go` | 398 | 0 | чисто · GDPR cascade abort HardDelete · daysRemaining · subject map |
| `backend/testframework/audit/service_trash_test.go` | 69 | 0 | чисто · unknown type err · empty list · SoftDelete not found |

Файлов в списке: **12**. Прочитано: **12**. Пропущено: **0**.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A3W6-5-001 | NEW | P2 | `service_purge_test.go:115–129`; `task114_residual_test.go:161–167`; prod `service.go:355–356` | `PurgeExpired_CapturesCutoffs` assert'ит только 6 ключей; `PurgeExpired_Succeeds` — только `NoError`. Prod зовёт ещё `crm_contacts`/`crm_deals` | Снять две CRM-ветки в `PurgeExpired` → CapturesCutoffs и Succeeds остаются зелёными (mock умеет `track("crm_*")`, но никто не проверяет) | Добавить `crm_contacts`/`crm_deals` в `wantKeys` (+ callCounts==1, cutoff±1m); Succeeds — либо удалить, либо `require.Equal` на callCounts×8 |
| A3W6-5-002 | NEW | P2 | `service_security_test.go:214–218,188–189`; `service_purge_test.go:191–205`; `task114_residual_test.go:176,186` | Mock `SoftDeleteCRM*` / `RestoreCRM*` не пишут в `softDeleteCalls`/`restoreCalls`; `RecordsAllTypes` без CRM; `SoftDelete/Restore_AllTypes` — только `NoError` | Switch `crm_contact`→`SoftDeleteIssue` (append `"issue"`) → AllTypes green, RecordsAllTypes не зовёт CRM | В mock append `"crm_contact"`/`"crm_deal"`; расширить RecordsAllTypes до 8 типов **или** отдельный тест map CRM |
| A3W6-5-003 | NEW | P2 | gap vs `repository_trash_test.go:209–236` / `repository_extra_test.go:52–97`; prod `repository.go:602–632` | Нет `RetentionHidesOlderThan30Days` для `crm_contact`/`crm_deal` (есть soft/purge в task118 вне A5, без −40d hide) | Снять `DeletedAtGT(cutoff)` в ListDeletedCRM* → list −40d зелёный; класс A2W6-A5-003 не доведён до CRM | Зеркало NewsPost: soft −40d, list(cutoff−30d) → total 0 |
| A3W6-5-004 | NEW | P2 | `service_deactivate_test.go:22–36`; prod `service.go:388–390` | `DeactivateUser_InvalidatesCache`: miniredis поднят, но ключ не seeded и `Del`/`Exists` не проверяются — только `deactivateIDs`/`deleteSessIDs` | Убрать `InvalidateUserStatusCache` → тест зелёный; имя теста врёт про cache (ME-13) | `rdb.Set(UserStatusCacheKey(target), …)`; после Deactivate — `assert.False(mr.Exists(key))` |
| A3W6-5-005 | RESIDUAL | P3 | `service_deactivate_test.go:48–63` | `ListTrash_WithProjectFilter` — только `StatusOK` (A2W6-A5-006 sibling) | Снятие `project_id` parse → 200 green; цемент в task114 `CapturesProjectID` | Удалить hollow **или** assert `gotProjectID` как в A2W6A5006 |
| A3W6-5-006 | RESIDUAL | P3 | `service_purge_test.go:115–129,185–206` | `PurgeExpired_Succeeds` / `SoftDelete_AllTypes` / `RestoreItem_AllTypes` — NoError-only после появления CapturesCutoffs/RecordsAllTypes | Дубли без mutation-value; создают ложное ощущение покрытия | Удалить или свести к одному mutation-proof тесту |
| A3W6-5-007 | RESIDUAL | P3 | `repository_extra_test.go:134–142` | `DeleteUserGDPR_Success` — только `gdprUserID` (A2W6-A5-009 sibling) | Wiring mock без cascade evidence; цемент в `SuccessCascadeCalls` | Удалить hollow **или** перевести на `gdprFullCascadeRepo` counters |

## Чисто (подтверждено)

- **Purge fail-closed:** `PropagatesError` ловит return-nil после slog (мутация A2W6-A5-002).
- **Trash cutoff:** `ListTrash_WithItems` — `cutoffCalled` + ±1m от `TrashRetentionDays`.
- **Retention hide 30d:** issue / project / test_case / file / newspost / scheduled_call.
- **Hard purge:** issue/tc/file/project/newspost/scheduled — `n==1` + `IsNotFound`.
- **Restore ClearedDeletedAt:** issue / file / test_case (+ IT newspost/scheduled/project).
- **ReassignTasks:** call counter + from/to UUID.
- **SkipsOnCreateError:** `createErr` → no panic, `createCalls==1`, logs empty.
- **Limits:** allowlist key, non-positive reject, invalid Atoi deny, batch single call.
- **GDPR:** cascade fail abort HardDelete (`service_test`); Ent rollback mid-cascade (task114, вне списка).
- **assert-free:** в 12 файлах нет `Test*` без `assert`/`require`/`t.Error*`.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 4 |
| P3 | 3 |
| **Total open** | **7** |

### Теги

| Tag | Count |
|-----|------:|
| NEW | 4 |
| REGRESS | 0 |
| RESIDUAL | 3 |
| FIXED-OK (A2 active set) | 9 (`A2W6-A5-001`…`009`, с оговорками partial → NEW) |
