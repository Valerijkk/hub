# MUTATION W6 slot 3 — TASK-128 (qa-auto1) · 2026-07-30

> Checker: **qa-auto1** (!= fe1 fixer) · Env: local Vitest · Evidence: `_mut-batch-w6-task128.json` + `_reprove_w6_task128.log`  
> Criteria: `.planning/backlog/TASK-128-2026-07-29-a3-10-5-fe-tms-onboarding-auth.md` `[MUTATION CRITERIA]`  
> Runner: `_run_mutations_w6_task128.py` (+ `_reprove_w6_task128.py` для `-t` regex escape)

## Verdict

**5/5 PROVEN** · статус TASK-128 → **Закрыта**

| ID | Verdict | RED test (`it`) | Mutation (BREAK) | Layer |
|----|---------|-----------------|------------------|-------|
| A3W10-5-001 | **PROVEN** | `viewer: hides Edit / link / unlink` | `cases_.$caseKey.lazy.tsx` — `canMutate = true` | frontend |
| A3W10-5-002 | **PROVEN** | `viewer: hides Новый прогон`; `viewer (canMutate=false): hides Start / Add cases`; `viewer: hides Complete / Reopen` | `runs.lazy.tsx` — `canMutate = true`; `run-header.tsx` — `{true ? (` (ignore prop) | frontend |
| A3W10-5-003 | **PROVEN** | `unauthenticated → redirect /login` | `onboarding.tsx` — удалить `beforeLoad` | frontend |
| A3W10-5-004 | **PROVEN** | `authenticated + completed → redirect /projects` (+ incomplete → `/onboarding`) | `login.tsx` — удалить `beforeLoad` | frontend |
| A3W10-5-005 | **PROVEN** | `production onboarding.lazy.tsx revokes blob URL in cleanup` | `onboarding.lazy.tsx` — убрать `URL.revokeObjectURL` из cleanup | frontend |

Production restored (`git checkout --` после каждого BREAK). Mock-gap: нет (RTL/route unit на prod-коде + source assert для revoke).

## Per-ID evidence

### A3W10-5-001 — case detail canMutate

```text
BREAK: cases_.$caseKey.lazy.tsx — const canMutate = true
→ FAIL viewer: hides Edit / link / unlink
RESTORE
→ PASS same test
```

### A3W10-5-002 — runs list + RunHeader

```text
BREAK list: runs.lazy.tsx — const canMutate = true
→ FAIL viewer: hides Новый прогон
RESTORE → PASS

BREAK header: run-header.tsx — {true ? ( instead of {canMutate ? (
→ FAIL viewer (canMutate=false): hides Start / Add cases
→ FAIL viewer: hides Complete / Reopen
RESTORE → PASS both
```

**Gap note (honest):** Vitest для detail бьёт в `RunHeader` с явным `canMutate={false}` prop — мутация только `runs_.$runId.tsx` (`canMutate = true`) **не** краснеет unit (GAP_NOTED). Gate в самом `RunHeader` доказан мутацией компонента (тот же UI-контракт, что страница прокидывает).

### A3W10-5-003 — onboarding beforeLoad

```text
BREAK: onboarding.tsx — remove beforeLoad
→ FAIL unauthenticated → redirect /login  (isRedirect=false)
RESTORE
→ PASS same test
```

### A3W10-5-004 — login beforeLoad

```text
BREAK: login.tsx — remove beforeLoad
→ FAIL authenticated + completed → redirect /projects
→ FAIL authenticated + incomplete → redirect /onboarding
RESTORE
→ PASS both
```

### A3W10-5-005 — avatar revokeObjectURL

```text
BREAK: onboarding.lazy.tsx — drop URL.revokeObjectURL(avatarPreview) in cleanup
→ FAIL production onboarding.lazy.tsx revokes blob URL in cleanup
RESTORE
→ PASS same test
```

Harness RTL (`revokes previous blob` / `revokes on unmount`) дублирует паттерн локально — prod-мутацию ловит **source assert**.

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs
python ../.planning/audits/total-2026-07/_run_mutations_w6_task128.py
python ../.planning/audits/total-2026-07/_reprove_w6_task128.py   # 002-header + 004 (vitest -t regex)

# regression after restore
npx vitest run \
  ../testframework/stages/unit/frontend-test/domains/tms/can-mutate-authz-ui.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/auth/onboarding-login-guards.test.ts \
  ../testframework/stages/unit/frontend-test/domains/auth/onboarding-avatar-revoke.test.tsx
→ Test Files 3 passed · Tests 20 passed (~3.6s)
```

## Assumptions / unresolved

- Vitest `-t` — regex: незаэкранированный `+` / `()` даёт false GREEN (0 matched) — в reprove фильтры экранированы.
- Коммит не создавался (инструкция слота).
- Criteria authored by fe1; mutation prove by qa-auto1 (W6 slot 3).

## Handoff

TASK-128 **Закрыта** · evidence этот файл · README синхронизирован · next: fe-ver / code-reviewer по желанию волны (mutation gate закрыт).
