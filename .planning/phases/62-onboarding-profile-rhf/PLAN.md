# Phase 62 — Onboarding Profile RHF+Zod (UX-P1-07)

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-25

## Goal

Close **UX-P1-07 / FB-12** remainder: profile step (step 1) with RHF + Zod validation, wired language select, E2E gate.

## Deliverables

### Frontend
- `onboarding-profile-schema.ts` + Vitest
- `onboarding.lazy.tsx` — `useForm` + `zodResolver`, disabled Next until valid, inline errors
- i18n validation keys (en/ru)
- `e2e/onboarding-profile.spec.ts` + CI e2e-full

## Verification

```bash
cd frontend && npx vitest run src/routes/onboarding-profile-schema.test.ts
cd frontend && npx tsc --noEmit
```

## Success criteria

- [x] Profile step uses RHF+Zod (matches issue-form pattern)
- [x] Next disabled until name ≥ 2 chars
- [x] Language select persisted to completeOnboarding
- [x] Vitest + E2E spec added
