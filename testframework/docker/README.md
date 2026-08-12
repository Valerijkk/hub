# Container contract boundary

## Belongs here

Immutable pinned images, isolated topologies, and resource, isolation, cleanup,
and other container contracts consumed by the future versioned launcher belong
here. They are execution assets, never the launcher API.

## Forbidden

Launcher code or entrypoints, unpinned images, direct or alternate launchers,
host execution, implicit fallbacks, and executable command snippets are
forbidden. Launcher implementation belongs only in `../runner/**`; no
executable fallback exists today.

## First consumer

The first accepted consumer must specify and selftest one pinned, bounded
container contract before the reviewed launcher consumes it. Package/launcher
status is **NOT_READY** and runtime is **NOT_RUN**. This directory currently
provides no launcher implementation or runtime authority; the docs-only
package boundary is [`../PACKAGE-CONTRACT.md`](../PACKAGE-CONTRACT.md).
