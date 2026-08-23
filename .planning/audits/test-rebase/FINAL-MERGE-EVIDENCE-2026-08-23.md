# Test rebase final merge evidence ? 2026-08-23

## Accepted scope

- Minimal root `testframework/` package and framework selftest skeleton.
- A bounded set of genuine backend and frontend unit-test repairs/additions.
- Hosted CI automation remains disabled; local Docker is the accepted gate.
- No product/module/production-readiness claim is made.

## Fresh Docker evidence

All commands ran sequentially with `COMPOSE_PARALLEL_LIMIT=1`, disposable
containers and Docker cache retained.

| Service | Result | Evidence |
|---|---|---|
| `backend-unit` | `PASS` | build, vet, unit and race receipts under `testframework/reports/backend-unit/`; target exit `0` |
| `frontend-unit` | `PASS` | TypeScript `PASS`; Vitest `338/338` suites and `419/419` tests; receipts under `testframework/reports/frontend-unit/` |
| `python-contract` | `PASS` | framework selftests only: `181/181` passed in 4.26s |

The owner-bound Calendar WIP test
`frontend/tests/calls/call-calendar-a11y.test.tsx` is deliberately excluded
from the frontend-unit image and this merge authority.

## Explicit scope reductions

By founder direction on 2026-08-23, five SUT contract tests and their five
OpenAPI/privacy/security case records were deleted. `python-contract` therefore
validates the installable framework skeleton, not Hub SUT contract behavior.
Python API, integration, e2e and load targets are outside this minimal merge
gate and remain `NOT_RUN`/not accepted. Hosted GitHub/GitLab CI is disabled and
was not executed.

## Legacy deletion provenance

- 14 founder-authorized legacy unit deletions: commit
  `83bf84946b0f7605e17eb164e55818f2f74ab699`, manifest digest
  `c5701a3ab72283caec9aa4b772582dff02799bff0975a7f0c9cc1e9e6306b37c`.
- 241 founder-authorized non-Mail/non-Calendar hook-dependent legacy
  deletions: commit `e85feac5c02f4652997c670e1c6fd4640f5760c0`, manifest digest
  `d6bc5ce2e1406e2baa2484b0e87155a4abdf102ffe0fdf96f2c6fbcbd0bc5cc9`.
- Machine-readable receipts:
  `.planning/audits/test-rebase/backend-unit-delete-supersession-2026-08-22.json`
  and
  `.planning/audits/test-rebase/backend-hook-dependent-delete-supersession.json`.

## Publication rule

Publication is allowed only as a non-force fast-forward of the reviewed HEAD
to `master`, followed by remote SHA verification. Protected Mail/Calendar WIP
and unrelated dirty-worktree edits are excluded from exact-path commits.
