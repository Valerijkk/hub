# Mail 0.1.8 machine-contract closure design

> **Status:** root-selected remediation design under the founder's autonomous-work mandate
> **Runtime:** `NOT_RUN`; the accepted Docker launcher remains `NOT_READY`

## Goal

Close the current Phase 142 OpenAPI, event-envelope and engine/persistence representation
findings without widening Mail product scope or claiming implementation/runtime acceptance.

## Decisions

1. `docs/mail/openapi.yaml` defines one versioned `x-hub-authz-vocabulary`. Every non-admin
   operation carries exactly one `x-hub-authz` policy id; admin operations continue to use
   explicit `x-hub-rbac: admin`. The vocabulary is default-deny, tenant-aware and maps a
   resource-scope miss to the shared non-disclosing `404` response.
2. Closed response DTOs reject undeclared fields. Summary/detail composition uses open field
   bases plus separately closed public summary/detail schemas, avoiding the invalid pattern
   where a closed base rejects fields added by a derived `allOf` schema.
3. Standard OpenAPI examples cover bounded cursor continuation, a recoverable `409` with
   `X-Hub-Error-Code`/`ETag`/`Location`, `423` with `Retry-After`, and identical-input
   idempotent replay. Examples are schema-shaped evidence, not runtime evidence.
4. Mail received/sent envelopes carry at most 200 normalized `to` addresses and calendar
   invite envelopes carry at most 200 attendee ids, matching their engine/domain bounds. The
   channel-scoped call-due variant adds a structured cross-field invariant that captures the
   UUID from `subject` and compares it with `/payload/channel_id`.
5. Draft recipient collections remain capped at 100 each. The deployment policy maximum is
   therefore 300 total recipients and the engine contract names the combined ceiling.
6. The deprecated legacy broadcast route remains a governance gap pending its already-decided
   disable/removal gate. While it remains in the compatibility contract, its request and
   response are closed and explicitly bounded; this does not re-authorize bulk marketing.

## Non-goals

- No product source, generated Ent code, migrations or protected QA files change.
- No permission is added and no existing ownership boundary is relaxed.
- No host test, validator, generator, linter or build runs.
- No Docker/Compose fallback is used while the accepted launcher is `NOT_READY`.

## Verification

Static evidence consists of exact-path review, operation-to-policy coverage, closed-object
inventory, bound synchronization searches, JSON/YAML textual consistency, `git diff --check`
and staged-path review. OpenAPI/JSON Schema executable validation and runtime behavior remain
explicitly `NOT_RUN` until the accepted task-owned Docker launcher exists.
