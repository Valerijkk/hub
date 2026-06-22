# Phase 18 — AI Privacy & v3.5 (M3-030+)

> **Branch:** `feat/v2.5-trust` · **Depends on:** Phase 17 AI Native

## Exit criteria

M3-030 privacy consent · user settings UI · enforcement on gateway + bot · tests + docs

---

## Wave 1 — M3-030 External AI consent

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W1-01 `allow_external_ai` on `ai_user_preference` + migration | be | arch-ver | ✓ |
| W1-02 Gateway enforcement (chat/stream/JQL/bot) | be | security-auditor | ✓ |
| W1-03 PATCH `/ai/preferences` + settings UI | fe | fe-ver | ✓ |
| W1-04 Unit tests + API.md | qa | qa-ver | ✓ |

## Wave 2 — E2E smoke + consent enforcement

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W2-01 Playwright `/settings/ai` consent toggle | qa | fe-ver | ✓ |
| W2-02 API smoke: preferences + consent 403 | qa | security-auditor | ✓ |
| W2-03 Playwright @hub-ai DM hook smoke | qa | qa-ver | ✓ |
| W2-04 E2E helpers `e2e/helpers/ai.ts` | qa | qa-ver | ✓ |

**Phase 18 exit criteria:** ✅ W1–W2 complete

---
