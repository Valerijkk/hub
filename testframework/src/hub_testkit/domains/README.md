# Domain boundary

## Belongs here

Each `domains/<domain>/` subtree owns that domain's models, facades, builders,
and oracles. If browser abstractions are required, pages and components live
only below that domain in `ui/pages/` and `ui/components/`.

## Forbidden

Cross-domain helper warehouses, suite-local reusable frameworks, generic page
object dumps, and imports from quarantined legacy content are forbidden.

## First consumer

The first accepted domain consumer must create one narrow domain module and
its framework selftests before implementation, without creating a shared
catch-all layer.
