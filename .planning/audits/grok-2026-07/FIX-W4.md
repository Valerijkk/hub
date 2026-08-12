# FIX-W4 — Residual IDOR tracker/TMS (A3-002…A3-012 hardening)

**Ветка:** `cursor-grok-4.5`  
**Дата:** 2026-07-27  
**Предшественник:** FIX-W3 (кодовые фиксы A3-002…012 уже влиты в ветку)

Docker: `postgres` / `mongo` / `redis` / `nats` via `docker-compose.local.yml`;  
`POSTGRES_TEST_URL` / `MONGODB_TEST_URL` set. Прогоны с живой БД.

---

## Цель волны

W3 закрыл первичные фиксы A3-002…012, но часть guard’ов жила только в service/mock.
W4: residual IDOR на том же контуре (issues / TC / runs / results / comments / Allure / links) —
**Ent Where + Postgres Ent-тест + mutation proof**.

---

## Самопроверка severity

| ID / finding | Tracked prod? | Вердикт |
|--------------|---------------|---------|
| A3-007 GetTestCaseVersion без HasTestCaseWith | Да | **P1, чинить** (service Edges only) |
| A3-012 / run UUID — `:projectId` игнорируется | Да | **P1, чинить** |
| A3-011-link DeleteIssueLink по linkId | Да | **P1, чинить** |
| A3-011-react ToggleReaction по commentId | Да | **P1, чинить** |
| A3-011-target CreateIssueLink чужой target | Да | **P1, чинить** |
| A3-003-link LinkTestCaseToIssue чужой issue | Да | **P1, чинить** |
| A3-004/008 Ent-тесты (mock-only gap) | Да | **test-only close** |
| A3-002/003/005/009/010/011 comment update | Уже CLOSED_SOLID в W3 | Не трогали |

---

## Baseline / after hubtest (с БД)

| | ok | fail | `?` |
|--|----|------|-----|
| **ДО (baseline)** | **41** | **1** (`testframework/tms` build failed mid-WIP) | **3** |
| **ПОСЛЕ** | **42** | **0** | **3** (factory, mongotest, steps) |

Артефакты: `w4-baseline-hubtest.txt`, `w4-after-hubtest.txt`, `w4-after-summary.txt`.

> Baseline fail = незавершённый WIP сигнатур `GetTestRunInProject` во время triage, не регрессия продукта. After: ok не уменьшился.

---

## Закрытые находки

### A3-007 — Rollback version ownership → Ent

**Коммит:** `ca5053c`  
**Фикс:** `GetTestCaseVersion(ctx, testCaseID, versionID)` + `HasTestCaseWith`.  
**Тест:** `TestEntRepository_GetTestCaseVersion_ScopedByTestCase_A3_007` (Postgres).  
**Mutation:** снять `HasTestCaseWith` → Ent FAIL; mock A3-007 остаётся зелёным.

### A3-011-link — DeleteIssueLink

**Коммит:** `058a47c`  
**Фикс:** `DeleteIssueLink(sourceIssueID, linkID)` + `HasSourceWith`.  
**Тест (sqlite Ent):** `TestEntRepository_DeleteIssueLink_Scoped_A3_011`.  
**Тест (Postgres Ent twin):** `TestEntRepository_DeleteIssueLink_Scoped_Postgres_A3_011` — Skip без `POSTGRES_TEST_URL`.  
**Mutation:** снять `HasSourceWith` → FAIL.

### A3-011-react — ToggleReaction

**Коммит:** `7a03fad`  
**Фикс:** `ToggleReaction(issueID, commentID, …)` + comment `HasIssueWith`.  
**Тест (sqlite Ent):** `TestEntRepository_ToggleReaction_Scoped_A3_011`.  
**Тест (Postgres Ent twin):** `TestEntRepository_ToggleReaction_Scoped_Postgres_A3_011` — Skip без `POSTGRES_TEST_URL`.  
**Mutation:** снять `HasIssueWith` → FAIL.

### A3-011-target — CreateIssueLink cross-project

**Коммит:** `ca8b8d4`  
**Фикс:** `FindIssueInProject` + `issueProjectIDEQ` + reject foreign target → 404.  
**Тест (sqlite Ent):** `TestEntRepository_FindIssueInProject_RejectsForeignProject_A3_011`.  
**Тест (Postgres Ent twin):** `TestEntRepository_FindIssueInProject_RejectsForeignProject_Postgres_A3_011` — Skip без `POSTGRES_TEST_URL`.  
**Mutation:** снять `issueProjectIDEQ` → FAIL (не `HasProjectWith` — A8-001 rewrite).

### A3-012 — GetTestRun project scope (+ run UUID class)

**Коммит:** `fe9fedc`  
**Фикс:** `GetTestRunInProject` / `GetTestRunWithResultsInProject`; все HTTP run-ops с `:projectId`.  
**Тест:** `TestEntRepository_GetTestRunInProject_RejectsForeign_A3_012`.  
**Mutation:** снять `HasProjectWith` → FAIL.

### A3-003-link — LinkTestCaseToIssue foreign issue

**Коммит:** `3d4d163`  
**Фикс:** TMS `FindIssueInProject` перед link/unlink.  
**Тест:** `TestLinkTestCaseToIssue_RejectsForeignIssue_A3_003`.  
**Mutation:** снять `HasProjectWith` → FAIL.

### A3-004 — Ent Postgres VersionEQ (W3 follow-up)

**Коммит:** `9d0cec1`  
**Тест:** `TestEntRepository_UpdateTestCase_OptimisticLock_A3_004`.  
**Mutation (ранее):** снять `VersionEQ` → Ent FAIL, mock PASS.

### A3-008 — Ent Postgres HasRunWith (W3 follow-up)

**Коммит:** `42ec64f`  
**Тест:** `TestEntRepository_UpdateTestResult_RejectsWrongRun_A3_008`.  
**Mutation (ранее):** снять `HasRunWith` → Ent FAIL, mock PASS.

---

## Коммиты W4

```text
ca5053c fix(tms): scope GetTestCaseVersion by owning test case (A3-007)
058a47c fix(tracker): scope DeleteIssueLink by source issue (A3-011-link)
7a03fad fix(tracker): scope ToggleReaction by issue comment (A3-011-react)
ca8b8d4 fix(tracker): reject cross-project target in CreateIssueLink (A3-011-target)
fe9fedc fix(tms): scope GetTestRun by project at Ent (A3-012)
3d4d163 fix(tms): scope LinkTestCaseToIssue by project issue (A3-003-link)
9d0cec1 test(tms): cover UpdateTestCase VersionEQ on Postgres Ent (A3-004)
42ec64f test(tms): cover UpdateTestResult HasRunWith on Postgres Ent (A3-008)
6342059 test(tracker): Postgres Ent twins for A3-011 IDOR guards
```

> **WEAK close (2026-07-28):** A3-011-link/react/target ранее числились как «Postgres Ent», но тесты шли через `env.New` → sqlite. Добавлены Postgres twins + Skip без `POSTGRES_TEST_URL`; mutation-комментарии target обновлены на `issueProjectIDEQ` (A8-001).

---

## Out of scope / next waves

| Item | Куда |
|------|------|
| A5-003…010 GitHub/GitLab/imex IDOR | **W5** |
| A2-002…007 calls/push | **W6** |
| A1-005/006, A7-001, A4-003/004, A9-012 | **W7** |
| Board/label UUID IDOR (P2 class) | Later / Block F |
| A3-005/006 reliability | **W8/W9** (уже закрыты в W3; re-verify там) |

*Merge в `master` — только пользователем.*
