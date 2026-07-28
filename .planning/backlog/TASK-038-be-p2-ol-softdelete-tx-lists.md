# TASK-038 — OL VersionEQ + soft-delete mutate + tx + list caps

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: На проверке (partial)  
> Источник: TASK-008 W12 · Classes: P2-BE-03..06 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P2-BACKEND.md`

## [OUTCOME]

Дефекты P2-BE-03..06 закрыты.

## [ROLE]

- Role: be1
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: backend/internal/{tracker,tms,sprint,bases,calls}/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P2-BE-03..06)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P2-BE-03..06 с мутацией «да» закрыты по реестру
- **Мутация (C):** без VersionEQ→оба PUT 200; soft-deleted Update→404; fail mid-tx→rollback; List без Limit→cap
- Regression: соответствующий `testframework` / Vitest domain package
- Evidence: вывод тестов, SHA коммитов
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое, не только mock
- Co-authored-by запрещён

---

## Progress (2026-07-28, be1)

**Статус:** На проверке (partial) — приоритетные in-scope баги с мутацией закрыты; остаток ниже.

### Closed Source IDs

| Class | Closed | Что сделано |
|-------|--------|-------------|
| P2-BE-03 | W1-A2-007, W1-A3-004, W2-A5-004 | `MoveIssue` VersionEQ+AddVersion; bases Update* VersionEQ в WHERE |
| P2-BE-04 | W1-A3-005, W2-A3-004, W2-A3-005, W3-A5-011 | UpdateIssue/UpdateTestCase + FindTestCase* + UpdateScheduledCall: `DeletedAtIsNil` |
| P2-BE-05 | W1-A3-006, W2-A5-001, W2-A5-002 | LogTime/StopTimer tx; CreateBase/CreateTable atomic tx; UpdateRecord cells merge |
| P2-BE-06 | W2-A5-005 | TableView `total`/`has_more` при cap MaxTableRecords |

### Remaining Source IDs

**In-scope (allowed modules), not done this wave:**

- W1-A4-008 — Sprint `version` schema + VersionEQ (**нужна Ent-миграция / approval**)
- W2-A2-009 — TMS `CreateTestCaseVersion`+`UpdateTestCase` одна tx
- W2-A2-011 — ImportAllureResults multi-step cleanup/tx
- W2-A4-006 — importer batch без all-or-nothing tx
- W2-A5-006 — bases sqlite DSN unique per test (flaky hygiene)
- W3-A5-009 — recording orphan egress compensation
- W3-A5-012 — finalizeCreate fail on AddParticipant
- W2-A2-006, W2-A2-007, W2-A2-008 — TMS reports N+1 batch
- W2-A3-006, W2-A3-007 — TMS mongo/list unbounded caps
- W2-A4-007 — export silent trunc / has_more

**Out of scope (files not allowed by TASK):**

- W7-A2-003 (intranet OL), W7-A3-008 (imex ZIP trash)
- W1-A5-008 (project cleanup), W3-A2-009/011, W3-A3-006/007 (messenger)
- W6-A3-008 (rbac), W7-A2-004/010 (intranet/calendar), W7-A4-008 (dashboard), W7-A5-003/004 (marketplace)

### Evidence

- `go test -tags hubtest -count=1 ./testframework/{tracker,bases,tms,calls}/` — PASS
- Тесты: `ol_softdelete_tx_task038_test.go`, `softdelete_find_update_task038_test.go`, `ol_tx_merge_hasmore_task038_test.go`, `softdelete_update_task038_test.go`
