# Inbound published cleanup execution

Status `DOCS_FROZEN / RUNTIME_NOT_RUN`; no merge/push. Hard prerequisites:
O1 generation PASS (`98e260...`), accepted/applied O2-O5, and frozen
`12-INBOUND-PUBLISHED-EXTERNAL-CLOSURE-MANIFEST.md`, sole authority for 82
ordered slots, 4767 bytes, SHA-256
`889c14127f47bdb73108bfef07573ff4b0bde7fd3b78b47122718a1f5b0a36b7`.
No dynamic catalog. Code packets are Terra/xhigh, review Terra/high, docs sync;
private types stay unexported; no source-adjacent/source-grep tests or fake Ent.

## Current source-acceptance receipt

`STATIC_ACCEPTED` source receipt for the inbound published-cleanup chain:
`f6a0b681` establishes deterministic published-fixture cleanup lifecycle and
`c171158e` dispatches it by lifecycle state (C9e). C8e's supervised QA child
protocol and recovery wiring are recorded by `18ef33e3`, `b010d39a`,
`ad1f1231`, `b94da019`, `772b11ec`, `93d68a84`, `ed630813` and `8ef3985c`:
the private recovery route is token-authenticated, available only in
`qa && !release`, preserves unsupervised wiring, uses the `qa-backend` /
`qa-runtime` supervisor selection, and hardens child supervision, bounded
signal shutdown and child-log preservation. This is source-static acceptance
only; it neither exposes a public API nor claims compile, Docker, integration
or product runtime evidence.

## Authority split

Atlas desired-state and full migration replay derive only canonical three-field
`root|table|column` projections. The frozen six-field
`root|table|column|mode|source|count` manifest is independent policy: grammar,
order, 82 lines, 4767 bytes and full SHA are validated separately. The schema
projection must equal the policy projection. No runtime catalog, Ent reflection
or source grep is authority.

## Atomic packets

- **C0** `controller_delete_published_types.go`: pure comparable V2 identity,
  legal states and cleanupProof presence plus six copied Calendar UUIDs; no I/O,
  pointers, rows, envelopes or physical Outbox IDs.
- **C1a** `controller_delete_published_roots_mail.go` (C0): caller-owned
  SERIALIZABLE transaction; exact-ID locks and immutable relation validation
  for primary MailMessageFolder, MailMessage, MailThread and MailFolder.
- **C1b** `controller_delete_published_roots_identity.go` (C1a): same caller
  transaction; exact-ID locks and immutable relation validation for
  MailIdentity, MailAccount and MailDomain.
- **C1c** `controller_delete_published_roots_owner.go` (C1b): same caller
  transaction; exact-ID locks for User and Organization and aggregate
  nine-root/primary-folder invariant; no Calendar or external-slot reads.
- **C2** `..._mail_closure.go` (C1c): exactly one inbound message/folder, zero
  drafts/submissions and false-sent/closure guards.
- **C3a** `..._calendar_receipt.go` (C2): one applied receipt/object state.
  **C3b** `..._calendar_event.go` (C3a): one constrained imported event.
  **C3c** `..._calendar_attendee.go` (C3b): one owner attendee accepted RSVP v2.
  **C3d** `..._calendar_effects.go` (C3c): one pending reply delivery and one
  completed `rsvp_series` mutation receipt. Earlier DELETE is 409.
- **C3e-1** `backend/docker/schema-maintenance/rootfk/frozen-external-closure.manifest`
  (C3d): literal 82-line six-field machine asset, byte-for-byte replica of the
  sole normative policy authority
  `12-INBOUND-PUBLISHED-EXTERNAL-CLOSURE-MANIFEST.md`; it carries the same
  expected metadata and fixed shards. No generator or machine asset owns the
  policy.
- **C3e-2** `backend/docker/schema-maintenance/rootfk/main.go` (C3e-1): stdlib
  helper with bounded Atlas-inspect SQL parser. It accepts only recognized direct
  root-FK DDL, rejects ambiguous/unrecognized input, emits sorted three-field
  projections, validates six-field grammar/order/count/bytes/full SHA and checks
  projection equality.
- **C3e-3** accepted genuine isolated backend test-root helper tests (C3e-2):
  desired/replay equality, parser rejection and exact asset metadata/projection
  drift; no Docker/generated-source test.
- **C3e-4** `backend/docker/Dockerfile.schema-maintenance` (C3e-2): build/copy
  helper and frozen asset into schema-maintenance image.
