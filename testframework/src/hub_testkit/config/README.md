# Configuration boundary

## Belongs here

Immutable, validated composition of environment inputs and framework
configuration belongs here. Consumers receive explicit configuration objects
rather than reading ambient state throughout the framework.

## Forbidden

Secret values in artifacts, mutable process-wide configuration, unvalidated
environment access, and domain-specific settings are forbidden.

## First consumer

The first accepted consumer must define one narrow configuration contract and
its framework selftests before adding its implementation. Until then this
directory is documentation-only.
