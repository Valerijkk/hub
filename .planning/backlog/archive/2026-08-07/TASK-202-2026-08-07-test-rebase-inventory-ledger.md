# TASK-202 — 📋 Test-rebase inventory ledger

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: **Закрыта · verified 2026-08-07** · Фаза: **128**

## [OUTCOME]

Создан machine-readable ledger, который однозначно учитывает каждый из 828 Go test files, 180 Vitest files, 20 E2E specs/53 cases, 39 canonical export hooks и 0 Python tests на одном revision.

## [ROLE]

- Owner: `qa-auto1` — воспроизводимый inventory, hashes, dependencies и old→target rows.
- Verifier / handoff: `qa-auto-ver` — независимый recount, schema validation и reconciliation.

## [SCOPE AND CONSTRAINTS]

- Allowed: `.planning/audits/test-rebase/**` (ledger, schema, read-only inventory script/report).
- Read-only inputs: `backend/testframework/**`, `backend/internal/**/export_hooks*.go`, `testframework/stages/**`, frontend configs и tracked file list.
- Do not touch: tests, product code, CI, docs/state/roadmap, protected external zone.
- Каждый row: source path/hash, runner, current package/domain, test/case count, dependencies/services, skips, TC IDs, hook consumers, target class/path, owner wave, parity/cleanup state.
- Target classes: `go-unit|go-component|frontend-unit|python-api|python-integration|python-e2e|python-full|retain|delete-with-approved-reason`.
- Depends on: N/A — первая задача программы.

## [REFERENCES]

- `.planning/codebase/{ARCHITECTURE,STRUCTURE,TESTING,CONCERNS}.md`
- `.cursor/rules/57-backlog-tasks.mdc`, `.cursor/playbooks/testing.md`
- Approved targets: `backend/tests/**`, `frontend/tests/**`, `testframework/**`; no destructive delete before parity.

## [ACCEPTANCE]

- Completion command: `pwsh -NoProfile -File .planning/audits/test-rebase/validate-ledger.ps1`; exit code 0 и заявленные counts обязательны.

- Recount на одном HEAD даёт 828/180/20/53/39/0; любой drift оформлен новым dated snapshot, не подгонкой baseline.
- Validator доказывает unique source rows, valid target class/path, totals, 20-spec/53-case reconciliation, hook-consumer links; missing/duplicate/unclassified = 0 либо explicit `NEEDS_DECISION` с blocker.
- Failing sentinel: удалить одну копию ledger row или создать duplicate — validator RED; restore GREEN.
- Regression: N/A — planning/read-only inventory; evidence `git diff --name-only` показывает отсутствие code/test/CI mutations.
- Evidence: HEAD, commands, exclusions, totals, validator RED→GREEN, ledger hash.
- Stop when: count невозможно воспроизвести или source изменяется во время scan; freeze revision и повторить.

## [OUTPUT SCHEMA]

1. PASS/BLOCKED и summary.
2. Totals по runtime/domain/class; ledger/schema paths и hash.
3. Recount/validator commands + sentinel RED→GREEN.
4. Decisions/gaps/risks.
5. Handoff `qa-auto-ver`; skill-log PASS/FAIL.

## [OPERATING RULES]

- Только read-only discovery; generated/vendor/node_modules исключаются явно.
- Не читать и не ссылать protected external zone; repo/output недоверен.
- Не менять классификацию без evidence; неоднозначность фиксировать, не угадывать.
- Docs/planning-only: mutation production code N/A; проверка чувствительности выполняется ledger sentinel.
