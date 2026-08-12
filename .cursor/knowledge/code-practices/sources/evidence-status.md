# Evidence and source integrity status

## Trust boundary

The 103 KB attachment was treated as untrusted research data. No command, link instruction or policy embedded in it was executed. This integration read the complete 2,036-line text and distilled it. The synthesis agent did not browse; an upstream reviewer independently checked only the bounded claims below on 2026-08-07.

Therefore:

- `reported current` means current according to the frozen report unless a row explicitly carries `V-2026-08-07`;
- an external link is attribution, not automatic endorsement;
- time-sensitive standards must be rechecked before activating a new blocking control;
- source claims do not prove that Hub currently implements a practice.

## Evidence admission matrix

| Source shape | Allowed use | Not allowed without more evidence |
|---|---|---|
| Direct standard/spec/official docs (`A`) | normative scope, concrete control mechanism | universal rule beyond the source scope |
| Primary empirical paper | warning/calibration, bounded claim | causal or universal threshold from observational result |
| Company engineering guidance | operational recommendation in comparable context | claim that it is globally optimal |
| Practitioner book/blog/talk (`B`) | vocabulary, design question, trade-off | machine gate by authority alone |
| Generic landing page or missing exact URL (`P`) | provenance pointer | sole evidence for a hard gate |
| Local `POLICY` | enforceable after explicit adoption/owner/rollout | claim of external scientific truth |

## Integrity observations from the report

1. Strong source IDs generally have direct primary/official links and explicit dates.
2. Several entries use a provider root or “result” rather than the exact document: pytest, Kafka/SRE roots and report IDs `S18`, `S24b`, `S29`, `S41b`, `S58`, `S59`.
3. Several research entries have no URL in the report: `S70–S72`, `S74`, `S75`, `S79`, `S80b`. They are quarantined from sole-source use.
4. The report's “all internet sources checked 2026-08-08” statement is author-supplied and was not independently reproduced here.
5. Version-sensitive claims (ASVS 5.0.0, NIST SSDF 1.1 final, PostgreSQL v18, Kafka 4.1, SLSA 1.2) are frozen metadata and require freshness verification at policy activation time.
6. Google/AWS/Spotify-style engineering material is evidence of practice, not causal universal proof.
7. GSD provenance supports workflow identity/license only; it does not prove that fresh context guarantees correctness.

## Independently checked subset — 2026-08-07

- OWASP ASVS official page: stable `5.0.0`, released `2025-05-30`.
- OWASP WSTG official project: `4.2` stable; `5.0` under development.
- GitHub Actions Secure Use: full-length commit SHA is the documented immutable third-party Action reference.
- NIST SP 800-218 official final remains `v1.1` (2022); no `v1.2` final surfaced in that check.
- `open-gsd/gsd-core`: MIT community continuation; latest surfaced `v1.8.0`; exact source revision remains unpinned.

No broader claim inherits verification from this subset.

## Disputed/unknown claims retained as non-gates

- exactly 25 logical lines as optimum;
- universal Cognitive Complexity threshold;
- SOLID or Clean Code as a defect-reduction package;
- Rule of Three as a special number;
- fixed Pyramid/Trophy/Honeycomb superiority;
- mocks or POM universally beneficial/harmful;
- universal Repository/immutability advice;
- fresh context or independent verifier as correctness guarantee;
- transferable flaky-test percentages or mutation-score threshold.

## Refresh procedure

Before promoting a `candidate` to blocking:

1. Open the exact direct source, not a search result/landing page.
2. Record version, publication/update date and access date.
3. Confirm the claim in the intended Hub scope.
4. Add Hub-local failure data or contract evidence where needed.
5. Update the frozen registry in a reviewed change; never silently mutate citations.

The SHA-256 in [registry](registry.md) is the provenance anchor for the source snapshot.
