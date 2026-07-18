# Hub — Technical Debt & Audit Findings
**Date:** 2026-06-23
**Status:** Pending Fixes

This document aggregates the findings from the Performance, Security, and Code Quality audits conducted by the AI agent team.

## 1. Performance Audit Findings
* **CRITICAL:** `backend/internal/audit/service.go:122` — N+1 query pattern in `GetResourceLimits`. 3 separate DB queries should be combined into a batch repository method `GetSystemSettingsBatch()`.
* **MEDIUM:** `backend/internal/rbac/repository.go` — N+1 risk in `GetRoleProjectID`. Suggest using `.WithProject()` to eager-load relationships.
* **MEDIUM:** Missing composite indices in Ent schemas (`issue.go`, `user.go`, `project.go`) for frequent filtering and sorting fields.
* **MEDIUM:** `frontend/src/components/messenger/message-item.tsx` — `MessageItem` is not wrapped in `React.memo`, causing unnecessary re-renders of the entire message list on state changes.
* **MEDIUM:** `messenger-store.ts` — `updateChannel` uses `.map()` (O(n)). Consider normalizing the state structure to `Record<string, Channel>`.
* **MEDIUM:** Missing Explicit GZIP/Brotli response compression middleware in `cmd/server/wiring_fiber.go`.

## 2. Code Review & Architecture Findings (DRY, KISS, OOP)
* **CRITICAL (Security/Quality):** Hardcoded API Key in `frontend/src/components/messenger/gif-picker.tsx:47`. `GIPHY_API_KEY` is hardcoded. It must be moved to `import.meta.env`.
* **WARNING (Bug):** `backend/internal/auth/handler.go` (`VerifyMagicLink`) validates string length by byte-count (`len(req.Code) != 6`), which can break on multibyte characters. Fix: `utf8.RuneCountInString()`.
* **WARNING (God Component):** `frontend/src/components/messenger/message-composer.tsx` is over 500 lines and violates SRP (mixes state logic with UI for typing, polls, audio, GIF, mentions). Break it down into subcomponents and hooks.
* **WARNING (Monolithic Routing):** `onboarding.lazy.tsx` and `cases_.$caseKey.lazy.tsx` are monolithic container files (around 30KB). Abstract the complex wizard logic into discrete step components.
* **INFO:** Many backend HTTP handlers return naked errors (`return err`), making tracing difficult. Use `fmt.Errorf("...: %w", err)` for context.

## 3. Security Audit Findings
* Security Audit agent did not find any *new* actionable vulnerabilities without a specific Threat Model. 
* Confirmed via `AUDIT-2026-06-23.md` that Wave C FA-02 Security Re-audit was passed previously with 0 CRITICAL vulnerabilities.

## Recommended Action Plan
1. Fix the GIPHY hardcoded API key (Security hygiene).
2. Fix the Magic Link byte-length validation bug.
3. Optimize the N+1 queries in the Go backend.
4. Implement `React.memo` in the Messenger frontend to reduce render thrashing.
5. Create follow-up tasks to refactor the monolithic React components.