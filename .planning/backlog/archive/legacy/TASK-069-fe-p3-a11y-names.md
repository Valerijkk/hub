# TASK-069 — FE a11y accessible names residual

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P3** · Статус: Закрыта
> Mutation C (W13 fe1 retry): 6/6 PROVEN · evidence `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-fe1-task069-retry.md`
> Источник: TASK-008 W12 · Classes: P3-FE-01 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P3-FRONTEND.md`

## [OUTCOME]

Дефекты P3-FE-01 закрыты.

## [ROLE]

- Role: fe1
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: frontend/src/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P3-FE-01)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P3-FE-01 с мутацией «да» закрыты по реестру
- **Мутация (C):** axe name пустой → FAIL; aria-label → PASS
- Regression: соответствующий `testframework` / Vitest domain package
- Evidence: вывод тестов, SHA коммитов
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое, не только mock
- Co-authored-by запрещён

## [MUTATION CRITERIA] — verifier-ready

Проверено по коду HEAD (2026-07-30, fe1 W13 slot 2 retry). Слой **frontend** (Class **P3-FE-01**). Typecheck: `cd frontend && npx tsc --noEmit` — не ловит aria-label.

| ID | file:line | Vitest path + `it` | kill recipe |
|----|-----------|----------------------|-------------|
| W9-A1-016 | `frontend/src/components/calls/CallPanel.tsx:65` — `aria-label={isFullscreen ? "Свернуть" : "На весь экран"}` | `domains/platform/a11y-names-p3.test.tsx` · describe `P3-FE-01 · accessible names residual (TASK-069)` > `it("CallPanel fullscreen control exposes aria-label")` | убрать `aria-label` (оставить только `title`) → `toHaveAttribute("aria-label", "На весь экран")` FAIL |
| W9-A1-017 | `frontend/src/components/calls/CallReactions.tsx:115` — `aria-label={label}` на emoji button | same file > `it("CallReactionBar emoji buttons expose aria-label")` | убрать `aria-label` → name Like/Heart FAIL |
| W9-A3-024 | `frontend/src/components/messenger/gif-picker.tsx:249` — `aria-label="Поиск GIF"` | same file > `it("GifPicker search input has accessible name when open")` | убрать `aria-label` с input → textbox name FAIL |
| W9-A3-025 | `frontend/src/components/messenger/sticker-picker.tsx:262-272` — `role="tablist"` + tab `aria-label={pack.name}`; `:308` sticker `aria-label` | same file > `it("StickerPicker pack tabs expose tablist and aria-label")` | убрать tablist/`aria-label` → tablist/tab/button name FAIL |
| W9-A3-026 | `frontend/src/components/messenger/create-channel-dialog.tsx:77` — `<Label htmlFor="channel-name">Название канала</Label>` | same file > `it("CreateChannelDialogContent name field is labelled")` | убрать `Label`/`htmlFor` → `getByLabelText("Название канала")` FAIL |
| W9-A4-011 | `frontend/src/components/tracker/checklist.tsx:133` — `aria-label="Удалить пункт"` | same file > `it("Checklist delete control has aria-label")` | убрать `aria-label` → button name FAIL |

### Gaps (mutation-proof)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| CallPanel / reactions / GIF / stickers / create-channel / checklist | ✅ RTL | 6 runnable · CallPanel via `toHaveAttribute("aria-label")` |
| W9-A3-027…030,038,039 · W9-A4-009,010 · W10-A2-022,023 · W10-A4-011 | ❌ GAP | Class source IDs без dedicated `it` в TASK-069 |
| `npx tsc --noEmit` | N/A | не primary |

**Confidence:** HIGH — 6/6 runnable PROVEN (retry); class residuals GAP_NOTED.