- **C3e-5** `backend/docker/entrypoints/schema-maintenance.sh` (C3e-4): every
  disposable generate/hash/verify independently creates desired and replay
  inspect artifacts and invokes helper; no shared tmpfs/run-order. Hash prints
  full expected metadata and actual desired/replay projected counts/bytes/SHA;
  verify recomputes independently and exits non-zero on drift.
- **C3e-6** `docker-compose.local.yml` (C3e-5): generate/hash use the
  schema-maintenance tool/internal DB; no product/QA ownership change.
- **C4a** `controller_delete_published_external_manifest.go` (C3e-1): literal
  canonical all-82-slot six-field `root|table|column|mode|source|count` block,
  including metadata, as a byte-for-byte replica of the sole normative policy
  authority `12-INBOUND-PUBLISHED-EXTERNAL-CLOSURE-MANIFEST.md`; no shard reads
  it and the code literal is never a policy authority.
- **C4b** `controller_delete_published_external_policy.go` (C4a): validated
  observation -> allowlisted/forbidden decision; no Ent shard.
- **C4c** `controller_delete_published_external_core.go` (C4b): shared typed
  exact-set/ForUpdate primitive; no slot enumeration.
- **C4d** `controller_delete_published_dependents_organization.go` (C4c): sole
  reader of slots 01-14.
- **C4e** `controller_delete_published_dependents_user_a.go` (C4c): sole reader
  of slots 15-27.
- **C4f** `controller_delete_published_dependents_user_b.go` (C4c): sole reader
  of slots 28-41.
- **C4g** `controller_delete_published_dependents_user_c.go` (C4c): sole reader
  of slots 42-55.
- **C4h** `controller_delete_published_dependents_user_d.go` (C4c): sole reader
  of slots 56-69.
- **C4i** `controller_delete_published_dependents_user_e.go` (C4c): sole reader
  of slots 70-82.
- **C4j** `controller_delete_published_external.go` (C4d-C4i): compose six
  shard observations/fail closed; no shard read.
- **C4k** `backend/tests/calendar/published_external_manifest_test.go`
  (C4a/C3e-1): genuine isolated hermetic drift test that emits C4a's canonical
  six-field code block and compares it byte-for-byte with the exact machine
  asset `backend/docker/schema-maintenance/rootfk/frozen-external-closure.manifest`,
  then independently verifies embedded expected constants line count `82`, byte
  count `4767` and SHA-256
  `889c14127f47bdb73108bfef07573ff4b0bde7fd3b78b47122718a1f5b0a36b7`.
  It must not read `.planning`, source-grep, use Ent reflection, Docker or
  generated source. Document-to-asset equivalence is established only by
  planning/static review and the schema-maintenance gate, never by this test.
- **C5** `controller_delete_published_calendar_delete.go` (C4j): exact-ID Calendar delete primitives.
- **C6** `controller_delete_published_outbox.go` (O2-O5/C4j): cardinality-one exact tuples P3
  mail_message/message/mail.received.v1, P4 calendar_event/event/
  calendar.invite.updated.v1, P5 calendar_attendee/attendee/
  calendar.rsvp.changed.v1; DecodeEnvelopeV1 after selection only; transient
  physical IDs; no payload query.
- **C7** `controller_delete_published_sequence.go` (C5/C6): after same-tx reproof delete P3 -> P4 -> P5 then
  ReplyDelivery -> MutationReceipt -> InboundReceipt -> ObjectState -> Attendee -> Event ->
  messageFolder -> message -> thread -> folder -> identity -> account -> domain -> user -> organization.
- **C8a** `controller.go` (C0): state fields/mutex-only V2 lifecycle, no I/O.
- **C8b** `controller_published_lifecycle.go` (C8a): CAS states; failed final
  CAS is 409/no retry.
- **C8c** `controller_published_boot.go` (C8a): fresh-boot invariant/recovered
  empty maps; does not restart a process.
- **C8d** `controller_published_engine.go` (C8b): one post-commit engine-destroy
  dispatch and terminal engine-failure state.
- **C8e-1** `backend/docker/entrypoints/qa-backend-supervisor.sh` (C8c):
  supervisor-only exit protocol, existing qa-backend child restart and new
  child PID/generation assertion; never in-process reset.
- **C8e-2** existing QA-only Dockerfile target and `docker-compose.local.yml`
  qa-backend selection/entrypoint (C8e-1): attach supervisor only; production
  target/entrypoint unchanged.
