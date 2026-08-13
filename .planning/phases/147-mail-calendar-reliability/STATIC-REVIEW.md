# Phase 147 — Mail 0.1.4 root static review

> **Reviewed semantic contract SHA:** `46cfa26c54800c27af89c9ccf92adf40d7e966e6`
> **Parent:** `2ac87fcb9722acb6fface1b8db4f9629f39a6793`
> **Branch:** `codex/mail-module-20260813`
> **Verdict:** `ROOT_STATIC_PASS`
> **Superseded by:** Mail 0.1.5 amendment; this pass is historical `STALE` for current
> contract acceptance and must not be reused as current-provenance evidence.
> **G148-02 state:** `ROOT_LENS_PASS` · `INDEPENDENT_REVIEW_PENDING`
> **Product/test code:** `NOT_IMPLEMENTED`
> **Runtime/build/test/schema-validator evidence:** `NOT_RUN`
> **Actual model telemetry:** `UNKNOWN`; this root review is not accepted as the required
> independent Terra provenance artifact

This verdict is deliberately narrow. It establishes internal static coherence for the exact
semantic snapshot above. The child evidence commit may change review/status prose only; it
cannot broaden this verdict to implementation, runtime, merge-readiness or production.

## Scope and method

The review compared the exact 35-path parent-to-SHA change across the Mail ADR/spec/OpenAPI,
calendar and mailbox dispatch contracts, Phase 147 state machines, Phase 148 gates, QA
handoff, UI design and routed system/architecture/security/test/product documentation.
Manual goal-backward inspection covered:

- reminder receipt/outbox/outcome/reason coherence and terminal-state closure;
- due-time and consumer-time current sequence/start/offset/eligibility checks;
- event → exception → attendee → stable user-policy guard → receipt → notification lock
  order, including an absent preference row and a concurrent quiet-hours update;
- strict invalid-policy handling, transient-failure retry and post-commit-only external
  channels;
- deterministic all-day day-boundary behavior across ordinary, overlap, gap and skipped
  civil dates without changing date-only persistence/API/iTIP;
- existing-message CRM linking from the ordinary owned reader with current-owner filtering
  as UX/privacy defense and independent server authorization as authority;
- legacy `/calendar/events` compatibility, NFR-13 errors, conditional RSVP recovery headers
  and same-origin canonical UUID `Location`;
- exact-file packet atomicity, protected-zone exclusions and honest blocked/not-run gates.

Read-only Git evidence on the exact reviewed SHA:

- isolated worktree was clean after commit;
- `git diff --check 2ac87fcb..46cfa26c`: PASS;
- exactly 35 modified paths, all inside the allowed Mail planning/documentation scope;
- protected `testframework/**`, legacy `backend/testframework/**`, test-rebase phases/audits/
  tasks and `pro-qa-ai/**`: zero changed paths;
- targeted added-line secret scan and forbidden commit-trailer scan: PASS;
- `docs/mail/events-v1.schema.json` blob stayed
  `46f7b153d63a8dfa5f08c99d4d1cdb681ec574e9`; the event payload shape did not drift;
- OpenAPI blob is `162ac0c7c1e90d32228aaee8dc74288f2a92a099` and contains the legacy path,
  canonical NFR-13 envelope and bounded relative RSVP recovery path.

Project policy forbids host tests, builds, linters and validators. The accepted versioned
Docker launcher is `NOT_READY`, so executable OpenAPI/JSON validation and every product
runtime claim remain `NOT_RUN` rather than being inferred from static inspection.

## Findings closed in Mail 0.1.4

| ID | Severity | Finding | Frozen resolution |
|---|---|---|---|
| R147-ROOT-11 | P1 | A reminder emitted from lookahead could survive an offset disable/change, reschedule, cancel/delete or attendee removal/decline | scheduler and consumer both recheck the current sequence, effective start, non-null offset, recomputed due and eligibility; mutation-first durably suppresses with no row |
| R147-ROOT-12 | P1 | Optional preference rows and the legacy invalid-timezone fallback left a policy race/fail-open path | the stable user row serializes first insert/update and quiet-hours mutation; malformed/incomplete policy becomes one `policy_invalid` suppression with no UTC/delivery fallback |
| R147-ROOT-13 | P1 | Receipt state, outbox, notification and suppression reason combinations were not closed | one explicit coherence table and closed safe-reason enum constrain pending/emitted/delivered/suppressed/expired tuples and forbid terminal reopen |
| R147-ROOT-14 | P1 | Event-local `00:00` was ambiguous in midnight overlaps, gaps and wholly skipped dates | all-day due uses the first representable instant; overlap picks earlier, gap/skipped date picks the first instant after the gap; QA packets carry exact fixtures |
| R147-ROOT-15 | P1 | Atomic packets did not explicitly prohibit Telegram/Web Push before database commit | keyed external work starts only after canonical commit; rollback invokes none and external results cannot mutate/rollback the durable outcome |
| R147-ROOT-16 | P1 | The contract allowed manual linking of an existing message but had no reader-side atomic implementation packet | packet 147-07H owns the current-user deal picker, explicit idempotent PUT, accessibility states and privacy oracles; `TC-CRM-MAIL-003` covers it |
| R147-ROOT-17 | P2 | RSVP recovery `Location` allowed a syntactically loose 36-character id | OpenAPI now requires `uri-reference` plus the exact same-origin canonical UUID event-path pattern |

## Historical 0.1.3 findings retained by ancestry

The former root review on `8daaf0d9a021e99e0f838dfdd765abd9dd9e8194` is stale as
current evidence, but its resolutions remain present and were checked for regression:

| IDs | Previously closed boundary |
|---|---|
| R147-ROOT-01..03 | lifetime CANCEL tombstone, immutable organizer ownership, series/occurrence sequencing and whole-series cancellation |
| R147-ROOT-04 | due-time authority, transient policy/store retry and post-commit realtime recovery |
| R147-ROOT-05 | all-users-first private availability authorization without admin bypass |
| R147-ROOT-06..07 | eight atomic conditional-feed packets and genuine-unit versus Python non-unit taxonomy |
| R147-ROOT-08 | machine-readable availability/CRM statuses, NFR-13 and legacy calendar compatibility |
| R147-ROOT-09 | ordinary idempotent CRM-to-Mail draft handoff with no recipient in URL and explicit post-send link |
| R147-ROOT-10 | bounded safe degraded behavior at the 10,000 unbound calendar-state cap |

## Remaining gates

No unresolved P0/P1 was found in the exact reviewed contract snapshot. G148-02 nevertheless
remains pending because it requires an independent current-provenance review; this root
review cannot self-certify that lens. Exact Spark implementation, migration/generation
authority, selected-engine evidence, the accepted Docker launcher, executable QA, broadcast
disposition, final protected-rebase signal and push/PR authority remain separately blocked,
unresolved, unauthorized or `NOT_RUN` in `GATE-MATRIX.md`.

The only supported acceptance claim is:
`Mail 0.1.4 root static contract coherence PASS on 46cfa26c`. The module remains unfinished
and is not implemented, runtime-verified,
merge-ready, complete, stable or production-ready.
