# MUTATION W13 slot 5 refill — TASK-074 (fe2) · 2026-07-30

> Checker: **fe2** · Env: local Vitest · Criteria: `.planning/backlog/archive/legacy/TASK-074-fe-p3-type-casts.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**7/8 PROVEN · 1 NOT_PROVEN · 1 GAP_NOTED** · TASK-074 → **Открыта** (W11-A1-010 NOT_PROVEN)

| ID | Verdict | RED |
|----|---------|-----|
| W9-A3-020 | **PROVEN** | `message-composer has no as unknown as Record double-cast` |
| W9-A3-021 | **PROVEN** | `sticker/render avoid bare children as unknown[] cast` |
| W9-A3-022 | **PROVEN** | `sticker/render avoid bare children as unknown[] cast` |
| W9-A2-016 | **PROVEN** | `widget data guards narrow DashboardWidgetData` |
| W10-A2-026 | **PROVEN** | `formatEnvConfigValue avoids [object Object]` |
| W11-A2-014 | **PROVEN** | `marketplace list calls use explicit generics` |
| W11-A1-010 | **NOT_PROVEN** | (kill: merge types / add `priority` to My) — Vitest остался GREEN |
| W11-A4-016 | **PROVEN** | `extractApiError does not treat plain Error as ApiError` |
| W9-A4-014 | **GAP_NOTED** | нет dedicated `it` в `type-casts-p3.test.ts` |

Production guards restored (`git checkout` / rewrite reverse). Leftover MUTATED: **нет**.

## Per-ID evidence

### W9-A3-020 — plateDoc / message-composer — PROVEN

```text
BREAK: message-composer.tsx — plateDocToContentJson →
  ((d: PlateDoc) => d as unknown as Record<string, unknown>)
→ FAIL message-composer has no as unknown as Record double-cast
     expected … not to match /as unknown as Record<string, unknown>/
RESTORE → PASS
```

Guard: `frontend/src/lib/plate-doc.ts` + `message-composer.tsx:317` / `:452`  
Vitest: `domains/platform/type-casts-p3.test.ts` · `it("message-composer has no as unknown as Record double-cast")`

### W9-A3-021 — sticker Array.isArray — PROVEN

```text
BREAK: sticker-picker.tsx:145/:156 — Array.isArray(children) →
  children as unknown[] length check
→ FAIL sticker/render avoid bare children as unknown[] cast
     expected … not to match /children as unknown\[\]/
RESTORE → PASS
```

Guard: `frontend/src/components/messenger/sticker-picker.tsx:145` / `:156`

### W9-A3-022 — render-content nodes typing — PROVEN

```text
BREAK: render-content.tsx:148 —
  const nodes: unknown[] = contentJson → const nodes = contentJson as unknown[]
→ FAIL sticker/render avoid bare children as unknown[] cast
     expected … to match /const nodes: unknown\[\] = contentJson/
RESTORE → PASS
```

Guard: `frontend/src/components/messenger/render-content.tsx:148`

### W9-A2-016 — widget data guards — PROVEN

```text
BREAK: widget-data.ts — isPersonalData/isProjectStats/isAdminStats → return !!data
→ FAIL widget data guards narrow DashboardWidgetData
     expected true to be false (isProjectStats(personal))
RESTORE → PASS
```

Guard: `frontend/src/components/dashboard/widget-data.ts:4-19`

### W10-A2-026 — formatEnvConfigValue JSON.stringify — PROVEN

```text
BREAK: format-env-config.ts — JSON.stringify(value) → String(value)
→ FAIL formatEnvConfigValue avoids [object Object]
     expected '[object Object]' to be '{"browser":"ff"}'
RESTORE → PASS
```

Guard: `frontend/src/lib/format-env-config.ts:2-12`

### W11-A2-014 — marketplace api.get generics — PROVEN

```text
BREAK: marketplace.ts:62/:68/:74 — снять <T> с api.get (оставить return-annotation)
→ FAIL marketplace list calls use explicit generics
     expected … to match /api\.get<\{ items: MarketplaceListing/
RESTORE → PASS
```

Guard: `frontend/src/api/marketplace.ts:62` / `:68` / `:74`

### W11-A1-010 — AdminAIProvider / MyAIProvider distinct — NOT_PROVEN

```text
BREAK attempts:
  1) MyAIProvider = AdminAIProvider
  2) priority/max_rpm/… lifted onto AIProviderBase; Admin=My=base
  3) swap: priority only on MyAIProvider; removed from AdminAIProvider
→ GREEN AdminAIProvider extends base; MyAIProvider is distinct shape (ожидался FAIL)
Причина: it() проверяет только локальные object literals в тесте
  (admin.priority / "priority" in mine) — изменение prod-типов не меняет
  runtime shape фикстур; source-guard на ai-types.ts в тесте нет.
RESTORE → PASS (baseline)
```

Guard: `frontend/src/api/ai-types.ts:20-28`  
Vitest: same file · `it("AdminAIProvider extends base; MyAIProvider is distinct shape")`  
Слой assert слабый (fixtures only, не readSrc / не expectTypeOf на My≠Admin) — kill recipe criteria не даёт RED.

### W11-A4-016 — extractApiError code/request_id — PROVEN

```text
BREAK: format-api-error.ts — убрать hasCode/hasRequestId; return error as ApiError
→ FAIL extractApiError does not treat plain Error as ApiError
     expected … to match /typeof candidate\.code === "number"/
RESTORE → PASS
```

Guard: `frontend/src/lib/format-api-error.ts:16-22`

### W9-A4-014 — GAP_NOTED

`comment-list` / `jql-search-page` / `automation-rules-panel` — FIXED в коде; **нет** dedicated `it` в `type-casts-p3.test.ts`. Criteria: no mutate.

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs
npx vitest run ../testframework/stages/unit/frontend-test/domains/platform/type-casts-p3.test.ts
→ Baseline GREEN: Test Files 1 passed · Tests 8 passed (~1.2s)
→ Per-ID RED: 7/8 FAIL as above; W11-A1-010 stayed GREEN
→ Final GREEN: Test Files 1 passed · Tests 8 passed (~1.4s)
```

## Gaps

| Guard | Ловит тест? | Примечание |
|-------|-------------|-----------|
| PlateDoc / sticker / render / widgets / env / marketplace / extractApiError | ✅ Vitest source/runtime | **7/8** runnable PROVEN |
| AdminAI / MyAI distinct shapes | ❌ NOT_PROVEN | fixtures-only assert |
| W9-A4-014 | ❌ GAP | нет dedicated `it` |

**Confidence:** HIGH для 7 PROVEN; W11-A1-010 — NOT_PROVEN (нужен source-guard / expectTypeOf на prod-типы).

## Handoff

Evidence: этот файл · TASK-074 → **Открыта** (W11-A1-010) · UI mutation targets restored · fe-ver: усилить `it` для ai-types (readSrc / expectTypeOf My≠Admin) перед re-run.
