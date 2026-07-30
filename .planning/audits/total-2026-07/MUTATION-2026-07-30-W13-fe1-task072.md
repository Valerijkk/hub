# MUTATION W13 slot 2 refill — TASK-072 (fe1) · 2026-07-30

> Checker: **fe1** · Env: local Vitest · Criteria: `.planning/backlog/archive/legacy/TASK-072-fe-p3-timers-cleanup.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**2/4 PROVEN** · **2 NOT_PROVEN** · TASK-072 → **Открыта**

| ID | Verdict | RED |
|----|---------|-----|
| W9-A1-014 | **PROVEN** | `CallRoom InviteBlock clears copy timer on unmount` |
| W9-A2-017 | **NOT_PROVEN** | — (unmount-only cleanup; Vitest stays green) |
| W10-A3-018 | **NOT_PROVEN** | — (unmount-only cleanup; Vitest stays green) |
| W11-A3-018 | **PROVEN** | `tick with corrupt startedAt stops timer instead of NaN elapsed` |

Production restored (`git checkout --`). Leftover MUTATED: **нет**.

`npx tsc --noEmit` — N/A (не primary per criteria).

## Per-ID evidence

### W9-A1-014 — CallRoom InviteBlock copyTimerRef — **PROVEN**

```text
BREAK: CallRoom.tsx InviteBlock — убрать copyTimerRef + unmount clearTimeout;
  голый setCopied(true); setTimeout(() => setCopied(false), 3000)
→ FAIL CallRoom InviteBlock clears copy timer on unmount
     AssertionError: expected … to match /copyTimerRef/
RESTORE → PASS 5/5
```

Guard: `frontend/src/components/calls/CallRoom.tsx:196-201` (+ copy path)  
Vitest: `domains/platform/timers-cleanup-p3.test.ts` · `it("CallRoom InviteBlock clears copy timer on unmount")`  
Note: it additionally asserts `not.toMatch(/setCopied(true);\s*setTimeout(...)/)` — ловит bare timer.

### W9-A2-017 — bot-markdown CodeBlock clearTimeout cleanup — **NOT_PROVEN**

```text
BREAK (kill recipe): bot-markdown.tsx — убрать unmount clearTimeout cleanup
  (оставить copyTimerRef + clearTimeout в onCopy)
→ PASS (source-guard всё ещё видит clearTimeout(copyTimerRef.current) в onCopy)

BREAK (stronger, beyond recipe): убрать весь copyTimerRef / bare setTimeout
→ FAIL expected … to match /copyTimerRef/
  (доказывает только наличие символа ref, не unmount-cleanup)
RESTORE → PASS 5/5
```

Verdict: **NOT_PROVEN** — stated kill («убрать clearTimeout cleanup») не краснеет; `it` не изолирует unmount от onCopy.

Guard: `frontend/src/components/messenger/bot-markdown.tsx:88-92`  
Vitest: same file > `it("bot-markdown CodeBlock clears copy timer on unmount")`

### W10-A3-018 — sprint-retro savedTimerRef cleanup — **NOT_PROVEN**

```text
BREAK (kill recipe): sprint-retro.tsx — убрать unmount clearTimeout cleanup
  (оставить savedTimerRef + clearTimeout в handleSave)
→ PASS (source-guard матчит handleSave clearTimeout)

BREAK (stronger): убрать весь savedTimerRef / bare setTimeout
→ FAIL expected … to match /savedTimerRef/
RESTORE → PASS 5/5
```

Verdict: **NOT_PROVEN** — unmount-only break не ловится; assert не требует cleanup в `useEffect`.

Guard: `frontend/src/modules/tracker/sprint-retro.tsx:52-56`  
Vitest: same file > `it("sprint-retro clears saved timer on unmount")`

### W11-A3-018 — timer-store Number.isFinite — **PROVEN**

```text
BREAK: timer-store.ts tick — убрать Number.isFinite(startedMs/elapsed) guards;
  считать elapsed as-is
→ FAIL tick with corrupt startedAt stops timer instead of NaN elapsed
     AssertionError: expected true to be false  (Number.isNaN(elapsedSeconds))
RESTORE → PASS 5/5
```

Guard: `frontend/src/stores/timer-store.ts:58-60` (+ elapsed guard)  
Vitest: same file > `it("tick with corrupt startedAt stops timer instead of NaN elapsed")`

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs

npx vitest run ../testframework/stages/unit/frontend-test/domains/platform/timers-cleanup-p3.test.ts
```

Baseline / final GREEN: Test Files 1 passed · Tests 5 passed · ~1.0–1.3s

## Gaps (criteria)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| CallRoom cleanup + bare-timeout ban | ✅ source-guard | PROVEN |
| bot-markdown unmount clearTimeout | ❌ | NOT_PROVEN — onCopy clearTimeout маскирует |
| sprint-retro unmount clearTimeout | ❌ | NOT_PROVEN — handleSave clearTimeout маскирует |
| timer-store NaN | ✅ unit store | PROVEN |
| `npx tsc --noEmit` | N/A | не primary |

**Confidence:** PARTIAL — 2/4 PROVEN; 2 source-guards не бьют unmount-слой.

## Remediation (для повторной мутации)

Ужесточить `it` для W9-A2-017 / W10-A3-018, например:

- assert `useEffect` cleanup блок с `clearTimeout(...Ref.current)`; или
- runtime unmount spy (`setState` after unmount / fake timers); или
- negative assert как у CallRoom (`not.toMatch` bare `setTimeout` без ref).

## Handoff

- Evidence: этот файл · TASK-072 **Открыта** · README sync  
- fe / qa-auto: починить source-guards → re-mutate W9-A2-017, W10-A3-018  
- Коммит не создавался  
