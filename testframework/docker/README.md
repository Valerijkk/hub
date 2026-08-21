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

## Static launcher boundary receipt

The committed Task 2A control bundle is
`e357e71ae445c797c686787e8cde2d643566bb68` (source
`cc0f3fcbf4e9a0adc70e946e739b81d21ac7da6c`; tree
`b16311222d3093ae4d2bb462a2e4d2835a3f38cf`). Its reviewed build-manifest and
source-policy blobs are `da1524cca5f68a4962680b4b0c8272fd22359df9` and
`04148a8a0d6ddb0836437bbf7b9c91d07b996fd2`. Contract (Terra/high), Security
(Terra/xhigh), and Lifecycle (Terra/high) precommit and immutable reviews are
PASS C/I/M `0/0/0`, dispatcher-enforced, no fallback, telemetry `UNKNOWN`.

This only documents the fixed-sibling public/private launcher boundary. No
`images.lock.v3.json`, dependency bytes, source image, Compose/admission
authority, Dockerfile authority, ready target, or runtime PASS exists. The
public launcher therefore remains `NOT_READY`/`NOT_RUN` and does not consume
these assets. This becomes governing only after the atomic Task 2B exact7 docs
commit plus immutable review; it cannot self-activate and Task 3 remains
separate.
