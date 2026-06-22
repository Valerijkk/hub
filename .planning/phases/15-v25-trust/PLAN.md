# Phase 15 — Trust (v2.5)

> **Branch:** `feat/v2.5-trust` · **PO LOCK:** `DISCUSS-ROADMAP-LOCKED.md`  
> **Exit:** AUTH + org + NOTIF bind + E2E FB-07..09 + WCAG + run_failed + Biome + LiveKit CI + coverage ≥50%

## Verification loop (each task)

```
be/fe → unit tests → qa-gate slice → *-ver (or inline checklist) → fix → re-verify → commit
```

---

## Wave 1 — AUTH-01 Magic Link only

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W1-01 Wire `MagicLinkForm` → `login.tsx` | fe | fe-ver | ✓ |
| W1-02 Remove TG login routes + handlers register | be | be-ver | ✓ |
| W1-03 Remove `/login` from auth bot; optional disable poll | be | be-ver | ✓ |
| W1-04 Update auth tests + e2e login.spec | qa | qa-ver | ✓ |
| W1-05 Docs sync vision, TESTING, TECHNICAL | docs | sa-ver | ✓ |

## Wave 2 — Organization v2.5

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W2-01 Ent `organization` + user.org_id + founder role | be | arch-ver | ✓ |
| W2-02 Pending user status + admin queue API | be | be-ver | ✓ |
| W2-03 7d pending purge job | be | be-ver | ✓ |
| W2-04 Admin UI pending users | fe | fe-ver | ✓ |
| W2-05 Invite-only registration default | be | qa-ver | ✓ |

## Wave 3 — NOTIF TG bind

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W3-01 Bind API + Redis keys | be | be-ver | ✓ |
| W3-02 Bot `/start bind_`, `/link` (notification pkg) | be | be-ver | ✓ |
| W3-03 `TelegramLinkCard` UI | fe | fe-ver | ✓ |
| W3-04 NOTIF e2e mock | qa | qa-ver | ✓ |

## Wave 4 — run_failed + TMS alerts

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W4-01 Types + QA-only recipients | be | be-ver | ✓ |
| W4-02 Consumer logic | be | qa-ver | ✓ |

## Wave 5 — E2E FB-07..09

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W5-01 project-lifecycle.spec | qa | qa-ver | ✓ |
| W5-02 messenger.spec | qa | qa-ver | ✓ |
| W5-03 tms.spec | qa | qa-ver | ✓ |

## Wave 6 — WCAG + RHF

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W6-01 axe expansion | qa | fe-ver | ✓ |
| W6-02 issue-form Zod | fe | fe-ver | ✓ |

## Wave 7 — Biome + LiveKit CI

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W7-01 Biome wave fix | fe | qa-ver | ✓ |
| W7-02 LiveKit compose CI job | devops | be-ver | ✓ |

## Wave 8 — Quick wins + coverage gate

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W8-01 FB-01..06 | fe | qa-ver | ✓ |
| W8-02 Coverage ≥50% backend | qa | qa-ver | ✓ |

## Wave 9 — Phase 15.2 hub-mail (after 15.1)

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W9-01 `cmd/hub-mail` skeleton | be | arch-ver | ✓ |
| W9-02 Admin mail settings | fe | fe-ver | ✓ |

## Wave 11 — Swagger + security close-out

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W11-01 Swagger: pending/org/TG bind/mail admin | docs | sa-ver | ✓ |
| W11-02 hub-mail `/v1/send` internal token auth | be | security-auditor | ✓ |
| W11-03 `swag init` regenerate docs | be | qa-ver | ✓ |

## Wave 10 — Test framework (BACKLOG → Phase 999.1)

> **Deferred:** promote from backlog when E2E velocity or coverage maintenance hurts.  
> **Spec:** `.planning/phases/999-test-framework/SPEC.md` · **Backlog ID:** FB-TEST-01

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W10-01 `tests/` skeleton + README + CI targets | qa | arch-ver | |
| W10-02 Playwright → `tests/e2e/` + POM/Flow refactor | qa | fe-ver | |
| W10-03 Vitest setup + component test templates | fe | fe-ver | |
| W10-04 Go integration suite + builders | qa | be-ver | |
| W10-05 Shared fixtures BE↔E2E + test design catalog | qa | qa-ver | |

---

*Swagger wave: end of Phase 15 · Security: gsd-secure-phase after phase complete*

## Phase 15 exit checklist

| Criterion | Status |
|-----------|--------|
| AUTH Magic Link only | ✓ |
| Organization + pending queue | ✓ |
| NOTIF TG bind + run_failed | ✓ |
| E2E FB-07..09 | ✓ |
| WCAG + RHF issue-form | ✓ |
| Biome 0 errors + LiveKit CI | ✓ |
| Coverage ≥50% (55.7%) | ✓ |
| hub-mail skeleton + admin UI | ✓ |
| Docs sync (W1-05) | ✓ |
| Swagger annotations (new endpoints) | ✓ |
| gsd-secure-phase audit | ✓ (HUB_MAIL_INTERNAL_TOKEN on /v1/send) |
| Wave 10 test framework | deferred → Phase 999.1 |

**Phase 15: SHIPPED (code complete, uncommitted)**
