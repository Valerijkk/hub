# W19-A3 — Combinatorial defects (Stage A × Stage B residual)

> 2026-07-28 · Агент **W19-A3** (security-auditor) · ветка `cursor-grok-4.5` · **no push master**  
> Задача: [TASK-008](../../backlog/TASK-008-total-audit-program.md) § Stage C / W19  
> Фокус: authz bypass chains · webhook+scope · soft-delete+search · FE empty+mutate  
> Метод: пары (finding X + finding Y) или (fix A + residual B), где **совместный** дефект тяжелее каждого по отдельности. Код — **текущий** после Stage B / W18.

## Сводка по severity

| Severity | Count |
|----------|------:|
| CRITICAL | 1 |
| HIGH     | 5 |
| MEDIUM   | 5 |
| LOW      | 2 |
| **Total** | **13** |

CRITICAL/HIGH блокируют «чисто» для Stage C red-team; MEDIUM/LOW — backlog / accepted с обоснованием.

---

## Комбинации

### C-001 · CRITICAL — GDPR Mongo no-op + HardDelete cascade

| | |
|--|--|
| **Пара** | W18-C5-001 (`gdpr_mongo.go` silent success) **+** P1-BE-34 cascade (`AuditService` erase → HardDelete) |
| **Почему хуже суммы** | Отдельно: no-op eraser = «0 deleted». Отдельно: cascade abort-on-err = защита. Вместе: `DeleteMessagesBySender` возвращает `(0, nil)` при `coll==nil` → cascade считает Mongo OK → **HardDelete user** при живых сообщениях в Mongo (right-to-erasure провален без ошибки). |
| **Файлы** | `backend/internal/audit/gdpr_mongo.go:19-30` · `backend/internal/audit/service.go:523-544` |
| **Stage A / B** | Catch-up W18-C5 · TASK-034 (GDPR cascade) PASS · residual TASK-079 |
| **Mutation idea** | `NewMongoMessageEraser(nil).DeleteMessagesBySender` → `err != nil`; cascade-тест: nil eraser wired → HardDelete **не** вызывается / erase returns error. |

---

### C-002 · HIGH — Soft-delete project + webhooks remain ListActive

| | |
|--|--|
| **Пара** | Project soft-delete (TASK-009 path) **+** webhook project-scope (TASK-012) без фильтра `project.deleted_at` |
| **Почему хуже суммы** | Soft-delete скрывает проект в UI/корзине. Scope не даёт cross-project leak. Вместе: `DeleteProject` не soft-delete'ит endpoints; `ListActiveWebhookEndpoints` только `IsActive` + `endpoint.DeletedAtIsNil` → **внешние URL продолжают получать события** с `project_id` корзины (outbox/delayed/residual mutations) до 30д. Пользователь ожидает «удалил = стоп интеграций». |
| **Файлы** | `backend/internal/project/service.go:207-228` · `backend/internal/apigateway/repository.go:168-176` · `backend/internal/apigateway/webhook.go:210-234` |
| **Stage A / B** | W1-A5-001 / W4-A5-002 · TASK-009 PASS · TASK-012 PASS · **residual combo** |
| **Mutation idea** | Soft-delete project → `ListActive` / `dispatchEvent` **не** enqueue для ep этого project; тест: endpoint active, project `deleted_at` set → 0 jobs. Fix: JOIN/`HasProjectWith(DeletedAtIsNil)` или cascade soft-delete endpoints в `DeleteProject`. |

---

### C-003 · HIGH — AI access checker fail-open + issue/review consumers

| | |
|--|--|
| **Пара** | W18-C5-004 (`access == nil` skip) **+** P1-BE-31 bot issue/review authz |
| **Почему хуже суммы** | `RBACProjectAccess` fail-closed на `Svc==nil`. Consumers (`IssueContext` / `authorizeReview`) при `access==nil` **пропускают** проверку → IDOR title/description issue и PR/MR context через @hub-ai. Wiring сейчас ставит checker (`wiring_audit_ai.go` / `wiring_integrations.go`), но любой забытый `SetAccessChecker` / тест-харнесс = полный authz bypass поверх «исправленного» P1-BE-31. |
| **Файлы** | `backend/internal/ai/issue_lookup.go:56-65` · `backend/internal/ai/code_review.go:66-69` · `backend/internal/ai/project_access.go:21-25` |
| **Stage A / B** | W5-A3-001..003 · TASK-021 PASS · residual W18-C5-004 |
| **Mutation idea** | Unit без `SetAccessChecker`: foreign issue key / MR URL → deny (NotFound / error), не 200 context. Prod consumers: `if access == nil { return deny }`. |

---

### C-004 · HIGH — Storage Confirm empty entity + Confirm-without-object residual (TASK-045 FAIL)

