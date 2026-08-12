# TASK-089 — 🔒 Messenger reaction + CRM contact + invoice serial

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Закрыта
> Источник: A2W3/W7 · Stage A2 dedup · TASK-008
> Mutation: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W8-qa-auto2.md` (3/3 PROVEN)

## [OUTCOME]

Concurrent AddReaction не плодит дубли; CRM deal→contact owner-scoped; invoice number не Count+1 race.

## [ROLE]

- Role: be2
- Responsibility: фикс + regression/mutation-ready тест
- Handoff to: be-ver/fe-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/messenger/mongo_repository.go
  - backend/internal/crm/service.go
  - backend/internal/accounting/invoice_repository.go
  - testframework messenger/crm/accounting
  - `.planning/backlog/README.md` + этот файл
- Do not touch:
  - FR-INTR U03/U05 greenfield
- Constraints: Conventional Commits RU description; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A (кроме schema)

## [REFERENCES]

- A2W3-3-004 · W7-A2-005 · W7-A2-007
- Dedup: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - Atomic reaction dedup
  - validateDealRefs GetContact owner scope
  - serial/unique invoice number
  - denial/race тесты + мутации
- Regression: затронутый `./testframework/...` или frontend unit
- Stop when: green + статус «Исправлена» в README

## [MUTATION CRITERIA] — verifier-ready

| ID | что ломать (path:line + фрагмент) | какой тест | слой |
|----|-----------------------------------|------------|------|
| A2W3-3-004 | `backend/internal/messenger/mongo_repository.go:242-254` — убрать `"reactions.emoji": bson.M{"$ne": emoji}` из push-фильтра (push без `$ne`) | `TestAddReaction_ConcurrentFirstReact_NoDuplicate_A2W3_3_004` | repository (Mongo) |
| W7-A2-005 | `backend/internal/crm/service.go:285-294` — в `validateDealRefs` заменить `authorizeContactWrite` на голый `GetContact` (без owner/admin) | `TestCreateDeal_RejectsStrangerContact_W7A2005` | service (Ent CRM) |
| W7-A2-007 | `backend/internal/accounting/invoice_repository.go:53-69` / `:89-106` — вернуть `Count(prefix)+1` **без** retry на `ent.IsConstraintError` | `TestCreateInvoice_ConcurrentUniqueNumbers_W7A2007` | repository (Ent invoice) |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Комментарий |
|----|-----------|-------------|
| A2W3-3-004 | `MongoRepository` (`mongotest.Connect`) | Guard в Mongo `$ne`/`$addToSet` — **Mongo OK** (не mock_repository) |
| W7-A2-005 | Ent CRM service (`enttest` sqlite) | Guard в service `authorizeContactWrite` — **Ent OK** |
| W7-A2-007 | Ent invoice + Postgres (`openAccountingPostgresEnt`) | MAX+1 + unique retry — **Ent OK** |

`layer_mismatch IDs: (none)`

### Gaps

- Reaction: мутировать только prod `mongo_repository.go`, не `mock_repository.AddReaction`.
- Invoice: concurrent IT требует Postgres; sqlite unique/retry может не воспроизвести race.

## [OUTPUT SCHEMA]

- Diff summary
- Test commands + PASS
- Mutation note
- Статус задачи обновлён

## [OPERATING RULES]

- Не закрывать соседние A2 ID «заодно» без evidence
- Docs sync только если контракт API изменился
