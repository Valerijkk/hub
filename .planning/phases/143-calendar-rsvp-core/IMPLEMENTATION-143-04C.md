# 143-04C — Calendar update/master-cancel immutable source receipt

> **Acceptance-sync gate:** `ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_EXACT12_COMMIT_REQUIRED /
> RUNTIME_NOT_RUN`. This dirty exact12 working tree grants no authority. The receipt below activates
> only after one atomic commit of exactly these 12 documents; until then prior 04C routing remains
> current. Activation selects no next Calendar packet and lifts no downstream blocker.

## Static acceptance (2026-08-20)

Calendar 143-04C is `STATIC_ACCEPTED / EXACT_SHA_PASS` only at immutable commit
`7fd061b5099b4608e0c33a6d99edbe1a66587d34`:

- parent: `bb84a47e12516c19682ff08783a5844306fc5766`
- tree: `0fb65fc72705bb145a1f2950c6efbe16607e6a04`
- subject: `feat(calendar): add update and cancel service`

The accepted manifest is exactly three test-first paths, in order:

1. `backend/tests/calendar/event_service_update_test.go` — `4421204141719fa1af586b3d6272aa130f185af2`
2. `backend/internal/calendar/event_repository.go` — `737207280cc23adb873ae8d36952a3e62a0a3a4d`
3. `backend/internal/calendar/event_service_update.go` — `094b68dd8369bdc9345d9348016b3bdc377aa7a9`

Contract/Spec/Oracle, Security and Lifecycle precommit plus immutable exact-SHA reviews PASS
C/I/M `0/0/0`. Author provenance: Terra/xhigh, `dispatcher-enforced`, no fallback; telemetry
`UNKNOWN` unless exposed.

This receipt accepts only the frozen optional update transaction ABI, update/master-cancel
behavior and authored-first static oracles in that exact3 manifest: capability-before-backend
guarding, receipt-free update scope, nondisclosing authorization order, copy-safe records,
merged recurrence identity, deterministic safe update/cancel notification intents, atomic
all-and-only persistence echo, attendee tombstones/re-invite behavior, and no raw iTIP.

Every preceding 143-04C candidate, `ROOT_SOURCE_AUTHORITY_CANDIDATE`, precommit,
source-unaccepted/uncommitted, or current-next clause is `CONSUMED/HISTORICAL` only for this
receipt. No next Calendar code packet is selected or authorized; root must separately select any
later packet. This does not authorize an adapter, Ent generation/migration, API/UI, transport
delivery, product/phase/release claims, raw iTIP, or runtime. 143-03E remains blocked for Ent
generation/migration; launcher `NOT_READY`, and all executable/runtime evidence remains `NOT_RUN`.
