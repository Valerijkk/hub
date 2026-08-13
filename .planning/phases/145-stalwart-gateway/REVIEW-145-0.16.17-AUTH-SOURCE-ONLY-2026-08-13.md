# Phase 145 Stalwart authentication source-only review — 2026-08-13

> **Status:** `SOURCE_ONLY_NO_FINDINGS / NOT_ACCEPTED_FOR_GATE`
>
> **Reviewed worktree HEAD:** `da00edc9cc1886fe5d69abd89bfc2cb65a754d22`
>
> **Selected upstream source:** Stalwart `v0.16.17`, commit
> `08509ddccbe6f978a1ffebeb250cab8588d765e5`

## Scope

Read-only review covered the seven-file Phase 145 authentication research, gate, plan, QA,
state and founder-question packet. No product source, runtime configuration or protected
test-rebase path was read or changed by the reviewer.

## Source findings

| Check | Result | Exact evidence |
|---|---|---|
| Management versus mail route | PASS | `crates/http/src/request.rs` routes authenticated mail JMAP under `/jmap` and `/api` to `handle_api_request`. |
| API-key boundary | PASS | Official API-key documentation makes the credential management-only and explicitly excludes JMAP mail and other mail/collaboration protocols. |
| OAuth flow set | PASS | Official documentation lists Authorization Code and Device Authorization. The packet correctly rejects an assumed client-credentials grant. |
| Shared ACL versus membership | PASS | `crates/jmap/src/api/request.rs` uses access checks for Email/Mailbox but membership checks for Identity, EmailSubmission, Sieve and Blob; `mailbox.rs` ACL rights do not remove those checks. |
| Recommendation wording | PASS | Authorization Code OAuth remains conditional and founder-selected; impersonation is described as broad/high-risk; enabled mail auth remains disabled before G145-5B. |

No P0, P1 or P2 content finding was reported.

## Gate limitation

The review was requested as `gpt-5.6-terra/high`, but the reviewer runtime exposed neither
actual model nor actual effort telemetry. Both are therefore `UNKNOWN`. Under the founder
model-routing and provenance lock, this evidence cannot accept G145-5B or authorize product
implementation. Runtime, tests, validators, image pull and deployment were not run.
