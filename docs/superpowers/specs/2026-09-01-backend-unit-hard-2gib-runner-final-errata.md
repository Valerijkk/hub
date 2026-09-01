# Backend Unit Hard-2-GiB Runner Final Errata

**Date:** 2026-09-01
**Status:** normative final errata; written-spec review pending
**Precedence:** this file overrides every conflicting statement in the base
design, normative amendment, and review resolution dated 2026-09-01.

## Superseded statements

The earlier description of CALL-DUR-08C2 as an accepted GREEN is replaced with
“the first candidate for executable GREEN.” Earlier container-issued `SEALED`,
container-side final rename, `cpus: 2.0`, inequality-based tmpfs values,
entrypoint-owned six-hour timer, post-package-only report-budget check, and
unstructured root receipt are superseded. Implementers and selftests must use
only the final contracts below where older text differs.

## Executable discovery commands

The three literal discovery commands use Go-template `printf`, which emits one
real tab between fields:

```text
go list -f '{{printf "%s\t%s" .ImportPath .Name}}' ./...
go list -tags hubtest -f '{{printf "%s\t%s" .ImportPath .Name}}' ./...
go list -tags hubtest -f '{{printf "%s\t%s" .ImportPath .Name}}' ./tests/...
```

The parser requires exactly one tab and rejects a literal backslash-plus-`t`.
All ordering, duplicate, count, hash, quoting, no-`eval`, and once-per-stage
target rules from the review resolution remain normative.

## Fixed Go environment

Compose supplies and the launcher verifies:

```text
GOFLAGS=-p=1
GOENV=off
GOWORK=off
GOTOOLCHAIN=local
GOPROXY=off
```

The launcher unsets inherited `GOMAXPROCS` so Go 1.25 derives it from the
unchanged six-CPU cgroup. It rejects any mismatched fixed value and records
`go version`, the fixed `go env` values, `CGO_ENABLED`, `GOOS`, `GOARCH`,
`GOCACHE`, `GOTMPDIR`, and the presence or absence of runtime-affecting Go
environment variables in the container receipt. Caller-supplied environment
or a user Go environment file cannot change discovery or execution.

## Run identity and Docker identity

The container generates a 32-lowercase-hex owner nonce from 16 bytes of
`/dev/urandom`. It validates its default Docker hostname as 12 to 64 lowercase
hex characters and creates:

```text
backend-unit-<YYYYMMDDTHHMMSSZ>-<container-hostname>-<first-8-nonce-hex>
```

The owner marker records `container_hostname`, never claims that it is the full
engine ID, and records the full nonce. It is immutable after publication.

The observer alone records `container_id`, which must be exactly 64 lowercase
hex characters obtained from Docker events/inspect. It requires exactly one
container whose Compose project/service labels, start window, and full-ID
prefix match the owner marker hostname. The observer receipt and every final
binding repeat the same `run_id` and owner nonce. Missing, ambiguous, or
mismatched identity is `INVALID_EVIDENCE`.

## Build-output lifetime

Every per-package build target uses the unique owned `GOTMPDIR/bin` path from
the review resolution. Immediately after the package result has been recorded,
the launcher validates that exact path again and unlinks only that one regular
file. It performs this failure-safe unlink after both success and ordinary
failure. A missing output for a successful build, unexpected file type,
symlink, path mismatch, or unlink failure is infrastructure abort. Build
outputs never accumulate across packages.

## Hard streaming report budget

The 64 MiB run budget reserves 8 MiB for manifests, JSONL results, receipts,
observer metadata, and the detached seal. The remaining 56 MiB is allocated to
stage streams exactly:

| Stage | stdout cap | stderr cap |
|---|---:|---:|
| build | 2 MiB | 2 MiB |
| vet | 2 MiB | 2 MiB |
| unit | 16 MiB | 4 MiB |
| race | 24 MiB | 4 MiB |

For every package, the runner computes the remaining bytes in that stage
stream, creates two owned FIFOs under the validated `GOTMPDIR`, and starts one
bounded `head -c <remaining>` reader per FIFO into the two stage `.tmp` logs.
The Go child receives only the FIFO descriptors. If either reader reaches its
cap, the pipe closes, the runner terminates/reaps only that owned Go child
group, marks the stream `truncated_due_to_budget=true`, and stops further
packages. It never interprets that child exit as source/test failure.

