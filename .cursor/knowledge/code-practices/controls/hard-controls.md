# Hard and hybrid controls

## Status

Это **candidate catalog**, перенесённый из research HANDOFF. Наличие строки не означает, что control уже включён в Hub. Перед активацией нужны owner, detector/test, rollout, remediation и exception path.

- `M` — machine-enforced/default hard candidate.
- `H` — hybrid: machine evidence обязательно, окончательная семантика требует review.

## Catalog C01–C44

| ID | Mode | Control | Minimum enforcement/evidence | Sources |
|---|---|---|---|---|
| C01 | M | No committed secrets | secret scanner + incident/history procedure | [S56, S61](../sources/registry.md) |
| C02 | M | Third-party GitHub Actions use reviewed full SHA | workflow policy/linter + controlled updater | [S60](../sources/registry.md) |
| C03 | M | Protected branches require CI/review | VCS branch policy | `POLICY` |
| C04 | M | Maintained TypeScript uses strict checking | `tsc` with staged migration only by exception | [S44](../sources/registry.md) |
| C05 | M | Public/core Python boundaries follow selected type policy | mypy/pyright config scoped to maintained modules | [S35](../sources/registry.md) |
| C06 | H | Concurrency-sensitive Go packages have race evidence | resource-scoped race suite; absence claim forbidden | [S38](../sources/registry.md) |
| C07 | M | Untrusted values do not construct SQL syntax | binding/SAST; dynamic identifiers allow-listed | [S59](../sources/registry.md) |
| C08 | M | Protected operations enforce server-side authz | architecture + negative security tests | [S56–S58](../sources/registry.md) |
| C09 | M | Cross-tenant authz has explicit negative tests | actor tenant A → object tenant B | [S58](../sources/registry.md) |
| C10 | M | External calls have timeout/deadline | static/custom check where precise + boundary tests | [S49, S51](../sources/registry.md) |
| C11 | M | Automated retries have finite budget | policy scan + timeout/retry tests | [S49, S53](../sources/registry.md) |
| C12 | H | Retry distinguishes transient/permanent failures | typed error contract + tests/review | [S49, S51, S53](../sources/registry.md) |
| C13 | H | Non-idempotent retry has idempotency/reconciliation rationale | persisted key/result or documented alternative | [S50](../sources/registry.md) |
| C14 | M | List API paginates or proves a bounded invariant | API contract/lint + documented bound | [S48b](../sources/registry.md) |
| C15 | M | API errors use stable structured contract | contract tests | [S47](../sources/registry.md) |
| C16 | M | Critical data invariants use DB constraints where practical | schema/migration tests | [S54, S55b](../sources/registry.md) |
| C17 | H | Schema migration is tested on realistic path | migration fixture/snapshot + compatibility review | [S54, S55](../sources/registry.md) |
| C18 | M | Production sleep is not a synchronization primitive | targeted lint/review; explicit exception for real delay semantics | `CONTROL` |
| C19 | M | UI/integration tests avoid arbitrary sleeps | test lint + semantic waiting review | [S31](../sources/registry.md) |
| C20 | M | Test resources are isolated and cleaned | fixture infrastructure + leak checks | [S25, S34](../sources/registry.md) |
| C21 | M | Build inputs/toolchains are declared | pinned config/lockfiles | [S64, S65](../sources/registry.md) |
| C22 | M | Release artifact has immutable identity/digest | pipeline assertion | [S64–S66](../sources/registry.md) |
| C23 | M | Environments promote the same built artifact | pipeline topology/evidence | [S65](../sources/registry.md) |
| C24 | M | Release provenance is retained | provenance artifact and retention policy | [S66](../sources/registry.md) |
| C25 | M | Agreed high-confidence scanner findings block | versioned severity policy, suppression expiry | [S56, S61](../sources/registry.md) |
| C26 | M | Dependency vulnerability/license policy is evaluated | supply-chain job + reviewed exception | [S61, S66](../sources/registry.md) |
| C27 | M | Accessible controls expose semantic name/role | automated a11y + review | [S45, S46](../sources/registry.md) |
| C28 | M | Critical auth/user flows have a11y evidence | automated checks + manual complement where required | [S45](../sources/registry.md) |
| C29 | H | Request cancellation propagates | language-specific review and integration tests | [S39–S40, S51](../sources/registry.md) |
| C30 | H | Fan-out concurrency is bounded | static/review + load/failure tests | [S40, S53](../sources/registry.md) |
| C31 | M | Tests do not share mutable global state | lint/review + isolated fixtures | [S25, S34](../sources/registry.md) |
| C32 | H | Behavior depending on time/randomness uses controllable source | deterministic tests + design review | [S25](../sources/registry.md) |
| C33 | M | AI tools run least privilege | platform/tool policy | [S76, S77](../sources/registry.md) |
| C34 | M | External/destructive AI writes have action-bound approval | tool ACL at exact call | [S76](../sources/registry.md) |
| C35 | M | Untrusted content cannot increase agent privilege | structured trust boundary | [S76, S77](../sources/registry.md) |
| C36 | M | Hard agent restrictions exist outside prompts | sandbox/hook/tool permissions | [S76, S78](../sources/registry.md) |
| C37 | M | Agent code passes the same CI as human code | branch policy | `POLICY` |
| C38 | H | High-risk agent changes get independent verification | risk classification + named verifier evidence | `POLICY`; medium |
| C39 | M | Generated dependency changes receive provenance/lockfile review | CI + diff ownership | [S61, S66](../sources/registry.md) |
| C40 | M | Observability context propagates across supported boundaries | integration tests | [S63](../sources/registry.md) |
| C41 | M | Logs exclude configured secret classes | redaction + scanning tests | [S56, S63](../sources/registry.md) |
| C42 | H | Failed async work has observable retry/terminal state | state contract + monitoring/test evidence | [S53, S63](../sources/registry.md) |
| C43 | H | Critical consumers tolerate replay per documented semantics | replay/idempotency tests | [S50, S52](../sources/registry.md) |
| C44 | H | Production migration has rollback or forward-recovery | release review + rehearsed evidence | [S55, S65](../sources/registry.md) |

## Admission and rollout

1. Mark a control `candidate`, `shadow`, `blocking` or `retired` in its owning implementation policy.
2. Start uncertain detectors in shadow/warning mode and measure false positives.
3. Blocking activation requires a fix path and [exception contract](exceptions-and-rationale.md).
4. Never cite this catalog as proof that the repository currently complies.

Design smells belong in [review guidelines](review-guidelines.md) and [warning signals](warning-signals.md), not here.
