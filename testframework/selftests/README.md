# Framework selftest boundary

## Belongs here

Contracts for framework configuration, runtime, schemas, policies, adapters,
and other reusable framework behavior belong here.

## Forbidden

SUT behavior, product acceptance scenarios, production credentials, and tests
that require a deployed Hub instance are forbidden.

## First consumer

Every first framework consumer must place a failing contract here before its
implementation. This structure-only slice adds no selftest or runtime claim.