- **C8e-3** `backend/cmd/server/wiring_calendar_mail_qa.go` (C8e-2): each child
  generates its opaque boot token; parent cannot preselect/reuse; recovery
  asserts pre/post token inequality.
- **C8e-4** qamailcalendar QA/release recovery route files (C8e-3): token-auth
  route only in qa&&!release; child suppresses Fake-mutating workers; release
  exposes no recovery capability.
- **C9a** `controller_delete_published_proof.go` (C1a-C1c/C2-C8): caller-owned
  full SERIALIZABLE proof assembly through commit.
- **C9b** `controller_delete_published_coordinator.go` (C9a): in one caller-
  owned SERIALIZABLE transaction, first assemble the complete full proof, then
  CAS Published -> CleanupClaimed, then reassemble and compare the complete
  full proof in the same transaction, and only then delete; no CAS occurs
  before the initial proof and no shard reads occur here.
- **C9c** `controller_delete_published_reconcile.go` (C9b): exact-present ->
  renewed full proof -> Published; exact-absent -> one engine step; partial 409.
- **C9d** `controller_delete_published_mapless.go` (C9c/C8e): mapless only in
  supervisor fresh child, maps empty/mutators suppressed/engine empty; partial
  and ambiguous shapes 409/no delete.
- **C9e** `controller_delete.go` (C9b-C9d): dispatch only; legacy untouched.
- **C10a** `backend/tests/platform/events/envelope_decode_test.go`: isolated decoder units.
- **C10b** Python delete client (C9e): public call/try-finally primary-error preservation.
- **C10e** `controller_faults_qa.go`/`controller_faults_release.go`
  (C9b/C8d/C9d): closed enum in qa&&!release; it may target only the exact
  lifecycle/coordinator states supplied by those packets.
- **C10f** `controller_fault_routes_qa.go`/`controller_fault_routes_release.go`
  (C10e):
  token-auth, no arbitrary IDs/data, one fault per nonce.
- **C10g** `mail_calendar_faults.py` (C10f): closed names/scoped setup-clear/no
  DB-NATS-raw ID access.
- **C10c** Python integration scenario (C10b/C8e-4/C10g): public and
  closed-QA evidence; it starts only after the public client, fresh-child
  recovery route and all three closed fault-capability packets exist.
- **C10d** TC-0002 traceability (C10c): requirement/evidence mapping after the
  completed integration scenario.

## Closed QA evidence capability

Only C10e-g is non-public: qa&&!release, child-local opaque boot token, absent
from release. Finite one-shot enum: `origin-p3-duplicate-v1`,
`origin-p4-malformed-envelope-v1`, `external-user-ai-provider-v1`,
`cleanup-commit-unknown-present-v1`, `cleanup-commit-unknown-absent-v1`,
`cleanup-serialization-rollback-v1`, `cleanup-engine-after-db-delete-failure-v1`,
`cleanup-final-cas-lost-v1`, `cleanup-mapless-partial-v1`,
`cleanup-mapless-ambiguous-v1`.

Serialization proves rollback then renewed full proof before sole Published
restore. Engine failure fires only after DB delete commits during/after one
engine-destroy dispatch and pins CleanupEngineFailed. Final-CAS loss fires only
after confirmed engine success and pins 409/no retry. Fixed mapless partial/
ambiguous shapes offer no arbitrary DB access and assert no deletion/engine destroy.

## Evidence matrix

Early DELETE 409/no mutation; happy 204; Cleaned retry 204. Fresh supervisor
mapless-present is DB-only 204 with new child PID/generation/unequal tokens.
Absent/partial/ambiguous mapless 409/no deletion. Duplicate/malformed origin
and external blocker 409/no delete. Commit-present renews proof/restores
Published; absent one engine step. Serialization rollback then renewed proof.
Committed-delete engine failure terminal 409. Confirmed-engine final-CAS loss
409/no DB or engine retry.

## C11 future serial Docker gates (`NOT_RUN`)

Run nonparallel after independent acceptance/docs sync; none has been run:

1. `docker compose -f docker-compose.local.yml run --rm schema-maintenance-generate`
2. `docker compose -f docker-compose.local.yml run --rm schema-maintenance-hash`
3. `docker compose -f docker-compose.local.yml run --rm schema-maintenance-verify`
4. `docker compose -f docker-compose.local.yml run --rm backend-unit`
5. `docker compose -f docker-compose.local.yml run --rm python-contract`
6. `docker compose -f docker-compose.local.yml run --rm python-integration`

Record gates independently; static evidence is not runtime acceptance.
