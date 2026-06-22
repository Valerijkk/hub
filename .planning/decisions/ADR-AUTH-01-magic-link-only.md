# ADR-AUTH-01 — Magic Link Only Authentication

> **Status:** Accepted · **Date:** 2026-06-21 · **Deciders:** PO

## Context

Hub historically supported Magic Link + Telegram login (US-M01). PO wants **Cloud Code / Cursor pattern**: email OTP only. Telegram remains valuable as **notification channel**, not identity.

## Decision

1. **Login:** Magic Link (Resend) only — no Telegram Widget, no `/auth/telegram`, no `/auth/telegram-code` for auth.
2. **Telegram bot:** Notification bind + delivery only (`NOTIF-TG-EMAIL-SPEC.md`).
3. **`user.telegram_id`:** Set via bind flow, not registration via Telegram OAuth.

## Consequences

### Positive
- Single auth UX; reduced security surface
- Clear separation: identity (email) vs channel (TG)

### Negative
- Users must have email; Telegram-only signup removed
- Migration banner for legacy TG-login users to add verified email

### Implementation
- Remove: `TelegramButton`, login bot `/login` handler
- Keep: `TelegramSender`, bind API NOTIF-02..04
- Update: `60-tz-constraints.mdc`, vision US-M01, security_guidelines

## Compliance

Supersedes TZ line «Magic Link + Telegram» for **authentication** only. Telegram bot token still required for notifications.
