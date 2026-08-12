# C-W19 — A-добор: sensitive gap files (STAGE C)

> TASK-008 · Stage C · Wave 19 · A-добор · READ-ONLY product · 2026-07-29  
> Агент: независимый аудитор (`cursor-grok-4.5`) · **не** закрывать TASK-008 · **не** коммитить  
> Scope: 7 product-файлов, пропущенных Stage A (см. [`W18-CATCHUP-PRODUCT.md`](W18-CATCHUP-PRODUCT.md))  
> Плюс быстрый residual-scan соседних authz / GDPR путей

---

## Вердикт

| Метрика | Значение |
|---------|----------|
| Файлов в списке | **7** |
| Прочитано | **7 / 7** |
| Пропущено | **0** |
| Новых P0/P1 в этих 7 файлах | **0** |
| Открытых findings в 7 файлах | **2×P3** (docs/coupling) |
| W18 P2 (C5-001/002) на HEAD | **FIXED** (fail-closed + hubtest IT) |
| Residual вне файлов (соседние пути) | **C-W19-002 / TASK-178** (post-Commit orphan) — уже в бэклоге |

**Итог:** gap Stage A закрыт по прочтению. Prod-контракты ACCESS / GDPR eraser / StructValidator / route-label / hubtest export-hooks на HEAD **в порядке**. Критичный residual GDPR — **не** в `gdpr_mongo.go`, а в post-Commit оркестрации `audit/service.go` (уже **TASK-178**).

---

## Per-file

### 1. `backend/internal/ai/project_access.go`

| | |
|--|--|
| **Прочитано** | ✅ 35 строк · целиком |
| **Вердикт** | **clean** (логика) · **1×P3** (godoc) |
| **Роль** | Адаптер `rbac.RBACService` → `ProjectAccessChecker` (P1-BE-31) |

**Проверено:**
- `Svc == nil` → `(false, nil)` — fail-closed (`:23-25`)
- `ent.IsNotFound` → deny без leak (`:28-30`)
- Иначе `eff.Has(rbac.PermProjectView)` (`:33`)
- System admin bypass **не** дублируется здесь — корректно делегирован в `rbac.GetEffectivePermissions` (`service.go:62-65`, full mask `^0`)
- Consumers: `issue_lookup.go:58-61`, `code_review.go:71-74` — `access == nil` → deny (W18-C5-004 **FIXED**)
- Wiring: `wiring_audit_ai.go:86`, `wiring_integrations.go:44` ставят `RBACProjectAccess{Svc: rbacSvc}`
- Unit: `TestRBACProjectAccess_NilSvcDenies` (`bot_authz_task021_test.go:83-88`)

| ID | Sev | file:line | Finding |
|----|-----|-----------|---------|
| **C-W19-GAP-001** | **P3** | `ai/project_access.go:21` | Godoc: «PermProjectView **(или system admin)**» — в адаптере только `Has(PermProjectView)`; admin/founder в upstream RBAC. Drift vs W18-C5-003 (docs) |

**Мутация (регрессия):** `return true` при `Svc==nil` → `TestRBACProjectAccess_NilSvcDenies` RED.

---

### 2. `backend/internal/audit/gdpr_mongo.go`

| | |
|--|--|
| **Прочитано** | ✅ 42 строки · целиком |
| **Вердикт** | **clean** (fail-closed контракт) · **1×P3** (coupling имени коллекции) |
| **Роль** | Mongo eraser messages по `sender_id` (P1-BE-34 / W18-C5) |

**Проверено:**
- `NewMongoMessageEraser(nil)` → eraser с `coll==nil` (`:24-26`)
- `e==nil \|\| coll==nil` → `ErrMongoMessageEraserUnavailable` — **не** `(0,nil)` (`:33-35`) → W18-C5-001 **FIXED**
- `DeleteMany(ctx, bson.M{"sender_id": senderID})` (`:36`) — поле совпадает с `MessageDocument.SenderID` (`messenger/dto.go:156`)
- Коллекция `"messages"` совпадает с `messenger.messagesCollection`
- IT: `gdpr_mongo_eraser_task079_test.go` (nil fail-closed + DeleteBySender EntMongo) → W18-C5-002 **FIXED**
- Cascade вызывает eraser **после** Ent Commit (`service.go:604-610`); fail eraser → error admin (но user уже HardDelete) → residual **C-W19-002 / TASK-178** (соседний путь, не дефект самого eraser)

| ID | Sev | file:line | Finding |
|----|-----|-----------|---------|
| **C-W19-GAP-002** | **P3** | `audit/gdpr_mongo.go:27` | Литерал `db.Collection("messages")` не шарит const с `messenger.messagesCollection` — rename коллекции → eraser бьёт мимо, `DeletedCount=0` без error |

**Мутация:** вернуть `(0,nil)` при nil coll → `TestMongoMessageEraser_NilColl_FailClosed` RED; `"sender_id"`→`"user_id"` → IT DeletedCount RED.

---

### 3. `backend/internal/platform/middleware/validator.go`

| | |
|--|--|
| **Прочитано** | ✅ 23 строки · целиком |
| **Вердикт** | **clean** |
| **Роль** | Fiber `StructValidator` ← go-playground/validator (P2-BE-01) |

