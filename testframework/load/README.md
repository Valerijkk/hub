# Load lifecycle boundary

## Belongs here

The separately measured Locust lifecycle, workload definitions, load profiles,
and load-specific result interpretation belong here.

## Forbidden

Treating load as a pytest layer, mixing functional assertions into workloads,
unbounded execution, and publishing measurements without exact provenance are
forbidden.

## First consumer

The first accepted consumer must define a bounded workload contract and its
framework selftests before adding Locust behavior. This slice runs no load
lifecycle.
