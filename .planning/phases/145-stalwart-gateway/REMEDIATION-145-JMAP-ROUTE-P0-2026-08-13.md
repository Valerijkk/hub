# Phase 145 — Stalwart v0.16.17 shared-JMAP route P0 remediation

> **Status:** `ROOT_CONTRACT_CORRECTED / SOURCE_ONLY_REREVIEW_NO_P0_P1 / P2_ROOT_FIXED / TELEMETRY_UNKNOWN / RUNTIME_NOT_RUN`
>
> **Selected upstream source:** Stalwart `v0.16.17`, commit
> `08509ddccbe6f978a1ffebeb250cab8588d765e5`

## P0

The prior Phase 145 packet treated management Registry JMAP as `POST /api` and standard
mail JMAP as `POST /jmap`. That route split is false for the selected version and would make
provisioning non-functional while encouraging authorization selection from a URL.

Exact pinned evidence:

- `UPGRADING/v0_16.md` states that previous-release `/api/...` management endpoints no
  longer exist and management operations are JMAP objects reachable at `/jmap`.
- `crates/http/src/request.rs` authenticates and dispatches JMAP at `/jmap`; `/api` enters a
  separate helper handler.
- `crates/jmap/src/api/request.rs` dispatches both standard mail methods and Registry
  Get/Query/Set branches for `x:*` objects.
- `crates/http/src/api/mod.rs` exposes auth/discovery/account/schema/token/live helpers;
  anonymous rate-limited `POST /api/auth` is OAuth login, not provisioning.
- Current object-document examples that show `/api` conflict with exact-version source and
  upgrade guidance. The pinned source contract wins.

## Corrected invariant

One trusted origin exposes one fixed `/jmap` route to Hub. The adapter has two typed,
non-interchangeable operation contexts:

1. Management Registry context: separately scoped API key; only exact accepted `x:*`
   methods/capabilities; fixed `/jmap` only.
2. Mailbox context: separately accepted mailbox credential; only standard identity/mail/
   submission/blob/Sieve methods and accepted same-origin upload/download URLs.

The client rejects before I/O:

- any JMAP batch mixing Registry and mailbox methods;
- a management key used for a standard mail method, upload/download or `/api/*` helper;
- a mailbox credential used for an `x:*` Registry method;
- provisioning through `/api`;
- context selection inferred from a URL rather than a typed operation.

`/api` remains an exact-path helper/auth namespace. If Authorization Code OAuth is selected,
only the independently accepted browser-facing paths and exact anonymous rate-limited
`POST /api/auth` exception may be exposed. Both Registry and mailbox `/jmap` remain
Hub/service-only.

## Acceptance state

The planning, dispatch, QA, state and founder-question contracts were corrected together.
The earlier source-only review is marked superseded. A fresh source-only re-review found no
remaining P0/P1; its sole P2 wildcard-wording concern was root-fixed with individually
allowlisted `/api` path+method pairs. The reviewer exposed no actual model/effort telemetry,
so this is not formal gate acceptance. No code packet may start until provenance and all
other G145 gates pass and the accepted Docker-only launcher exists. No runtime, image pull,
build, test, validator or deployment was performed by this remediation.
