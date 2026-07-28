# W18-C4 — мутация FE (fe2 runtime)

> 2026-07-28 · Автор Stage B: **fe1** · Проверяющий: **fe2 (C4)** · Правило: проверяет ≠ автор  
> Задачи: TASK-026, 027, 028, 029, 030, 031, 035, 047, 049, 050 (10)

## Итог

| Метрика | Значение |
|---------|----------|
| PASS | **10 / 10** |
| FAIL | **0** |
| FAIL IDs | — |

Все мутации: сломан guard в `frontend/src/**` → Vitest RED (имя ниже) → restore → GREEN (53/53 в пакете). Guard не mock-only.

## Протокол

1. `[ACCEPTANCE]` → критерий мутации  
2. Слом ровно этой защиты в production UI/logic  
3. `cd frontend && node scripts/link-testframework-node-modules.mjs && npm run test:run -- <path>`  
4. RED → restore → GREEN  
5. Mock-only? нет (все бьют в `frontend/src`)  
6. PASS → «Исправлена»

## Результаты по задачам

| ID | Class | Мутация (что сломали) | RED test | GREEN | Verdict |
|----|-------|------------------------|----------|-------|---------|
| TASK-026 | FE-P1-API-CLIENT | `client.ts`: second 401 → extra `tryRefresh` (one-retry broken) | `api client > one-retry: second 401…`; `getBlob one-retry…` (`refreshCalls` 2≠1) | ✅ | **PASS** |
| TASK-027 | FE-P1-CALLS-CHAT | `CallChat.tsx`: remove `subscribe(call:)` + drop `event==="chat.message"` filter | `subscribes to call:{callId}…` (timeout); `appends chat.message… ignores call.joined` | ✅ | **PASS** |
| TASK-028 | FE-P1-CALLS-LIFECYCLE | `upcoming-calls.tsx`: Войти → `startCall`; `call-store`: `checkActiveCall` catch → `null` | `calls joinCall(call.id) on Войти…`; `checkActiveCall propagates non-404…` | ✅ | **PASS** |
| TASK-029 | FE-P1-MESSENGER | composer: ignore `replyToMessage.id`; store: `setUnreadCounts` merge not replace | `sendMessage includes replyToMessage.id…`; `setUnreadCounts drops channels…` | ✅ | **PASS** |
| TASK-030 | FE-P1-BOARD-DND | board: dual `moveCardMutation.mutate` + `monitorForElements`; `computeInsertRank` → `"0|aaaaaa:"` | `board page source has no monitorForElements dual mutate`; all `computeInsertRank` cases | ✅ | **PASS** |
| TASK-031 | FE-P1-SPRINT-UI | poker: null before load → «не найдена»; `fetchSprints` catch → silent `[]` | `shows loading before response…`; `sets error on catch…` | ✅ | **PASS** |
| TASK-035 | FE-P1-TMS-DOCS | `environment-comparison`: env change without `setSubmitted(false)` | `hides comparison table after env change…` | ✅ | **PASS** |
| TASK-047 | P2-FE-01 | `bases-page`: ignore `basesError` → empty sidebar | `shows alert + retry instead of empty sidebar…` | ✅ | **PASS** |
| TASK-049 | P2-FE-03/04 | remove `aria-label` on PinnedMessages; Button without default `type="button"` | `PinnedMessages trigger has accessible name`; `defaults to type=button…` | ✅ | **PASS** |
| TASK-050 | P2-FE-05/06/16 | Presence JIT `bg-${green}-500`; `filterBoardIssue` ignore `myTasks`; webhooks `canMutate=true` always | Presence tokens; `myTasks hides…`; `WebhooksPage hides create for viewer…` | ✅ | **PASS** |

### TASK-035 residual (не FAIL)

Мутация docs (`./internal/...` vs `ci.yml`) **вне** `frontend/src/**` (протокол C4 step 2). Vitest на docs-path нет. FE stale-cache guard покрыт и мутирован успешно → PASS по FE-критерию.

## Команды evidence

```bash
cd frontend
node scripts/link-testframework-node-modules.mjs
npm run test:run -- \
  ../testframework/stages/unit/frontend-test/domains/auth/client-api.test.ts \
  ../testframework/stages/unit/frontend-test/domains/calls/call-chat-ws.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/calls/upcoming-calls-join.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/calls/call-store-lifecycle.test.ts \
  ../testframework/stages/unit/frontend-test/domains/messenger/message-composer-reply.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/messenger/reply-thread-unread.test.ts \
  ../testframework/stages/unit/frontend-test/domains/tracker/board-dnd-single-path.test.ts \
  ../testframework/stages/unit/frontend-test/domains/tracker/board-dnd-rank.test.ts \
  ../testframework/stages/unit/frontend-test/domains/tracker/planning-poker-loading.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/tracker/sprint-store-error.test.ts \
  ../testframework/stages/unit/frontend-test/domains/tms/environment-comparison-stale-cache.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/v4/bases-page-error.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/messenger/a11y-names-p2.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/platform/a11y-button-type-p2.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/platform/tokens-dead-ui-authz-p2.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/tracker/board-my-tasks-filter-p2.test.ts
```

- **RED (мутации активны):** 14 files failed, 21 failed tests (см. таблицу)  
- **GREEN (после restore):** `Test Files 16 passed · Tests 53 passed`

## Mock-only check

| Task | Guard layer | Test hits production? |
|------|-------------|----------------------|
| 026 | `api/client.ts` | да (stub `fetch`, не mock api) |
| 027 | `CallChat.tsx` | да |
| 028 | widget + `call-store` | да |
| 029 | composer + `messenger-store` | да |
| 030 | board route + `board-rank` | да (в т.ч. source scan) |
| 031 | poker + `sprint-store` | да |
| 035 | `environment-comparison` | да |
| 047 | `bases-page` | да |
| 049 | pinned-messages + `ui/button` | да |
| 050 | presence + `board-filters` + webhooks | да |

## Handoff

- Статусы задач → **Исправлена** (README + frontmatter)  
- Production-код после restore чист (мутации не оставлены)  
- FAIL IDs: **none**
