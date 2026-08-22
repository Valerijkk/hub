# 144-05D — plain-content normalization source receipt

> **Receipt lifecycle:** `SOURCE_RECEIPT_SYNC_CANDIDATE / ATOMIC_EXACT9_COMMIT_REQUIRED / IMMUTABLE_REVIEW_REQUIRED / RUNTIME_NOT_RUN`. This dirty docs-only exact9 receipt is non-self-activating. Until one atomic manifest-equal exact9 receipt commit and its immutable Contract/Security/Lifecycle receipt triple PASS, the immutable source below is not `STATIC_ACCEPTED` and no former 05D authority/source-unimplemented route is historical.

## Immutable source

- Commit: `acc593fe4f20669a21a6f5ba70a8a20293111017`
- Sole parent: `98ebedc9dd0b6ad90dbe06e1faf5810376dce392`
- Tree: `7de89d1237bc0c8f640729f607a2dbfd9189f3c4`
- Subject: `feat(mail): add plain content normalization`
- Authored-first test: `backend/tests/mail/mailbox_content_test.go` —
  `37c29c421d21e7c173142647d894de0d890894de`
- Source: `backend/internal/mail/mailbox_content.go` —
  `b064a2f6f944d5ffb10651ccb068a77634a9267b`

The immutable commit changes exactly this ordered exact2, test first then source. Source
Contract/Security/Lifecycle precommit and immutable `EXACT_SHA_PASS` reviews are C/I/M `0/0/0`.
Provenance: Contract Terra/high, Security Terra/xhigh, Lifecycle Terra/high,
dispatcher-enforced/no fallback/telemetry `UNKNOWN`.

## Ordered exact9 receipt-sync manifest

The atomic receipt-sync commit diff MUST set-equal these nine paths, in this order, with no
addition, omission, rename, source, schema, OpenAPI, persistence, `go.mod`, testframework, Ent,
API/UI, launcher, or runtime file:

1. `.planning/STATE.md`
2. `.planning/decisions/specs/MAIL-SPEC.md`
3. `.planning/phases/144-mailbox-domain/PLAN.md`
4. `.planning/phases/144-mailbox-domain/DISPATCH-INDEX.md`
5. `.planning/phases/144-mailbox-domain/DISPATCH-REMAINING.md`
6. `.planning/phases/144-mailbox-domain/IMPLEMENTATION-144-05D.md`
7. `.planning/phases/144-mailbox-domain/REVIEW-144-05D-CONTENT-READINESS.md`
8. `docs/mail/CONTRACT-CHANGELOG.md`
9. `docs/mail/QA-HANDOFF.md`

Only after that atomic exact9 commit and immutable receipt triple PASS is this source
`STATIC_ACCEPTED / EXACT_SHA_PASS`; then former 05D authority, acceptance, source-unimplemented
and current-next routes are `CONSUMED/HISTORICAL`. This receipt selects no next Mail packet.

## Retained boundaries

Acceptance is static only: runtime remains `NOT_RUN`. The source remains plain-only and grants no
rich sanitizer, token/proxy, remote-image, 05D-Remote, 05E, schema/OpenAPI/persistence/Ent/API/UI,
testframework, launcher, or runtime authority. Existing dependencies and blockers remain unchanged.
