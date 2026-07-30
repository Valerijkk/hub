# MUTATION W13 slot 5 refill — TASK-052 (fe2) · 2026-07-30

> Checker: **fe2** · Env: local Vitest · Criteria: `.planning/backlog/TASK-052-fe-p2-api-dto-gaps.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**10/10 PROVEN** · TASK-052 → **Закрыта**

| ID | Verdict | RED |
|----|---------|-----|
| W11-A1-005 | **PROVEN** | `tsc fixture: omit version is @ts-expect-error (mutation optional → unused)` (+ source `version: number`) |
| W10-A5-013 | **PROVEN** | `GuestCallSession is assignable to ActiveCallSession without CallResponse cast` |
| W11-A2-006 | **PROVEN** | `exposes create/delete environment and updateNews/updateMemberRole` |
| W11-A2-007 | **PROVEN** | same `it` (import/`updateNews` Match) |
| W11-A2-008 | **PROVEN** | same `it` (import/`updateMemberRole` Match) |
| W11-A2-010 | **PROVEN** | `buildMessageFromWsEvent rejects null/empty/missing sender` |
| W10-A2-018 | **PROVEN** | `parsePlateValue accepts plate nodes and rejects garbage` |
| W9-A5-007 | **PROVEN** | `user-picker disables members query without projectId` |
| W10-A2-017 | **PROVEN** | `case-filters updateFilter uses Partial filters not Record unknown` |
| W10-A2-020 | **PROVEN** | `environment-comparison uses formatApiError not message cast` |

Production guards restored. Leftover MUTATED: **нет**.  
Baseline / final: Vitest **9/9** GREEN (`api-dto-gaps-p2.test.ts`).

## Per-ID evidence

### W11-A1-005 — bases `version: number` — PROVEN

```text
BREAK: bases.ts updateBase/updateRecord → version?: number
→ FAIL tsc fixture: omit version is @ts-expect-error (unused @ts-expect-error)
RESTORE → PASS 9/9
```

Guard: `frontend/src/api/bases.ts:74,116`  
Канон: tsc fixture (`updateBase-version-required.fixture.ts`); `expectTypeOf` alone = runtime no-op.

### W10-A5-013 — GuestCallSession без CallResponse cast — PROVEN

```text
BREAK: call.$code.lazy.tsx — activeCall object `as import("@/api/calls").CallResponse`
→ FAIL GuestCallSession is assignable… (source not.toMatch CallResponse cast)
RESTORE → PASS
```

Guard: `frontend/src/api/calls.ts:40-52`; `routes/call.$code.lazy.tsx`

### W11-A2-006 — create/deleteEnvironment paths — PROVEN

```text
BREAK: tms.ts create/delete → `/environments-MUTATED`
→ FAIL exposes create/delete… (source path Match)
RESTORE → PASS
```

Guard: `frontend/src/api/tms.ts:778-795`

### W11-A2-007 — updateNews — PROVEN

```text
BREAK: intranet.ts `updateNews` → `updateNews_REMOVED`
→ FAIL suite (import / source Match `export function updateNews`)
RESTORE → PASS
```

Guard: `frontend/src/api/intranet.ts:48-53`

### W11-A2-008 — updateMemberRole — PROVEN

```text
BREAK: projects.ts `updateMemberRole` → `updateMemberRole_REMOVED`
→ FAIL suite (import / source Match)
RESTORE → PASS
```

Guard: `frontend/src/api/projects.ts:139-147`

### W11-A2-010 — buildMessageFromWsEvent null-guards — PROVEN

```text
BREAK: messenger.ts — убрать object/id/sender null returns
→ FAIL buildMessageFromWsEvent rejects null… (null/{} / empty sender)
RESTORE → PASS
```

Guard: `frontend/src/api/messenger.ts:99-110`

### W10-A2-018 — parsePlateValue zod — PROVEN

```text
BREAK: issue-description.tsx — `return raw as Value` (no safeParse)
→ FAIL parsePlateValue accepts… (null/"x" → undefined)
RESTORE → PASS
```

Guard: `frontend/src/modules/tracker/issue-description.tsx:34-37`

### W9-A5-007 — user-picker enabled — PROVEN

```text
BREAK: user-picker.tsx `enabled: Boolean(projectId)` → `enabled: true`
→ FAIL user-picker disables members query…
RESTORE → PASS
```

Guard: `frontend/src/components/tracker/user-picker.tsx:54`

### W10-A2-017 — Partial filters — PROVEN

```text
BREAK: case-filters.tsx → `Record<string, unknown>`
→ FAIL case-filters updateFilter uses Partial…
RESTORE → PASS
```

Guard: `frontend/src/modules/tms/components/case-filters.tsx:24`

### W10-A2-020 — formatApiError — PROVEN

```text
BREAK: environment-comparison.tsx → `(compError as { message?: string }).message`
→ FAIL environment-comparison uses formatApiError…
RESTORE → PASS
```

Guard: `frontend/src/modules/tms/components/environment-comparison.tsx:317`

## Rework notes

- Добавлен `[MUTATION CRITERIA]` (10 Class source IDs).
- Vitest усилен source-guards + 3 новых `it` (picker / filters / formatApiError); было 6 → **9**.
- Prior W1-fe1 закрывал только tsc/version; W18-C5 FAIL по version закрыт fixture + Class-wide prove.

## Mock-only?

Нет — runtime guards бьют prod functions; type guard — real `tsc -p tsconfig.task052.json`; остальные — `readFileSync` production sources.
