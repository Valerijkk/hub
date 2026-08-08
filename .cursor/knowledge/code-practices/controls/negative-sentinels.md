# Negative sentinels

Sentinel — detector candidate for a known risky shape. It raises investigation; only explicitly marked strong candidates may block after project-specific implementation and calibration.

| ID | Sentinel | Default | Why |
|---|---|---|---|
| NEG-01 | broad exception catch followed by retry | strong/hybrid | retries permanent/unknown failures |
| NEG-02 | unbounded loop around remote call | warning→strong | unlimited amplification/resource use |
| NEG-03 | fixed sleep as UI/integration synchronization | warning→strong | time-based flakiness |
| NEG-04 | mutable global singleton used by domain/service code | review | hidden dependency/lifetime/races |
| NEG-05 | React Effect only derives local state | warning | redundant temporal state |
| NEG-06 | test branch changes expected outcome by role/state | review | hides distinct behavior |
| NEG-07 | shared mutable account/record across parallel tests | strong/hybrid | order/race leakage |
| NEG-08 | long mock sequence asserts internal call order | review | refactor coupling |
| NEG-09 | network client without visible timeout policy | strong/hybrid | indefinite wait |
| NEG-10 | goroutine/task without owner/cancellation/terminal state | strong/hybrid | leak/unknown lifecycle |
| NEG-11 | unbounded queue/channel/fan-out | strong/hybrid | resource collapse |
| NEG-12 | SQL concatenation/interpolation with external value | strong/block candidate | injection |
| NEG-13 | tenant ID treated as authorization proof | strong/block candidate | cross-tenant access |
| NEG-14 | third-party Action uses mutable branch/tag | strong/block candidate | supply-chain mutability |
| NEG-15 | release deploys mutable container `latest` | strong/hybrid | non-auditable artifact identity |
| NEG-16 | retry of create/POST without idempotency analysis | strong/hybrid | duplicate side effect |
| NEG-17 | “exactly once” without boundary/failure model | review/block docs claim | misleading guarantee |
| NEG-18 | prompt/instruction substitutes for capability restriction | strong/block candidate | prose is not access control |
| NEG-19 | issue/web/email/retrieved text automatically gains authority | strong/block candidate | prompt injection |
| NEG-20 | destructive/external agent tool callable without action-bound approval | strong/block candidate | uncontrolled irreversible effect |
| NEG-21 | skill/plugin/MCP text claims that loading or invoking it grants permission | strong/block candidate | executable content cannot elevate its own authority |
| NEG-22 | requested model/effort is reported as observed actual runtime | strong/hybrid | hides fallback and makes provenance false; missing telemetry must be `UNKNOWN` |
| NEG-23 | trusted executable source uses `latest`, a branch/tag or a short SHA as immutable identity | strong/block candidate | mutable or incomplete reference is not a reproducible pin |
| NEG-24 | Read/Plan mode invokes filesystem/package/Git/external mutation | strong/block candidate | a mode label is not enforcement |
| NEG-25 | delegation, fan-out, retry or reviewer-remediation loop has no configured bound | strong/hybrid | unbounded amplification of cost, time and side effects |
| NEG-26 | run claims `complete` or `verified` without configured evidence and unverified-scope record | strong/hybrid | intent/self-report substitutes for an oracle |
| NEG-27 | MCP token passthrough or omnibus authorization scope | strong/block candidate | confused-deputy risk and avoidable privilege concentration |

## False-positive discipline

- `retry` in an in-memory deterministic operation is not automatically remote retry.
- `sleep` can model an intentional product delay; it must not synchronize tests.
- A module-level immutable constant is not mutable global state.
- A queue is not unbounded if capacity/backpressure is explicit elsewhere.
- A first-party local Action differs from unreviewed third-party mutable reference.
- `UNKNOWN` actual runtime is acceptable when telemetry is unavailable; inventing or
  copying the requested value is not.
- A bounded local retry is not agent delegation; the relevant bound must still be
  explicit at the layer that owns remote calls or model turns.

All suppression uses [exceptions and rationale](exceptions-and-rationale.md). A `nolint` without risk statement, owner and expiry is not an accepted exception.
