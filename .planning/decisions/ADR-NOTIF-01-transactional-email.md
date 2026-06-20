# ADR-NOTIF-01 — Transactional Email Alerts

> **Status:** Accepted · **Date:** 2026-06-21 · **Deciders:** PO

## Context

TZ forbade «Email-уведомления (кроме Magic Link через Resend)». PO wants alerts e.g. **tests failed in run** to email.

## Decision

Extend Resend usage to **transactional system alerts** when:
- User opt-in (`channel=email` in notification preferences)
- Email verified (not `telegram_*@hub.local` placeholder)
- Template whitelist: `run_failed`, `run_regression`, `new_login`, `assigned_urgent`
- No marketing, digests, newsletters

Admin configures `RESEND_API_KEY` + `RESEND_FROM_ALERTS`.

## Consequences

- New `EmailSender` in `notification` package
- Update `60-tz-constraints.mdc` FORBIDDEN list: clarify «no marketing email» vs «transactional alerts allowed»

## Out of scope

User SMTP passwords, bulk mail, SSO magic link changes.
