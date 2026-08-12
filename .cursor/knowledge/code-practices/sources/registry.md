# Frozen source registry

## Snapshot

This registry freezes the external research input used to create this KB; it is not a live claim that every URL or revision remains current.

| Field | Value |
|---|---|
| Attachment | `715b1413-e8c2-44ee-8559-2e173e86bb34/pasted-text.txt` |
| Bytes | `103288` |
| Lines | `2036` |
| SHA-256 | `0A777BB233D0B14BF0E980FB0B4CCB7CF7E8019626CC3A44932FD9A620518FD5` |
| Attachment mtime UTC | `2026-08-07T21:23:18.0738301Z` |
| Research access date claimed by report | `2026-08-08` |
| Independent URL re-fetch in this integration | Partial upstream verification; see below |

`A` means the report supplied a direct primary/official URL. `B` means a direct practitioner/context source. `P` means partial metadata, a generic landing page or no exact URL; it cannot be the sole basis of a new hard gate. `V-2026-08-07` means an upstream reviewer independently checked only the bounded claim shown in the row on that date.

## Foundations and design

| ID | Frozen source | Reported date/status | Evidence | Integrity |
|---|---|---|---|---|
| S01 | [Dijkstra — Notes on Structured Programming](https://www.cs.utexas.edu/~EWD/transcriptions/EWD02xx/EWD249/EWD249.html) | 1970, historical | foundational | A |
| S02 | [Parnas — On the Criteria To Be Used in Decomposing Systems into Modules](https://doi.org/10.1145/361598.361623) | 1972, historical | foundational/empirical case | A |
| S03 | [Liskov & Wing — A Behavioral Notion of Subtyping](https://dl.acm.org/doi/10.1145/197320.197383) | 1994 | normative concept | A |
| S04 | [Eiffel — Design by Contract and Assertions](https://www.eiffel.org/doc/solutions/Design_by_Contract_and_Assertions) | current page per report | norm/recommendation | A |
| S05 | [Design Patterns publisher record](https://www.informit.com/store/design-patterns-elements-of-reusable-object-oriented-9780321770462) | 1994 | practitioner catalog | A |
| S06 | [Martin — SOLID Relevance](https://blog.cleancoder.com/uncle-bob/2020/10/18/Solid-Relevance.html) | 2020-10-18 | opinion/context | B |
| S07 | [Martin — Single Responsibility Principle](https://blog.cleancoder.com/uncle-bob/2014/05/08/SingleReponsibilityPrinciple.html) | 2014-05-08 | opinion/context | B |
| S08 | [Martin — Dependency Inversion Principle](https://objectmentor.com/resources/articles/dip.pdf) | 1996 | practitioner/foundational | B |
| S09 | [Dan North — Why Every Element of SOLID is Wrong](https://speakerdeck.com/tastapod/why-every-element-of-solid-is-wrong) | n.d. | critique/opinion | B |
| S10 | [Pearson — Clean Code](https://www.pearson.com/en-us/subject-catalog/p/clean-code-a-handbook-of-agile-software-craftsmanship/P200000009044/9780136083252) | 2008 | publisher/practitioner | B |
| S11 | [qntm — Clean Code critique](https://qntm.org/clean) | 2020-06-28 | critique/opinion | B |
| S12 | [Maintainable Method Size in Java](https://arxiv.org/abs/2205.01842) | 2022 | empirical, context-limited | A |
| S13 | [Cognitive Complexity validation](https://arxiv.org/abs/2007.12520) | 2020 | empirical | A |
| S14 | [How do Developers Improve Code Readability?](https://arxiv.org/abs/2309.02594) | 2023 | empirical | A |
| S14b | [To Clean-Code or Not To Clean-Code](https://arxiv.org/abs/2208.07056) | 2022 | empirical/perception | A |
| S15 | [Sandi Metz — The Wrong Abstraction](https://sandimetz.com/blog/2016/1/20/the-wrong-abstraction) | 2016-01-20 | practitioner recommendation | B |
| S16 | [The Pragmatic Programmer — DRY excerpt](https://media.pragprog.com/titles/tpp20/dry.pdf) | 20th Anniversary edition | practitioner/foundational | B |
| S17 | [Fowler — YAGNI](https://martinfowler.com/bliki/Yagni.html) | n.d. | practitioner recommendation | B |
| S18b | [Bernhardt — Boundaries](https://www.destroyallsoftware.com/talks/boundaries) | historical | practitioner recommendation | B |
| S19 | [Fowler — Value Object](https://martinfowler.com/bliki/ValueObject.html) | 2016 | practitioner recommendation | B |
| S20 | [Cockburn — Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/) | 2005-09-04 | practitioner/foundational | B |
| S21 | [Fowler — Dependency Injection](https://martinfowler.com/articles/injection.html) | 2004-01-23 | practitioner/foundational | B |
| S22 | [Fowler catalog — Repository](https://martinfowler.com/eaaCatalog/repository.html) | 2003 | practitioner/context | B |
| S23 | [Fowler — Mocks Aren't Stubs](https://martinfowler.com/articles/mocksArentStubs.html) | 2007-01-02 | practitioner/context | B |
| S24 | [Google Testing Blog — Root Cause of Singletons](https://testing.googleblog.com/2008/08/root-cause-of-singletons.html) | 2008-08-27 | historical opinion | B |

## Testing

| ID | Frozen source | Reported date/status | Evidence | Integrity |
|---|---|---|---|---|
| S25 | [Software Engineering at Google — Testing Overview](https://abseil.io/resources/swe-book/html/ch11.html) | 2020 book | official engineering practice | A |
| S26 | [Google — Tests Too DRY? Make Them DAMP!](https://testing.googleblog.com/2019/12/testing-on-toilet-tests-too-dry-make.html) | 2019-12-03 | local recommendation | A |
| S27 | [Google — Just Say No to More End-to-End Tests](https://testing.googleblog.com/2015/04/just-say-no-to-more-end-to-end-tests.html) | 2015-04-22 | historical/context | A |
| S28 | [Google — SMURF: Beyond the Test Pyramid](https://testing.googleblog.com/2024/10/smurf-beyond-test-pyramid.html) | 2024-10-15 | current recommendation per report | A |
| S30 | [Google Code Review Standard](https://google.github.io/eng-practices/review/reviewer/standard.html) | current per report | local policy | A |
| S30b | [Google — Small CLs](https://google.github.io/eng-practices/review/developer/small-cls.html) | current per report | local policy | A |
| S31 | [Playwright Best Practices](https://playwright.dev/docs/best-practices) | current docs per report | official docs | A |
| S32 | [Selenium Page Object Models](https://www.selenium.dev/documentation/test_practices/encouraged/page_object_models/) | current docs per report | official docs | A |
| S33 | [Fowler — Page Object](https://martinfowler.com/bliki/PageObject.html) | 2013-09-10 | practitioner concept | B |
| S34 | [pytest fixtures](https://docs.pytest.org/) | current docs per report | official docs | P: docs root |
| S34b | [pytest flaky tests](https://docs.pytest.org/) | current docs per report | official docs | P: docs root |

## Languages and UI

| ID | Frozen source | Reported date/status | Evidence | Integrity |
|---|---|---|---|---|
| S35 | [PEP 544 — Protocols](https://peps.python.org/pep-0544/) | final | norm | A |
| S36 | [PEP 557 — Data Classes](https://peps.python.org/pep-0557/) | final | norm | A |
| S37 | [Python contextlib](https://docs.python.org/3/library/contextlib.html) | Python 3.14 docs per report | official docs | A |
| S38 | [Go Memory Model](https://go.dev/ref/mem) | revised 2022 | norm | A |
| S39 | [Go Concurrency Patterns: Context](https://go.dev/blog/context) | 2014-07-29 | official recommendation | A |
| S40 | [Go pipelines and cancellation](https://go.dev/blog/pipelines) | 2014-03-13 | official recommendation | A |
| S41 | [Go fuzzing](https://go.dev/doc/security/fuzz/) | current per report | official docs | A |
| S42 | [React — You Might Not Need an Effect](https://react.dev/learn/you-might-not-need-an-effect) | current per report | official recommendation | A |
| S43 | [React — Choosing State Structure](https://react.dev/learn/choosing-the-state-structure) | current per report | official recommendation | A |
| S44 | [TypeScript `strict`](https://www.typescriptlang.org/tsconfig/strict.html) | current per report | norm/docs | A |
| S44b | [TypeScript Narrowing](https://www.typescriptlang.org/docs/handbook/2/narrowing.html) | current per report | official docs | A |
| S45 | [WCAG 2.2](https://www.w3.org/TR/WCAG22/) | W3C Recommendation 2023-10-05 | norm | A |
| S46 | [W3C — Using ARIA](https://www.w3.org/WAI/ARIA/apg/practices/read-me-first/) | current per report | norm/docs | A |

## API, resilience and storage

| ID | Frozen source | Reported date/status | Evidence | Integrity |
|---|---|---|---|---|
| S47 | [RFC 9457 — Problem Details for HTTP APIs](https://www.rfc-editor.org/info/rfc9457/) | 2023-07 | norm | A |
| S48 | [Google AIP-180 — Backwards Compatibility](https://google.aip.dev/180) | 2019-07-23 | local API guidance | A |
| S48b | [Google AIP-158 — Pagination](https://google.aip.dev/158) | n.d. | local API guidance | A |
| S49 | [AWS — Timeouts, retries and backoff with jitter](https://aws.amazon.com/builders-library/timeouts-retries-and-backoff-with-jitter/) | updated 2026-06-12 per report | official engineering experience | A |
| S50 | [AWS — Idempotent APIs](https://aws.amazon.com/builders-library/making-retries-safe-with-idempotent-APIs/) | current per report | official engineering experience | A |
| S51 | [gRPC Deadlines](https://grpc.io/docs/guides/deadlines/) | updated 2025-07-07 per report | official docs | A |
| S52 | [Apache Kafka documentation](https://kafka.apache.org/documentation/) | 4.1 docs per report | official docs | P: broad docs page |
| S53 | [Google SRE](https://sre.google/) | current edition per report | official engineering guidance | P: book root |
| S54 | [PostgreSQL Transaction Isolation](https://www.postgresql.org/docs/current/transaction-iso.html) | v18 current per report | official docs | A |
| S55 | [PostgreSQL ALTER TABLE](https://www.postgresql.org/docs/current/sql-altertable.html) | current per report | official docs | A |
| S55b | [PostgreSQL INSERT / ON CONFLICT](https://www.postgresql.org/docs/current/sql-insert.html) | current per report | official docs | A |

## Security, delivery and observability

| ID | Frozen source | Reported date/status | Evidence | Integrity |
|---|---|---|---|---|
| S56 | [OWASP ASVS](https://owasp.org/www-project-application-security-verification-standard/) | stable 5.0.0, released 2025-05-30 | norm/baseline | A; V-2026-08-07 |
| S57 | [OWASP WSTG 4.2](https://owasp.org/www-project-web-security-testing-guide/v42/) | 4.2 stable; 5.0 under development | norm/guidance | A; V-2026-08-07 |
| S58 | [OWASP Cheat Sheet Series](https://cheatsheetseries.owasp.org/) — authorization automation named by report | current per report | official guidance | P: exact page absent |
| S59 | [OWASP Cheat Sheet Series](https://cheatsheetseries.owasp.org/) — SQL injection prevention named by report | current per report | official guidance | P: exact page absent |
| S60 | [GitHub Actions secure use](https://docs.github.com/en/actions/reference/security/secure-use) | full-length commit SHA is the documented immutable Action reference | official docs/control | A; V-2026-08-07 |
| S61 | [NIST SP 800-218 SSDF 1.1](https://csrc.nist.gov/pubs/sp/800/218/final) | v1.1 final, 2022-02-03; no v1.2 final surfaced in verification | norm | A; V-2026-08-07 |
| S62 | [NIST SP 800-218A](https://csrc.nist.gov/pubs/sp/800/218/a/final) | 2024-07-26, final | norm | A |
| S63 | [OpenTelemetry Signals](https://opentelemetry.io/docs/concepts/signals/) | modified 2026-03-10 per report | official docs | A |
| S64 | [Reproducible Builds Definition](https://reproducible-builds.org/docs/definition/) | current per report | norm | A |
| S65 | [Google — Hermetic/Reproducible builds](https://google.github.io/building-secure-and-reliable-systems/raw/ch14.html) | 2020 book | official engineering guidance | A |
| S66 | [SLSA Specification v1.2](https://slsa.dev/spec/v1.2/) | v1.2 per report | norm | A |

## Test design

| ID | Frozen source | Reported date/status | Evidence | Integrity |
|---|---|---|---|---|
| S67 | [ISTQB CTFL 4.0.1](https://istqb.org/wp-content/uploads/2024/11/ISTQB_CTFL_Syllabus_v4.0.1.pdf) | 2024-09-15 | standardized definitions | A |
| S68 | [NIST ACTS](https://csrc.nist.gov/projects/automated-combinatorial-testing-for-software) | current per report | official project | A |
| S69 | [Hypothesis](https://hypothesis.readthedocs.io/) | current per report | official project docs | A |
| S73 | [Microsoft Research — Lifecycle of Flaky Tests](https://www.microsoft.com/en-us/research/publication/a-study-on-the-lifecycle-of-flaky-tests/) | 2020-07-01 | empirical/context-limited | A |

## AI agents and workflow

| ID | Frozen source | Reported date/status | Evidence | Integrity |
|---|---|---|---|---|
| S76 | [OpenAI — Safety in building agents](https://developers.openai.com/api/docs/guides/agent-builder-safety) | current per report | official guidance/control | A |
| S77 | [OpenAI — Running Codex safely](https://openai.com/index/running-codex-safely/) | 2026 | local operational evidence | A |
| S77b | [OpenAI — Agents resisting prompt injection](https://openai.com/index/designing-agents-to-resist-prompt-injection/) | 2026 | official guidance | A |
| S78 | [Anthropic — Claude Code memory/instructions](https://code.claude.com/docs/en/memory) | current per report | official docs | A |
| S78b | [Anthropic — Claude Code security](https://code.claude.com/docs/en/security) | current per report | official docs | A |
| S80 | [open-gsd/gsd-core](https://github.com/open-gsd/gsd-core) | MIT community continuation; latest surfaced v1.8.0; exact revision not pinned | project provenance | A; V-2026-08-07 |

## Quarantined/incomplete report entries

These report IDs are preserved for integrity but are not cited as sole evidence in this KB until an exact primary URL/edition is resolved: `S18`, `S24b`, `S29`, `S41b`, `S70`, `S71`, `S72`, `S74`, `S75`, `S79`, `S80b`, `S81`, `S82`, `S83`, `S84`, `S86`.

See [evidence status](evidence-status.md) for admission rules and integrity limitations.
