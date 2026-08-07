# Cargo-cults

| Не делать обязательным | Вместо этого | Mode |
|---|---|---|
| Function `≤N` lines | warning + cohesion/nesting/effects review | `WARNING` |
| SRP по числу методов | group by real change reasons | `REVIEW` |
| Interface for every class | consumer capability/variation seam | `REVIEW` |
| Repository for every entity | protect a useful domain/query decision | `REVIEW` |
| Ports/adapters for every service | use at significant external boundary | `REVIEW` |
| Functional core everywhere | separate decisions from I/O where it buys testability | `REVIEW` |
| Immutable everything | prefer values; measure buffers/hot paths | `REVIEW` |
| Literal Rule of Three | wait for enough evidence; three is heuristic | `REVIEW` |
| POM for every UI suite | semantic locator first, extract repeated UI knowledge | `REVIEW` |
| Always/never mocks | choose by observable contract | `REVIEW` |
| One assertion per test | one behavioral concept | `REVIEW` |
| One physical call per Act | one meaningful operation/action | `REVIEW` |
| Fixed Pyramid/Trophy/Honeycomb | risk/latency/resource model | `REVIEW` |
| 100% line coverage | coverage is visibility, not oracle quality | `WARNING` |
| Universal mutation score | targeted changed/critical logic probe | `WARNING` |
| Zero duplication | deduplicate shared knowledge | `WARNING` |
| No comments | preserve rationale/contracts; remove restatement | `REVIEW` |
| No inheritance | allow stable behavioral subtype | `REVIEW` |
| All Singleton evil | reject hidden mutable global reachability | `REVIEW/HYBRID` |
| Fixed Cognitive Complexity number | calibrate locally before any threshold | `WARNING` |
| Exact agent context-file length | optimize scope/clarity; measure tool-specific behavior | `REVIEW` |
| “CAP means pick two” | state exact model/guarantees during partition | `NEGATIVE_SENTINEL` |
| “Exactly once” without scope | define operation, boundaries, failure/restart and effects | `NEGATIVE_SENTINEL` |

Evidence for numerical absolutes and universal testing shapes is `UNKNOWN`, `DISPUTED` or context-limited. Они не переходят в hard gate без нового Hub-specific evidence и admission process из [evidence model](../01-evidence-model.md).
