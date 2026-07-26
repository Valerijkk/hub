# FIX-W3 — W2 debts + A3-002…A3-012 IDOR wave

**Ветка:** `cursor-grok-4.5`  
**Дата:** 2026-07-26  
**Предшественник:** FIX-W2 (`e2390d5`)

Docker: `postgres` / `mongo` / `redis` healthy; `POSTGRES_TEST_URL` / `MONGODB_TEST_URL` set.  
Прогоны: `go test -tags hubtest ./testframework/...` с живой БД (без SKIP из‑за Mongo/PG).

---

## Самопроверка severity (до фиксов волны)

| ID | Tracked source? | CI / prod? | Вердикт |
|----|-----------------|------------|---------|
| A3-002 | `tracker/repo_issues.go` FindIssueByKey | Да | **P1 IDOR, чинить** |
| A3-003 | `tms` FindTestCaseByKey | Да | **P1 IDOR, чинить** |
| A3-004 | `tms` UpdateTestCase без VersionEQ | Да | **P1 lost update, чинить** |
| A3-005 | `tracker/repo_time.go` LogTime/StopTimer | Да | **P1 panic, чинить** |
| A3-006 | `tracker/automation_consumer.go` Ack-on-error | Да | **P1 event loss, чинить** |
| A3-007 | `tms` RollbackTestCase version ownership | Да | **P1 IDOR/corruption, чинить** |
| A3-008 | `tms` result ↔ run / bug-from-result | Да | **P1 IDOR, чинить** |
| A3-009 | `tms` CreateBugIssue(uuid.Nil) | Да | **P1 fail/panic path, чинить** |
| A3-010 | `allure_matcher` FindTestCaseByKey | Да | **P1 IDOR** — закрыт кодом A3-003 |
| A3-011 | `tracker` comment update/delete | Да | **P1 IDOR, чинить** |
| A3-012 | `tms` ImportAllureToExistingRun | Да | **P1 IDOR, чинить** |

Завышенных нет. Все в tracked production paths.

---

## Baseline / after hubtest (с БД)

| | ok | fail | `?` (no tests) |
|--|----|------|----------------|
| **ДО (baseline)** | **41** | **1** (flaky `platform/cron` TestSchedule_RunsFnOnTick) | **3** |
| **ПОСЛЕ** | **42** | **0** | **3** (factory, mongotest, steps) |

Артефакты: `.planning/audits/grok-2026-07/w3-baseline-hubtest.txt`, `w3-after-hubtest.txt`.

Frontend A11: `import-multipart-api.test.ts` — **1/1 PASS**.

---

# ДОЛГИ W2

## Долг 1 — A6-002 Postgres concurrent pay (P0)

**Коммит:** `12cdbad`

**Проблема:** sqlite-тест + `simulateConcurrentPay` не ловили снятие CAS; гонка маскировалась сериализацией записи.

**Что:** переписан `TestMarkInvoicePaid_ConcurrentSingleJournal_A6_002` на изолированную Postgres DB; `simulateConcurrentPay` удалён. Защита — `PayInvoiceAtomic` + `Where(StatusEQ(Issued))` в prod.

**Mutation (prod CAS):** снят `invoice.StatusEQ(invoice.StatusIssued)` в `UpdateOneID` mark-paid:

```text
# mutated → FAIL
expected journals=1 actual=2; okCount=2 errCount=0
--- FAIL: TestMarkInvoicePaid_ConcurrentSingleJournal_A6_002

# restored → PASS
ok  github.com/.../accounting  (1 ok pay, 1 err, 1 posted journal)
```

---

## Долг 2 — A11-002 FormData через client.ts

**Коммит:** `1d6d81a`

**Проблема:** сырой `fetch` обходил single-flight refresh (FE-H1).

**Что:** `client.ts` — `instanceof FormData` → без `JSON.stringify` и без ручного `Content-Type`; `importTestCases` → `api.post`.

**Тест:** `testframework/stages/unit/frontend-test/domains/tms/import-multipart-api.test.ts`  
FormData доходит до fetch неизменной; `Content-Type` не выставлен вручную. **PASS.**

---

## Долг 3 — sprint Ent + аудит mock-дыр W1

**Коммит:** `d3def38`

**Что:**
- Postgres/Ent: `GetSprintInProject` — спринт чужого проекта не находится (`TestEntRepository_GetSprintInProject_RejectsForeign_A3_001`).
- Messenger (A2-001): empty `channelIDs` → empty search на repo-слое (закрыта дыра «guard только в моке»).

**Аудит W1 mock-only:** аналогичная дыра у comments/results закрыта в этой волне Ent-тестами A3-008/A3-011 (не только моки).

---

# НОВАЯ ВОЛНА A3-002…A3-012

## A3-002 — FindIssueByKey без project

