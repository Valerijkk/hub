# TASK-237 — Mailbox domain and engine port

> Date: **2026-08-13** · Priority: **P0** · Status: **Open** · Phase: **144**

## Outcome

Implement the engine-neutral mailbox domain: account/identity/folder/thread/message
projections, labels, attachments, rules, contacts, drafts, submission idempotency, fake
engine, services, handlers, and NATS outbox contracts.

## Owned paths

- `backend/ent/schema/mail*.go` and `backend/ent/schema/recentcontact.go`
- mail-owned generated Ent/migration artifacts
- mail-owned additions under `backend/internal/mail/**`
- mail-owned additions to storage/search/notification/ws/audit/wiring
- `backend/tests/mail/**`
- backend Swagger additions generated only through accepted Docker execution

The existing Resend alert files under `backend/internal/mail/**` remain behaviorally
unchanged unless a narrowly documented compatibility seam is required.

## Acceptance

- All engine calls are server-bound to the authenticated owner's account.
- Threading, addressing, reply/forward quoting, folder constraints, rule compilation,
  submission idempotency, rate/loop policy, and event mapping have genuine units.
- PostgreSQL stores no raw MIME/body and no engine credentials.
- SeaweedFS uses existing presign/confirm infrastructure plus fail-closed scan state.
- Search is asynchronous through versioned NATS events and owner-filtered hydration.
- API behavior matches `docs/mail/openapi.yaml` or the contract is amended before code.
- Same-session docs sync and atomic commits are maintained.
