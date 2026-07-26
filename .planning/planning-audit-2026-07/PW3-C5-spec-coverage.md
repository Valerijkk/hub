# PW3-C5 — `spec_coverage` metric for STATE.md

> Planning-audit **C5** · docs only · 2026-07-27  
> Source: [`SPEC-COVERAGE.md`](SPEC-COVERAGE.md) (C3)  
> Do **not** fight other agents on `.planning/STATE.md` if locked — paste block below when free.

## Definition

```
spec_coverage = (FR with numbered GSD phase OR planned phase 126+) / (all numbered FR-* in docs/system_specification.md)
```

- **Denominator:** 28 (`FR-INTR-01..11` + `FR-CAL-01..08` + `FR-WIKI-01..09`)
- **Numerator:** 9 (see SPEC-COVERAGE «FRs that do have a numbered/planned phase»)
- July wave tasks alone do **not** count (Wiki W5/W8 → still uncovered for this metric)

## Computed value

| Field | Value |
|-------|-------|
| `fr_total` | 28 |
| `fr_with_phase` | 9 |
| `spec_coverage` | **9/28** |
| decimal | **0.321** |
| percent | **32.1%** |

## Paste-ready YAML (STATE.md frontmatter)

Insert under `progress:` (alongside `total_phases` / `percent` → prefer renaming `percent` → `phases_percent` per A6 when editing STATE):

```yaml
progress:
    total_phases: 125
    completed_phases: 125
    total_plans: 250+
    completed_plans: 250+
    phases_percent: 100
    # C5 · 2026-07-27 · source: .planning/planning-audit-2026-07/SPEC-COVERAGE.md
    # Fraction of FR-* in docs/system_specification.md that map to a numbered
    # GSD phase (.planning/ROADMAP.md / phases/) or planned gap phase 126+.
    # July waves alone do not count. 9/28 = FR-INTR-07..11 + FR-CAL-03..06.
    spec_coverage: "9/28"
    spec_coverage_decimal: 0.321
    spec_coverage_as_of: "2026-07-27"
    spec_coverage_source: ".planning/planning-audit-2026-07/SPEC-COVERAGE.md"
```

### Minimal one-liner (if only adding the field)

```yaml
    spec_coverage: "9/28"   # C5 2026-07-27 · SPEC-COVERAGE.md · 9 FR with GSD/planned phase of 28
```

## Not applied to STATE.md in this PW

`STATE.md` may be locked by parallel agents (A1/A5/A6). This file is the **canonical paste source** for C5. After paste, verify frontmatter still parses and session note cites C5.

---

*C5 complete when this file exists and matches SPEC-COVERAGE numerator/denominator.*
