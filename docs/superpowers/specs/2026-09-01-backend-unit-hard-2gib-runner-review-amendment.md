# Backend Unit Hard-2-GiB Runner Normative Amendment

**Date:** 2026-09-01
**Applies to:** `2026-09-01-backend-unit-hard-2gib-runner-design.md`
**Status:** written-spec review pending

This amendment is part of the design contract. Its exact rules override a less
specific statement in the base design.

## Frozen memory constants

The implementation uses these exact tmpfs values:

```text
GOCACHE tmpfs ceiling = 768 MiB
GOCACHE eviction watermark = 512 MiB
GOTMPDIR tmpfs ceiling = 384 MiB
/tmp tmpfs ceiling = 64 MiB
sum of tmpfs ceilings = 1216 MiB
```

The runner samples filesystem usage after each exited Go child. It invokes
`go clean -cache` before the next package when used GOCACHE space is at least
512 MiB. It never recursively scans the cache to make that decision.

## Frozen concurrency and deadline rules

Only package processes are serialized. Every `go build`, `go vet`, and
`go test` target receives explicit `-p=1`. The runner does not set
`GOMAXPROCS`, does not add `go test -parallel`, and does not modify test-owned
parallelism. Unit and race retain the existing 150-second per-package Go test
timeout.

The whole entrypoint has a six-hour wall-clock deadline. Expiry terminates and
reaps only the exact owned Go child process group, records
`INFRA_ABORTED_TIMEOUT`, leaves the report non-GREEN, and never skips work while
claiming PASS.

## Frozen package equivalence rule

For each stage the launcher records the complete lexical, duplicate-free
import-path list produced inside the accepted source/image context:

```text
build: go list ./...
vet:   go list -tags hubtest ./...
unit:  go list -tags hubtest ./tests/...
race:  go list -tags hubtest ./tests/...
```

Every listed package must appear exactly once as a stage target. No unlisted
package may appear as a target. Dependency compilation is allowed and is not a
duplicate target. A list error, empty list, duplicate, omitted target, extra
target, or manifest/result count mismatch fails closed. This equivalence is a
mandatory Python architecture selftest and independent review gate.

## Bounded report file shape

Each stage owns exactly five logical artifacts:

1. one ordered package manifest;
2. one package-results JSONL file with one record per manifest entry;
3. one complete stdout log;
4. one complete stderr log;
5. one stage result document.

The runner does not create a host file per package. Stage files use `.tmp`
names until closed and hashed, then rename within the same run directory. This
keeps the physical report write set compact while retaining complete raw
output.

## Seal and observer authority

Container `SEALED` means only that the container-side artifact set is complete
and internally hash-bound. It is neither PASS nor source acceptance. A root
receipt remains mandatory and binds the sealed directory to Git HEAD/status,
source/oracle hashes, image and launcher identity, resolved Compose identity,
container/HostConfig, resource and concurrency timeline, terminal state, and
container removal.

This phase does not create a generic host runner, Docker event daemon, Docker
socket mount, or second public QA API. Root captures the observer receipt
during the canonical command as it does for current Phase-150 gates. If the
container disappears before required engine metadata is captured, the packet
is `OBSERVER_UNBOUND / INVALID_EVIDENCE` even when container-side files are
sealed.

## Image authority

Executable acceptance uses only the image-baked launcher. A tracked launcher
bind is rejected because the host file can change while a container runs and
would create separate development and acceptance semantics. TDD may reach a
source-level implementation before the image build, but Docker GREEN requires
one admitted, resource-monitored rebuild of the affected local QA image and
an exact launcher hash match afterward.

## Scope economy

No new generic runner framework or retired `hub.qa-runner/v2` asset is revived.
The existing exact Python source-contract test remains the primary static
oracle for Compose and launcher bytes. A new schema or automated observer is
outside this atom unless implementation demonstrates that a deterministic
sealed manifest cannot otherwise be validated; that outcome requires a new
contract review rather than silent scope expansion.
