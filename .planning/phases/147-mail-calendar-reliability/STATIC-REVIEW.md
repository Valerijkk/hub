# Phase 147 — Mail 0.1.3 root static review

> **Reviewed contract SHA:** `8daaf0d9a021e99e0f838dfdd765abd9dd9e8194`
> **Parent:** `5c12e907b5c5e1a9f484b3fe46031e7aee4a5879`
> **Branch:** `codex/mail-module-20260813`
> **Verdict:** `ROOT_STATIC_PASS`
> **Independent G148-02 verdict:** `PENDING`
> **Runtime/build/test/schema-validator evidence:** `NOT_RUN`
> **Actual model telemetry:** `UNKNOWN`; this root review is not accepted as the required
> independent Terra provenance artifact

## Scope and method

The review compared the exact 28-path change in the reviewed SHA across ADR, Mail spec,
OpenAPI, event schema, UI design, Phase 146/147 plans, QA handoff, security/system/test
documentation and the Phase 148 gate. It checked source-of-truth ownership, state-machine
ordering, privacy, recovery, HTTP/event shapes, compatibility, unit taxonomy and packet
atomicity.

Read-only Git evidence on the exact SHA:

- clean isolated worktree after commit;
- `git diff --check` from parent to reviewed SHA: PASS;
- exactly 28 changed paths, all inside the Mail planning/documentation scope;
- protected `testframework/**`, test-rebase phases/audits/tasks and `pro-qa-ai/**`: no
  changed path;
- targeted trailing-whitespace and common secret-token scan: no match.

Project policy forbids host tests, builds, linters and validators. The accepted versioned
Docker launcher is `NOT_READY`, so OpenAPI/JSON executable validation and all product
runtime claims remain `NOT_RUN` rather than being inferred from this review.

## Findings closed before the reviewed SHA

| ID | Severity | Finding | Frozen resolution |
|---|---|---|---|
| R147-ROOT-01 | P1 | Expiring an early CANCEL tombstone contradicted the promise that delayed mail cannot resurrect an object | unbound tombstone persists for mailbox/account life; deferred missing-master work alone expires after seven days |
| R147-ROOT-02 | P1 | An occurrence with no prior row could bypass immutable series organizer ownership | occurrence admission locks series first, requires organizer equality and defers while the master is absent/cancelled |
| R147-ROOT-03 | P1 | Whole-series cancellation and occurrence sequencing were underspecified | sequence remains per UID + optional RECURRENCE-ID; series CANCEL cancels materialized exceptions and retains high-watermarks |
| R147-ROOT-04 | P1 | Reminder lookahead could suppress too early or turn a repository failure into a false policy decision | only due-time state is terminal; transient user/event/preferences/store failure leaves `pending` and retries; post-commit WebSocket is best-effort |
| R147-ROOT-05 | P1 | Initial availability wording widened existing HR/calendar privacy to organization-wide visibility | all users authorize before queries under self/shared-project-or-channel scope, with no admin bypass and target-user IANA leave expansion |
| R147-ROOT-06 | P1 | A single L1 feed packet mixed schema, crypto, network, parsing, DB transaction, HTTP and scheduling | split into eight dependency-ordered packets 147-06A..H; code remains founder-gated |
| R147-ROOT-07 | P1 | Planned Go DB/HTTP “units” would violate the founder greenfield taxonomy | repository/handler packets contain no Go component tests; genuine pure/fake-port units are exact and higher layers remain Python QA handoff |
| R147-ROOT-08 | P1 | Machine contract lacked complete availability/CRM status and stable error coverage | OpenAPI 0.1.3 adds closed availability/CRM shapes, 404/422/409 codes, opaque cursors, legacy calendar path and same-origin RSVP recovery URI |
| R147-ROOT-09 | P1 | CRM “open composer” mixed API/UI/send/link behavior and could leak recipient state or duplicate a draft after lost response | ordinary draft POST uses one key/body, URL carries only draft UUID, Mail owns send, explicit link waits for owned message, and UI is split into 07D..G |
| R147-ROOT-10 | P1 | The 10,000 unbound-state cap had no explicit safe degraded behavior | unseen UIDs quarantine before mutation, known UIDs continue, health/ops signal is bounded/content-free, and tombstones are never auto-evicted |

## Remaining gates

No unresolved P0/P1 was found in the root-reviewed contract snapshot. This does not make
G148-02 pass: an independent current-SHA review with accepted actual model/effort provenance
is still required. Exact Spark implementation, migration/generation authority, selected
engine evidence, accepted Docker launcher, executable QA, broadcast disposition, final
rebase and push/PR authority remain separately blocked, unresolved or `NOT_RUN` in
`GATE-MATRIX.md`.

The only claim supported here is: Mail 0.1.3 is internally coherent enough for independent
static review. It is not implemented, runtime-verified, merge-ready, complete, stable or
production-ready.
