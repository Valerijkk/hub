# TASK-060 — UTF-8 truncate by runes

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P3** · Статус: Закрыта
> Источник: TASK-008 W12 · Classes: P3-BE-04 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P3-BACKEND.md`
> Mutation C (W13 slot 4 refill / qa-auto1 · 2026-07-30): **5/5 PROVEN** — `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-qa1-task060.md`

## [OUTCOME]

Дефекты P3-BE-04 закрыты.

## [ROLE]

- Role: be2
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: backend/internal/{search,messenger,imex,marketplace}/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P3-BE-04)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P3-BE-04 с мутацией «да» закрыты по реестру
- **Мутация (C):** byte truncate → битый UTF-8; rune helper → ok
- Regression: соответствующий `testframework` / Vitest domain package
- Evidence: вывод тестов, SHA коммитов
- Stop when: green + статус «Закрыта» в README · **done 2026-07-30** (mutation 5/5 PROVEN)

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое, не только mock
- Co-authored-by запрещён

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 4 refill / qa-auto1 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified · Class **P3-BE-04**

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| W5-A4-010 | `backend/internal/search/truncate.go:6-14` — `truncateRunes` via `[]rune` (+ `ExportTruncateRunes`) | helper / search | `TestTruncateRunes_Cyrillic_NotMidRune`; `TestTruncateRunes_Emoji_KeepsWholeRunes` · `backend/testframework/search/rune_truncate_task060_test.go` | вернуть `s[:N]` bytes → `utf8.ValidString==false` / rune count assert RED |
| W3-A2-017 | `backend/internal/messenger/truncate.go:5-14` — `truncateRunes` для promote title (500) | helper / messenger | `TestTruncateRunes_PromoteTitle_Cyrillic`; `TestTruncateRunes_PromoteTitle_Emoji` · `backend/testframework/messenger/rune_truncate_task060_test.go` | `title[:500]` bytes → mid-rune / invalid UTF-8 → ValidString/RuneCount RED |
| W7-A3-012 | `backend/internal/imex/truncate.go:6-18` — `truncateCell` rune-safe (+ PDF `pdf.go:205`) | helper / imex | `TestTruncateCell_Cyrillic_NotMidRune`; `TestTruncateCell_Emoji_KeepsWholeRunes` · `backend/testframework/imex/rune_truncate_task060_test.go` | `cellVal[:47]` bytes → invalid UTF-8 → ValidString RED |
| W7-A5-009 | `backend/internal/marketplace/service.go:376-378` — `utf8.RuneCountInString(title/description)` limits | service | `TestValidateListingInput_CyrillicTitle_RuneLimit`; `…_OverRuneLimit`; `…_EmojiDescription_RuneLimit` · `backend/testframework/marketplace/rune_truncate_task060_test.go` | `len(title)` bytes → кириллица/emoji ложно reject/accept → RuneCount assert RED |
| W2-A2-019 | `backend/internal/tms/service_converters.go:396-403` — `normalizeErrorPattern` первые 100 **рун** | converter | `TestNormalizeErrorPattern_Cyrillic_NotMidRune` · `backend/testframework/tms/rune_truncate_task060_residual_test.go` | `msg[:100]` bytes → invalid UTF-8 pattern → ValidString/RuneCount RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| W5-A4-010 | pure `ExportTruncateRunes` | **OK** — helper (не mock≠Ent) |
| W3-A2-017 | pure `ExportTruncateRunes` messenger | **OK** |
| W7-A3-012 | pure `ExportTruncateCell` | **OK** |
| W7-A5-009 | `validateListingInput` service | **OK** — no DB |
| W2-A2-019 | pure `normalizeErrorPattern` | **OK** |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)`

### Gaps

- Docs-only: **N/A**.
- Class source IDs = **5/5** с dedicated RED (полный coverage P3-BE-04).
- Kill = byte truncate; short-string unchanged tests — anti-regression, не primary kill.
