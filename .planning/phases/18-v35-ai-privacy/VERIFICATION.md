# Phase 18 — AI Privacy & v3.5 (M3-030) — VERIFICATION

**Date:** 2026-07-27  
**Method:** Schema/field + enforcement + UI + E2E helper/spec inspection  
**Sources:** `PLAN.md` (W1–W2 ✅), live AI privacy paths

## Goal (exit criteria)

`allow_external_ai` preference · gateway/bot enforcement · settings UI · tests + docs · E2E consent smoke

## Measured evidence

| Criterion | Evidence (2026-07-27) | Result |
|-----------|----------------------|--------|
| Preference field | Ent generated + service: `AllowExternalAI` / `allow_external_ai` on `AIUserPreference`; `preference_input.go`, `repository_ent.go` | PASS |
| Enforcement | `backend/internal/ai/privacy.go` defines `ErrExternalAIConsentRequired`; handler maps consent errors on chat/stream paths | PASS |
| Settings UI | `frontend/src/modules/settings/ai-privacy.tsx`; API types in `frontend/src/api/ai.ts` | PASS |
| Unit / Vitest | `testframework/…/ai/ai-privacy.test.tsx`; backend `testframework/ai/service_test.go` preference fixtures | PASS |
| E2E consent | `ai-hub.spec.ts`: settings/ai consent toggle, API `allow_external_ai`, consent 403 path; helpers in `e2e/shared/helpers/ai.ts` | PASS |

## Verdict: **PASS**
