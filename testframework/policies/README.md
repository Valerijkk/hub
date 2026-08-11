# Policy authority boundary

## Belongs here

Marker, environment, redaction, and execution-profile policy authorities
belong here as explicit, reviewable data contracts.

## Forbidden

Secrets, mutable local overrides, launcher implementations, implicit defaults,
and policy copies embedded in individual suites are forbidden.

## First consumer

The first accepted consumer must define one narrow policy contract and its
framework selftests before any runtime code reads it.
