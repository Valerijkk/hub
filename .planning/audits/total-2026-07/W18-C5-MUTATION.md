# W18-C5 — мутация FE + docs (fe1 runtime)

> 2026-07-28 · Автор Stage B: **fe2 / docs-writer** · Проверяющий: **fe1 (C5)** · Правило: проверяет ≠ автор  
> Задачи: TASK-032, 033, 034, 048, 051, 052, 053, 054, 055, 056 (10)

## Итог

| Метрика | Значение |
|---------|----------|
| PASS | **8 / 10** |
| FAIL | **2** |
| FAIL IDs | **TASK-052**, **TASK-054** |

Production после restore чист (`git status` без diff в `frontend/` / `docs/`).

## Протокол

1. `[ACCEPTANCE]` → критерий мутации  
2. Слом ровно этой защиты в production  
3. `cd frontend && node scripts/link-testframework-node-modules.mjs && npx vitest run <path>`  
4. RED → restore → GREEN  
5. Mock-only? (да → недостаточно)  
6. PASS → «Исправлена» / FAIL → «Открыта»

## Результаты по задачам

| ID | Class | Мутация (что сломали) | RED test | GREEN | Verdict |
|----|-------|------------------------|----------|-------|---------|
| TASK-032 | FE-P1-TRASH-WEBHOOKS | `listTrash` → `GET /projects` | `listTrash calls GET /projects/trash…`; `mutation guard: listProjects path…` | ✅ 4/4 | **PASS** |
| TASK-033 | FE-P1-ERROR-AS-EMPTY | `use-auth`: убран `cancelled` / token guard перед `setUser` | `ignores late getMe setUser after logout` | ✅ 5/5 | **PASS** |
| TASK-034 | FE-P1-A11Y-CONTROLS | collapsed `SidebarUserProfile`: `button` ⊃ `Link` | `collapsed SidebarUserProfile is button without nested link` | ✅ 4/4 | **PASS** |
| TASK-048 | P2-FE-02 | ConfirmDialog: `onOpenChange(false)` в `finally`; `useMarkRead` без rollback | `keeps dialog open when onConfirm rejects`; `rolls back badge when markChannelRead fails` | ✅ | **PASS** |
| TASK-051 | P2-FE-07 | `gif-picker.tsx`: удалены `clearTimeout(debounceRef…)` | `gif-picker clears debounceRef on unmount` | ✅ | **PASS** |
| TASK-052 | P2-FE-08 | `updateBase` `version?: number` → Vitest/`tsc` **не краснеют** (`expectTypeOf` runtime no-op). Sub: null-guard `buildMessageFromWsEvent` → RED | version criterion **не пойман**; `buildMessageFromWsEvent rejects null…` ловит | partial | **FAIL** |
| TASK-053 | P2-FE-09..11 | убран `revokeSessionBestEffort`; `clearSessionLocal`+broadcast; WS null-guard off; `dragOver`++counter; paste props | `refresh failure best-effort POSTs…`; `clearSessionLocal…`; `useMessages: payload null…`; dragCounter; paste zone | ✅ | **PASS** |
| TASK-054 | P2-FE-12 | docs: `618` → `515` в TROUBLESHOOTING/test_plan/TESTING/strategy | **ни один** CI/Vitest/script не упал | n/a | **FAIL** |
| TASK-055 | P2-FE-13/14 | `sidebar-user`: `bg-emerald-400` + fake Online | `sidebar-user does not claim fake Online` | ✅ | **PASS** |
| TASK-056 | P2-FE-15 | `formatRelativeTime` locale=`"ru"`; push dismiss forever; `roomsKey="static"` | `locale=en does not return Russian…`; `dismissLater hides banner until TTL…`; `depends on roomsKey join…` | ✅ | **PASS** |

## FAIL детали

### TASK-052 — Открыта

- Критерий ACCEPTANCE: «`updateBase` без `version` → TS fail».
- Мутация: `version: number` → `version?: number` в `frontend/src/api/bases.ts`.
- `npx vitest run …/api-dto-gaps-p2.test.ts` и `npx tsc --noEmit` — **зелёные** (`expectTypeOf` не enforce без typecheck-режима Vitest; в конфиге typecheck не включён).
- Другой подкласс (null `buildMessageFromWsEvent`) мутируется и краснеет — но **канонический** критерий version не защищён автотестом → статус **Открыта**.
- Нужно: `vitest typecheck` **или** compile-fail fixture / `// @ts-expect-error` harness на omit `version`.

### TASK-054 — Открыта

- Критерий: вернуть «515 / 39 hooks» → расхождение; канон counts → grep/script зелёный.
- Мутация docs counts `618`→`515` — **ни один** automated check не упал (нет Vitest/CI gate на inventory drift).
- Вердикт по протоколу C5: **«нет автотеста на drift»** → **Открыта**.
- Следующий шаг: TASK-079 (скрипт/CI assert counts).

## Mock-only check

| Task | Guard layer | Test hits production? |
|------|-------------|----------------------|
| 032 | `api/projects.ts` + `WebhooksPage` | да (spy на `api.get`, реальный page) |
| 033 | `use-auth` / ActivityFeed / ErrorBoundary / dashboard-store | да |
| 034 | `sidebar-user` / SidebarLink / NotificationPanel | да |
| 048 | `confirm-dialog` + `use-unread` | да |
| 051 | source scan `gif-picker.tsx` | да (prod source) |
| 052 | `bases.ts` types + `messenger.ts` | types **не** enforced; messenger — да |
| 053 | `client.ts` / auth-store / use-messages / use-file-upload | да |
| 054 | docs only | **нет автотеста** |
| 055 | source scan `sidebar-user.tsx` | да |
| 056 | `format-time` / `push-notifications` / `use-websocket` source | да |

## Команды evidence

```bash
cd frontend
node scripts/link-testframework-node-modules.mjs
npx vitest run \
  ../testframework/stages/unit/frontend-test/domains/platform/trash-webhooks-p1.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/platform/error-empty-auth-p1.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/platform/a11y-sidebar-notifications-p1.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/platform/confirm-dialog-mutation.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/messenger/mark-read-rollback.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/platform/effects-hooks-race-p2.test.ts \
  ../testframework/stages/unit/frontend-test/domains/platform/api-dto-gaps-p2.test.ts \
  ../testframework/stages/unit/frontend-test/domains/auth/client-api.test.ts \
  ../testframework/stages/unit/frontend-test/domains/auth/auth-store.test.ts \
  ../testframework/stages/unit/frontend-test/domains/messenger/ws-payload-guards.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/platform/file-upload-drag-paste.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/platform/shell-tracker-ux-p2.test.ts \
  ../testframework/stages/unit/frontend-test/domains/platform/i18n-push-rooms-p2.test.ts
```

## Handoff

- PASS → **Исправлена**: 032, 033, 034, 048, 051, 053, 055, 056  
- FAIL → **Открыта**: 052, 054 (+ TASK-079 на docs drift gate)  
- Catch-up product: `W18-CATCHUP-PRODUCT.md`
