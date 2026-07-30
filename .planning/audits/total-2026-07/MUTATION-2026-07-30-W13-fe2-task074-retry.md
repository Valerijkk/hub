# MUTATION W13 slot 5 refill retry — TASK-074 (fe2) · 2026-07-30

> Checker: **fe2** · Env: local Vitest · Criteria: `.planning/backlog/TASK-074-fe-p3-type-casts.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).  
> Prior: `MUTATION-2026-07-30-W13-fe2-task074.md` (W11-A1-010 NOT_PROVEN — fixtures-only).

## Verdict

**8/8 PROVEN · 1 GAP_NOTED** · TASK-074 → **Закрыта**

| ID | Verdict | RED |
|----|---------|-----|
| W9-A3-020 | **PROVEN** | (prior W13) message-composer double-cast |
| W9-A3-021 | **PROVEN** | (prior W13) sticker children cast |
| W9-A3-022 | **PROVEN** | (prior W13) render-content nodes |
| W9-A2-016 | **PROVEN** | (prior W13) widget data guards |
| W10-A2-026 | **PROVEN** | (prior W13) formatEnvConfigValue |
| W11-A2-014 | **PROVEN** | (prior W13) marketplace generics |
| W11-A1-010 | **PROVEN** | source-guard on `ai-types.ts` (3 kill recipes RED) |
| W11-A4-016 | **PROVEN** | (prior W13) extractApiError |
| W9-A4-014 | **GAP_NOTED** | нет dedicated `it` (criteria: no mutate) |

Production `ai-types.ts` restored. Leftover MUTATED: **нет**. Diff production: только усиление Vitest.

## Fix applied (before re-mutate)

`type-casts-p3.test.ts` · `it("AdminAIProvider extends base; MyAIProvider is distinct shape")`:

- `readSrc("api/ai-types.ts")` — Admin has `priority` in extends body; `MyAIProvider = AIProviderBase`; not `= AdminAIProvider`; base body has no `priority`
- `expectTypeOf<AdminAIProvider>().toHaveProperty("priority")` / `MyAIProvider` not
- fixtures kept as secondary

## W11-A1-010 evidence (retry)

```text
BREAK 1) MyAIProvider = AdminAIProvider
→ FAIL expected … toMatch /export type MyAIProvider\s*=\s*AIProviderBase\b/
RESTORE (into next kill)

BREAK 2) priority lifted onto AIProviderBase; Admin without priority in body
→ FAIL expected Admin interface body to contain priority: number
RESTORE (into next kill)

BREAK 3) priority only on My (My = AIProviderBase & { priority })
→ FAIL expected Admin interface body to contain priority: number
RESTORE → PASS
```

Guard: `frontend/src/api/ai-types.ts:20-28`  
Vitest: `domains/platform/type-casts-p3.test.ts` · same `it` + source-guard

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs
npx vitest run ../testframework/stages/unit/frontend-test/domains/platform/type-casts-p3.test.ts
→ Baseline GREEN (after test strengthen): Test Files 1 passed · Tests 8 passed (~1.4s)
→ Kill 1/2/3: Tests 1 failed | 7 passed (W11-A1-010 RED)
→ Final GREEN: Test Files 1 passed · Tests 8 passed (~1.8s)
```

## Gaps

| Guard | Ловит тест? | Примечание |
|-------|-------------|-----------|
| 8 runnable Class IDs | ✅ | **8/8 PROVEN** |
| W9-A4-014 | ❌ GAP | FIXED в коде; нет dedicated `it` |

**Confidence:** HIGH.

## Handoff

Evidence: этот файл · TASK-074 → **Закрыта** · UI/prod restored · only Vitest strengthened · fe-ver optional.
