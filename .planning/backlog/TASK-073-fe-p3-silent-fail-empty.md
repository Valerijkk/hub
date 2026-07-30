# TASK-073 — FE silent fail → empty residual

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P3** · Статус: Закрыта
> C-мутация: 2026-07-30 fe2 W13 slot 5 · **5/5 PROVEN** · отчёт `MUTATION-2026-07-30-W13-fe2-task073.md`
> Источник: TASK-008 W12 · Classes: P3-FE-06 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P3-FRONTEND.md`

## [OUTCOME]

Дефекты P3-FE-06 закрыты.

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

- Реестр W12 (P3-FE-06)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P3-FE-06 с мутацией «да» закрыты по реестру
- **Мутация (C):** reject → isError/toast, не empty
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

Проверено по коду HEAD (2026-07-30, fe2 W13 slot 5 criteria-only). Слой **frontend** (Class **P3-FE-06**). Typecheck: `cd frontend && npx tsc --noEmit` — не ловит toast/error≠empty.

| ID | file:line | Vitest path + `it` | kill recipe |
|----|-----------|----------------------|-------------|
| W9-A1-015 | `frontend/src/components/ai/byok-settings.tsx:28-30` / `:70-75` — `toastApiError` + `QueryErrorAlert` на reject | `domains/platform/silent-fail-empty-p3.test.tsx` · describe `P3-FE-06 silent fail → error UI (TASK-073)` > `it("BYOK refresh calls toastApiError on reject (not silent empty)")` | вернуть silent `catch {}` / убрать `loadError`+alert → toast spy + `role=alert` FAIL; empty «нет byok» снова ок |
| W9-A1-019 | `frontend/src/components/calls/CallChat.tsx:62-63` / `:190-204` — `isPending`/`isError` разделены от empty | same file > `it("CallChat shows error≠empty when history fails")` (+ `it("CallChat shows empty only after successful empty load")`) | убрать ветку `isError` (reject → empty «Нет сообщений») → alert/retry FAIL; empty assert на reject проходит ошибочно |
| W9-A2-018 | `frontend/src/components/messenger/bookmarks-panel.tsx:43` / `:66` — `toastApiError` + `QueryStateView` | same file > `it("poll-message and bookmarks use toastApiError on mutate fail")` | убрать `toastApiError` из delete / `QueryStateView` → source-guard FAIL |
| W9-A3-037 | `frontend/src/components/messenger/poll-message.tsx:23` / `:57` / `:100` / `:131` — `channelId` WS-scope + vote/close `toastApiError` | same `it` (poll half) | убрать `channelId` filter / `toastApiError` vote\|close → source-guard FAIL |
| W10-A2-024 | `frontend/src/modules/tracker/issue-message-links-panel.tsx:34` / `:45` — `onError: toastApiError` + `linksQuery.isError` alert | same file > `it("issue-message-links shows alert on list reject")` + `it("issue unlink wires onError toastApiError")` | убрать `isError` UI / `onError` toast → alert FAIL + `onError:[\s\S]*toastApiError` FAIL |

### Gaps (mutation-proof)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| BYOK / CallChat / issue-links list | ✅ RTL | 3 runnable PROVEN-ready |
| poll + bookmarks + unlink onError | ✅ source-guard | 2 runnable (W9-A2-018, W9-A3-037; unlink half of W10-A2-024) |
| `npx tsc --noEmit` | N/A | не primary |

**Confidence:** HIGH — **5/5** Class source IDs с runnable criteria (нет docs-only).
