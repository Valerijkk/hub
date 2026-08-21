# Autonomous Mail + Calendar questions — 2026-08-21

> **143-04J source-receipt sync:** `SOURCE_ACCEPTANCE_SYNC_CANDIDATE /
> ATOMIC_EXACT14_COMMIT_REQUIRED / RUNTIME_NOT_RUN`. These proposed receipt facts grant no
> authority before one atomic commit of the exact14 documentation manifest. After that commit,
> immutable exact2 `e74efa827b52f35e3d3fbe8346adf0c0121ba48d` is
> `STATIC_ACCEPTED / EXACT_SHA_PASS`; lower 04J authority/source-unaccepted routes become
> `CONSUMED/HISTORICAL`. No next Calendar packet is selected. AQ-02 remains only a legacy
> migration/backfill question and does not alter the pure accepted source boundary.

> **143-04J acceptance-sync gate:** `ACCEPTANCE_SYNC_CANDIDATE /
> ATOMIC_EXACT13_COMMIT_REQUIRED / SOURCE_IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN`.
> These proposed receipt facts have no authority before one atomic commit of the same exact13
> documentation manifest. After that commit, immutable candidate
> `5841621720cc03682b3d1f1e2af9cba2953797cf` is `SOURCE_AUTHORITY_ACCEPTED /
> EXACT_SHA_PASS`; its Contract/Spec/Oracle, Security and Lifecycle immutable reviews are PASS
> C/I/M `0/0/0`, and lower 04J candidate/precommit/exact-review-next clauses become
> `CONSUMED/HISTORICAL`. Only the ordered test-first exact2 may then be dispatched; its source is
> still unimplemented and unaccepted. AQ-02 remains migration/backfill-only, and every 04K/04F,
> Ent, adapter, API/UI, launcher and runtime boundary remains unchanged.

These questions require founder authority and are intentionally deferred while autonomous work
continues on unblocked pure-domain packets. They do not authorize source, migration, launcher, or
runtime work by themselves.

## 143-04K pure call-due source receipt-sync — conditional

> **Source receipt-sync gate:** `SOURCE_ACCEPTANCE_SYNC_CANDIDATE /
> ATOMIC_EXACT15_COMMIT_REQUIRED / RUNTIME_NOT_RUN`. This dirty exact15 documentation candidate
> grants no authority. Only after one atomic exact15 receipt-sync commit and immutable acceptance is
> source `572f911227b9d50e21aa73b87c530ddcaf0f8934` (parent `26b377eb`, tree `ac96e2f5`, subject
> `feat(calendar): add call due policy`) `STATIC_ACCEPTED / EXACT_SHA_PASS` for authored-first test
> blob `437db3af0263cedb10dcf217cd53fbbcac03cf23` then source blob
> `376f46c1cc79956fab3479efc38750fd88161b16`. Contract/Spec/Oracle, Security/Adversarial and
> Lifecycle/Dependency/Boundary precommit plus immutable reviews are PASS C/I/M `0/0/0`;
> Terra/xhigh, dispatcher-enforced/no fallback, telemetry `UNKNOWN`. The dirty candidate is not an
> acceptance and selects no next packet.

`REVIEW-143-04K-CALL-DUE-AUTHORITY.md` is the accepted source-authority predecessor. Its lower
authority-candidate/precommit/test-first dispatch text is `CONSUMED/HISTORICAL` only after this
source receipt activates. Source `572f9112` consumes accepted 04J (`d7926a27` receipt for
`e74efa827`), 04A's `EventOutboxDescriptor` and 04L's `EnvelopeV1`; it authorizes neither a
Calendar-subject admission from 04G nor persistence.

