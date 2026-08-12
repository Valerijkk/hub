# Phase 144 — Mailbox Domain

> Task: TASK-237 · Status: PLANNED · Depends on: Phase 142

## Goal

Implement the engine-independent domain and HTTP behavior against a deterministic fake.

## Ordered tasks

1. `144-01` — public `MailEngine` port, DTOs, error taxonomy and fake; write adapter
   contract units under `backend/tests/mail/**`.
2. `144-02` — Ent schemas for account, identity, folder, thread, message, label,
   attachment, rule, contact and submission; Docker-only generation.
3. `144-03` — pure address normalization, threading, reply/forward quoting, folder tree,
   typed rule compiler, rate/loop and attachment policies with failing units first.
4. `144-04` — repositories/services for folders, labels, list/read/bulk, draft revisions,
   send idempotency and ownership; fake ports and service units.
5. `144-05` — storage presign/confirm adapter, scanner interface, search/event projection,
   recent contacts, notifications/ws/audit seams.
6. `144-06` — Fiber handlers/routes matching `docs/mail/openapi.yaml`, admin policy DTOs,
   error mapping and handler units where genuine.
7. `144-07` — system/architecture/security docs sync and independent static review.

Each item is one or more atomic commits split further when generated schema, service,
handler, or docs cannot be reviewed independently. Existing Resend alert behavior stays
separate and green.