**Severity:** P1 IDOR · **Коммит:** `58a5f56`  
**Фикс:** `FindIssueByKey(ctx, projectID, key)` + `HasProjectWith`.  
**Тест:** `TestEntRepository_FindIssueByKey_RejectsForeignProject_A3_002`  
**Follow-up:** `47a097e` — AI `TrackerIssueLookup` резолвит project по prefix ключа, затем `GetIssue(projectID, key)` (сборка `testframework/ai` после смены сигнатуры).

## A3-003 — FindTestCaseByKey без project

**Severity:** P1 IDOR · **Коммит:** `510fb30`  
**Фикс:** scoped find; allure matcher передаёт `projectID`.  
**Тест:** `TestEntRepository_FindTestCaseByKey_RejectsForeignProject_A3_003`

## A3-004 — UpdateTestCase без optimistic lock

**Severity:** P1 · **Коммит:** `cd852a8`  
**Фикс:** `VersionEQ` + 409 на mismatch.

## A3-005 — LogTime / StopTimer panic

**Severity:** P1 · **Коммит:** `d280eaa`  
**Фикс:** LogTime пишет `StartedAt`+`EndedAt`; StopTimer — `StartedAtNotNil()`.  
**Тест:** `TestEntRepository_LogTime_DoesNotMatchStopTimer_A3_005`

## A3-006 — automation Ack-on-error

**Severity:** P1 · **Коммит:** `067a566`  
**Фикс:** Nak на unmarshal error; Ack только после успеха.  
**Тест:** `TestAutomationConsumer_NakOnBadJSON_A3_006`  
**Mutation:** всегда Ack → `naks==0` FAIL.

## A3-007 — Rollback чужой version

**Severity:** P1 · **Коммит:** `60c974e`  
**Фикс:** `version.Edges.TestCase.ID == existing.ID` (Get с `WithTestCase`).  
**Тест:** `TestRollbackTestCase_RejectsForeignVersion_A3_007`

## A3-008 — result не привязан к run / project

**Severity:** P1 · **Коммит:** `20cd8e0`  
**Фикс:** `UpdateTestResult(runID, id)` + `HasRunWith`; `CreateBugFromFailedResult` сверяет `run.Edges.Project`.  
**Тесты:** `TestExecuteTestResult_RejectsWrongRun_A3_008`, `TestCreateBugFromFailedResult_RejectsForeignProject_A3_008`

## A3-009 — CreateBugIssue(uuid.Nil)

**Severity:** P1 · **Коммит:** `c9ce31a`  
**Фикс:** link к TC только если `testCaseID != uuid.Nil`.  
**Тест:** `TestEntRepository_CreateBugIssue_NilTestCase_A3_009`

## A3-010 — Allure match без project

**Severity:** P1 · **Отдельный коммит не нужен** — закрыт в `510fb30` (matcher → `FindTestCaseByKey(ctx, projectID, key)`). Покрытие — A3-003 Ent-тест + matcher path.

## A3-011 — comment UUID IDOR

**Severity:** P1 · **Коммит:** `ece4b6b`  
**Фикс:** Update/Delete Where(issueID, authorID, commentID); handlers передают project/issueKey/userID.  
**Тест:** `TestEntRepository_UpdateDeleteComment_Scoped_A3_011` (+ мок enforce edges).

## A3-012 — Allure import в чужой run

**Severity:** P1 · **Коммит:** `28a8383`  
**Фикс:** `GetTestRun` `WithProject()`; compare `run.Edges.Project.ID == projectID`.  
**Тест:** `TestImportAllureToExistingRun_RejectsForeignProject_A3_012`

---

## Коммиты W3

```text
12cdbad test(accounting): rewrite A6-002 concurrent pay on Postgres
1d6d81a fix(frontend): FormData through api client for TMS import (A11-002)
d3def38 test(sprint,messenger): close W1/W2 repo-layer gaps (A3-001, A2-001)
58a5f56 fix(tracker): scope FindIssueByKey by projectID (A3-002)
510fb30 fix(tms): scope FindTestCaseByKey by projectID (A3-003)
cd852a8 fix(tms): enforce VersionEQ optimistic lock on UpdateTestCase (A3-004)
d280eaa fix(tracker): close manual LogTime open-timer panic (A3-005)
067a566 fix(tracker): Nak automation JetStream msgs on unmarshal error (A3-006)
ece4b6b fix(tracker): scope comment update/delete by issue and author (A3-011)
60c974e fix(tms): reject rollback of foreign test-case version (A3-007)
20cd8e0 fix(tms): bind test result updates and bugs to run/project (A3-008)
c9ce31a fix(tms): skip TC link when CreateBugIssue gets uuid.Nil (A3-009)
28a8383 fix(tms): scope Allure import to run project (A3-012)
47a097e fix(ai): resolve issue lookup via project key after A3-002
```

*Merge в `master` — только пользователем.*
