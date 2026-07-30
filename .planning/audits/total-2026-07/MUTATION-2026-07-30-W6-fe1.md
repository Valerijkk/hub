# MUTATION W6 slot 5 — TASK-174 (fe1) · 2026-07-30

> Checker: **fe1** · Env: local Vitest · Criteria: `.planning/backlog/TASK-174-2026-07-29-a3-9-3-components.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**3/3 PROVEN** (001–003) · **004 = N/A** (GAP: только `rg`, нет Vitest source-guard — не fake-PROVEN)  
TASK-174 → **Закрыта** (004 mutation explicit N/A; code-fix 004 остаётся FIXED по grep)

| ID | Verdict | RED test (`it`) | Mutation (BREAK) | Layer |
|----|---------|-----------------|------------------|-------|
| A3W9-3-001 | **PROVEN** | `collapse button has aria-expanded and aria-controls pointing at panel id` | drop `aria-expanded` + `aria-controls` on ChannelSection `<button>` | frontend |
| A3W9-3-002 | **PROVEN** | `falls back to neutral label instead of UUID`; `mutation: returning ID:uuid would fail neutral-label assert` | `return \`ID: ${messageId}\`` вместо `"Закреплённое сообщение"` | frontend |
| A3W9-3-003 | **PROVEN** | `does not render preview card for javascript: URL`; `does not render preview card for data: URL` | drop `if (!isSafeUrl(url)) return null` in `LinkPreview` | frontend |
| A3W9-3-004 | **N/A** | — (нет Vitest) | criteria: `rg "rgba\\(52,211,153|rgba\\(13,14,18"` → empty; `tsc` не ловит | frontend |

Production restored (`git checkout --` на 3 файла). Mock-gap: нет (RTL/unit на prod-коде).

## Per-ID evidence

### A3W9-3-001 — collapse aria-*

```text
BASELINE GREEN → PASS collapse button has aria-expanded…
BREAK: channel-list.tsx — remove aria-expanded={!collapsed} + aria-controls={panelId}
→ FAIL collapse button has aria-expanded… (Received: null for aria-expanded)
RESTORE: git checkout -- channel-list.tsx
→ PASS same it
```

### A3W9-3-002 — pin preview ≠ UUID

```text
BREAK: pinned-messages.tsx resolvePinPreview — return `ID: ${messageId}`
→ FAIL falls back to neutral label instead of UUID
→ FAIL mutation: returning ID:uuid would fail neutral-label assert
RESTORE: git checkout -- pinned-messages.tsx
→ PASS both (3/3 file)
```

### A3W9-3-003 — LinkPreview isSafeUrl gate

```text
BREAK: link-preview.tsx — remove if (!isSafeUrl(url)) return null
→ FAIL does not render preview card for javascript: URL (skeleton, not empty)
→ FAIL does not render preview card for data: URL
(note: sanitizeUrl на href всё ещё #; early-null gate — то, что ловит empty-DOM assert)
RESTORE: git checkout -- link-preview.tsx
→ PASS javascript:/data: empty + https OK
```

### A3W9-3-004 — token rgba literals — **N/A**

```text
GAP per [MUTATION CRITERIA]: нет Vitest source-guard.
Manual only: rg "rgba\(52,211,153|rgba\(13,14,18" на
  channel-list / chat-view / message-composer / message-item → EMPTY (exit 1).
npx tsc --noEmit не ловит литералы.
Verdict: N/A — не ставить PROVEN без слоя теста. Не блокер code-FIXED (grep clean).
```

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs
# baseline + post-restore:
npx vitest run \
  ../testframework/stages/unit/frontend-test/domains/messenger/channel-list-collapse-a11y.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/messenger/pinned-messages-preview.test.ts \
  ../testframework/stages/unit/frontend-test/domains/messenger/link-preview-xss-whitelist.test.tsx
→ Test Files 3 passed · Tests 9 passed (~3.8s) after full restore

rg "rgba\(52,211,153|rgba\(13,14,18" frontend/src/components/messenger/{channel-list,chat-view,message-composer,message-item}.tsx
→ empty
```

## Assumptions / unresolved

- 004 mutation остаётся GAP до появления source-guard `it` (не делали в этом слоте — scope = prove only).
- `message-item` extract `isSafeUrl` не ломали отдельно: criteria допускает **и/или**; LinkPreview early-gate уже дал RED на обязательных `it`.

## Handoff

Evidence: этот файл · TASK-174 **Закрыта** (004 mut N/A) · fe-ver → code-reviewer (003 whitelist path).
