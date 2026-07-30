# MUTATION W13 slot 2 refill — TASK-124 (fe1) · 2026-07-30

> Checker: **fe1** · Env: local Vitest · Criteria: `.planning/backlog/archive/2026-07-29/TASK-124-2026-07-29-a3-10-1-fe-routes-modules.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**5/5 PROVEN** · TASK-124 → **Закрыта**

| ID | Verdict | RED |
|----|---------|-----|
| A3W10-1-001 | **PROVEN** | `shows P&L alert…` / `shows invoices alert…` / `shows trial balance alert…` |
| A3W10-1-002 | **PROVEN** | `shows QueryErrorAlert when /project-templates fails` |
| A3W10-1-003 | **PROVEN** | `passes noopener,noreferrer as third arg to window.open` |
| A3W10-1-004 | **PROVEN** | `shows moderation queue alert when listPendingListings fails` |
| A3W10-1-005 | **PROVEN** | `shows alert in main when getBase fails for selected base (A3W10-1-005)` |

Production restored (`git checkout --`). Leftover MUTATED: **нет**. Mock-gap: нет (RTL на prod UI).

`npx tsc --noEmit` — N/A (не primary per criteria; не ловит error≠empty / noopener).

## Per-ID evidence

### A3W10-1-001 — accounting secondary QueryErrorAlert

```text
BREAK: accounting-page.tsx — убрать ветки incomeStatementError / invoicesError / trialBalanceError
  (рендер empty table / скрытый P&L без alert)
→ FAIL shows P&L alert when getIncomeStatement fails
→ FAIL shows invoices alert when listInvoices fails
→ FAIL shows trial balance alert when getTrialBalance fails
     TestingLibraryElementError: Unable to find role="alert" (×3)
RESTORE → PASS 3/3 (accounting-secondary-error.test.tsx)
```

Guard: `frontend/src/modules/accounting/accounting-page.tsx:68` + `:208-212` / `:322-326` / `:419-423`  
Vitest: `domains/v4/accounting-secondary-error.test.tsx` · describe `AccountingPage secondary query error≠empty (A3W10-1-001)`

### A3W10-1-002 — TemplateSelector loadError

```text
BREAK: project-templates.tsx TemplateSelector — при loadError рендерить grid вместо QueryErrorAlert
→ FAIL shows QueryErrorAlert when /project-templates fails
     TestingLibraryElementError: Unable to find role="alert"
RESTORE → PASS 2/2 (template-selector-error.test.tsx)
```

Guard: `frontend/src/modules/settings/project-templates.tsx:27` + `:69-73`  
Vitest: `domains/v4/template-selector-error.test.tsx` · describe `TemplateSelector error≠empty (A3W10-1-002)`

### A3W10-1-003 — recordings window.open noopener

```text
BREAK: recordings-list.tsx — window.open(data.url, "_blank") без 3-го arg
→ FAIL passes noopener,noreferrer as third arg to window.open
     AssertionError: expected … to be called with … "noopener,noreferrer"
RESTORE → PASS 1/1 (recordings-download-noopener.test.tsx)
```

Guard: `frontend/src/modules/calls/recordings-list.tsx:65`  
Vitest: `domains/calls/recordings-download-noopener.test.tsx` · describe `RecordingsList download noopener (A3W10-1-003)`

### A3W10-1-004 — marketplace pending admin gate

```text
BREAK: marketplace-page.tsx — гейт только pendingItems.length > 0 (скрыть секцию при fail)
→ FAIL shows moderation queue alert when listPendingListings fails
     TestingLibraryElementError: Unable to find role="alert"
RESTORE → PASS 1/1 (marketplace-pending-error.test.tsx)
```

Guard: `frontend/src/modules/marketplace/marketplace-page.tsx:71` + `:245` / `:255-257`  
Vitest: `domains/v4/marketplace-pending-error.test.tsx` · describe `MarketplacePage pending error≠empty (A3W10-1-004)`

### A3W10-1-005 — bases detailError in main

```text
BREAK: bases-page.tsx — убрать detailError → QueryErrorAlert в <main>
→ FAIL shows alert in main when getBase fails for selected base (A3W10-1-005)
     TestingLibraryElementError: Unable to find role="alert"
  (list-sidebar it остался PASS — prior TASK-047)
RESTORE → PASS 2/2 (bases-page-error.test.tsx)
```

Guard: `frontend/src/modules/bases/bases-page.tsx:54` + `:288-292`  
Vitest: `domains/v4/bases-page-error.test.tsx` · describe `BasesPage error≠empty (W10-A1-006 / TASK-047)`

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs

# RED (после BREAK) / GREEN (после RESTORE) — per file:
npx vitest run ../testframework/stages/unit/frontend-test/domains/v4/accounting-secondary-error.test.tsx
npx vitest run ../testframework/stages/unit/frontend-test/domains/v4/template-selector-error.test.tsx
npx vitest run ../testframework/stages/unit/frontend-test/domains/calls/recordings-download-noopener.test.tsx
npx vitest run ../testframework/stages/unit/frontend-test/domains/v4/marketplace-pending-error.test.tsx
npx vitest run ../testframework/stages/unit/frontend-test/domains/v4/bases-page-error.test.tsx
```

GREEN (все 5 файлов после restore): Test Files 5 passed · Tests 9 passed · ~2.0s

## Gaps (criteria)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| secondary accounting alerts | ✅ RTL | PROVEN |
| TemplateSelector loadError | ✅ RTL | PROVEN |
| recordings noopener | ✅ RTL spy | PROVEN |
| marketplace pending admin | ✅ RTL | PROVEN |
| bases detail fail | ✅ RTL | PROVEN |
| `npx tsc --noEmit` | N/A | не primary |

**Confidence:** CONFIRMED — 5/5 runnable PROVEN.

## Handoff

- Evidence: этот файл · TASK-124 **Закрыта** · README sync  
- fe-ver → code-reviewer (fix code already FIXED; mutation PROVEN)  
- Коммит: не создавался
