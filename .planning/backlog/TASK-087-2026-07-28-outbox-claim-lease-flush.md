# TASK-087 — 🐛 Outbox: короткий claim lease на время flush

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Закрыта · MUTATION-2026-07-30-W8-be2
> Источник: A2W4-2-001 · Stage A2 dedup · TASK-008

## [OUTCOME]

SKIP LOCKED claim не держит строку на всё время NATS publish; lease/heartbeat или publish вне длинной tx.

## [ROLE]

- Role: be1
- Responsibility: фикс + regression/mutation-ready тест
- Handoff to: be-ver/fe-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/platform/outbox/**
  - backend/testframework/**/outbox*
  - `.planning/backlog/README.md` + этот файл
- Do not touch:
  - менять JetStream subjects без ADR
- Constraints: Conventional Commits RU description; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A (кроме schema)

## [REFERENCES]

- A2W4-2-001 NEW · outbox/worker.go:62-196
- Dedup: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - Документированный lease или publish after commit claim
  - тест concurrent worker
  - Мутация: hold-during-flush → тест краснеет
- Regression: затронутый `./testframework/...` или frontend unit
- Stop when: green + статус «Исправлена» в README

## [MUTATION CRITERIA] — verifier-ready

| ID | что ломать (path:line + фрагмент) | какой тест | слой |
|----|-----------------------------------|------------|------|
| A2W4-2-001 | `backend/internal/platform/outbox/worker.go:100-141` / `processBatch:75-97` — hold-during-flush: не `Commit` claim-tx до PublishMsg/Flush (вернуть publish внутрь claim-tx **до** Commit; `afterClaimCommitHook` сработает только после Flush) | `TestProcessBatch_ClaimReleasedBeforePublish_TASK087` (+ `TestProcessBatch_ConcurrentWorkerDuringFlush_TASK087`) | worker (Ent outbox claim) |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Комментарий |
|----|-----------|-------------|
| A2W4-2-001 | Ent OutboxEvent + embedded NATS hooks (`ExportSetAfterClaimCommitHook` / `ExportSetAfterCorePublishHook`) | Guard = Commit claim **до** publish — **Ent OK** |

`layer_mismatch IDs: (none)`

### Gaps

- Concurrent flush-тест на SQLite: write-lock семантика ≠ PG `SKIP LOCKED`; канон порядка Commit→publish — `ClaimReleasedBeforePublish`.
- Lease/`stampClaimLease` (`published_at`) — A3W4-2-002 / TASK-149; **не** мутировать как доказательство TASK-087.

## [OUTPUT SCHEMA]

- Diff summary
- Test commands + PASS
- Mutation note
- Статус задачи обновлён

## [OPERATING RULES]

- Не закрывать соседние A2 ID «заодно» без evidence
- Docs sync только если контракт API изменился
