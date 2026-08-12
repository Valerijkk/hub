# MUTATION W13 slot 2 refill — TASK-078 (fe1) · 2026-07-30

> Checker: **fe1** · Env: local Vitest · Criteria: `.planning/backlog/archive/legacy/TASK-078-fe-p3-docs-spa-ux.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался**.

## Verdict

**4/4 FE PROVEN · 3/3 docs N/A** · TASK-078 → **Закрыта**

| ID | Verdict | RED |
|----|---------|-----|
| W11-A4-017 | **N/A** | docs-only (architecture.md routes pin) |
| W11-A5-020 | **N/A** | docs-only (stack_and_tools stores/motion/pwa) |
| W11-A5-021 | **N/A** | docs-only (API.md Caddy :8080) |
| W10-A2-021 | **PROVEN** | `security settings uses TanStack Link, not <a href>` |
| W10-A5-015 | **PROVEN** | `trash column is Удалено (date), not Удалил` |
| W10-A5-018 | **PROVEN** | `run-results-table has correct UTF-8 Исполнитель` |
| W10-A1-022 | **PROVEN** | `message-bubble extracts URLs from displayText and contentText` |

Production restored (`git checkout --` 4 FE files). Leftover MUTATED: **нет**.

## Per-ID evidence

### W11-A4-017 / W11-A5-020 / W11-A5-021 — docs-only — **N/A**

```text
Mutation surface = docs/*.md; criteria flag docs-only.
Vitest source-guards exist and baseline PASS; FE agent does not break docs.
Verdict: N/A (per slot refill rule).
```

### W10-A2-021 — security TanStack Link — **PROVEN**

```text
BREAK: security.tsx — <Link to="/settings/profile"> → <a href="/settings/profile">
→ FAIL security settings uses TanStack Link, not <a href>
     expected toMatch /<Link\s*\n?\s*to="\/settings\/profile"/
RESTORE → PASS
```

### W10-A5-015 — trash «Удалено» — **PROVEN**

```text
BREAK: trash.tsx — column «Удалено» → «Удалил»
→ FAIL trash column is Удалено (date), not Удалил
     expected toMatch /Удалено/
RESTORE → PASS
```

### W10-A5-018 — UTF-8 «Исполнитель» — **PROVEN**

```text
BREAK: run-results-table.tsx — «Исполнитель» → «Испол\uFFFD\uFFFDитель»
→ FAIL run-results-table has correct UTF-8 Исполнитель
     expected toMatch /Исполнитель/
RESTORE → PASS
```

### W10-A1-022 — extractUrls displayText ∪ contentText — **PROVEN**

```text
BREAK: message-bubble.tsx — drop extractUrls(displayText); keep only extractUrls(contentText)
  (comment must not contain literal extractUrls(displayText) — else source-guard false GREEN)
→ FAIL message-bubble extracts URLs from displayText and contentText
     expected toMatch /extractUrls\(displayText\)/
RESTORE → PASS
```

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs

npx vitest run ../testframework/stages/unit/frontend-test/domains/platform/docs-spa-ux-p3.test.ts
```

Baseline / final: **7/7 PASS** (~1.2–1.4s).

## Gaps (criteria)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| W11-A4-017..W11-A5-021 | docs Vitest readFile | **N/A** docs-only |
| W10-A2-021 Link | ✅ source-guard | PROVEN |
| W10-A5-015 trash copy | ✅ source-guard | PROVEN |
| W10-A5-018 UTF-8 | ✅ source-guard | PROVEN |
| W10-A1-022 URL extract | ✅ source-guard | PROVEN; comment self-match caveat noted |

**Confidence:** CONFIRMED — all runnable FE guards PROVEN; docs rows N/A.

## Handoff

- Evidence: этот файл · TASK-078 **Закрыта** · README sync  
- Vitest: `domains/platform/docs-spa-ux-p3.test.ts`  
- Коммит не создавался  