**Проверено:**
- `Validate` → `validate.Struct(out)` — теги исполняются (`:15-17`)
- Wiring: `wiring_fiber.go:33` `StructValidator: middleware.NewStructValidator()`
- Тесты: `struct_validator_test.go`, tracker `handler_struct_validator_test.go`

**Findings:** нет.

---

### 4. `backend/internal/platform/observability/route_label.go`

| | |
|--|--|
| **Прочитано** | ✅ 23 строки · целиком |
| **Вердикт** | **clean** |
| **Роль** | Низкая cardinality метрик/трейсов (W4-A3-001/002) |

**Проверено:**
- `r==nil` / empty Path → `"unmatched"` (`:14-16`)
- Catch-all Fiber v3: `r.Path=="/"` && `c.Path()!="/"` → `"unmatched"` (`:18-20`) — сырой UUID path не попадает в label
- Consumers: `otel.go:110,128`, `prometheus.go:104`

**Findings:** нет.

---

### 5. `backend/internal/platform/cron/export_hooks.go`

| | |
|--|--|
| **Прочитано** | ✅ 19 строк · целиком |
| **Вердикт** | **clean** |
| **Роль** | `//go:build hubtest` re-export `defaultTickTimeout` / `runTick` |

**Проверено:** prod-логики нет; только тестовые обёртки (W4-A4-008).

**Findings:** нет.

---

### 6. `backend/internal/platform/database/export_hooks.go`

| | |
|--|--|
| **Прочитано** | ✅ 21 строка · целиком |
| **Вердикт** | **clean** |
| **Роль** | hubtest: `applyPoolLimits` / MaxOpenConns / `envBool` (P1-BE-25, W4-A4-003) |

**Проверено:** prod-логики нет.

**Findings:** нет.

---

### 7. `backend/internal/platform/middleware/export_requestid_hooks.go`

| | |
|--|--|
| **Прочитано** | ✅ 9 строк · целиком |
| **Вердикт** | **clean** |
| **Роль** | hubtest wrap `validRequestID` (W4-A1-002) |

**Проверено:** prod-валидация в `requestid.go:36-47` (UUID / ASCII ≤128 рун); export только для тестов.

**Findings:** нет.

---

## Residual scan (соседние authz / GDPR)

Быстрый обход call-sites и каскада **вне** 7 файлов:

| Зона | Результат | Уже учтено? |
|------|-----------|-------------|
| AI consumers nil-access | fail-closed (`issue_lookup`, `code_review`) | W18-C5-004 FIXED |
| GDPR post-Commit Mongo fail → orphan messages | admin error, **нет** retry/outbox | **C-W19-002 → TASK-178** (Открыта) |
| `msgEraser == nil` → шаг Mongo **пропускается** (`service.go:605`) | Wiring всегда `SetMessageEraser` при `auditSvc!=nil` (`wiring_messenger.go:66-68`); nil Mongo → fail-closed eraser после Commit | residual покрыт TASK-178 + fail-closed eraser |
| Founder GDPR guard | `ErrCannotDeleteFounder` (`service.go:498-500`) | OK |
| Ent cascade atomicity | `WithTx` + HardDelete внутри tx | A3W6-4-001 / TASK-160 |
| Mongo `allure_results` | keyed by `run_id`, **не** user; GDPR cascade **не** чистит Allure raw | наблюдение P3/informational — не IDOR; PII в raw Allure вне scope eraser |
| Deactivate ≠ WS kill | mid-session socket | **C-W19-001 → TASK-177** (соседний authz, не эти 7 файлов) |
| Collection name drift eraser↔messenger | литерал vs const | **C-W19-GAP-002** |

Новых orphan P0/P1 по residual-scan **нет**.

---

## Сводка findings

| ID | Sev | file:line | Статус |
|----|-----|-----------|--------|
| C-W19-GAP-001 | P3 | `backend/internal/ai/project_access.go:21` | NEW (docs; = W18-C5-003 residual) |
| C-W19-GAP-002 | P3 | `backend/internal/audit/gdpr_mongo.go:27` | NEW (coupling) |
| W18-C5-001 | P2 | `gdpr_mongo.go` nil silent success | **FIXED** на HEAD |
| W18-C5-002 | P2 | нет IT eraser | **FIXED** (`gdpr_mongo_eraser_task079_test.go`) |
| W18-C5-004 | P3 | consumers fail-open | **FIXED** |
| C-W19-002 | P2 | `audit/service.go:604-610` | residual → **TASK-178** (не в этих 7) |

### Чистые файлы (5/7 без open findings)

`validator.go` · `route_label.go` · `cron/export_hooks.go` · `database/export_hooks.go` · `export_requestid_hooks.go`

---

## Stop / next

- Отчёт записан: `.planning/audits/total-2026-07/C-W19-sensitive-gap-files.md`
- **Коммит не создавался**
- **TASK-008 не закрывать**
- P3 GAP-001/002 — косметика; отдельный TASK не обязателен (по желанию team-lead смержить в hygiene)
- GDPR retry — исполнять **TASK-178**, не переоткрывать TASK-079
