# Task 143-03C — content review and provenance blocker

> **Status:** P1_REMEDIATION_REQUIRED · MODEL_PROVENANCE_BLOCKED · NOT_STATIC_ACCEPTED
> **Implementation commit:** `1288d5dfacd6ee93d78a07da765d6ca3d75cb535`
> **Implementation dependency:** 143-03A `c6f5ec8b54b7e35327125dd4445634c5d1764491` · later reopened

## Scope and result

- Authoring packet: `/root/mail_143_03c_impl`.
- Requested authoring route: exact `gpt-5.6-terra` / `xhigh`; no fallback.
- Sole owned file: `backend/ent/schema/calendareventexception.go`.
- Source blob: `100f9d1e3c6448996f5ec0d02be68be64100375c` (file SHA-256
  `bf263669f894dddc160925bf1a664197f94e7b3bcb5aef0c226a3b0d9c5971dd`).

The first source-only review found no P0, P1 or P2: the occurrence override schema has the
unique event plus recurrence identity, typed timed/all-day override fields,
timezone/location/status overrides and the accepted no-cascade event edge. A later root
deep audit found the same P1 as 143-03A: nullable canonical hash version/hash fields lack
the required schema-level pair check. The frozen fix is in
`REMEDIATION-143-03-HASH-PAIR.md`.

## Why acceptance is blocked

The review runtime did not expose its actual model/effort. A second independent provenance
attempt explicitly requested `gpt-5.6-terra` / `high`, reported actual `UNAVAILABLE`, used no
fallback and stopped before reading files. The founder routing lock says unavailable runtime
provenance is `UNKNOWN/BLOCKED`. Therefore content quality is recorded, but packet 143-03C
is not statically accepted and dependent 143-03D must not start.

Ent generation, schema compilation, migrations and executable tests remain `NOT_RUN`; the
accepted launcher is `NOT_READY`.
