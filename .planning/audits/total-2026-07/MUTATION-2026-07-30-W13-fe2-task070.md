# MUTATION W13 slot 5 refill — TASK-070 (fe2) · 2026-07-30

> Checker: **fe2** · Env: local Vitest · Criteria: `.planning/backlog/TASK-070-fe-p3-a11y-structure.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота). TASK-069 файлы не трогались.

## Verdict

**6/7 PROVEN · 1 NOT_PROVEN · 2 GAP_NOTED · tsc N/A** · TASK-070 → **Открыта** (W9-A5-016 NOT_PROVEN)

| ID | Verdict | RED |
|----|---------|-----|
| W9-A1-018 | **PROVEN** | `TrashManager tabs use tablist/tab/aria-selected` |
| W9-A4-013 | **PROVEN** | `RouteErrorBoundary fallback has role=alert` |
| W9-A5-013 | **PROVEN** | `CardTitle renders as heading` |
| W9-A5-016 | **NOT_PROVEN** | (kill recipe: убрать `aria-hidden`) — тест остался GREEN |
| W9-A5-012 | **PROVEN** | `Avatar announces online status when data-online` |
| W9-A5-014 | **PROVEN** | `Badge renders as span without focus ring classes` |
| W9-A5-019 | **PROVEN** | `FormField required sets aria-required on control` |
| W9-A5-015 | **GAP_NOTED** | нет dedicated `it` |
| W9-A5-017 | **GAP_NOTED** | нет dedicated `it` |
| `npx tsc --noEmit` | **N/A** | не primary (не ловит roles/aria-*) |

Production restored (`git checkout` UI files; `git diff` empty). Leftover MUTATED: **нет**. Mock-gap: нет (RTL на prod компонентах).

## Per-ID evidence

### W9-A1-018 — TrashManager tablist/tab/aria-selected

```text
BREAK: trash-manager.tsx — убрать role="tablist"/role="tab"/aria-selected (plain <button>)
→ FAIL TrashManager tabs use tablist/tab/aria-selected
     TestingLibraryElementError: Unable to find role="tablist"
RESTORE → PASS
```

Guard: `frontend/src/components/admin/trash-manager.tsx:80-86`  
Vitest: `domains/platform/a11y-structure-p3.test.tsx` · describe `P3-FE-02 · a11y structure residual (TASK-070)`

### W9-A4-013 — RouteErrorBoundary role=alert

```text
BREAK: route-error-boundary.tsx:61 — убрать role="alert"
→ FAIL RouteErrorBoundary fallback has role=alert
     Unable to find role="alert"
RESTORE → PASS
```

Guard: `frontend/src/components/route-error-boundary.tsx:61`

### W9-A5-013 — CardTitle as heading

```text
BREAK: card.tsx:36 — <h3> → <div>
→ FAIL CardTitle renders as heading
     Unable to find role="heading" name "Заголовок карточки"
RESTORE → PASS
```

Guard: `frontend/src/components/ui/card.tsx:36`

### W9-A5-016 — EmptyState aria-hidden — NOT_PROVEN

```text
BREAK: empty-state.tsx:24-25 — убрать aria-hidden с wrap и Icon
→ GREEN EmptyState decorative icon is aria-hidden (ожидался FAIL)
Причина: lucide-react SVG по умолчанию рендерит aria-hidden="true";
  тест делает querySelector("[aria-hidden='true']") по всему container —
  находит SVG, не проверяет guard на wrap.
RESTORE → PASS (baseline)
```

Guard: `frontend/src/components/ui/empty-state.tsx:24-25`  
Слой assert слабый (any aria-hidden in subtree) — kill recipe criteria не даёт RED.

### W9-A5-012 — Avatar online sr-only label

```text
BREAK: avatar.tsx:37 — не рендерить AvatarOnlineLabel при data-online
→ FAIL Avatar announces online status when data-online
     Unable to find text "В сети"
RESTORE → PASS
```

Guard: `frontend/src/components/ui/avatar.tsx:37`

### W9-A5-014 — Badge span / no focus:ring

```text
BREAK: badge.tsx:33 — <span> → <button type="button"> + class focus:ring
→ FAIL Badge renders as span without focus ring classes
     expected 'button' to be 'span'
RESTORE → PASS
```

Guard: `frontend/src/components/ui/badge.tsx:33`

### W9-A5-019 — FormField aria-required

```text
BREAK: form-field.tsx:31/:39 — не прокидывать aria-required при required
→ FAIL FormField required sets aria-required on control
     expected aria-required="true", received null
RESTORE → PASS
```

Guard: `frontend/src/components/ui/form-field.tsx:31` / `:39`

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs

# RED (после BREAK) / GREEN (после RESTORE)
npx vitest run ../testframework/stages/unit/frontend-test/domains/platform/a11y-structure-p3.test.tsx
# per-ID: -t "TrashManager tabs" | -t "RouteErrorBoundary" | -t "CardTitle"
#         | -t "EmptyState" | -t "Avatar announces" | -t "Badge renders" | -t "FormField required"
```

Baseline / final GREEN: Test Files 1 passed · Tests **7 passed** · ~1.7s

## Gaps (criteria)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| trash tabs / route alert / card / avatar / badge / form-field | ✅ RTL | PROVEN |
| EmptyState aria-hidden wrap | ❌ слабый assert | NOT_PROVEN — lucide default aria-hidden |
| W9-A5-015, W9-A5-017 | ❌ GAP | Class source IDs без dedicated `it` |
| `npx tsc --noEmit` | N/A | не primary |

**Confidence:** PARTIAL — 6/7 runnable PROVEN; W9-A5-016 NOT_PROVEN; 015/017 GAP_NOTED.

## Handoff

- Evidence: этот файл · TASK-070 **Открыта** (нужен усиленный assert EmptyState wrap) · README sync  
- Fix path: в `a11y-structure-p3.test.tsx` проверять `aria-hidden` на icon wrap (например первый child / конкретный selector), не any `[aria-hidden]` в container  
- Коммит: не создавался
