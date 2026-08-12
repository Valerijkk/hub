# MUTATION W13 slot 2 refill — TASK-069 (fe1) · 2026-07-30

> Checker: **fe1** · Env: local Vitest · Criteria: `.planning/backlog/archive/legacy/TASK-069-fe-p3-a11y-names.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался**.

## Verdict

**5/6 PROVEN** · **1 NOT_PROVEN** (W9-A1-016) · Class gaps **GAP_NOTED**  
TASK-069 → **Открыта** (kill recipe CallPanel: `title` alone still names the control).

| ID | Verdict | RED |
|----|---------|-----|
| W9-A1-016 | **NOT_PROVEN** | — (убрать только `aria-label`, оставить `title` → `getByRole(…, { name: "На весь экран" })` остаётся PASS) |
| W9-A1-017 | **PROVEN** | `CallReactionBar emoji buttons expose aria-label` |
| W9-A3-024 | **PROVEN** | `GifPicker search input has accessible name when open` |
| W9-A3-025 | **PROVEN** | `StickerPicker pack tabs expose tablist and aria-label` |
| W9-A3-026 | **PROVEN** | `CreateChannelDialogContent name field is labelled` |
| W9-A4-011 | **PROVEN** | `Checklist delete control has aria-label` |
| W9-A3-027…030,038,039 · W9-A4-009,010 · W10-A2-022,023 · W10-A4-011 | **GAP_NOTED** | нет dedicated `it` в TASK-069 |

Production restored (`git checkout --`). Leftover MUTATED: **нет**.

## Per-ID evidence

### W9-A1-016 — CallPanel fullscreen — **NOT_PROVEN**

```text
BREAK (kill recipe): CallPanel.tsx — убрать aria-label, оставить title
→ PASS CallPanel fullscreen control exposes aria-label
     (accessible name берётся из title на icon-only кнопке с SVG)
RESTORE → PASS 6/6
```

Причина: Testing Library / a11y tree использует `title` как имя, когда нет текста/`aria-label`. Kill recipe «оставить только title» не даёт RED.

### W9-A1-017 — CallReactionBar — **PROVEN**

```text
BREAK: CallReactions.tsx — убрать aria-label={label} (оставить title)
→ FAIL CallReactionBar emoji buttons expose aria-label
     Unable to find … name "Like" (имя = emoji "👍", title не перекрывает text content)
RESTORE → PASS
```

### W9-A3-024 — GifPicker search — **PROVEN**

```text
BREAK: gif-picker.tsx — убрать aria-label="Поиск GIF"
→ FAIL GifPicker search input has accessible name when open
     Unable to find textbox name "Поиск GIF"
RESTORE → PASS
```

### W9-A3-025 — StickerPicker tabs — **PROVEN**

```text
BREAK: sticker-picker.tsx — убрать role="tablist"/role="tab"/aria-label на tabs + sticker
→ FAIL StickerPicker pack tabs expose tablist and aria-label
     Unable to find role "tablist"
RESTORE → PASS
```

### W9-A3-026 — CreateChannelDialog Label — **PROVEN**

```text
BREAK: create-channel-dialog.tsx — убрать <Label htmlFor="channel-name">
→ FAIL CreateChannelDialogContent name field is labelled
     Unable to find a label with the text of: Название канала
RESTORE → PASS
```

### W9-A4-011 — Checklist delete — **PROVEN**

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

Baseline / final GREEN: Test Files 1 passed · Tests **6** passed · ~1.8s

## Gaps (criteria)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| CallPanel fullscreen aria-label | ❌ title fallback | NOT_PROVEN — усилить kill/assert (убрать title тоже / запретить title-only name) |
| CallReactionBar emoji aria-label | ✅ RTL | PROVEN |
| GifPicker search aria-label | ✅ RTL | PROVEN |
| StickerPicker tablist/aria-label | ✅ RTL | PROVEN |
| CreateChannelDialog Label | ✅ RTL | PROVEN |
| Checklist delete aria-label | ✅ RTL | PROVEN |
| W9-A3-027…030,038,039 · W9-A4-009,010 · W10-A2-022,023 · W10-A4-011 | ❌ GAP | Class source IDs без dedicated `it` |
| `npx tsc --noEmit` | N/A | не primary |

**Confidence:** PARTIAL — 5/6 runnable PROVEN; 1 NOT_PROVEN; class residuals GAP_NOTED.

## Handoff

- Evidence: этот файл · TASK-069 **Открыта** · README sync  
- Vitest: `domains/platform/a11y-names-p3.test.tsx` (6 its)  
- Follow-up: ужесточить тест/kill для CallPanel (не полагаться на `title` как name) или расширить assert  
- Коммит не создавался  
