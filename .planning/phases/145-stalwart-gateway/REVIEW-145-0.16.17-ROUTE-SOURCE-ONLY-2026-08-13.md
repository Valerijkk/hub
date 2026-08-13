# Phase 145 Stalwart shared-JMAP route source-only re-review — 2026-08-13

> **Status:** `SOURCE_ONLY_NO_P0_P1 / P2_ROOT_FIXED / ACTUAL_TELEMETRY_UNKNOWN / NOT_ACCEPTED_FOR_GATE`
>
> **Reviewed worktree base HEAD:** `a8addaf2`
>
> **Selected upstream source:** Stalwart `v0.16.17`, commit
> `08509ddccbe6f978a1ffebeb250cab8588d765e5`

## Scope

The independent reviewer read only the Phase 145 state, research, remediation, historical
review, gate, plan, dispatch, founder-question and QA documents. Product source, runtime,
Docker and protected test-rebase paths were excluded.

## Result

No P0 or P1 remained after the root route remediation. The reviewed packet consistently
states that:

- management Registry `x:*` and standard mailbox methods both use fixed `POST /jmap`;
- `/api` is a helper/auth namespace, with anonymous rate-limited `POST /api/auth` as the
  exact OAuth-login exception currently evidenced;
- management API-key and mailbox credentials are selected by typed operation context,
  never by URL;
- Registry and mailbox methods cannot share one JMAP batch, and wrong-context credentials
  fail before I/O;
- conflicting current `/api` object-document examples are recorded as subordinate to the
  pinned source and `UPGRADING/v0_16.md`;
- the earlier source-only review is explicitly superseded and supplies no gate evidence.

The reviewer reported one P2 wording ambiguity: the positive phrase “exact `/api/*`
helpers” could be interpreted as a prefix wildcard despite the default-deny contract. Root
closed it by requiring “individually allowlisted `/api` path+method pairs” throughout the
positive authorization contract. Negative statements may still use `/api/*` to mean every
unrelated helper path that must remain blocked.

## Gate limitation

The re-review was requested as `gpt-5.6-terra/high`, but the reviewer runtime exposed
neither actual model nor actual effort telemetry. Both remain `UNKNOWN`; therefore the
review cannot accept G145-2, G145-5 or G145-5B. Runtime, tests, validators, image pull and
deployment were not run.
