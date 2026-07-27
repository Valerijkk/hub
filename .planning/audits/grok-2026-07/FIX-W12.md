# FIX-W12 — Frontend A10-001…A10-004

**Ветка:** `cursor-grok-4.5`  
**Дата:** 2026-07-27  
**Агент:** fe2  
**Правила:** 1 commit / finding · FE tests where possible · no push master

Verify vs `10-frontend.md` / `17-verification.md` — все четыре finding были OPEN в `frontend/src` до фикса.

| ID | Status | SHA | Notes |
|----|--------|-----|-------|
| A10-001 | FIXED | `903c3b7` + WS Vitest | WS `channel_created`/`channel_updated` → `queryClient.setQueryData` (`upsertChannelInList` / `patchChannelInList`); UI читает Query. Helpers: `use-channels-cache.test.ts`. **Handler path (WEAK close):** `use-channels-ws.test.tsx` — FakeWS → `setQueryData` / `useChannels().data` |
| A10-002 | FIXED | `35ad363` | Issues list → `useInfiniteQuery` + `flattenIssuePages` / `nextIssuesCursor`; «Загрузить ещё» мержит все pages. Тест: `issues-cursor.test.ts` |
| A10-003 | FIXED | `9b5e1ad` | WS backoff reconnect on unexpected `onclose`; `disconnect` cancels timer (no leak). Тест: `ws-store-reconnect.test.ts` |
| A10-004 | FIXED | `257e3f6` | `presenceQueryKeyPart` = `[...userIds].sort().join(',')` — без mutate props. Тест: `use-presence-key.test.ts` |

## A10-001 detail

`ChannelList` читает `useChannels()` → React Query. Раньше WS писал только в zustand → store↔query drift. Теперь WS обновляет `["messenger","channels"]`; sync effect по-прежнему копирует query → store для unread/side panels.

> **WEAK close (2026-07-28):** ранее покрывались только pure helpers. Добавлен `use-channels-ws.test.tsx` (FakeWS как A10-003): `channel_created` / `channel_updated` → spy `queryClient.setQueryData` + обновление `useChannels().data`.

## A10-002 detail

Старый `loadMore` делал `setAllIssues(prev => [...prev, ...data.items])` + `setCursor(next)` на *текущем* ответе: новая страница оказывалась в Query, но `displayIssues` оставался на page1 до следующего клика. `useInfiniteQuery` + flat pages закрывает off-by-one.

## A10-003 detail

`onclose` больше не оставляет UI без real-time: exponential backoff (1s…30s), token retained; intentional `disconnect` (logout / effect cleanup) сбрасывает timer и `lastToken`. Rooms re-subscribe через существующие effects на `connected`.

## A10-004 detail

`userIds.sort()` мутировал входной массив (props). Копия перед sort.

## Regression

```
cd frontend
npx vitest run ../testframework/stages/unit/frontend-test/domains/messenger \
  ../testframework/stages/unit/frontend-test/domains/tracker/issues-cursor.test.ts \
  ../testframework/stages/unit/frontend-test/domains/platform/ws-store-reconnect.test.ts
```

**Результат:** 6 files / 20 tests PASS · `tsc --noEmit` PASS на затронутых путях.

## Docs

Session note в `.planning/STATE.md` (этот wave).
