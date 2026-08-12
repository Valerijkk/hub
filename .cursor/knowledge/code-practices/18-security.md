# 18 — Security

Security rules защищают trust boundaries и поэтому чаще являются controls, а не style preferences.

## Authorization

- Принцип: каждый protected resource проверяется server-side; UI visibility не является authorization.
- Evidence: `CONTROL`, `NORM`; high; [S56–S58](sources/registry.md).
- Режим: `HARD` — server enforcement + negative authz tests.
- BAD: доверять `tenant_id` из request body или скрытой кнопке.
- GOOD: actor, tenant membership, role и object relation проверяются на server boundary; tenant A → tenant B regression test обязателен для relevant flow.

## Input and SQL boundary

- Принцип: untrusted input валидируется на входе, data не смешивается с executable syntax.
- Evidence: `CONTROL`, `NORM`; high; [S56, S59](sources/registry.md).
- Режим: `HARD` — parameterized queries/ORM binding; dynamic identifiers только через allow-list.
- BAD: string interpolation внешнего значения в SQL.
- GOOD: bind parameter + typed/validated domain value.

## Secrets and least privilege

- Принцип: credential exposure и privilege определяют blast radius.
- Evidence: `CONTROL`, `NORM`; high; [S56, S61](sources/registry.md).
- Режим: `HARD` — no committed/logged secrets, least privilege, rotation/audit process.
- BAD: broad shared credential в repo/log или agent tool.
- GOOD: scoped identity, secret manager, redaction, rotation evidence.

## Supply chain

- Принцип: dependency identity/provenance должны быть reviewable и immutable в release path.
- Evidence: `CONTROL`, `NORM`; high; [S60, S61, S64–S66](sources/registry.md).
- Режим: `HARD/HYBRID` — reviewed full SHA для third-party Actions, lockfile review, immutable artifact digest, retained provenance.
- BAD: third-party `@main`, mutable `latest`, generated lockfile без review.
- GOOD: pinned identity + provenance + controlled update automation.

## Baseline interpretation

- OWASP ASVS — verification baseline, не linter и не доказательство implementation.
- WSTG — testing guide, а не fixed checklist для каждого endpoint.
- NIST SSDF — process framework; version/date обязаны быть зафиксированы.
- Automated a11y/security scan дополняется boundary tests и review; scanner green не доказывает отсутствие vulnerability.

## Exceptions

Security exception требует owner, exact scope, threat statement, compensating control, expiry и removal test. «False positive» без evidence не является исключением. Формат: [exceptions and rationale](controls/exceptions-and-rationale.md).
