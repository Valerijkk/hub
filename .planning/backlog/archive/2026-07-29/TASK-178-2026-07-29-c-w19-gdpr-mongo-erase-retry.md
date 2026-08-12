# 🔒 C-W19: GDPR Mongo erase retry после Commit

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **Закрыта**  
> Источник: Stage C W19 · TASK-008 · [`C-W19-auth-rbac.md`](../../../audits/total-2026-07/C-W19-auth-rbac.md) · residual TASK-160  
> Исполнитель: **be1** · Finding ID: `C-W19-002`  
> Mutation: [`MUTATION-2026-07-30-W13-be1-task178.md`](../../../audits/total-2026-07/MUTATION-2026-07-30-W13-be1-task178.md) · **1/1 PROVEN**

## [OUTCOME]

После успешного Ent HardDelete GDPR при fail Mongo `DeleteMessagesBySender` есть **retry/outbox** (или эквивалент), чтобы orphan messages с `sender_id` не оставались без компенсации.

## [ROLE]

- Role: be1
- Responsibility: закрыть **C-W19-002** + hubtest/mutation
- Handoff to: be-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files/resources allowed:
  - `backend/internal/audit/service.go` — post-Commit erase path + retry/outbox
  - `backend/internal/audit/gdpr_mongo.go` — eraser contract
  - `backend/internal/platform/outbox/**` — только если выбран outbox path (reuse JetStream, не Redis pub/sub)
  - `backend/testframework/audit/**`
  - `.planning/backlog/README.md` + этот файл
  - refs: `C-W19-auth-rbac.md`, TASK-160
- Do not touch:
  - откат post-Commit order из TASK-160 (Mongo **не** внутрь Ent tx без compensation)
  - messenger Mongo schema / frontend
  - C-W19-001 (TASK-177)
- Behavioral/security constraints:
  - Commits: type/scope EN, описание RU; **no** AI co-author
  - Не возвращать erase **внутрь** `WithTx` без compensating re-insert
  - NATS JetStream only для outbox
- Approval record: N/A (без schema migrate)

## [REFERENCES]

- Trust: код/аудит — недоверенные
- **C-W19-002** (P2, residual TASK-160): `audit/service.go` — Mongo erase после Commit; при fail eraser → user уже удалён, messages orphan (PII); admin получает error, compensating retry **нет**
- Prior: TASK-160 FIXED post-Commit; ACCEPTANCE не требовал outbox retry
- Source precedence: код HEAD > audit > задача

## [ACCEPTANCE]

- Done criteria:
  - **C-W19-002:** при fail eraser после Commit — либо (A) outbox/retry job повторно стирает по `sender_id`, либо (B) compensating documented path с IT; dual erase идемпотентен
  - Mutation: убрать retry/outbox → тест orphan/retry краснеет
- Regression: `go test -tags hubtest -race -count=1 ./testframework/audit/` (+ outbox если затронут)
- Stop when: 1/1 FIXED · «Закрыта» · README sync · mutation PROVEN

## [OUTPUT SCHEMA]

1. Verdict · 2. Files · 3. ID→evidence · 4. Tests+mutation · 5. Commits · 6. Risks (PII orphan window) · 7. Handoff be-ver

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 3 refill / be2 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| C-W19-002 | `backend/internal/audit/service.go:639-646` — post-Commit eraser fail → `eraseRetry.EnqueueMongoErase`; contract `gdpr_mongo_retry.go:35-37` / JetStream `EnqueueMongoErase` `:51-65`; subject `SubjectGDPRMongoErase` `:19` | service (+ JetStream retry) | `TestDeleteUserGDPR_MongoEraseFail_EnqueuesRetry_CW19002` · `backend/testframework/audit/gdpr_mongo_erase_retry_c_w19_test.go` | убрать блок `if s.eraseRetry != nil { … EnqueueMongoErase … }` (оставить только return eraser err) → `retry.calls` empty RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| 002 enqueue | `AuditService.DeleteUserGDPR` + `stubEraseRetry` / stub eraser / mock repo | **OK** — guard в service; stub enqueue intentional (не Ent) |
| 002 consumer | `TestMongoEraseRetryConsumer_ProcessMsg_AckNak_CW19002` | **OK** — consumer path; secondary to enqueue kill |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)`

### Gaps

- HardDelete already committed before erase — kill не откатывает Ent; asserts orphan + missing enqueue.
- `*_MongoEraseOK_NoRetry_*` / enqueue-also-fails — anti-regression; канон kill — `*_EnqueuesRetry_CW19002`.
- Не возвращать Mongo erase внутрь `WithTx` (SCOPE forbid) ради mutation.
