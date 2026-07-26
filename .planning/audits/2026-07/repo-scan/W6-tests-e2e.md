# W6-tests-e2e — Playwright E2E

> **Provenance class:** FALLBACK_SUMMARY · **Confidence:** PRELIMINARY · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Роль:** qa-auto-ver (readonly) · **Дата:** 2026-07-23  
> **Scope:** `testframework/stages/e2e/domains/**/*.spec.ts` · **Код не менялся**

## Summary

На диске **20** E2E specs (полный набор). Smoke-набор 3 (auth/platform). **`waitForTimeout(3000)`** в `ai/ai-hub.spec.ts:228` — flaky anti-pattern (HIGH/CRITICAL по qa-auto-ver).

**Вердикт:** `verified: false` из-за waitForTimeout.

## Findings

| Severity | file:line | Issue | Fix |
|----------|-----------|-------|-----|
| **HIGH** | `ai/ai-hub.spec.ts:228` | `await page.waitForTimeout(3_000)` | Web-first `expect` / `expect.poll` |
| **PASS** | inventory | **20** specs на диске | Держать PIPELINE sync |
| **MEDIUM** | coverage | Нет E2E denial IDOR (ожидаемо после BE) | Post-fix security e2e |
| **MEDIUM** | calls | Нет dedicated calls E2E (LiveKit hard) | Smoke join mocked |
| **INFO** | domains | auth×3, admin×3, platform×5, ai×2, nav×3, messenger/tms/tracker/v4 | — |
| **PASS** | guides | Явный запрет waitForTimeout в guides | Enforce в review |
| **LOW** | e2e-full CI minutes | Студенческий пакет — не раздувать | — |

## Specs (20)

auth: login, logout, onboarding-profile · admin: admin-ai, admin-audit, admin-metrics · ai: ai-hub, automation · messenger, tms, tracker/project-lifecycle, v4-superapp · navigation×3 · platform: a11y, critical-gate, health, metrics, notifications
