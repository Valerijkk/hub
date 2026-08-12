# Contradictions and reconciliations

| Apparent contradiction | Reconciliation | Evidence / confidence |
|---|---|---|
| DRY vs DAMP | DRY shared knowledge; DAMP local test narrative | [S16, S26](../sources/registry.md); high |
| OCP vs YAGNI | Create seam after credible variation, not hypothetical future | `REC`; medium |
| Small functions vs local reasoning | Split when extraction creates meaningful named abstraction | `EMP/REC`; high |
| Composition vs inheritance | Composition for orthogonal variation; inheritance for stable behavioral subtype | [S03](../sources/registry.md); high |
| Contract vs defensive validation | Trusted internal contract may assert; untrusted boundary validates/rejects | `NORM`; high |
| Pure core vs rich domain objects | Pure decisions and encapsulated values can coexist | `REC`; medium |
| State tests vs mocks | Verify state/output unless collaboration itself is observable contract | [S23](../sources/registry.md); high |
| POM vs plain locators | Add POM when repeated UI knowledge/change pays for indirection | [S31, S32](../sources/registry.md); high |
| Hermeticity vs realism | Use both at suitable scopes; realism does not excuse uncontrolled dependency | [S25, S28](../sources/registry.md); high |
| Retry vs fail fast | Retry classified transient safe operation; fail permanent/invalid fast | [S49–S53](../sources/registry.md); high |
| Immutability vs performance | Default for shared values; measure hot-path mutation need | `REC`; medium |
| YAGNI vs security | Security/compatibility are known constraints, not speculative features | [S56, S61](../sources/registry.md); high |
| Repository vs direct ORM | Boundary only when it hides domain/query decision | `REC`; high |
| Interface segregation vs concrete simplicity | Consumer capability drives interface | `REC`; high |
| Pyramid vs Trophy/Honeycomb | Different architectures optimize different scopes; no universal ratio | [S25, S28](../sources/registry.md); high |

Если reconciliation не объясняет failure model и scope, claim остаётся `DISPUTED`.