| | |
|--|--|
| **Пара** | `validateEntityScope("","")` → nil **+** P2-BE-21 / TASK-045 mutation FAIL (Confirm без S3 object) |
| **Почему хуже суммы** | Empty entity alone = orphan confirmed file. Missing HEAD/404 alone = oversized/ghost object. Вместе: Confirm может пройти **без** entity binding и **без** доказанного object-exists guard → NF-09-класс (presign) + невидимый file metadata в project scope. |
| **Файлы** | `backend/internal/storage/service.go:144-171` · `:247-248` · TASK-045 / W18-C2b FAIL |
| **Stage A / B** | W7-A4-001/002 · TASK-045 **Открыта** (mutation gap) |
| **Mutation idea** | Confirm с `entityType=""`/`entityId=""` → 400; Confirm без object в S3 → 404 (RED при снятии HEAD). Усилить тест на production guard, не mock. |

---

### C-005 · HIGH — CallChat empty default + send mutation (FE empty+mutate)

| | |
|--|--|
| **Пара** | P3-FE-06 / W9… CallChat silent/empty **+** sendMutation always enabled |
| **Почему хуже суммы** | List fail alone = empty UI. Send alone = normal chat. Вместе: `useQuery` → `data: messages = []` **без** `isError` → лента «пуста», пользователь шлёт sensitive в звонок, считая историю пустой (история могла быть; ошибка сети скрыта). WS listener всё ещё пишет в cache. |
| **Файлы** | `frontend/src/components/calls/CallChat.tsx:60-64` · `:98-116` |
| **Stage A / B** | W12 P3-FE-06 · TASK-073 **Открыта** |
| **Mutation idea** | Mock `listChatMessages` reject → UI error/retry, **не** empty; send disabled или banner пока `isError`. Vitest: reject → no «пустой чат» + data-testid error. |

---

### C-006 · HIGH — Soft-delete search SQL OK + Meili upsert orphans

| | |
|--|--|
| **Пара** | P1-BE-32 SQL `deleted_at IS NULL` (TASK-022 PASS) **+** Meili `IndexAllIssues` upsert-only без delete |
| **Почему хуже суммы** | Ctrl+K (SQL) корректно скрывает trash. Meili после soft-delete **оставляет** старые docs (`upsert` live-only, orphan delete нет). Операторский/будущий Meili query / SEARCH-01 CLI → утечка корзины; рассинхрон двух индексов маскирует регресс SQL. |
| **Файлы** | `backend/internal/search/service.go:249+` (SQL filters) · `backend/internal/search/meilisearch_indexer.go:84-114` · `backend/cmd/meilisearch-indexer/main.go` |
| **Stage A / B** | W5-A4-001..003 · TASK-022 PASS · P2-BE-18 Meili SSRF fixed · **residual delete path** |
| **Mutation idea** | Soft-delete issue → reindex → Meili GET doc 404 / search 0 hits; или explicit `DELETE /documents` на soft-delete hook. |

---

### C-007 · MEDIUM — Search shard total-outage empty + soft-delete UX

| | |
|--|--|
| **Пара** | W5-A4-007 / P2-BE-18 partial flag **+** soft-delete «0 hits = корзина чиста» |
| **Почему хуже суммы** | При `shardOK==0` код возвращает `results, partial, nil` (пустая выдача, не hard error) — `service.go:178-183`. Soft-delete filters дают те же 0 hits. Пользователь/FE без `partial` banner путает **outage** с «ничего / всё в корзине». |
| **Файлы** | `backend/internal/search/service.go:164-184` · FE Ctrl+K consumer |
| **Stage A / B** | TASK-044 (часть FAIL per W18) · residual comment «hard-error follow-up» |
| **Mutation idea** | Все шарды down → HTTP 503 / error, не empty; FE показывает banner при `partial`. |

---

### C-008 · MEDIUM — WebhooksPage list fail → [] + create mutate

| | |
|--|--|
| **Пара** | FE list fail / missing catch **+** `handleCreate` без блокировки |
| **Почему хуже суммы** | `fetchEndpoints` try/finally без catch: reject → `endpoints` остаются `[]`, loading false. UI «нет вебхуков» + Create всё ещё доступен → дубли URL/secret, ложное «интеграций нет» при 403/5xx. |
| **Файлы** | `frontend/src/modules/settings/webhooks.tsx:54-67` · `:73-80` |
| **Stage A / B** | FE-P1-TRASH-WEBHOOKS / TASK-032 (projectId) PASS · residual empty+mutate |
| **Mutation idea** | `listWebhooks` reject → alert/retry, Create disabled; Vitest. |

---

### C-009 · MEDIUM — BYOK silent refresh empty + add provider mutate

