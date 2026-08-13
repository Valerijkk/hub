# Phase 147 — Mail 0.1.7 canonical-hash static review

> **Current use:** HISTORICAL_CVE_V1_PASS_ONLY. A later root deep audit reopened a distinct
> P1 in engine message/blob/admin byte bounds. This review still supports the unchanged
> calendar canonical-identity semantics, but no longer proves global Mail contract readiness.
> See Phase 144 `REVIEW-144-01A-ENGINE-BOUNDS.md`.

> **Reviewed semantic contract SHA:** `b3fb6f483b1335c4ef9ec3efbb0ff49c6d94ac3d`
> **Mail 0.1.7 amendment base:** `e2f42b9b469b080291c12bacbc1622c047a401d3`
> **Initial 0.1.7 snapshot:** `7a6ffc1c340e8b701e37388a871035976620b915`
> **Remediation:** `b3fb6f483b1335c4ef9ec3efbb0ff49c6d94ac3d`
> **Branch:** `codex/mail-module-20260813`
> **Root verdict:** `ROOT_STATIC_PASS`
> **Independent verdict:** `TERRA_HIGH_STATIC_PASS`
> **G148-02 state:** historical `PASS` for this exact reviewed SHA; current gate `BLOCKED`
> **Phase 143 source:** 143-01 and 143-02A `STATIC_ACCEPTED`; 143-02B `NOT_IMPLEMENTED`
> **Runtime/build/test/schema-validator evidence:** `NOT_RUN`

This evidence freezes only static Mail 0.1.7 contract coherence. It does not claim parser,
schema, service, handler, migration, executable QA, merge or production readiness. Any
semantic contract change makes this evidence `STALE` and reopens G148-02.

## Scope and method

Mail 0.1.7 separates inbound canonical duplicate identity from the accepted 143-02A
outbound serializer. Root reviewed the complete amendment and remediation across ADR,
Mail spec, parser dispatch, Phase 143/144/146/147/148 routing, QA handoff, changelog and the
OpenAPI version marker. An independent read-only Terra/high reviewer then inspected the
exact clean remediation SHA.

Read-only Git/text evidence on the exact SHA:

- `git diff --check e2f42b9b..b3fb6f48`: PASS;
- the amendment contains 26 changed paths, all under Mail planning/docs ownership;
- protected-zone matches: zero; backend/frontend/CLI/deployment source matches: zero;
- added-line secret and forbidden AI/coauthor-trailer scan: zero matches;
- Mail spec blob: `d9e7a070efa434ff7e976ff0af9f53ea1c474889`;
- ADR blob: `09fff1cd829d9f5808937045984fd7da532e631e`;
- OpenAPI blob: `df0af182f381d894b616d9d287698453e74849f7`;
- event schema stayed `dd98b5587ef38dddf20ba61d895464e001fb3e10`;
- QA handoff blob: `efa266fffee2d04a2c3888e7e0ec92ef36bc102e`;
- Phase 147 contract blob: `69984a2114292cd130f9774ed0c82bd8d5b7e17f`;
- 143-02B dispatch blob: `c4d57296e0e35b31857645aebb1b8e743dbc4f8c`.

The OpenAPI diff from the amendment base changes only `info.version` from 0.1.6 to 0.1.7;
no HTTP path, shape, code, example or component changed. The already reviewed 0.1.6
machine contract is therefore wire-identical for this amendment, and its event schema is
byte-identical.

Project policy forbids host tests, builds, linters and validators. Docker Desktop is
available, but the accepted versioned launcher remains `NOT_READY`; no direct Docker/
Compose fallback was used and executable evidence remains `NOT_RUN`.

## Review finding closure

The first independent review of `7a6ffc1c` found three P1 ambiguities and no P0/P2. The
remediation commit `b3fb6f48` closes all three:

| ID | Finding | Resolution |
|---|---|---|
| R147-017-01 | Invalid RFC TEXT/RFC 6868 escapes and decoded controls could be preserved, normalized or rejected differently | wire decoding now has exact allowlists, terminal/unknown escapes fail closed, post-decode `Cc` handling is explicit, and `CanonicalITIPIdentity` never decodes already-domain strings a second time |
| R147-017-02 | decoded field maxima could be measured as runes or bytes | every decoded maximum is a post-normalization UTF-8 byte count before CVE framing; the parser dispatch requires multibyte boundary oracles |
| R147-017-03 | migration and Phase 147 review status still described the old contract as current | the semantic remediation made 0.1.6 historical `STALE` and routed every current status through exact-SHA re-review; this evidence-only follow-up records the resulting `b3fb6f48` PASS across Phase 143–148/STATE/ROADMAP/QA |

## Accepted static contract

- CVE-v1 begins with the fixed domain prefix, NUL and version byte 1, then streams the
  exact tagged, length-delimited, big-endian frames into SHA-256.
- Every enum, integer, nested string, attendee/EXDATE record and sort order is fixed; JSON,
  map iteration, locale, random values, clocks, MIME and physical folding are excluded.
- The bounded wire grammar, singleton cardinality, accepted parameters, temporal union,
  RRULE subset, exact limits and stable safe sentinels are closed.
- Occurrence REQUEST is accepted inbound without changing the outbound serializer.
- All-day VALUE=DATE uses the semantic `UTC` marker with start/exclusive-end dates and no
  midnight inference; founder question 27 preserves the only future policy fork.
- Event, exception, object-state and inbound-REPLY hashes store canonical version; hashes
  from different versions are never equal. Sender-bound REPLY receipts use a separate
  domain prefix and length-delimited normalized sender.
- A test-owned literal CVE byte vector is required, but no ad-hoc host-calculated digest is
  accepted as executable evidence.

## Independent re-review

The exact clean `b3fb6f48` snapshot received read-only `gpt-5.6-terra` / `high` re-review,
with no fallback. Verdict: PASS, no P0/P1/P2. The reviewer explicitly rechecked all three
remediations, one-time decode, protected paths, wire grammar, cryptographic framing,
occurrence REQUEST, all-day UTC semantics, versioned persistence, sender binding, OpenAPI
non-drift and honest `NOT_RUN`/blocked status. It ran no test, build, validator, generator,
Docker/Compose, network command or write.

## Remaining gates

G148-02 is closed only for `b3fb6f48`. Parser implementation/review, schema/migration
authority, accepted launcher, generation, executable QA, engine/legal/broadcast decisions,
final protected-rebase signal and push/PR authority remain pending, blocked, unauthorized
or `NOT_RUN` in `GATE-MATRIX.md`.

The only supported contract claim is:
`Mail 0.1.7 static contract coherence PASS on b3fb6f48`.
