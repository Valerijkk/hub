# FIX-W5 — IDOR integrations A5-003…A5-010

**Ветка:** `cursor-grok-4.5`  
**Дата:** 2026-07-27  
**Предшественник:** FIX-W4 (`684bbc1`)

Docker: `postgres` / `mongo` / `redis` / `nats`; `POSTGRES_TEST_URL` / `MONGODB_TEST_URL` set.

---

## Цель

IDOR в GitHub/GitLab/imex: delete/trigger/CreateBranch, webhook any-secret, PDF unscoping.

---

## Самопроверка severity

| ID | Tracked prod? | Вердикт |
|----|---------------|---------|
| A5-003 | Да | **P1, чинить** |
| A5-004 | Да | **P1, чинить** |
| A5-005 | Да | **P1, чинить** |
| A5-006 | Да | **P1, чинить** |
| A5-007 | Да | **P1, чинить** |
| A5-008 | Да | **P1, чинить** |
| A5-009 | Да | **P1, чинить** |
| A5-010 | Да | **P1, чинить** |

---

## Baseline / after hubtest

| | ok | fail | `?` |
|--|----|------|-----|
| **ДО** | **42** | **0** | **3** (= W4 after) |
| **ПОСЛЕ** | **42** | **0** | **3** (factory, mongotest, steps) |

Артефакты: `w5-baseline-hubtest.txt`, `w5-after-hubtest.txt`, `w5-after-summary.txt`.

Доп. коммит: `1469922` — mock `FindIssueByKeyInProject` в `testframework/ai` (сборка после A5-007).

---

## Закрытые находки

### A5-003 — GitHub DeleteIntegration
**Коммит:** `7937060` (+ Ent в `2ffd60f`)  
**Фикс:** `DeleteIntegration(ctx, projectID, id)` + `HasProjectWith`.  
**Mutation:** снять `HasProjectWith` → Ent FAIL.

### A5-004 — GitLab DeleteIntegration
**Коммит:** `c0df516`  
**Фикс:** scoped delete + `HasProjectWith`.  
**Mutation:** снять `HasProjectWith` → Ent FAIL.

### A5-005 — GitLab TriggerPipeline
**Коммит:** `bd7a898`  
**Фикс:** `GetIntegrationInProject` перед trigger.  
**Mutation:** снять `HasProjectWith` → Ent FAIL.

### A5-006 — GitHub CreateBranch
**Коммит:** `2ffd60f`  
**Фикс:** route under `/projects/:projectId/...` + `PermTaskEdit` + Ent scope.  
**Mutation:** снять `HasProjectWith` → Ent FAIL.

### A5-007 — GitLab CreateBranch
**Коммит:** `72f55e7`  
**Фикс:** аналогично A5-006.  
**Mutation:** снять `HasProjectWith` → Ent FAIL.

### A5-008 — GitHub webhook any-secret
**Коммит:** `c2dd881`  
**Фикс:** HMAC только секретом интеграции, найденной по repo.  
**Mutation:** вернуть any-of secrets → 200 вместо 403.

### A5-009 — GitLab webhook any-token
**Коммит:** `5318fed`  
**Фикс:** token только секретом интеграции по `path_with_namespace`.  
**Mutation:** any-of → 200 вместо 403.

### A5-010 — imex PDF unscoping
**Коммит:** `0d5d862`  
**Фикс:** `GetTestRunInProject` / `GetSprintInProject`.  
**Mutation:** снять `HasProjectWith` → Ent FAIL.

---

## Коммиты W5

```text
0d5d862 fix(imex): scope PDF export run/sprint by project (A5-010)
c0df516 fix(gitlab): scope DeleteIntegration by projectID (A5-004)
bd7a898 fix(gitlab): scope TriggerPipeline by project integration (A5-005)
c2dd881 fix(github): bind webhook HMAC to matched integration (A5-008)
5318fed fix(gitlab): bind webhook token to matched integration (A5-009)
2ffd60f fix(github): scope CreateBranch to project RBAC (A5-006)
7937060 fix(github): scope DeleteIntegration by projectID (A5-003)
72f55e7 fix(gitlab): scope CreateBranch to project RBAC (A5-007)
```

*Merge в `master` — только пользователем. Next: W6.*
