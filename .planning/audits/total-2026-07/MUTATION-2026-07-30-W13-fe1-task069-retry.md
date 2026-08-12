# MUTATION W13 slot 2 refill retry — TASK-069 (fe1) · 2026-07-30

> Checker: **fe1** · Env: local Vitest · Criteria: `.planning/backlog/archive/legacy/TASK-069-fe-p3-a11y-names.md` `[MUTATION CRITERIA]`  
> Prior: [MUTATION-2026-07-30-W13-fe1-task069.md](./MUTATION-2026-07-30-W13-fe1-task069.md) — W9-A1-016 **NOT_PROVEN**  
> Коммит: **не создавался**.

## Verdict

**6/6 PROVEN** · Class gaps **GAP_NOTED** · TASK-069 → **Закрыта**

| ID | Prior | Retry | RED |
|----|-------|-------|-----|
| W9-A1-016 | NOT_PROVEN | **PROVEN** | `toHaveAttribute("aria-label", …)` after drop aria-label (title kept) |
| W9-A1-017 | PROVEN | **PROVEN** (hold) | `CallReactionBar emoji buttons expose aria-label` |
| W9-A3-024 | PROVEN | **PROVEN** (hold) | `GifPicker search input has accessible name when open` |
| W9-A3-025 | PROVEN | **PROVEN** (hold) | `StickerPicker pack tabs expose tablist and aria-label` |
| W9-A3-026 | PROVEN | **PROVEN** (hold) | `CreateChannelDialogContent name field is labelled` |
| W9-A4-011 | PROVEN | **PROVEN** (hold) | `Checklist delete control has aria-label` |
| W9-A3-027…030,038,039 · W9-A4-009,010 · W10-A2-022,023 · W10-A4-011 | GAP_NOTED | **GAP_NOTED** | нет dedicated `it` в TASK-069 |

Production restored (`git checkout --`). Leftover MUTATED: **нет**.

## Gap closed

Prior assert used only `getByRole("button", { name: "На весь экран" })`. On icon-only buttons Testing Library takes accessible name from `title`, so kill «убрать aria-label, оставить title» оставался GREEN.

Strengthened `a11y-names-p3.test.tsx` CallPanel `it`:

```ts
const fullscreen = screen.getByRole("button", { name: "На весь экран" });
expect(fullscreen).toHaveAttribute("aria-label", "На весь экран");
```

Теперь требуется именно атрибут `aria-label`, а не name через `title`.

## Per-ID evidence

### W9-A1-016 — CallPanel fullscreen — **PROVEN**

```text
BREAK (kill recipe): CallPanel.tsx — убрать aria-label, оставить title
→ FAIL CallPanel fullscreen control exposes aria-label
     expect(element).toHaveAttribute("aria-label", "На весь экран")
     Received: null
RESTORE → PASS 6/6
```

### W9-A1-017 — CallReactionBar — **PROVEN** (hold)

```text
BREAK: CallReactions.tsx — убрать aria-label={label} (оставить title)
→ FAIL CallReactionBar emoji buttons expose aria-label
     Unable to find … name "Like"
RESTORE → PASS
```

### W9-A3-024 — GifPicker search — **PROVEN** (hold)

```text
BREAK: gif-picker.tsx — убрать aria-label="Поиск GIF"
→ FAIL GifPicker search input has accessible name when open
     Unable to find textbox name "Поиск GIF"
RESTORE → PASS
```

### W9-A3-025 — StickerPicker tabs — **PROVEN** (hold)

```text
BREAK: sticker-picker.tsx — убрать role="tablist"/role="tab"/aria-label на tabs + sticker
→ FAIL StickerPicker pack tabs expose tablist and aria-label
     Unable to find role "tablist"
RESTORE → PASS
```

### W9-A3-026 — CreateChannelDialog Label — **PROVEN** (hold)

```text
BREAK: create-channel-dialog.tsx — убрать <Label htmlFor="channel-name">
→ FAIL CreateChannelDialogContent name field is labelled
     Unable to find a label with the text of: Название канала
RESTORE → PASS
```

### W9-A4-011 — Checklist delete — **PROVEN** (hold)

```text
BREAK: checklist.tsx — убрать aria-label="Удалить пункт" (оставить title="Удалить")
→ FAIL Checklist delete control has aria-label
     Unable to find button name "Удалить пункт"
RESTORE → PASS
```

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs

npx vitest run ../testframework/stages/unit/frontend-test/domains/platform/a11y-names-p3.test.tsx
```

Baseline / final GREEN: Test Files 1 passed · Tests **6** passed · ~2.1–2.4s

## Gaps (criteria)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| CallPanel fullscreen aria-label | ✅ `toHaveAttribute` | PROVEN (retry) |
| CallReactionBar emoji aria-label | ✅ RTL | PROVEN (hold) |
| GifPicker search aria-label | ✅ RTL | PROVEN (hold) |
| StickerPicker tablist/aria-label | ✅ RTL | PROVEN (hold) |
| CreateChannelDialog Label | ✅ RTL | PROVEN (hold) |
| Checklist delete aria-label | ✅ RTL | PROVEN (hold) |
| W9-A3-027…030,038,039 · W9-A4-009,010 · W10-A2-022,023 · W10-A4-011 | ❌ GAP | Class source IDs без dedicated `it` |
| `npx tsc --noEmit` | N/A | не primary |

**Confidence:** HIGH — 6/6 runnable PROVEN после retry; class residuals GAP_NOTED.

## Handoff

- Evidence: этот файл · TASK-069 **Закрыта** · README sync  
- Vitest: `domains/platform/a11y-names-p3.test.tsx` (6 its; CallPanel assert усилен)  
- Коммит не создавался  
