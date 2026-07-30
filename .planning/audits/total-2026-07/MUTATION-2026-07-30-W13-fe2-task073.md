# MUTATION W13 slot 5 refill — TASK-073 (fe2) · 2026-07-30

> Checker: **fe2** · Env: local Vitest · Criteria: `.planning/backlog/TASK-073-fe-p3-silent-fail-empty.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**5/5 PROVEN** · `tsc` N/A · TASK-073 → **Закрыта**

| ID | Verdict | RED |
|----|---------|-----|
| W9-A1-015 | **PROVEN** | `BYOK refresh calls toastApiError on reject (not silent empty)` |
| W9-A1-019 | **PROVEN** | `CallChat shows error≠empty when history fails` |
| W9-A2-018 | **PROVEN** | `poll-message and bookmarks use toastApiError on mutate fail` (bookmarks half) |
| W9-A3-037 | **PROVEN** | `poll-message and bookmarks use toastApiError on mutate fail` (poll half) |
| W10-A2-024 | **PROVEN** | `issue-message-links shows alert on list reject` + `issue unlink wires onError toastApiError` |
| `npx tsc --noEmit` | **N/A** | не primary (не ловит toast/error≠empty) |

Production restored (`git checkout` UI files; `git diff` empty). Leftover MUTATED: **нет**. Mock-gap: нет (RTL + source-guard на prod компонентах).

## Per-ID evidence

### W9-A1-015 — BYOK toast + QueryErrorAlert — PROVEN

```text
BREAK: byok-settings.tsx refresh catch → silent catch {} (без setLoadError / toastApiError)
→ FAIL BYOK refresh calls toastApiError on reject (not silent empty)
     AssertionError: expected "spy" to be called at least once
     UI: «Нет BYOK провайдеров» (empty на reject)
RESTORE → PASS
```

Guard: `frontend/src/components/ai/byok-settings.tsx:28-30` / `:70-75`  
Vitest: `domains/platform/silent-fail-empty-p3.test.tsx` · describe `P3-FE-06 silent fail → error UI (TASK-073)`

### W9-A1-019 — CallChat isError ≠ empty — PROVEN

```text
BREAK: CallChat.tsx — убрать ветку isError (reject → «Нет сообщений»)
→ FAIL CallChat shows error≠empty when history fails
     Unable to find role="alert"
RESTORE → PASS
```

Guard: `frontend/src/components/calls/CallChat.tsx:62-63` / `:190-204`

### W9-A2-018 — bookmarks toastApiError + QueryStateView — PROVEN

```text
BREAK: bookmarks-panel.tsx — убрать toastApiError import/call + QueryStateView
→ FAIL poll-message and bookmarks use toastApiError on mutate fail
     expected … toMatch /toastApiError/
RESTORE → PASS
```

Guard: `frontend/src/components/messenger/bookmarks-panel.tsx:43` / `:66`  
Слой: source-guard.

### W9-A3-037 — poll channelId + vote/close toast — PROVEN

```text
BREAK: poll-message.tsx — убрать channelId WS-filter; silent catch на vote/close
→ FAIL poll-message and bookmarks use toastApiError on mutate fail
     expected … toContain "Не удалось проголосовать"
RESTORE → PASS
```

Guard: `frontend/src/components/messenger/poll-message.tsx:57` / `:100` / `:131`  
Слой: source-guard. Примечание: `toMatch(/channelId/)` всё ещё матчит prop; kill vote/close toast даёт RED по recipe.

### W10-A2-024 — issue-links isError + onError toast — PROVEN

```text
BREAK: issue-message-links-panel.tsx — убрать isError UI + onError toastApiError (+ imports)
→ FAIL issue-message-links shows alert on list reject
     Unable to find role="alert" (waitFor)
→ FAIL issue unlink wires onError toastApiError
     expected … toMatch /onError:[\s\S]*toastApiError/
RESTORE → PASS
```

Guard: `frontend/src/modules/tracker/issue-message-links-panel.tsx:34` / `:45`

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs

# RED (после BREAK) / GREEN (после RESTORE)
npx vitest run ../testframework/stages/unit/frontend-test/domains/platform/silent-fail-empty-p3.test.tsx
# per-ID: -t "BYOK refresh" | -t "CallChat shows error" | -t "poll-message and bookmarks"
#         | -t "issue"
```

Baseline / final GREEN: Test Files 1 passed · Tests **6 passed** · ~1.8s
