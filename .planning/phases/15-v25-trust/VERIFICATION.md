# Phase 15 — Trust (v2.5) — VERIFICATION

**Date:** 2026-07-27  
**Method:** File existence + route/grep inspection against PLAN.md exit checklist  
**Sources:** `PLAN.md` (all waves marked ✓ / SHIPPED), live `backend/` + `frontend/` + `testframework/`

## Goal (exit criteria)

AUTH Magic Link only · organization + pending · NOTIF TG bind · E2E FB-07..09 · WCAG · hub-mail · coverage ≥50% · Swagger/security close-out. Wave 10 test-framework deferred → 999.1 (accepted).

## Measured evidence

| Criterion | Evidence (2026-07-27) | Result |
|-----------|----------------------|--------|
| Magic Link only (HTTP) | `handler.go` registers only `POST /magic-link` + `/magic-link/verify`; wiring hardens verify rate-limit; FE `MagicLinkForm` + `api/auth.ts` magic-link endpoints | PASS |
| Organization + pending | `backend/ent/schema/organization.go` exists; `admin.Get("/users/pending", …)`; `cron.Schedule(…, "pending-user-purge", …)` in `wiring_auth.go` | PASS |
| NOTIF TG bind (not login) | `backend/internal/notification/telegram_bind.go`, `handler_telegram_bind.go`; FE `telegram-link-card.tsx` | PASS |
| E2E FB-07..09 | Specs present: `project-lifecycle.spec.ts`, `messenger.spec.ts`, `tms.spec.ts`, `login.spec.ts` under `testframework/stages/e2e/domains/` | PASS |
| WCAG | `testframework/stages/e2e/domains/platform/a11y.spec.ts` exists | PASS |
| hub-mail + internal token | `backend/cmd/hub-mail` exists; `InternalTokenAuth` + `HUB_MAIL_INTERNAL_TOKEN` in `backend/internal/mail/` | PASS |
| Coverage ≥50% | Project baseline COV **83.0%** (AGENTS.md / COVERAGE.md, 2026-06-22 Docker) ≫ 50% | PASS |
| Wave 10 deferred | PLAN explicitly defers to Phase 999.1 — not a phase-15 failure | N/A |

## Notes

- Legacy helpers `LoginWithTelegram*` / `TelegramBotAuth` remain in auth package (back-compat/tests); **HTTP login routes for Telegram are not registered** — aligns with TZ Magic-Link-only.

## Verdict: **PASS**
