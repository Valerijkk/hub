# Warning signals

Warnings направляют attention reviewer-а и не определяют correctness.

| Signal | Что может обнаружить | Что пропускает | Default action |
|---|---|---|---|
| Logical lines > local budget | navigation/size burden | semantic cohesion | спросить о responsibilities; не fail |
| Cyclomatic complexity | control-flow paths | domain/data complexity | подобрать path tests и inspect branching |
| Cognitive Complexity | nesting/flow proxy | naming/contracts/domain load | review, local calibration before threshold |
| Nesting depth | mental stack | long flat God procedure | consider guards/extraction |
| Fan-out/dependency count | change surface | stability of dependencies | inspect boundary/lifetime |
| Constructor dependency count | orchestration breadth | legitimate composition root | inspect role, not numeric fail |
| Duplication | repeated syntax | whether knowledge is shared | investigate only |
| Mock count/interactions | potential test coupling | meaningful collaboration contract | inspect assertions |
| Interface/class/module count | fragmentation | semantic boundary quality | inspect navigation/change reasons |
| Coverage | executed locations | assertion/oracle value | visibility, not quality score |
| Mutation survival | weak oracle candidates | equivalent/low-value mutants | targeted probe, not universal target |
| Comment density | unusual explanation burden | rationale value | read comments; never numeric gate |

## Function review budget

Suggested local warning:

```text
if logical_lines > 25:
    WARN function-review-budget
    collect nesting, branches, effect categories, dependencies
```

Evidence is `EMP` but context-limited: S12 studied Java SLOC observationally. It does not prove a universal 25 logical-line optimum. Keep cohesive linear code together when splitting would worsen locality.

## Promotion rule

Warning may become a gate only after:

- a precise failure class is identified;
- Hub-local false-positive telemetry is acceptable;
- remediation does not incentivize ritual wrappers/helpers;
- owner and exception lifecycle are defined.

Until then, warnings never block solely because a scalar crossed a number.
