# 144-02F-R — mutation-receipt source authority receipt

> **144-02B boundary:** This accepted 02F-R receipt permits only `MailMutationReceipt.folders` inverse solely for `MailFolder.control_receipt_id -> receipt RESTRICT`; actor/message receipt edges remain scalar-only and every other receipt edge/check is unchanged. Combined `45fd3369` + `4f3713590fee455fae1c94f2464d7de548e73e76` is now 02B `SOURCE_AUTHORITY_ACCEPTED` / `EXACT_SHA_PASS` after Contract/Security/Lifecycle corrected-child exact reviews PASS C/I/M `0/0/0`; it authorizes root dispatch only of the separately frozen exact13 projection manifest, not source implementation acceptance. Candidate/precommit/correction/failure-next and lower 02B eligibility/no-dispatch clauses are `CONSUMED/HISTORICAL`, while the immutable `45fd3369` failure remains history. 02C waits accepted 02B source; 02F-D/02E/Ent/generation/migration/runtime/services/API/UI/release remain blocked.

> **Historical authority status:** `CONSUMED/HISTORICAL` at
> `a877848ef4d1ce80a30a4389f97ef0f863615159` (parent
> `25cfb36720f9a35c96268a62f7ee81fd91f83503`, tree
> `917fed11d1316678008229193aef6186ebaa0abb`; subject
> `docs(mail): authorize receipt schema packet`).

> **Source acceptance supersession:** `ce542cb5de839a1fb5200295f710b0ae45bcdbe6` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; parent `1792cdc381eeb4db0ab1ede5b512ba1222eb90b2`, tree `2cadf80c9ae0af99b5c8027c1d78c3ec55af7071`, exact eight blobs in `IMPLEMENTATION-144-02F-R.md`. Tests first; Contract/Security/Lifecycle PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Authority-only/root-dispatch/source-unaccepted and 02B-separate-authority/02C-predecessor route below `CONSUMED/HISTORICAL`; 02B `SOURCE_AUTHORITY_ACCEPTED` for root exact13 dispatch only, implementation unaccepted, 02C waits accepted 02B source, 02F-D blocked; launcher `NOT_READY`, executables/generation/migration/runtime `NOT_RUN`.

## Historical consumed authority

This is a docs-only authority acceptance. It does not accept source implementation, tests, Ent
generation, migration, runtime, service, API, UI, deployment, release, or a version change. The
accepted 02F normative documents and readiness facts remain authoritative. The former broad
no-source routing and candidate/precommit/commit/review-next clauses are
`CONSUMED/HISTORICAL` only for 02F-R; they do not authorize any other 02F source, including rule
deployment.

Root may now dispatch only one test-first, static source packet with these exact eight paths; no
path may be added, omitted, renamed, or substituted:

1. `backend/ent/schema/mailmutationreceipt.go` (new)
2. `backend/ent/schema/mailmutationreceipttarget.go` (new)
3. `backend/ent/schema/mailmutationreceiptmember.go` (new)
4. `backend/ent/schema/mailaccount.go` (edit)
5. `backend/ent/schema/organization.go` (edit)
6. `backend/tests/mail/mail_mutation_receipt_schema_test.go` (new; authored first)
7. `backend/tests/mail/mail_account_identity_schema_test.go` (edit)
8. `backend/tests/mail/mail_domain_policy_schema_test.go` (edit)

The packet has no `mailruledeployment.go`, User schema, generated Ent, migration, persistence
contract, OpenAPI, service, handler, worker, deployment, or runtime authority.

## Frozen source contract

- Receipt has a required organization relation with `NO ACTION` delete behavior.
- Receipt has a nullable account relation with `NO ACTION`; absence is permitted only for
  `domain_create`.
- `actor_user` is a scalar only: no User edge is introduced.
- Receipt target belongs to receipt with `CASCADE` delete behavior.
- Receipt member belongs to both target and receipt with `CASCADE` delete behavior.
- `message_id` is scalar only.
- The required local child inverses inside these three new schemas are `receipt.targets`,
  `receipt.members`, and `target.members`.
- Outside those three new receipt schemas, `MailAccount` receives only the inverse
  `mutation_receipts` edge and `Organization` only the inverse `mail_mutation_receipts` edge.
- No other foreign keys or inverse relations are authorized: in particular, no User, Message, or
  other entity edge is introduced.

## Accepted authority evidence

The exact authority-document boundary is these eight paths: this receipt,
`.planning/STATE.md`, `DISPATCH-INDEX.md`, `DISPATCH-REMAINING.md`, `PLAN.md`, `MAIL-SPEC.md`,
`docs/mail/QA-HANDOFF.md`, and `docs/mail/CONTRACT-CHANGELOG.md`. Independent immutable
exact-SHA Contract/Spec/Oracle `gpt-5.6-terra`/`high`, Security/Adversarial
`gpt-5.6-terra`/`xhigh`, and Lifecycle/Dependency/Boundary `gpt-5.6-terra`/`high` reviews PASS
C/I/M `0/0/0`. Provenance source is `dispatcher-enforced`; fallback is none; actual telemetry is
`UNKNOWN` unless exposed.

This acceptance authorizes root dispatch of the exact source/test manifest above, but does not
accept its implementation. The dispatched packet must still be authored test-first, independently
reviewed, atomically committed, and separately accepted by immutable exact-SHA review.

## Dependency and blocker preservation

This 02B-eligibility/02C-predecessor route is `CONSUMED/HISTORICAL`: 02B is `SOURCE_AUTHORITY_ACCEPTED` for root exact13 dispatch only, implementation remains unaccepted, and 02C waits accepted 02B source.
02F-D (rule deployment) is `BLOCKED` / `NOT_AUTHORIZED`, separate, and nonblocking. 02E, Ent
generation, migration, runtime, services, APIs, UI, release, Phase 142/148, launcher `NOT_READY`,
and executable evidence `NOT_RUN` remain unchanged.
