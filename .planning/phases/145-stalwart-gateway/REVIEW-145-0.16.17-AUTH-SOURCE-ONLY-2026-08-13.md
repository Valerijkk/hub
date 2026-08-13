# Phase 145 Stalwart authentication source-only review — 2026-08-13

> **Status:** `SUPERSEDED_BY_P0 / NOT_ACCEPTED_FOR_GATE`
>
> **Reviewed worktree HEAD:** `da00edc9cc1886fe5d69abd89bfc2cb65a754d22`
>
> **Selected upstream source:** Stalwart `v0.16.17`, commit
> `08509ddccbe6f978a1ffebeb250cab8588d765e5`

## Scope

Read-only review covered the seven-file Phase 145 authentication research, gate, plan, QA,
state and founder-question packet. No product source, runtime configuration or protected
test-rebase path was read or changed by the reviewer.

## Superseding finding

This historical review accepted a false route premise. Exact pinned source plus
`UPGRADING/v0_16.md` show that `v0.16.17` routes both management Registry `x:*` methods and
standard mail JMAP through `/jmap`; `/api` is a helper/auth/account/schema/telemetry
namespace. The corrected contract is recorded in
`REMEDIATION-145-JMAP-ROUTE-P0-2026-08-13.md`. Nothing in this file may be used as positive
gate evidence.

## Historical source findings

| Check | Result | Exact evidence |
|---|---|---|
| Management versus mail route | **P0 FAIL** | The review mistook `handle_api_request` for management Registry dispatch. Registry Get/Query/Set branches are in the JMAP handler reached at `/jmap`; pinned upgrade guidance says prior `/api/...` management endpoints no longer exist. |
| API-key boundary | PASS | Official API-key documentation makes the credential management-only and explicitly excludes JMAP mail and other mail/collaboration protocols. |
| OAuth flow set | PASS | Official documentation lists Authorization Code and Device Authorization. The packet correctly rejects an assumed client-credentials grant. |
| Shared ACL versus membership | PASS | `crates/jmap/src/api/request.rs` uses access checks for Email/Mailbox but membership checks for Identity, EmailSubmission, Sieve and Blob; `mailbox.rs` ACL rights do not remove those checks. |
| Recommendation wording | PASS | Authorization Code OAuth remains conditional and founder-selected; impersonation is described as broad/high-risk; enabled mail auth remains disabled before G145-5B. |

The former “no findings” conclusion is withdrawn. The route finding above is P0 because it
would send provisioning requests to a non-provisioning helper namespace and build credential
selection around a false URL split.

## Gate limitation

The review was requested as `gpt-5.6-terra/high`, but the reviewer runtime exposed neither
actual model nor actual effort telemetry. Both are therefore `UNKNOWN`. Under the founder
model-routing and provenance lock, this evidence cannot accept G145-5B or authorize product
implementation. Runtime, tests, validators, image pull and deployment were not run.
