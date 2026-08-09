# Traceability boundary

## Belongs here

Loading validated `TC-*` metadata and schemas and constructing the explicit
risk-linked `task -> requirement -> TC/case` graph belong here.

## Forbidden

Test execution, SUT access, result fabrication, suite-local metadata parsers,
and case records missing task, requirement, or risk linkage are forbidden.

## First consumer

The first accepted consumer must add a single schema-backed loader test-first,
preserve closed-field validation, and prove the complete
`task -> requirement -> TC/case` chain with risk linkage. This boundary
currently performs no loading or execution.
