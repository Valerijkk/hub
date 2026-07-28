# W6 агент A5 — audit remaining tests

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: trash/purge tests, assert quality, mock vs repo guards.

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/testframework/audit/repository_extra_test.go` | 197 | 2 | замечания |
| `backend/testframework/audit/repository_newspost_trash_test.go` | 113 | 0 | чисто |
| `backend/testframework/audit/repository_scheduled_call_trash_test.go` | 117 | 0 | чисто |
| `backend/testframework/audit/repository_test.go` | 115 | 0 | чисто |
| `backend/testframework/audit/repository_trash_test.go` | 203 | 1 | замечания |
| `backend/testframework/audit/retention_test.go` | 175 | 0 | чисто |
| `backend/testframework/audit/service_deactivate_test.go` | 117 | 2 | замечания |
| `backend/testframework/audit/service_limits_test.go` | 278 | 1 | замечания |
| `backend/testframework/audit/service_purge_test.go` | 174 | 3 | замечания |
| `backend/testframework/audit/service_security_test.go` | 445 | 1 | замечания |
| `backend/testframework/audit/service_test.go` | 355 | 0 | чисто |
| `backend/testframework/audit/service_trash_test.go` | 69 | 0 | чисто |

Файлов в списке: 12. Прочитано: 12. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W6-A5-001 | P2 | `service_purge_test.go:76-90` | `TestAuditService_PurgeExpired_Succeeds` — только `require.NoError`; `purgeMockRepo` игнорирует `cutoff` (`_ time.Time`); нет счётчиков вызовов 6 purge-методов; `scheduled_calls` нет в `purgeCounts` | Prod `PurgeExpired` всегда `return nil` и глотает ошибки (`service.go:285-305`). Удаление любого `purge("…")` / передача `time.Time{}` → тест всё ещё зелёный | Captured cutoffs ≈ `now-30d`; `assert.Equal` вызовов issues/test_cases/files/projects/newsposts/scheduled_calls = 1; mutation: убрать одну строку purge → FAIL |
| W6-A5-002 | P2 | `repository_trash_test.go:95-114`; `repository_extra_test.go:19-46,89-130`; `service_deactivate_test.go:75-99` | `RetentionHidesOlderThan30Days` есть только у newspost + scheduled_call | Guard списка корзины — `DeletedAtGT(cutoff)` в **репозитории** (`repository.go:269` и аналоги). Issue/test_case/file/project: soft-delete/list с `when.Add(-1h)`, не «старше 30д → скрыто». Снятие `DeletedAtGT` на этих типах → suite зелёный | Зеркало NewsPost: soft-delete −40d, list с cutoff −30d → `total==0`; mutation: drop `DeletedAtGT` → FAIL |
| W6-A5-003 | P2 | `service_purge_test.go:136-157` | `SoftDelete_AllTypes` / `RestoreItem_AllTypes` — только `NoError` на `mockAuditRepo`, методы SoftDelete*/Restore* не пишут вызовы | Неверный `case` (все типы → `SoftDeleteIssue`) или no-op `return nil` без switch → AllTypes зелёный (`UnknownType` ловит только default). Switch в **service**, но без call-recording тест не доказывает маршрутизацию | Как `RestoreItem_Issue` (`restoreID`): map[type]called; mutation: свести все case к одному методу → FAIL |
| W6-A5-004 | P2 | `repository_extra_test.go:48-67,89-130` | `PurgeExpiredTestCases` / `Files` / `Projects` — assert только `n==1`, без `ent.IsNotFound` после hard delete | Issues/NewsPost/ScheduledCall проверяют отсутствие строки. Здесь `n==1` при soft-delete без DELETE всё ещё возможен при баге счётчика; нет доказательства hard delete | После purge: `Get` → `IsNotFound` (как `repository_trash_test.go:147-149`) |
| W6-A5-005 | P2 | `service_deactivate_test.go:48-63` | `TestHandler_ListTrash_WithProjectFilter` — только `StatusOK`; `trashHandlerMockRepo.ListDeletedIssues` игнорирует `projectID` | Имя обещает project filter; снятие парсинга `project_id` / передачи в service → 200, тест зелёный | Captured `*uuid.UUID` в mock; `assert.Equal(projectID, *got)`; decode body `total==1` |
| W6-A5-006 | P3 | `repository_trash_test.go:109-113`; `repository_extra_test.go:41-45`; `service_deactivate_test.go:98` | SoftDelete+Restore issue/test_case/file: после restore нет `DeletedAt==nil` (и file — даже list empty) | NewsPost/ScheduledCall/Project проверяют entity. Сломанный `ClearDeletedAt` + list-баг могут дать ложный PASS на `total==0` | После restore: `Get` + `assert.Nil(DeletedAt)`; file — list empty |
| W6-A5-007 | P3 | `service_purge_test.go:71-73,101-123` | `ListTrash_WithItems` на mock: cutoff игнорируется; нет assert, что service передаёт ~`now-30d` | Cutoff считается в service (`service.go:195`), фильтр — в repo. Mock-тест не бьёт guard и не фиксирует аргумент cutoff | Captured cutoff в mock ±1m от `now-30d`; retention — repo-тесты (W6-A5-002) |
| W6-A5-008 | P3 | `service_limits_test.go:267-278` | `ReassignTasks` на `mockAuditRepo` (возвращает 0) → `assert.Equal(0, count)` | Зелёный, даже если service не вызывает repo | Mock counters assigned/reported; real-repo reassign уже в trash tests — связать или усилить mock |
| W6-A5-009 | P3 | `service_security_test.go:424-433` | `TestRecordHTTPRequest_SkipsOnCreateError` не инжектит ошибку Create; mock всегда success | Имя обещает skip-on-error; проверяет лишь «create вызван» | `createErr` в mock; assert createCalls==1 + no panic; или переименовать |
| W6-A5-010 | P3 | `repository_extra_test.go:79-87` | `DeleteUserGDPR_Success` — чистый mock (`gdprUserID`), без каскада | Hard-delete/cascade живут в repo+service; mock success не ловит пропуск `HardDeleteUser`/реакций. Частично закрыто `WithReassign` + `JoinsCascadeErrors` | Оставить mock как wiring; evidence каскада — только real-repo / stub Join |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| **assert-free** | Во всех 12 файлах у каждого `Test*` есть `assert.*`/`require.*` или `t.Fatal`/`t.Errorf` (`service_test.go`). Строго assert-free **нет**. |
| `repository_newspost_trash_test.go` | Soft-delete → list (ID+EntityType) → restore → `DeletedAt` nil; retention −40d/cutoff −30d; purge + IsNotFound; **service** `ListTrash`/`RestoreItem("newspost")` на **реальном** Ent — эталон mock-vs-repo. |
| `repository_scheduled_call_trash_test.go` | То же для scheduled_call + `IsActive` после restore; ViaTrashAPI на real repo. |
| `repository_trash_test.go` (кроме retention gap) | Soft/restore issue+project; purge issues + IsNotFound; reassign assignee; deactivate+sessions; HardDeleteUser. Guard на **repo** — тесты бьют Ent. |
| `repository_test.go` | Real Ent ListAuditLogs: pagination, filters, date range; SlowOnly skip на sqlite (честный) — SQL-guard не подменён вечным PASS. |
| `retention_test.go` | SystemSetting days + fallback; PurgeOldCallRecordings old vs fresh + RunRetentionCleanup; NilMongo no-op — asserts на Ent state. |
| `service_limits_test.go` (кроме Reassign) | Settings/batch/IsSystemAdmin на real repo; `UpdateResourceLimit` allowlist A1-013; CheckLimit under/at/over; batchCountingMock — guard batch в **service**, слой верный. |
| `service_trash_test.go` | ListTrash empty на real repo; SoftDelete/Restore unknown → error; SoftDeleteIssue NotFound на Ent. |
| `service_security_test.go` (кроме misnamed) | BanUser reason/self/deact/audit на mock — guards в **service**, слой верный; RecordHTTP status→action; List filters/pagination/SlowOnly на in-memory mock (SQL SlowOnly — `repository_test`). |
| `service_test.go` | GDPR `errors.Join` cascade на stub — service-layer; daysRemaining/mapSubject/CRUD/extract — pure helpers. |
| `service_deactivate` DeactivateUser_InvalidatesCache | Mock + miniredis: service orchestration deactivate+sessions — OK. `RestoreItem_Issue` пишет `restoreID`. |

### Вердикт по фокусу

| Чек-лист TASK-008 (тесты) | Статус |
|---------------------------|--------|
| assert-free тест | **Не найдено** |
| Тест на моке там, где guard в репозитории | **Да:** retention list (`DeletedAtGT`) для issue/file/project/test_case не покрыт repo-тестами (W6-A5-002); ListTrash/Purge service-моки игнорируют cutoff (W6-A5-001/007). Newspost/scheduled_call ViaTrashAPI — корректный real-repo слой |
| Тест, который пройдёт и на сломанном коде | **Да:** W6-A5-001 (purge orchestration), W6-A5-003 (AllTypes routing), W6-A5-005 (project filter HTTP) |
| trash/purge quality | Repo soft/purge для 6 типов **есть**; service PurgeExpired — **hollow**; retention-hide — только 2 из 6 типов |

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | 12 / 12 / 0 |
| Находок | **10** |
| P0 | 0 |
| P1 | 0 |
| P2 | 5 |
| P3 | 5 |

### Топ-3

1. **W6-A5-001** — `PurgeExpired_Succeeds` не доказывает вызов всех purge и cutoff 30 дней (всегда `return nil` в prod).  
2. **W6-A5-002** — retention-hide 30d в list-корзине не покрыт для issue/test_case/file/project (guard в repo).  
3. **W6-A5-003 / W6-A5-004** — SoftDelete/Restore AllTypes без call-recording; purge test_case/file/project без IsNotFound после hard delete.

P0/P1 в remaining audit tests **нет**. Newspost/scheduled_call trash+ViaTrashAPI — сильный образец; главные риски — hollow service purge и дыры retention на основных типах корзины.
