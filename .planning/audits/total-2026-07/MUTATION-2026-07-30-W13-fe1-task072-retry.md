# MUTATION W13 slot 2 refill retry — TASK-072 (fe1) · 2026-07-30

> Checker: **fe1** · Env: local Vitest · Criteria: `.planning/backlog/archive/legacy/TASK-072-fe-p3-timers-cleanup.md` `[MUTATION CRITERIA]`  
> Prior: [MUTATION-2026-07-30-W13-fe1-task072.md](./MUTATION-2026-07-30-W13-fe1-task072.md) — W9-A2-017 / W10-A3-018 **NOT_PROVEN**  
> Коммит: **не создавался**.

## Verdict

**4/4 PROVEN** · TASK-072 → **Закрыта**

| ID | Prior | Retry | RED |
|----|-------|-------|-----|
| W9-A1-014 | PROVEN | **PROVEN** (hold) | `CallRoom InviteBlock clears copy timer on unmount` |
| W9-A2-017 | NOT_PROVEN | **PROVEN** | source-guard `unmountCopyTimerCleanup` + runtime `clearTimeout` spy on unmount |
| W10-A3-018 | NOT_PROVEN | **PROVEN** | source-guard `unmountSavedTimerCleanup` + runtime `clearTimeout` spy on unmount |
| W11-A3-018 | PROVEN | **PROVEN** (hold) | `tick with corrupt startedAt stops timer instead of NaN elapsed` |

Production restored (`git checkout --`). Leftover MUTATED: **нет**.

## Gap closed

Prior source-guards matched any `clearTimeout(...Ref.current)` — including `onCopy` / `handleSave` — so unmount-only kill stayed green.

Strengthened `timers-cleanup-p3.test.tsx` (renamed from `.test.ts`):

1. **useEffect-scoped source-guard** — regex requires `clearTimeout` inside `useEffect(() => () => { … }, [])`, not merely anywhere in the file.
2. **Runtime unmount spy** — after copy/save schedules a timer, `clearSpy.mockClear()` then `unmount()` → `clearTimeout` must run (catches missing cleanup even if source text is obfuscated).
3. **Bare-timeout ban** — `not.toMatch` bare `setTimeout(() => setCopied/setSaved(false))` without ref (same pattern as CallRoom).

## Per-ID evidence

### W9-A2-017 — bot-markdown CodeBlock — **PROVEN**

```text
BREAK (kill recipe): bot-markdown.tsx — убрать только unmount useEffect clearTimeout
  (оставить copyTimerRef + clearTimeout в onCopy)
→ FAIL bot-markdown CodeBlock clears copy timer on unmount
     expected … to match /useEffect(() => () => { if (copyTimerRef.current) clearTimeout…/
→ FAIL bot-markdown CodeBlock clearTimeout runs on unmount after copy
     expected "clearTimeout" to be called at least once
RESTORE → PASS 7/7
```

### W10-A3-018 — sprint-retro — **PROVEN**

```text
BREAK (kill recipe): sprint-retro.tsx — убрать только unmount useEffect clearTimeout
  (оставить savedTimerRef + clearTimeout в handleSave)
→ FAIL sprint-retro clears saved timer on unmount
→ FAIL sprint-retro clearTimeout runs on unmount after save
RESTORE → PASS 7/7
```

### W9-A1-014 — CallRoom (hold) — **PROVEN**

```text
BREAK: InviteBlock — bare setTimeout, без copyTimerRef / cleanup
→ FAIL CallRoom InviteBlock clears copy timer on unmount
RESTORE → PASS
```

### W11-A3-018 — timer-store (hold) — **PROVEN**

```text
BREAK: tick — убрать Number.isFinite guards
→ FAIL tick with corrupt startedAt … expected true to be false (NaN elapsed)
RESTORE → PASS
```

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs

npx vitest run ../testframework/stages/unit/frontend-test/domains/platform/timers-cleanup-p3.test.tsx
```

Baseline / final GREEN: Test Files 1 passed · Tests **7** passed · ~1.4–1.6s

## Gaps (criteria)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| CallRoom cleanup + bare-timeout ban | ✅ | PROVEN (hold) |
| bot-markdown unmount clearTimeout | ✅ useEffect guard + runtime spy | PROVEN |
| sprint-retro unmount clearTimeout | ✅ useEffect guard + runtime spy | PROVEN |
| timer-store NaN | ✅ unit store | PROVEN (hold) |
| `npx tsc --noEmit` | N/A | не primary |

**Confidence:** HIGH — 4/4 PROVEN после retry.

## Handoff

- Evidence: этот файл · TASK-072 **Закрыта** · README sync  
- Vitest: `domains/platform/timers-cleanup-p3.test.tsx` (7 its)  
- Коммит не создавался  
