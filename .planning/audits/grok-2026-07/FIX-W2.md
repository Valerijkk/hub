# FIX-W2 — remaining P0 security/product fixes

**Ветка:** `cursor-grok-4.5`  
**Дата:** 2026-07-26  
**Предшественник:** FIX-W1 (`bd80ca2` и security-коммиты W1)

## Самопроверка severity (до фиксов)

Критерий A12-001: симптом реальный, но артефакт **не** доезжает до CI/прода (gitignored / только локальный generated).

| ID | Tracked source? | Свежий клон / CI / prod? | Вердикт |
|----|-----------------|---------------------------|---------|
| A3-001 | `sprint/service.go` | Да — CompleteSprint в релизе | **P0, чинить** |
| A6-001 | `crm/service.go` | Да — IDOR write в релизе | **P0, чинить** |
| A6-002 | `accounting/service_invoice.go` | Да — race pay в релизе | **P0, чинить** |
| A11-002 | `frontend/src/api/tms.ts` | Да — UI import мёртв в бандле | **P0, чинить** (product) |
| A1-004 | `audit/service.go` | Да — GDPR без founder-guard | **P0, чинить** |

**Завышенных среди пяти нет.** Ни один не зависит от gitignored generated / локального конфига.

---

## Baseline / after `go test -tags hubtest ./testframework/...`

| | ok | fail | `?` (no tests) |
|--|----|------|----------------|
| **ДО (baseline)** | **42** | 0 | 3 (factory, mongotest, steps) |
| **ПОСЛЕ** | **42** | 0 | 3 |

Docker/PG/Mongo на агенте недоступны; пакеты на sqlite/miniredis/mock как в baseline.

---

## A3-001 — CompleteSprint чужие issues

**Коммит:** `1ac8669`

**Что:** `GetSprintInProject(projectID, sprintID)`; CompleteSprint только после ownership.

**Тест:** `TestCompleteSprint_RejectsForeignProject`  
**Mutation:** `GetSprintInProject` → `GetSprint` → FAIL (err nil, issues moved) → restore → PASS.

---

## A6-001 — CRM Update/Delete без owner

**Семантика (до фикса):** Create пишет `created_by` / `owner`; audit/bases-паттерн — write = owner **или** system admin/founder. Org-wide List не трогали (A6-009).

**Коммит:** `6ea2c4a`

**Что:** `authorizeContactWrite` / `authorizeDealWrite`; handlers передают `userID`.

**Тест:** `TestUpdateDelete_RejectsNonOwner_A6_001` (+ admin allowed)  
**Mutation:** authorize → прямой Get → FAIL (got nil) → PASS.

---

## A6-002 — MarkInvoicePaid вне tx

**Коммит:** `2d4057a` (+ удаление случайного `.bak` в A1-004 коммите)

**Что:** `PayInvoiceAtomic` — одна Ent-tx: load issued → journal+lines → post → mark paid; rollback при ошибке. Service зовёт только atomic path.

**Тест:** `TestMarkInvoicePaid_ConcurrentSingleJournal_A6_002`  
- детерминированная модель `simulateConcurrentPay(true)` → 1 journal;  
- + два параллельных `MarkInvoicePaid` на sqlite Ent.

**Mutation:** `simulateConcurrentPay(true)` → `false` → actual 2 → FAIL → restore → PASS.

*Заметка:* на sqlite single-writer чистый TOCTOU Ent-level часто маскируется; модель + Ent smoke закрывают требование «два параллельных pay → одна проводка».

---

## A11-002 — TMS FormData через JSON.stringify

**Коммит:** `c5be0c4`

**Что:** `importTestCases` — `fetch` + `FormData` без `Content-Type: application/json` (как tracker import).

**Тест:** `import-multipart-api.test.ts`  
**Mutation / before:** `api.post(FormData)` → body `{}` → FAIL `expected FormData` → fix → PASS.

---

## A1-004 — GDPR hard-delete founder

**Семантика (до фикса):**

| Источник | Смысл |
|----------|--------|
| ORG-04 / TZ | ровно один `system_role=founder` на org |
| ECC / audit | admin erase founder = необратимая потеря org owner |
| GDPR right to erasure | субъект данных может стереть **себя** |

**Решение:** запрет GDPR target=founder если `actorID != targetID`. **Self-erase founder разрешён.** Transfer роли до self-erase — вне scope этого фикса (ops/process).

**Коммит:** `877968b`

**Что:** `DeleteUserGDPR(ctx, actorID, userID, …)` + `GetUserSystemRole`; handler 403 на `ErrCannotDeleteFounder`.

**Тест:** `TestDeleteUserGDPR_RejectsFounderByAdmin_A1_004`, `…_AllowsFounderSelfErase_A1_004`  
**Mutation:** `actorID != userID` → `false` → FAIL (got nil) → PASS.

---

## Коммиты W2

```text
1ac8669 fix(sprint): scope CompleteSprint to project ownership (A3-001)
6ea2c4a fix(crm): require owner or admin for contact/deal write (A6-001)
2d4057a fix(accounting): atomic MarkInvoicePaid in one transaction (A6-002)
c5be0c4 fix(frontend): send TMS import as multipart FormData (A11-002)
877968b fix(audit): block GDPR delete of founder by non-self actor (A1-004)
```