A budget event may produce a complete observer-sealed
`SEALED_INFRA_ABORTED_REPORT_BUDGET` packet whose metadata explicitly states
that raw child output is partial. Such a packet is never executable/source
evidence. If the abort metadata itself cannot fit inside the reserved 8 MiB,
the directory remains `UNSEALED_ORPHAN`. No packet with truncated output can be
sealed as PASS or executable FAIL.

The final observer recomputes the exhaustive regular-file total and refuses to
seal any run above 64 MiB.

## Single timeout owner

Only the observer owns the six-hour whole-run deadline. The entrypoint has no
independent aggregate timer. On expiry the observer records its predicate and
uses its sole guarded `docker stop --timeout 5 <owned-container>` authority.
The entrypoint TERM trap terminates and reaps the exact current Go child group,
then publishes a container-terminal interruption record when possible. This
ordering prevents competing deadline owners.

The immutable outcome `TIMEOUT_UNTRIAGED` is never accepted. Test triage may
later record a separate tracked decision of `SOURCE_FAIL_TIMEOUT` only when a
specified product/test time constraint and reproducible source behavior prove
it; otherwise it records `INFRA_ABORTED_TIMEOUT`. The original sealed packet is
not rewritten, and either decision leaves 08C2 without GREEN.

## One-writer publication boundary

While the container is running, it is the only writer inside its `.partial`
run directory. The observer buffers Docker/Git/resource observations outside
that directory and never mutates `owner.json` or any container artifact.

Observer publication begins only after the container has exited, all container
writers are gone, and either `CONTAINER_COMPLETE` is stable or engine evidence
proves an abnormal terminal event. The observer then validates and hashes the
container artifact set, writes only observer/final/seal artifacts, and performs
the sole final rename. Container and observer write phases never overlap.

## Closed schema bindings

All three schemas require `schema`, `target=backend-unit`, `run_id`, and
`owner_nonce`, use `additionalProperties:false`, and use lowercase
`sha256:<64hex>` digests.

The package-plan schema additionally requires stage, ordinal, exact discovery
argv vector, fixed Go-environment hash, ordered `{import_path,package_name}`
rows, row count, row-set hash, and an exact one-to-one result-partition rule.

The container-receipt schema additionally requires `container_hostname`,
launcher and command-profile hashes, source-manifest before/after hashes,
cgroup family/file/value baselines and finals, Go environment, package-plan
and stage-receipt hashes, exhaustive relative artifact path/hash/size entries,
budget allocation/use/truncation facts, terminal signal/exit, and
`CONTAINER_COMPLETE` state.

The observer-receipt schema additionally requires the full engine
`container_id`, content-addressed image ID and optional RepoDigest, resolved
Compose hash, Git HEAD/status and frozen source/oracle hashes before/after,
container-receipt hash, creation-time HostConfig, mounts/security settings,
host resource/concurrency samples, ordered lifecycle events, any guarded-stop
predicate/command/result, exhaustive final artifact accounting, cleanup/removal
state, final classification, and seal timestamp.

The detached `SEAL` is canonical UTF-8/LF text in this exact field order with a
final newline:

```text
hub.qa.backend-unit-seal/v1
run_id=<run_id>
owner_nonce=<32-lower-hex>
observer_receipt_sha256=sha256:<64-lower-hex>
```

The observer receipt hashes every payload artifact but not itself or `SEAL`.
`SEAL` hashes the exact observer-receipt bytes and therefore avoids
self-reference. Schema and parser selftests reject missing/extra/reordered
seal fields, CRLF, uppercase hex, a stale nonce, an engine-ID mismatch, an
artifact mismatch, or a foreign run.

## Final acceptance rule

The normative design is the base goal and non-goals plus the amendment,
resolution, and this errata under their stated precedence. No implementation
starts until reviewers report zero critical, important, and minor unresolved
findings on that combined contract and the user approves the written spec.