The future policy takes explicit UTC whole-second `Now`, call, raw 04J projection and envelope
values, returns an optional descriptor, derives a domain-separated occurrence key and validates the
existing plural nine-field due envelope. Active input is deep-cloned and produces exactly two fixed
04J reconcile calls; inactive/deleted input requires nil projection/zero envelope and immediately
noops. It has
no I/O or clock. AQ-02 remains legacy recurring-row backfill-only. CAS, event→call locking,
rollback and same-transaction descriptor append remain only for future 04F-K. EventID/OccurredAt/
CorrelationID are fresh active candidates and non-due outcomes discard them; only a due outcome may
be reserved/reloaded by 04F-K and rerun byte-identically on retry/redelivery. Ent/generation/
migration, adapters/API/UI/raw iTIP, legacy subject cutover, launcher and runtime remain blocked or
`NOT_RUN`.

## Pending founder questions

### AQ-01 — Ent migration and launcher authority

Choose the canonical production migration mechanism and task-owned Docker launcher authority for
Calendar `143-03E` and Mail `144-02E`. Until this is decided and accepted, Ent generation,
migrations, concrete Ent adapters, and the final integrated Docker acceptance remain
`BLOCKED / NOT_READY / NOT_RUN`.

### AQ-02 — existing recurring ScheduledCall interpretation

Confirm the canonical timezone/recurrence interpretation for already persisted recurring
`ScheduledCall` rows (founder question 18), including DST gap/fold handling and backfill behavior.
This blocks production backfill and concrete persistence integration only. It does not block the
pure `143-04J` ScheduledCall projection policy, which will receive normalized caller-owned inputs.

## Root decisions made autonomously

- The exact13 `143-04J` docs-only candidate freezes a pure caller-owned projection ABI,
  including event/target/creator identity, series-only participant projection, canonical timed
  recurrence boundary and opaque snapshot result. It is not authority: only its atomic exact13
  commit, Contract/Spec/Oracle + Security + Lifecycle immutable exact-SHA PASS, and separate
  docs-only acceptance can authorize test-first exact2 `event_call_policy_test.go` then
  `event_call_policy.go`. AQ-02 remains migration/backfill-only.
- Do not implement Calendar `143-04D` in its current two-file form: a separate optional RSVP
  transaction-backend ABI and expanded exact manifest are required to preserve authorization,
  receipt replay, atomic attendee mutation, and outbox boundaries.
- Do not start Mail `144-04H` or Calendar `143-04F-*` before the accepted `02E` / `03E` generation
  gates; no handwritten generated Ent or migration fallback is allowed.

## Founder resolution — 2026-08-21

The founder instructed root to choose the most maintainable architecture and continue autonomously
to Docker readiness without merging. This resolves AQ-01 and AQ-02 for branch implementation;
shared/production migration apply and final merge retain their normal personal review gates.

### AQ-01 resolved

Use one closed public `qa-gate` entrypoint backed by the existing private supervisor. The caller
supplies only `target`, `task`, and `run`; committed internal authority selects contract/catalog
versions. It derives worktree-local canonical roots, materializes an exact commit into offline
image builds, and exposes only fixed versioned targets. Ent generation is image-local. An admitted,
attested exact Atlas CLI owns a bounded feature migration lineage: a versioned full pre-feature
baseline fixture followed by immutable Mail/Calendar SQL and an allowlisted Ent parity comparison.
Python framework/closed vertical targets are accepted before persistence/API packets need them.
No direct Docker/Compose fallback exists.

### AQ-02 resolved

Do not guess an original timezone or `DTSTART`. Non-recurring coherent rows may be migrated. Any
non-deleted recurring legacy row aborts automatic backfill with a read-only report and zero
mutation; an operator must recreate or explicitly map it through a later per-row reviewed artifact.
Docker migration fixtures must prove both the successful non-recurring path and this fail-closed
recurring path. Pure `143-04J` remains unchanged.

Controlling design and execution plan:

- `docs/superpowers/specs/2026-08-21-mail-calendar-docker-readiness-design.md`
- `docs/superpowers/plans/2026-08-21-mail-calendar-docker-readiness.md`

These documents authorize neither merge/push/deploy nor shared/production migration.
