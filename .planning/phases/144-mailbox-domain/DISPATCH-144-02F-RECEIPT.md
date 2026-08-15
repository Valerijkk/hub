# 144-02F-R — mutation-receipt source authority candidate

> **Status:** `ROOT_AUTHORITY_CANDIDATE` / `PRECOMMIT_REVIEW_REQUIRED` / `IMPLEMENTATION_BLOCKED` / `RUNTIME_NOT_RUN`.

## Narrow authority sought

This is a docs-only amendment. It does not accept source, tests, Ent generation, migration,
runtime, service, API, UI, deployment, release, or a version change. The accepted 02F normative
documents and readiness facts remain authoritative. The prior broad no-source routing clause is
consumed **only for 02F-R**, and only after this authority amendment has itself been committed and
accepted; it does not authorize any other 02F source, including rule deployment.

After that acceptance, root may dispatch only one test-first, static source packet with these exact
eight paths; no path may be added, omitted, renamed, or substituted:

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

## Required lifecycle before source dispatch

1. Run the current-policy independent precommit triple: Contract/Spec/Oracle
   `gpt-5.6-terra`/`high`, Security/Adversarial `gpt-5.6-terra`/`xhigh`, and
   Lifecycle/Dependency/Boundary `gpt-5.6-terra`/`high`; record C/I/M and actual telemetry as
   `UNKNOWN` unless exposed.
2. Make one atomic commit containing exactly these eight authority documents: this receipt plus
   `.planning/STATE.md`, `DISPATCH-INDEX.md`, `DISPATCH-REMAINING.md`, `PLAN.md`,
   `MAIL-SPEC.md`, `docs/mail/QA-HANDOFF.md`, and `docs/mail/CONTRACT-CHANGELOG.md`.
3. Perform the independent immutable exact-SHA review. Its scope/diff check, stale-clause scan,
   and C/I/M result are candidates until recorded; no acceptance is implied by this receipt.
4. Only after that exact-SHA review accepts the authority amendment may root dispatch the exact
   02F-R source packet above. That packet requires its own test-first authoring, review, atomic
   commit, and exact-SHA acceptance.

## Dependency and blocker preservation

02B becomes eligible only after exact 02F-R static acceptance; 02C becomes eligible only after
02B. 02F-D (rule deployment) is separate and nonblocking. 02E, Ent generation, migration,
runtime, services, APIs, UI, release, Phase 142/148, launcher `NOT_READY`, and executable evidence
`NOT_RUN` remain unchanged.
