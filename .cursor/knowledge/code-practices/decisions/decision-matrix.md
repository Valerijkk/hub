# Decision matrix

| Decision | Prefer A when | Prefer B when | Smell / evidence |
|---|---|---|---|
| Function / class | function transforms explicit inputs | class owns invariant/state/lifecycle | stateless class with unrelated one-off method; `REC` |
| Composition / inheritance | behavior/lifetime varies independently | stable behavioral subtype preserves contract | override disables base behavior; `NORM` [S03](../sources/registry.md) |
| Duplicate / abstraction | concepts may diverge, evidence weak | same knowledge must change together | flags in “generic” abstraction; `REC` [S15, S16](../sources/registry.md) |
| Concrete / interface | one internal implementation and no seam | consumer capability, variation or boundary | interface quota; `REC` |
| Direct ORM / repository | ORM model already sufficient for CRUD | domain/query decision needs isolation | generic CRUD wrapper; `REC` |
| Module value / injected lifetime / singleton | immutable deterministic value | mutable/external collaborator needs explicit lifetime | hidden global access; `REC` |
| Mutex / channel | protected shared state has clear owner | ownership transfer/coordination fits messages | ideological channel-only design; `NORM/REC` |
| Retry / fail fast | known transient + safe/idempotent + deadline | invalid/auth/permanent/unsafe effect | broad catch+retry; `CONTROL` [S49–S51](../sources/registry.md) |
| Fixture / factory / builder / Object Mother | lifecycle / defaults / expressive overrides / named canonical persona | choose smallest sufficient mechanism | hidden fixture graph; `REC` |
| Fake / stub / spy / mock | lightweight semantics / input / meaningful interaction | prefer state/output when contract permits | private call-order assertions; `DISPUTED` [S23](../sources/registry.md) |
| Plain UI flow / Component / Page Object | tiny flow / repeated widget / repeated page service | minimize layers | class-per-URL; `REC` [S31, S32](../sources/registry.md) |
| Unit / integration / contract / E2E | pure decision / boundary / compatibility / user journey | combine by risk and latency | one level expected to prove everything; `REC` [S25, S28](../sources/registry.md) |
| Split / keep >25 lines | mixed levels/change reasons/nesting | cohesive linear algorithm clearer together | split solely for counter; `EMP/REC` [S12–S14](../sources/registry.md) |
| Extract helper / DAMP duplication | hide irrelevant mechanics | repeat behavior-relevant values/steps | tests read as framework DSL; `REC` [S26](../sources/registry.md) |

Decision must name the axis of change or failure. Pattern names alone are insufficient rationale.
