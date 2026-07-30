# MUTATION W13 slot 5 refill — TASK-077 (fe2) · 2026-07-30

> Checker: **fe2** · Env: local Vitest · Criteria: `.planning/backlog/TASK-077-fe-p3-api-auth-naming.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**6/6 PROVEN** · TASK-077 → **Закрыта**

| ID | Verdict | RED |
|----|---------|-----|
| W11-A1-008 | **PROVEN** | `getBlob: maps wire error→message and request_id` (+ `sends Bearer…` / describe `downloadCalendarExport via api.getBlob`) |
| W11-A1-009 | **PROVEN** | `2xx invalid JSON throws ApiError code 502` |
| W11-A2-011 | **PROVEN** | `auth-store broadcasts userId from setTokens/setUser` |
| W11-A2-012 | **PROVEN** | `broadcastAuth mirrors login to localStorage key` |
| W11-A2-013 | **PROVEN** | `github.ts no longer owns /link-preview client`; `messenger owns fetchLinkPreview / getLinkPreview alias` |
| W11-A3-017 | **PROVEN** | `composer helpers live in lib/, not hooks/use-*` |

Production guards restored (mutation targets only). Leftover MUTATED: **нет**.  
`hooks/use-composer-ai-suggest.ts` удалён после RED.

## Per-ID evidence

### W11-A1-008 — calendar getBlob / ApiError path — PROVEN

```text
BASELINE GREEN → PASS client-api + api-auth-naming-p3
KILL A (calendar.ts): throw new Error("export failed") вместо api.getBlob
→ FAIL downloadCalendarExport via api.getBlob > sends Bearer and downloads blob without raw fetch
     Error: export failed
RESTORE calendar
KILL B (client.ts getBlob): throw new Error("export failed") вместо mapWireToApiError
→ FAIL getBlob: maps wire error→message and request_id
     expected undefined to be 403 (нет code/request_id)
RESTORE getBlob → PASS
```

Guard: `frontend/src/api/calendar.ts:38-40` (+ `client.ts` getBlob wire map)  
Vitest: `domains/auth/client-api.test.ts`

### W11-A1-009 — 2xx invalid JSON → ApiError 502 — PROVEN

```text
BREAK: убрать try/catch вокруг JSON.parse → сырой SyntaxError
→ FAIL 2xx invalid JSON throws ApiError code 502
     expected SyntaxError … to match object { code: 502, message: 'invalid json' }
RESTORE → PASS
```

Guard: `frontend/src/api/client.ts:173-180`  
Vitest: `domains/platform/api-auth-naming-p3.test.ts`

### W11-A2-011 — setTokens/setUser broadcast userId — PROVEN

```text
BREAK: setTokens всегда userId: null; убрать setUser broadcast
→ FAIL auth-store broadcasts userId from setTokens/setUser
     expected … to match /const userId = useAuthStore\.getState\(\)\.user\?\.id/
RESTORE → PASS
```

Guard: `frontend/src/stores/auth-store.ts:155-167`  
Vitest: same file · source-guard

### W11-A2-012 — broadcastAuth localStorage mirror — PROVEN

```text
BREAK: убрать setItem/removeItem mirror в broadcastAuth
→ FAIL broadcastAuth mirrors login to localStorage key
     expected "setItem" to be called with … Number of calls: 0
RESTORE → PASS
```

Guard: `frontend/src/stores/auth-store.ts:67-82`  
Vitest: same file · runtime spy

### W11-A2-013 — link-preview ownership messenger-only — PROVEN

```text
BREAK: вернуть getLinkPreview в github.ts; rename messenger fetchLinkPreview/getLinkPreview → *_REMOVED
→ FAIL github.ts no longer owns /link-preview client (not.toMatch /link-preview/)
→ FAIL messenger owns fetchLinkPreview / getLinkPreview alias
RESTORE → PASS
```

Guard: `frontend/src/api/messenger.ts:484-490`; `github.ts` без `/link-preview`  
Vitest: same file · source-guard

### W11-A3-017 — composer helpers in lib/ — PROVEN

```text
BREAK: создать hooks/use-composer-ai-suggest.ts (naming drift)
→ FAIL composer helpers live in lib/, not hooks/use-*
     expected [Function] to throw an error
DELETE hooks file → PASS
```

Guard: `frontend/src/lib/composer-ai-suggest.ts`  
Vitest: same file · import + source-guard

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs
npx vitest run ../testframework/stages/unit/frontend-test/domains/platform/api-auth-naming-p3.test.ts \
  ../testframework/stages/unit/frontend-test/domains/auth/client-api.test.ts
→ Baseline GREEN: Test Files 2 passed · Tests 18 passed (~1.4s)
→ Per-ID RED: 6/6 FAIL as above
→ Final GREEN: Test Files 2 passed · Tests 18 passed (~2.2s)
```

## Gaps

| Guard | Ловит тест? | Примечание |
|-------|-------------|-----------|
| calendar getBlob / ApiError | ✅ | calendar kill → describe RED; getBlob map kill → Exact it RED |
| 2xx JSON 502 / auth broadcast / naming | ✅ | **6/6** Class source IDs |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)`

**Confidence:** HIGH — 6/6 runnable PROVEN.

## Handoff

Evidence: этот файл · TASK-077 → **Закрыта** · UI/guards restored · fe-ver может принимать.
