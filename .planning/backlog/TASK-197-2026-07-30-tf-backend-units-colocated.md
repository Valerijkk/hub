# TASK-197 — 🧪 Backend target ledger и shared support

> Заведено: 2026-07-30 · Revision: **2026-08-07** · Приоритет: **P1** · Статус: Открыта · Фаза: **132**

## [OUTCOME]

`backend/tests/**` имеет рабочий Go test harness/shared support и утверждённый file-level migration ledger для всех 828 legacy файлов без их перемещения или удаления.

## [ROLE]

- Owner: `be1` — Go target harness, shared support и destination ledger.
- Verifier / handoff: `be-ver` — package semantics, compile/discovery, hook rules и ledger completeness.

## [SCOPE AND CONSTRAINTS]

- Allowed: `backend/tests/**`, backend migration ledger/report, минимальный test-only config внутри `backend/`.
- Read-only: `backend/testframework/**`, `backend/internal/**/export_hooks*.go`, `backend/go.mod`.
- Do not touch: product code, legacy files/hooks, CI, Python/FE trees.
- Target остаётся external package tree внутри Go module: импорт `internal/**` разрешён, private symbols недоступны. `hubtest` hooks сохраняются consumer-by-consumer.
- Классификация 828 rows: `go-unit|go-component|python-api|python-integration|delete-with-approved-reason`; current directory не является доказательством типа.
- Depends on: TASK-202, TASK-203, TASK-204.

## [REFERENCES]

- `.planning/codebase/{ARCHITECTURE,STRUCTURE,TESTING,CONVENTIONS}.md`
- `backend/testframework/{env,factory,mongotest,steps}/`, `backend/internal/**/export_hooks*.go`
- `.cursor/playbooks/testing/backend-go.md`, `.cursor/rules/56-regression-verify.mdc`

## [ACCEPTANCE]

- Completion command: `cd backend && go test -tags hubtest -race -count=1 ./tests/...`; exit code 0 и заявленные counts обязательны.

- Completion command: `cd backend && go test -tags hubtest -race -count=1 ./tests/...`; exit code 0 и заявленные counts обязательны.

- Ledger содержит все 828 уникальных legacy paths, target class/path, package, dependencies, test count, hook consumers и owner wave; missing/duplicate = 0.
- `backend/tests/shared/**` обеспечивает минимальный representative unit test; target command с `-tags hubtest` собирает ненулевой пакет без запуска legacy второй раз.
- Failing sentinel: временно добавить заведомо падающий target test — target command non-zero; restore GREEN.
- Regression: canonical legacy domain/full discovery unchanged; hook files/count и consumers не удалены.
- Evidence: ledger validator output, target/legacy commands and counts, sentinel RED→GREEN, `go vet -tags hubtest` для затронутого scope.
- Stop when: row нельзя классифицировать по evidence или support требует production API expansion; пометить `NEEDS_DECISION`, не угадывать.

## [OUTPUT SCHEMA]

1. PASS/BLOCKED и резюме.
2. Ledger totals по taxonomy/domain/destination и exceptions.
3. Harness/support changes, commands/counts, sentinel evidence.
4. Hook consumer baseline и regression.
5. Handoff `be-ver`, risks, skill-log.

## [OPERATING RULES]

- Сначала ledger, затем минимальный harness; legacy не переносить и не удалять.
- Test triage для каждого FAIL; same instrumentation с TASK-203.
- Hook не удаляется в backend waves: zero-consumer proof только помечает candidate; физическое удаление разрешено исключительно TASK-219 после CI/parity gates.
- Docs-sync ledger/readme; atomic commit со ссылкой на TASK, без AI attribution.