| | |
|--|--|
| **Пара** | P3-FE-06 BYOK catch empty **+** `handleAdd` create |
| **Почему хуже суммы** | Fail list → `providers=[]` (looks like no keys). User adds again → duplicate BYOK secrets / wrong mental model of what is stored. |
| **Файлы** | `frontend/src/components/ai/byok-settings.tsx:20-26` · `:32-51` |
| **Stage A / B** | TASK-073 **Открыта** |
| **Mutation idea** | Reject list → error state + disable Add; toast on load fail. |

---

### C-010 · MEDIUM — In-flight webhook delivery + soft-delete endpoint

| | |
|--|--|
| **Пара** | Soft-delete endpoint (`DeletedAt`) **+** `deliverWithRetry` holds `webhookJob{ep}` |
| **Почему хуже суммы** | ListActive перестаёт видеть ep (OK). Уже queued job доставляет с plaintext secret в памяти ещё 3 retry (1s/5s/30s) после «удалил вебхук». Scope/match уже пройдены. Revocation lag = data exfil after UI delete. |
| **Файлы** | `backend/internal/apigateway/webhook.go:254-295` · `repository.go:260-261` |
| **Stage A / B** | P2-BE-08 soft-delete webhooks · TASK-039 PASS |
| **Mutation idea** | Перед каждой attempt: re-check `DeletedAtIsNil`/`IsActive` by ID; else abort. Test: soft-delete mid-retry → no further POST. |

---

### C-011 · MEDIUM — Storage non-issue entity types without membership + soft-delete target

| | |
|--|--|
| **Пара** | `validateEntityScope` message/avatar/channel = only non-empty ID **+** soft-deleted message/channel still addressable by UUID |
| **Почему хуже суммы** | Issue path checks project match + `DeletedAtIsNil`. Legacy types bind file to **any** entityID string without membership/live check → attach to soft-deleted / foreign message id. |
| **Файлы** | `backend/internal/storage/service.go:240-245` · messenger SoftDeleteMessage |
| **Stage A / B** | W7-A4-002 partial · TASK-045 residual |
| **Mutation idea** | Confirm `entityType=message` + soft-deleted / non-member message → 404/403. |

---

### C-012 · LOW — Poll silent load fail + vote gated (near-miss)

| | |
|--|--|
| **Пара** | Poll `catch` → `poll=null` **+** vote UI |
| **Почему не CRITICAL** | `handleOptionClick` требует `poll` — vote не идёт вслепую. Остаётся UX: «пропал опрос» vs error. |
| **Файлы** | `frontend/src/components/messenger/poll-message.tsx:32-39` · `:66-68` |
| **Stage A / B** | TASK-073 |
| **Mutation idea** | Reject → error UI; optional. |

---

### C-013 · LOW — Trash FE isError OK + restore mutate (near-miss after TASK-032)

| | |
|--|--|
| **Пара** | Исторически trash=`listProjects` empty **+** restore |
| **Текущее** | `trash.tsx` использует `listTrash` + `isError` + restoreMutation toast — **combo закрыт** Stage B. Фиксируем как closed reference. |
| **Файлы** | `frontend/src/routes/_authenticated/trash.tsx:36-53` |
| **Stage A / B** | FE-P1-TRASH · TASK-032 PASS (W18-C5) |

---

## Матрица фокуса → combo IDs

| Focus | Combo IDs |
|-------|-----------|
| Authz bypass chains | C-003, C-004, C-011 |
| Webhook + scope | C-002, C-008, C-010 |
| Soft-delete + search | C-006, C-007 |
| FE empty + mutate | C-005, C-008, C-009, C-012 |
| GDPR / erase | C-001 |
| Closed reference | C-013 |

---

## Рекомендуемый порядок фикса (handoff)

1. **be** — C-001 (TASK-079), C-002 (project-deleted webhook filter), C-003 (fail-closed nil access)  
2. **be** — C-004 / reopen TASK-045 mutation, C-006 Meili delete, C-010 revoke-on-retry  
3. **fe** — C-005 / C-008 / C-009 через TASK-073 (+ webhooks list isError)  
4. **be** — C-007 hard-error on total shard outage, C-011 entity live/membership  

---

## Вердикт

- Комбинаторных дефектов на текущем коде: **13** (CRITICAL 1 · HIGH 5 · MEDIUM 5 · LOW 2; из них C-013 = closed reference).  
- Открытых actionable: **12** (без C-013).  
- `verified: false` для «чистого» Stage C — есть CRITICAL/HIGH без mitigation.  
- Дата: **2026-07-28 19:30** (UTC+3).

*Артефакт W19-A3 · TASK-008 · readonly product review · docs-only commit.*
