# TASK-072 — FE timers cleanup + NaN guard

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P3** · Статус: **Закрыта**
> Источник: TASK-008 W12 · Classes: P3-FE-05,P3-FE-16 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P3-FRONTEND.md`
>
> Mutation C retry (2026-07-30, fe1 W13 slot 2): **4/4 PROVEN**. Evidence: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-fe1-task072-retry.md` (prior NOT_PROVEN: `MUTATION-2026-07-30-W13-fe1-task072.md`)

## [OUTCOME]

Дефекты P3-FE-05/16 закрыты.

## [ROLE]

- Role: fe2
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: frontend/src/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P3-FE-05,P3-FE-16)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P3-FE-05,P3-FE-16 с мутацией «да» закрыты по реестру
- **Мутация (C):** unmount → no setState; NaN timer blocked
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

Проверено по коду HEAD (2026-07-30, fe1 W13 slot 2 retry). Слой **frontend** (Classes **P3-FE-05**, **P3-FE-16**). Typecheck: `cd frontend && npx tsc --noEmit` — не ловит cleanup/NaN guards.

| ID | file:line | Vitest path + `it` | kill recipe |
|----|-----------|----------------------|-------------|
| W9-A1-014 | `frontend/src/components/calls/CallRoom.tsx:196-201` — `copyTimerRef` + unmount `clearTimeout(copyTimerRef.current)` | `domains/platform/timers-cleanup-p3.test.tsx` · describe `P3-FE-05/16 timer cleanup + NaN (TASK-072)` > `it("CallRoom InviteBlock clears copy timer on unmount")` | убрать cleanup / `copyTimerRef` (голый `setTimeout(() => setCopied(false))`) → source-guard FAIL |
| W9-A2-017 | `frontend/src/components/messenger/bot-markdown.tsx:88-92` — `copyTimerRef` + unmount `clearTimeout` | same file > `it("bot-markdown CodeBlock clears copy timer on unmount")` (+ runtime `it("… clearTimeout runs on unmount after copy")`) | убрать `clearTimeout` cleanup (оставить onCopy) → useEffect source-guard FAIL + runtime spy FAIL |
| W10-A3-018 | `frontend/src/modules/tracker/sprint-retro.tsx:52-56` — `savedTimerRef` + unmount `clearTimeout` | same file > `it("sprint-retro clears saved timer on unmount")` (+ runtime `it("… clearTimeout runs on unmount after save")`) | убрать `savedTimerRef` cleanup (оставить handleSave) → useEffect source-guard FAIL + runtime spy FAIL |
| W11-A3-018 | `frontend/src/stores/timer-store.ts:58-60` — `if (!Number.isFinite(startedMs)) { get().stopTimer(); return; }` | same file > `it("tick with corrupt startedAt stops timer instead of NaN elapsed")` | убрать `Number.isFinite` guard (считать `elapsed` as-is) → NaN / `isRunning` asserts FAIL |

### Gaps (mutation-proof)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| CallRoom cleanup | ✅ source-guard (+ bare-timeout ban) | P3-FE-05; W9-A1-014 **PROVEN** |
| bot-markdown cleanup | ✅ useEffect-scoped guard + runtime unmount spy | W9-A2-017 **PROVEN** |
| sprint-retro cleanup | ✅ useEffect-scoped guard + runtime unmount spy | W10-A3-018 **PROVEN** |
| timer-store NaN | ✅ unit store | P3-FE-16; W11-A3-018 **PROVEN** |
| `npx tsc --noEmit` | N/A | не primary |

**Confidence:** HIGH после C-retry (2026-07-30) — **4/4 PROVEN** (см. MUTATION-2026-07-30-W13-fe1-task072-retry.md).
