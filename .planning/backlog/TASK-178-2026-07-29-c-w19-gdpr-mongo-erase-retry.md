# 🔒 C-W19: GDPR Mongo erase retry после Commit

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **На проверке**  
> Источник: Stage C W19 · TASK-008 · [`C-W19-auth-rbac.md`](../audits/total-2026-07/C-W19-auth-rbac.md) · residual TASK-160  
> Исполнитель: **be1** · Finding ID: `C-W19-002`

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
- Stop when: 1/1 FIXED · «На проверке» · README sync

## [OUTPUT SCHEMA]

1. Verdict · 2. Files · 3. ID→evidence · 4. Tests+mutation · 5. Commits · 6. Risks (PII orphan window) · 7. Handoff be-ver
