# 144-04I-A — bounded mailbox-sync worker source receipt

> **Current status:** `STATIC_ACCEPTED / EXACT_SHA_PASS` for only immutable source `091f2ef286a71f506e1b048f5386fd37cefa06bc`; exact10 receipt `03674eace728642fc2c9b39d2e8e2c81bf016033` is ancestral to HEAD. The conditional receipt status below is `CONSUMED/HISTORICAL` only for 04I-A. Contract/Security/Lifecycle PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback, telemetry `UNKNOWN`. No next packet is selected; Ent/concrete adapter, generation/migration, API/UI, launcher and runtime remain blocked or `NOT_RUN`.
>
> **Status:** `SOURCE_RECEIPT_SYNC_CANDIDATE / ATOMIC_EXACT10_COMMIT_REQUIRED /
> RUNTIME_NOT_RUN`. This dirty exact10 documentation candidate grants no acceptance. Only one atomic
> commit whose diff equals this manifest, followed by immutable receipt review, may mark the immutable
> source below `STATIC_ACCEPTED / EXACT_SHA_PASS`.

## Immutable source receipt

- Commit: `091f2ef286a71f506e1b048f5386fd37cefa06bc`
- Parent: `6b27643fcde3b702fa13c5fcc9b4bfefcb9fc87a`
- Tree: `b01be92acf01033935b85446454eaee16b84e01b`
- Subject: `feat(mail): add bounded sync worker`
- Authored-first test: `backend/tests/mail/mailbox_sync_worker_test.go` —
  `79c2691971287e3a62c830a5d12c78699414bb86`
- Source: `backend/internal/mail/mailbox_sync_worker.go` —
  `6c20fadf563d4d146b2fc1cbacdddea485266eb7`

The immutable commit changes exactly those two paths, in test-first order. Contract/Spec/Oracle,
Security/Adversarial and Lifecycle/Dependency/Boundary precommit reviews PASS C/I/M `0/0/0`; the
immutable Contract review and combined Security/Lifecycle exact-SHA reviews PASS C/I/M `0/0/0`.
Author Terra/xhigh; Contract Terra/high; Security/Lifecycle Terra/xhigh;
dispatcher-enforced/no fallback; telemetry `UNKNOWN`.

## Ordered exact10 receipt-sync manifest

The atomic receipt-sync commit must change exactly these paths, in this order, with no substitution
or extra path:

1. `.planning/STATE.md`
2. `.planning/decisions/specs/MAIL-SPEC.md`
3. `.planning/phases/144-mailbox-domain/PLAN.md`
4. `.planning/phases/144-mailbox-domain/DISPATCH-INDEX.md`
5. `.planning/phases/144-mailbox-domain/DISPATCH-REMAINING.md`
6. `.planning/phases/144-mailbox-domain/IMPLEMENTATION-144-04I-A.md`
7. `.planning/phases/144-mailbox-domain/REVIEW-144-04I-A-SYNC-WORKER-READINESS.md`
8. `docs/mail/CONTRACT-CHANGELOG.md`
9. `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md`
10. `docs/mail/QA-HANDOFF.md`

After that atomic exact10 commit and its immutable receipt review, only this source becomes
`STATIC_ACCEPTED / EXACT_SHA_PASS`; former 04I-A authority-candidate, precommit,
source-unaccepted and current-next clauses are `CONSUMED/HISTORICAL`. This receipt selects no next
source packet.

## Boundary retained

144-04F and 144-04G are accepted static prerequisites. 144-04H and 144-02E, Ent/concrete adapter,
generation, migration, SQL/two-replica DB-time proof, API/OpenAPI/UI, composition/wiring, launcher
readiness and runtime remain blocked, `NOT_READY` or `NOT_RUN`. The receipt adds no schema, event,
delivery, direct engine, adapter, raw body or runtime claim.
