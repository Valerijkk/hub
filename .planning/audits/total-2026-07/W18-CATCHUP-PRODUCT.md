# W18 Catch-up — product files (Stage A)

> TASK-008 · READ-ONLY product · 2026-07-28 · Агент **W18-C5** (fe1 + auditor)  
> Приоритет: ACCESS (`project_access.go`) · GDPR (`gdpr_mongo.go`) · 5 platform stubs.

## Прочитанные файлы

| Файл | Строк (≈) | Находок | Вердикт |
|------|----------:|--------:|---------|
| `backend/internal/ai/project_access.go` | 35 | 1 | замечания |
| `backend/internal/audit/gdpr_mongo.go` | 37 | 2 | замечания |
| `backend/internal/platform/middleware/validator.go` | 23 | 0 | чисто |
| `backend/internal/platform/observability/route_label.go` | 23 | 0 | чисто |
| `backend/internal/platform/cron/export_hooks.go` | 19 | 0 | чисто |
| `backend/internal/platform/database/export_hooks.go` | 21 | 0 | чисто |
| `backend/internal/platform/middleware/export_requestid_hooks.go` | 9 | 0 | чисто |

Файлов в списке: **7**. Прочитано: **7**. Пропущено: **0**.  
Residual TF (68) + docs (13) — **не** разобраны в этой сессии (время после 7 product).

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Критерий мутации | Как чинить |
|----|----------|-------------|---------|-------------------|------------------|------------|
| W18-C5-001 | **P2** | `gdpr_mongo.go:19-30` | `NewMongoMessageEraser(nil)` / `coll == nil` → `DeleteMessagesBySender` возвращает **`(0, nil)`** (silent success) | Wiring всегда делает `SetMessageEraser(NewMongoMessageEraser(mongoDB))`. Если `mongoDB == nil` (или eraser с пустым coll), GDPR cascade в `service.go` считает шаг Mongo успешным и идёт в HardDelete → **right to erasure неполный без ошибки** | Заменить no-op на `return 0, errMongoUnavailable`; тест: `NewMongoMessageEraser(nil).Delete…` → `err != nil`; cascade abort (как `MessageEraserFailAbortsHardDelete`) | Fail-closed: nil db/coll → error; wiring не вызывать `SetMessageEraser` без живого Mongo **или** явный feature-flag с логом |
| W18-C5-002 | **P2** | `gdpr_mongo.go` (весь файл) + testframework | **Нет** hubtest на реальном `MongoMessageEraser` / `DeleteMany` по `sender_id` | Cascade-тесты бьют только в **stub** `MessageEraser`. Мутация фильтра (`"sender_id"`→`"user_id"`) / BSON UUID encoding **не краснеет** на слое eraser | Docker Mongo: вставить N docs с `sender_id`, вызвать eraser, assert `DeletedCount` + count remaining; мутация поля фильтра → count не 0 | `backend/testframework/audit/gdpr_mongo_eraser_test.go` (`-tags hubtest`) |
| W18-C5-003 | P3 | `project_access.go:21-22` | Godoc: «PermProjectView **(или system admin)**» — в адаптере только `eff.Has(PermProjectView)` | Admin/founder bypass живёт в `rbac.GetEffectivePermissions` (`^0`), не в этом файле. Комментарий вводит в заблуждение при чтении изоляции | N/A (docs) | Уточнить godoc: «delegate to RBAC; system admin → full mask upstream» |
| W18-C5-004 | P3 | consumers вне файла (`issue_lookup.go:56`, `code_review.go:67`) — контекст ACCESS | При `access == nil` lookup/review **fail-open** (пропуск проверки) | Wiring ставит `RBACProjectAccess`; забытый `SetAccessChecker` → IDOR bot/review. Сам `project_access.go` fail-closed на `Svc==nil` | Убрать `SetAccessChecker` в wiring-тесте / unit → expect deny; или fail-closed в consumers | Consumers: `if access == nil { return notFound/err }` |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| `project_access.go` | `Svc==nil` → `(false,nil)`; `ent.IsNotFound` → false; иначе `Has(PermProjectView)`. Wiring: `wiring_audit_ai.go` + `wiring_integrations.go` ставят checker. Unit: `TestRBACProjectAccess_NilSvcDenies`, bot_authz deny/allow |
| `validator.go` | Тонкий adapter playground→Fiber; `wiring_fiber.go` включает `StructValidator`. Тесты: `struct_validator_test.go`, tracker handler mutation |
| `route_label.go` | Unmatched `/` + path≠`/` → `"unmatched"`; иначе `r.Path` (низкая cardinality). Тест: `pii_task043_test.go` unmatched |
| `cron/export_hooks.go` | `//go:build hubtest` re-export tick timeout / `runTick` — prod-логики нет |
| `database/export_hooks.go` | hubtest: pool limits / `envBool` — prod-логики нет |
| `export_requestid_hooks.go` | hubtest wrap `validRequestID`; prod в `requestid.go` + `requestid_task043_test.go` |

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **7 / 7 / 0** |
| Находок | **4** |
| P0 | **0** |
| P1 | **0** |
| P2 | **2** |
| P3 | **2** |

### Топ

1. **W18-C5-001 (P2)** — GDPR Mongo eraser silent success при nil coll.  
2. **W18-C5-002 (P2)** — нет integration-теста реального eraser (stub-only).  

## Задачи бэклога

| Finding | Task |
|---------|------|
| W18-C5-001 + 002 | [TASK-079](../../backlog/archive/legacy/TASK-079-gdpr-mongo-eraser-fail-closed.md) |
| TASK-054 FAIL (docs drift) | [TASK-080](../../backlog/archive/legacy/TASK-080-docs-metrics-drift-autotest.md) |

P3 (003/004) — без отдельных задач (косметика / смежный consumer; зафиксированы здесь).
