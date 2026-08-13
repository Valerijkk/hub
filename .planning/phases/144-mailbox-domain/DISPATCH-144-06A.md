[РОЛЬ И TASK CONTRACT]

Ты — `codex-micro`, атомарный Go-исполнитель Hub. Реализуй только public mailbox HTTP
DTO/error mapping, Fiber handlers and narrow composition wiring matching OpenAPI 0.1.6.

- User outcome: authenticated user uses folders/labels/threads/messages/drafts/send/
  submission/attachments/search/settings through `/api/mail/**` without choosing mailbox
  or organization ids.
- Completion: failing genuine handler units precede minimal bindings; exact route/method,
  NFR-13 and recovery headers match OpenAPI; legacy routes remain unchanged.
- Non-goals: service/domain logic, Ent, Stalwart, UI, admin mailbox policy/domain/health,
  legacy Resend behavior, migration, accepted runtime claim.

[OWNERSHIP]

- Entry: root-reviewed 144-04* and 144-05* public service interfaces on clean HEAD.
- Allowed mutations only:
  - `backend/internal/mail/mailbox_dto.go`
  - `backend/internal/mail/mailbox_handler.go`
  - `backend/cmd/server/wiring_mailbox.go`
  - the smallest explicitly approved call site in existing server wiring needed to invoke
    `wireMailbox`; no other wiring refactor
  - `backend/tests/mail/mailbox_handler_test.go`
- Do not edit existing `dto.go`, `hub_handler.go`, `admin_handler.go`,
  `admin_broadcast.go`, `/calendar/events`, legacy `/admin/mail/settings`, generated docs or
  protected test roots.

[FROZEN HTTP RULES]

- Exact source: `docs/mail/openapi.yaml` 0.1.6. Implement no undocumented endpoint and
  silently omit no documented public mailbox endpoint.
- Principal middleware supplies `user_id`; service resolves organization/account.
  Requests carrying `organization_id`, `mail_account_id` or engine ids are rejected or
  ignored only as specified DTO unknown-field policy; never trusted.
- JSON errors always follow Hub NFR-13 `{error,request_id}`. Machine detail uses approved
  headers only. Do not invent endpoint-specific JSON envelopes.
- `Idempotency-Key` is required/validated for every contracted create/bulk route, draft
  create and send. `If-Match` quoted
  positive integer is required for draft patch.
- `GET /mail/submissions/{submissionId}` invokes owner-scoped reconciliation behavior.
- Cross-account opaque ids do not disclose existence and produce zero engine call through
  the service oracle.
- Pagination cursors are opaque/bounded; limits default 50/max 100.
- Folder and thread-list responses always serialize `projection_state` plus nullable
  `projection_as_of` exactly as OpenAPI. `bootstrapping`, `degraded` and `rebuilding` keep
  owner filtering and current rows but are never mapped to fresh/ready. Opaque engine
  cursor/checkpoint/base-cursor and safe internal retry state never leave the service.
- Engine command/reconciliation lease contention maps only to `423` plus
  `X-Hub-Error-Code: MAIL_ACCOUNT_BUSY` and bounded `Retry-After`; the service guarantees
  zero engine call for that response. Unknown/indeterminate outcomes never map to busy.
- `GET /mail/account` is read-only and returns only safe provisioning state; it never
  creates an intent or exposes local part, domain id, lease, engine id or provider text.
- Rule create/update/delete changes Hub desired state only and therefore never maps 423.
  Rule responses carry `pending|applying|ready|unknown|error` deployment state honestly.
- Remote image route accepts only server token paired with owned message id; never URL.
- Response serializes draft/message bodies returned by authorized engine hydration but
  handler logging/error mapping must not retain them.

[REQUIRED ORACLES]

1. Route inventory matches every `/mail/**` path/method in OpenAPI 0.1.6, including
   submission GET, without duplicate registration.
1A. Folder/thread list fixtures cover all four projection states and nullable freshness;
    stale states remain explicit and expose no engine cursor/checkpoint/lease/error text.
2. Missing/malformed auth-derived context stops before service call.
3. Missing/invalid idempotency key on any contracted create/bulk route and malformed
   `If-Match` return safe NFR-13 error.
4. Generic receipt conflicts serialize reused/unavailable/expired exactly; specialized
   draft/create and legacy conflicts expose only `IDEMPOTENCY_KEY_REUSED` where specified.
5. Draft update/delete/send conflicts use their exact codes and never overwrite, delete,
   submit or expose body from a newer revision.
6. 204 endpoints produce empty body; JSON endpoints produce documented schema shape.
7. Cursor/limit/bulk cardinality/address/body/attachment bounds are enforced before
   service call.
8. Cross-account not-found/forbidden mappings do not reveal which case occurred.
9. Handler/request logs and error values omit raw key, message/draft body, recipient list,
   upload URL query and engine error.
10. Existing Resend status/settings/broadcast and legacy calendar route registrations are
    unchanged by exact diff.
11. Folder create over the fixed 100-folder engine cap maps to 422 plus
    `X-Hub-Error-Code: MAIL_FOLDER_LIMIT_EXCEEDED`; no partial projection is returned.
12. Attachment confirm mismatch/state and expiry map to the frozen 409/410 codes; non-clean
    download is 409. A clean 302 has required fixed-storage HTTPS Location plus
    `Cache-Control: no-store` and `Referrer-Policy: no-referrer`.
13. Every new mailbox request rejects unknown properties before service invocation;
    forged organization/account/engine fields and invalid rule or bulk field/operator/value/
    action/target combinations make zero repository/engine call.
14. Folder DTO always carries control state/version. Create/update/delete map definitive
    success to 201/200/204 and safely persisted indeterminate intent to 202; stale/protected/
    name/pending conflicts use frozen 409 codes and no new engine command.
15. Every new mailbox 409 serializes the exact route-specific `X-Hub-Error-Code` from
    OpenAPI. Global domain/address collision maps only to the non-disclosing unavailable
    code; no handler substitutes the generic Conflict response.
16. Draft recipient request objects accept only email/name and reject response-only
    `user_id` before service invocation; enriched response ids remain server-derived.
17. Submission, rule, domain and bulk-target error codes serialize only the frozen allowlist;
    arbitrary provider strings are rejected/redacted before DTO construction.
18. Every OpenAPI engine-command route maps pre-I/O lease contention to the exact 423 busy
    response; fixtures reject missing/invalid code or Retry-After and prove no busy mapping
    after an indeterminate engine call.
19. Account-status fixtures prove read-only not-configured/provisioning/active/suspended/
    error shapes and no engine/reservation/provider fields.
20. Rule writes never emit 423 or call the engine in handler/service fakes; deployment
    state and idempotency conflict serialize exactly.

[EXECUTION / PROVENANCE]

- TDD first with fake service; no real DB/engine/network/listener.
- Host validation/direct Docker forbidden; accepted launcher absent => `NOT_RUN`.
- No dependency/generated code/commit/push/PR. max_depth 0; fan_out 0; retries 2;
  wall clock 75m.
- requested_model `gpt-5.6-terra`; reasoning `xhigh`; fallback none; mismatch => `BLOCKED`.

[ОТВЕТ]

На русском: route/method inventory; changed files/symbols; error/header matrix;
compatibility evidence; verification; model provenance; risks.
