# TASK-046 — Ent unique/FK/CHECK (без VersionMixin ADR)

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Закрыта
> Источник: TASK-008 W12 · Classes: P2-BE-22 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P2-BACKEND.md`

## [OUTCOME]

Дефекты P2-BE-22 закрыты. U02/U03 — после ADR.

## [ROLE]

- Role: be1
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: backend/ent/schema/**; backend/ent/migrate/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P2-BE-22)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P2-BE-22 с мутацией «да» закрыты по реестру
- **Мутация (C):** duplicate (project,name)→unique; XOR debit/credit CHECK
- Regression: соответствующий `testframework` / Vitest domain package
- Evidence: вывод тестов, SHA коммитов
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое, не только mock
- Co-authored-by запрещён

## [REWORK 2026-07-28]

enttest Component unique + JournalLine XOR CHECK runtime.

## [W18-C3 MUTATION — 2026-07-28 be2]

**FAIL.** Evidence: `.planning/audits/total-2026-07/W18-C3-MUTATION.md`

- `TestSchema_P2BE22_UniqueFKCheck` RED when Unique/CHECK strings removed from schema source — but test is **source grep only**, not Ent-runtime duplicate Create / XOR insert.
- Insufficient vs ACCEPTANCE «duplicate (project,name)→unique; XOR debit/credit CHECK». Need enttest like TASK-025.

## [W1-be2 MUTATION — 2026-07-29]

**PASS.** Evidence: .planning/audits/total-2026-07/MUTATION-2026-07-29-W1-be2.md

- migrate/schema.go Unique false / drop XOR Checks → TestComponent_UniqueNamePerProject_P2BE22, TestJournalLine_DebitXorCredit_P2BE22 RED
- GREEN after restore. Runtime enttest (not schema grep).

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 4 refill / qa-auto1 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified · Class **P2-BE-22**

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| W8-A2-006 | `backend/ent/schema/component.go:59-60` — `index.Edges("project").Fields("name").Unique()` (+ migrate DDL Unique) | schema / Ent runtime | `TestComponent_UniqueNamePerProject_P2BE22` · `backend/testframework/ent/runtime_p2be22_task046_test.go` | Unique false / drop index → duplicate Create `(project,name)` `err==nil` → constraint assert RED |
| W8-A3-008 | `backend/ent/schema/journalline.go:63-64` — `entsql.Checks` `journalline_debit_xor_credit` `(debit_cents > 0) = (credit_cents = 0)…` | schema / Ent runtime | `TestJournalLine_DebitXorCredit_P2BE22` · `runtime_p2be22_task046_test.go` | drop Checks map / XOR SQL → Create debit+credit `err==nil` → CHECK assert RED |
| W8-A5-004 | `backend/ent/schema/testcaseversion.go:52-53` — `index.Edges("test_case").Fields("version_number").Unique()` | schema (source Contains) | `TestSchema_P2BE22_UniqueFKCheck` · `backend/testframework/ent/schema_p2be22_task046_test.go` | убрать Unique index string → Contains assert RED (**не** Ent-runtime duplicate Create) |
| W8-A1-007 | `backend/ent/schema/bookmark.go:36-48` — `edge.From("channel", Channel.Type).Field("channel_id")` | schema (source Contains) | `TestSchema_P2BE22_UniqueFKCheck` · `schema_p2be22_task046_test.go` | убрать channel edge / `Field("channel_id")` → Contains assert RED (**не** runtime FK insert) |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| W8-A2-006 | enttest sqlite Create duplicate → ConstraintError | **OK** — Ent runtime |
| W8-A3-008 | enttest sqlite XOR CHECK | **OK** — Ent runtime |
| W8-A5-004 | schema source grep only | **layer_mismatch** (GAP_NOTED): schema Contains ≠ Ent-runtime unique Create |
| W8-A1-007 | schema source grep only | **layer_mismatch** (GAP_NOTED): schema Contains ≠ runtime FK |

`layer_mismatch IDs: W8-A5-004, W8-A1-007` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)`

### Gaps

- Docs-only: **N/A**.
- ACCEPTANCE kill-канон = runtime 006+008 (см. W1-be2 PASS / W18 schema-only FAIL).
- FIXED без dedicated RED в TASK-046: W8-A2-007 (environment unique), W8-A4-005 (role unique) — schema есть, runtime/schema-assert TASK-046 **нет** → **GAP**.
- Class residuals still open (не kill targets этой волны): W8-A1-005/006, W8-A2-008…010, W8-A2-014…018, W8-A3-006/013, …
- U02/U03 VersionMixin/OnDelete — после ADR (вне mutate «да» этой TASK).

## [W13 slot 4 refill MUTATION — 2026-07-30 qa-auto1]

**PASS → Закрыта.** Evidence: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-qa1-task046.md`

| ID | Verdict | RED |
|----|---------|-----|
| W8-A2-006 | **PROVEN** | `TestComponent_UniqueNamePerProject_P2BE22` |
| W8-A3-008 | **PROVEN** | `TestJournalLine_DebitXorCredit_P2BE22` |
| W8-A5-004 | **N/A** (layer_mismatch source-assert) | `TestSchema_P2BE22_UniqueFKCheck` |
| W8-A1-007 | **N/A** (layer_mismatch source-assert) | `TestSchema_P2BE22_UniqueFKCheck` |

2 PROVEN + 2 N/A · 0 NOT_PROVEN · leftover MUTATED: нет.
