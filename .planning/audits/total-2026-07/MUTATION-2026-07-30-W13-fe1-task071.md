# MUTATION W13 slot 2 refill — TASK-071 (fe1) · 2026-07-30

> Checker: **fe1** · Env: local Vitest · Criteria: `.planning/backlog/archive/legacy/TASK-071-fe-p3-tokens-i18n.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался**.

## Verdict

**9/9 PROVEN** · TASK-071 → **Закрыта**

| ID | Verdict | RED |
|----|---------|-----|
| W9-A5-018 | **PROVEN** | `SelectTrigger uses hub radius/border tokens (not shadcn rounded-md/border-input)` |
| W9-A3-031 | **PROVEN** | `messenger residual files have no raw Tailwind palette colors` (voice-recorder) |
| W9-A3-032 | **PROVEN** | same `it` (poll-message) |
| W9-A3-033 | **PROVEN** | same `it` (channel-list via `bg-green-500`) |
| W9-A3-034 | **PROVEN** | same `it` (thread-issue-links) |
| W9-A3-035 | **PROVEN** | same `it` (invite-link-dialog) |
| W9-A5-009 | **PROVEN** | `Dialog Close uses i18n btn.close (RU)` |
| W9-A5-010 | **PROVEN** | `CommandDialog title follows locale` |
| W9-A5-011 | **PROVEN** | `ErrorBoundary fallback uses i18n (EN locale)` |

Production restored (`git checkout --` + manual restore). Leftover MUTATED: **нет**.

## Per-ID evidence

### W9-A5-018 — SelectTrigger hub tokens — **PROVEN**

```text
BREAK: select.tsx — rounded-[var(--r2)] → rounded-md; border-b1 → border-input
→ FAIL SelectTrigger uses hub radius/border tokens …
     expected className to match /rounded-\[var\(--r2\)\]/
RESTORE → PASS
```

### W9-A3-031 — voice-recorder — **PROVEN**

```text
BREAK: voice-recorder.tsx — bg-[var(--red)] → bg-red-500
→ FAIL messenger residual files …
     voice-recorder.tsx matched palette regex
RESTORE → PASS
```

### W9-A3-032 — poll-message — **PROVEN**

```text
BREAK: poll-message.tsx — text-[var(--orange)] → text-orange-500 (line 175)
→ FAIL messenger residual files … (poll-message.tsx)
RESTORE → PASS
```

### W9-A3-033 — channel-list — **PROVEN**

```text
BREAK (kill recipe): channel-list.tsx — bg-[var(--gG)] → bg-green-500
→ FAIL messenger residual files … (channel-list.tsx)
RESTORE → PASS

Note: naive border-l-green-500 alone does NOT match
/\b(bg|text|hover:text|border)-(red|orange|yellow|green)-\d{2,3}\b/
(border-l- ≠ border-). Kill recipe bg-green-* / border-green-500 — PROVEN.
```

### W9-A3-034 — thread-issue-links — **PROVEN**

```text
BREAK: thread-issue-links.tsx — hover:text-[var(--red)] → hover:text-red-500
→ FAIL messenger residual files … (thread-issue-links.tsx)
RESTORE → PASS
```

### W9-A3-035 — invite-link-dialog — **PROVEN**

```text
BREAK: invite-link-dialog.tsx — text-[var(--orange)] → text-orange-500 (badge)
→ FAIL messenger residual files … (invite-link-dialog.tsx)
RESTORE → PASS
```

### W9-A5-009 — Dialog Close i18n — **PROVEN**

```text
BREAK: dialog.tsx — t("btn.close") → hardcode "Close"
→ FAIL Dialog Close uses i18n btn.close (RU)
     Unable to find button name "Закрыть" (EN Close still PASS)
RESTORE → PASS
```

### W9-A5-010 — CommandDialog title — **PROVEN**

```text
BREAK: command.tsx — t("search.globalSearch") → «Глобальный поиск»
→ FAIL CommandDialog title follows locale
     Unable to find text "Global search"
RESTORE → PASS
```

### W9-A5-011 — ErrorBoundary i18n — **PROVEN**

```text
BREAK: error-boundary.tsx — t("error.somethingWrong")/t("btn.refresh") → RU hardcode
→ FAIL ErrorBoundary fallback uses i18n (EN locale)
     Unable to find "Something went wrong"
RESTORE → PASS
```

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs

npx vitest run ../testframework/stages/unit/frontend-test/domains/platform/tokens-i18n-p3.test.tsx
```

Baseline / final GREEN: Test Files 1 passed · Tests **6** passed · ~1.6–1.8s

## Gaps (criteria)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| Select hub tokens | ✅ RTL className | PROVEN |
| messenger palette (5 files) | ✅ source-guard `readFileSync` | PROVEN; `border-l-*` вне regex — kill via `bg-green-*` / `border-green-*` |
| Dialog / Command / EB i18n | ✅ RTL + i18n.changeLanguage | PROVEN |
| `npx tsc --noEmit` | N/A | не primary |

**Confidence:** CONFIRMED — 9/9 runnable PROVEN.

## Handoff

- Evidence: этот файл · TASK-071 **Закрыта** · README sync  
- Vitest: `domains/platform/tokens-i18n-p3.test.tsx` (6 its → 9 Class IDs)  
- Коммит не создавался  
