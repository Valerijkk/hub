# Exceptions and rationale

Exception — bounded engineering decision, not silent bypass.

## Required record

```yaml
control_id: Cxx | NEG-xx | warning-name
scope: exact file/module/endpoint/pipeline
owner: accountable team/person
rationale: why the rule does not fit this scope
failure_model: what can still go wrong
evidence: test, measurement, design note, or source
compensating_controls:
  - concrete protection
expiry: YYYY-MM-DD | permanent-with-review-date
removal_condition: measurable trigger
reviewed_by: role/name
```

## Acceptable patterns

- Cohesive parser exceeds the function review budget; complexity/path tests and benchmark show extraction would harm locality.
- External API cannot support idempotency keys; reconciliation record and operator recovery are documented/tested.
- Third-party tool cannot be pinned by SHA; artifact is mirrored, verified by digest and owned update process.
- Mutable process cache is required; ownership, synchronization, reset/lifetime and concurrency tests are explicit.

## Unacceptable patterns

- «Legacy» without bounded scope and removal condition.
- «False positive» without detector evidence.
- «Too hard» without compensating control.
- Permanent security suppression with no owner/review date.
- Split helpers or wrapper interfaces created only to silence a metric.

## Lifecycle

1. Record before merge/release when possible.
2. Keep exception adjacent to the enforcing system or linked from it.
3. Fail visibly when expiry passes.
4. Re-evaluate when source, architecture or failure model changes.
5. Delete suppression and record when removal condition is met.

Evidence class is normally `POLICY`; an exception never downgrades an external `NORM`, it documents how risk is handled locally.
