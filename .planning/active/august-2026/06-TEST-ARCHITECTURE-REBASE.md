# Test Architecture Rebase — greenfield Python QA

**Founder decision:** 2026-08-07 · **Status:** Phase 128 REPLAN · **Priority:** P1

Это единственный канон фаз 128–141. Старый parity/dual-run план отменён:
переписывать component/E2E/static tests один-к-одному дольше и хуже, чем
спроектировать проверки заново по требованиям и рискам.

## 1. Неподвижный target layout

```text
backend/tests/<domain>/**       только service-free Go unit tests
frontend/tests/<domain>/**      только Vitest/RTL unit tests
frontend/tests/support/**       только support для frontend unit tests
testframework/                  один Python QA-проект в корне
  pyproject.toml
  {core,api,integration,contract,e2e,load,docker,reports}/**
```

- Из legacy сохраняются и перемещаются только доказанные genuine units.
- Все legacy non-unit Go tests, static/source guards, TS E2E, их helpers,
  configs, linkers и obsolete runners удаляются по exact allowlist.
- Новые API/integration/contract/E2E/load tests пишутся с нуля на Python.
- В `testframework/**` исполняемый test/load code только Python. YAML/TOML,
  Docker assets, docs, locks и отчёты разрешены.
- `pro-qa-ai/**` не читается, не меняется и не попадает в discovery/evidence.

## 2. Founder overrides

1. **Parity migration отменена.** Legacy case/file counts и assertion parity не
   являются acceptance gate. Ledger нужен только для exact retain/delete
   allowlists, provenance и доказательства отсутствия мусора.
2. **Go coverage ≥82% не блокирует rebase.** После сохранения genuine units
   публикуется новый честный baseline без искусственного удержания component
   tests. Coverage не подменяет requirement/risk coverage.
3. **Cutover ранний, не финальный.** После Phase 128 exact-manifest review
   Phase 129 физически переносит units, удаляет legacy non-units/runners,
   создаёт root Python scaffold и переключает entrypoints.
4. **Новые проверки не являются переводом старых.** Test design выводится из
   актуальных requirements, API/specs, задач, threat model, accessibility и
   observable runtime behavior.
5. **Один безопасный outer runner.** Все target jobs идут через versioned
   launcher с outer wall-clock timeout, exit-code preservation, task-owned
   resource names и unconditional exact cleanup в `finally`. До его отдельной
   проверки runtime gate `NOT_READY`; прямые snippets не fallback.
6. **Task-run reports isolated.** Каждый запуск пишет только в
   `testframework/reports/<task-run>/` после проверки owner marker с task и
   exact revision. Shared reports root и чужие artifacts read/write запрещены.

## 3. Обязательный контракт каждого нового Python test case

Каждый test case имеет стабильный `TC-*` ID и metadata/marker:

- `requirement`: `FR-*`, `NFR-*`, `GZ-*`, ADR/API section или approved spec;
- `task`: backlog/phase TASK, который ввёл или менял поведение;
- `risk`: защищаемый failure mode;
- `layer`, `domain`, `priority`, `owner`;
- предусловия, test data, observable oracle и cleanup;
- ссылка на human-readable case или generated catalog entry.

Для каждого применимого requirement/risk test-design checklist рассматривает:

1. happy path;
2. negative/invalid input;
3. authentication и authorization/tenant isolation;
4. boundaries и equivalence classes;
5. edge/corner cases, empty/null/duplicates/order/time/timezone/concurrency;
6. dependency failure, timeout, partial failure, retry/recovery/idempotency;
7. security/privacy/secret redaction;
8. accessibility для UI;
9. performance/resource behavior для hot paths.

`N/A` допустим только с кратким rationale. Количество тестов само по себе не
является качеством; gate — закрытые requirements/risks и чувствительность
oracle, подтверждённая temporary failing sentinel или mutation.

## 4. Docker-only execution

Все tests, validators, linters, build/vet, Vite, backend/frontend processes,
pytest, Playwright и Locust запускаются только в task-owned Docker
containers/Compose с CPU/RAM limits. Host — только чтение/редактирование,
Git и Docker CLI. Контейнер/volume/network/image удаляется сразу после задачи,
если явно не передан следующей фазе через owner+TTL handoff.

Минимум для каждого evidence: image digest, command, source SHA, service
profile, CPU/RAM limits, collect/pass/fail/skip, duration, artifact URI и
cleanup result. Неожиданный skip, zero collection или leaked container = FAIL.

## 5. Фазовая матрица

| Phase | Outcome | Main TASKs | Primary owner → verifier |
|---:|---|---|---|
| [128](../../phases/128-tf-python-scaffold/PLAN.md) | Founder contract + exact retain/delete manifests; TASK-202 archived/verified; ordered `cutover-manifest.json` | 202, 203, 204 | arch + qa-auto1 → arch-ver + qa-auto-ver |
| [129](../../phases/129-tf-e2e-shared/PLAN.md) | Seven ordered atomic plans: operation + affected CI/docs switch + manifest evidence per batch; 350 hook-free units move, 245 protected through TASK-234, legacy non-units deleted, Python root scaffold | 193, 197, 198, 201, 219, 234 | be/fe/qa/devops → independent verifiers |
| [130](../../phases/130-tf-traceability-gates/PLAN.md) | Traceability, test-case catalog, quality gates | 206, 218 | qa-auto1 + ba1 → qa-auto-ver + ba-ver |
| [131](../../phases/131-tf-python-core-auth/PLAN.md) | Python core + health/auth/account API foundation | 205, 207 | qa-auto1/2 → qa-auto-ver + security-auditor |
| [132](../../phases/132-tf-security-platform/PLAN.md) | AuthZ, tenant isolation, admin/platform and abuse cases | 208, 209, 227 | qa-auto1/2 → security-auditor + qa-auto-ver |
| [133](../../phases/133-tf-infrastructure-integration/PLAN.md) | Infra/data integration: PostgreSQL, Mongo, Redis, NATS, jobs | 210, 211 | qa-auto1/2 → qa-auto-ver + be-ver |
| [134](../../phases/134-tf-work-realtime/PLAN.md) | Work + realtime domains from requirements | 212, 213, 228 | qa-auto1/2 → qa-auto-ver + product-ver |
| [135](../../phases/135-tf-business-integrations/PLAN.md) | Business + integrations from requirements | 214, 215, 229 | qa-auto1/2 → qa-auto-ver + product-ver |
| [136](../../phases/136-tf-python-contracts/PLAN.md) | Contract/static/schema compatibility suite | 216, 230 | qa-auto1/2 → arch-ver + qa-auto-ver |
| [137](../../phases/137-tf-python-e2e-foundation/PLAN.md) | Python Playwright foundation + critical smoke journeys | 194, 217 | qa-auto1/2 → qa-auto-ver + security-auditor |
| [138](../../phases/138-tf-python-e2e-journeys/PLAN.md) | End-to-end domain journeys and recovery | 195, 196 | qa-auto1/2 → qa-auto-ver + qa-manual-ver |
| [139](../../phases/139-tf-accessibility-security/PLAN.md) | Accessibility + cross-cutting security/privacy | 231, 232 | qa-auto1 + security-auditor → qa-auto-ver |
| [140](../../phases/140-tf-load-resilience/PLAN.md) | Load/performance/resilience baselines | 200 | qa-auto2 + devops-engineer → qa-auto-ver |
| [141](../../phases/141-tf-ci-final-verification/PLAN.md) | CI quality gates, final audit/docs and post-rebase handoff | 220, 233 | devops-engineer + qa-auto-ver → team-lead |

Live slugs 130-141 follow phase outcomes. Historical snapshots using an old slug are **non-executable** and must resolve through this mapping before routing; old directories are not a fallback:

| Old snapshot slug | Live execution slug |
|---|---|
| `130-tf-e2e-mechanical` | `130-tf-traceability-gates` |
| `131-tf-e2e-rewrite` | `131-tf-python-core-auth` |
| `132-tf-backend-units-colocated` | `132-tf-security-platform` |
| `133-tf-frontend-units-colocated` | `133-tf-infrastructure-integration` |
| `134-tf-integration-python` | `134-tf-work-realtime` |
| `135-tf-locust-load` | `135-tf-business-integrations` |
| `136-tf-ci-cleanup` | `136-tf-python-contracts` |
| `137-tf-python-e2e-foundation` | unchanged |
| `138-tf-python-e2e-migration` | `138-tf-python-e2e-journeys` |
| `139-tf-load` | `139-tf-accessibility-security` |
| `140-tf-ci-dual-run` | `140-tf-load-resilience` |
| `141-tf-cutover-cleanup` | `141-tf-ci-final-verification` |

## 6. Dependency graph

```text
128 contract/manifests
  -> 129 immediate cutover/scaffold
      -> 130 traceability gates
          -> 131 core/auth API
              -> 132 security/platform
              -> 133 infra integration -> 134 work/realtime
                                       -> 135 business/integrations
              -> 136 contract/static
              -> 137 E2E foundation -> 138 E2E journeys -> 139 security/a11y cross-cutting
              -> 140 load/resilience
{132..140} -> 141 CI/final verification
```

После Phase 131 фазы 132, 133, 136 и 137 могут идти параллельно только при
непересекающихся owned paths/data. **Phases 134 и 135 формально зависят от
Phase 133.** Phase 138 зависит от 137 и green соответствующих 134/135 domains.
**Phase 139 formally depends on Phase 138** and therefore transitively on Phases 134/135.
Phase 140 зависит от стабильных API/auth fixtures и hot paths из 134–135.

## 7. Phase 128/129 destructive gate

Founder уже разрешил стратегию удаления; исполнитель всё равно обязан:

- пересчитать ledger на текущем SHA и получить exact POSIX row partitions:
  hook-free move, hook-dependent protected keep, frontend move и delete-no-port;
- reject absolute paths, traversal, symlinks/reparse points и protected zone;
- перед удалением доказать, что move/keep/delete sets disjoint и покрывают весь
  pinned scope; manual spot-check по каждому domain и ambiguous class;
- переносить genuine units без semantic rewrite, затем Docker-run их нового
  entrypoint; failure triage не восстанавливает component tests ради coverage;
- выполнять безопасную partition: 350 hook-free Go units переносятся сразу,
  245 hook-dependent Go units блокируются от deletion и проходят отдельный
  domain refactor/move (TASK-234); это не блокирует независимое удаление
  manifest-proven legacy non-units;
- удалять малыми атомарными commits по manifest batch; после каждого batch
  выполнять Docker-only static/collection gates;
- до Phase 129 создать `.planning/audits/test-rebase/cutover-manifest.json`;
  каждый move/delete batch обновляет его и переключает все затронутые
  CI/scripts/docs в **том же atomic commit**, поэтому committed state никогда
  не указывает на удалённый path; owner ставит `applied`, independent verifier
  — `verified`;
- удалить task-owned Docker resources после evidence capture;
- не push/merge/deploy без отдельного поручения founder.

Rollback — обычный revert конкретного атомарного commit; копии legacy tests в
дереве запрещены. Git history является recovery mechanism.

## 8. Definition of Done программы

- `backend/tests/**` содержит только доказанные service-free Go units.
- `frontend/tests/**` содержит только Vitest/RTL units/support.
- `testframework/**` — один root Python project без legacy executable code.
- Все Python cases проходят schema/traceability gate; orphan TC/requirement/task
  links отсутствуют или находятся в явно approved backlog.
- Required suites имеют non-zero collection, unexpected skip=0, sentinel
  sensitivity и Docker provenance; launcher enforced outer timeout, а все
  task-owned resources cleaned в `finally` с post-cleanup proof.
- Каждый report artifact находится только в `reports/<task-run>/` с валидным
  owner marker; shared-root output и cross-run overwrite отсутствуют.
- API/integration/contract/E2E/load coverage спроектирована заново по current
  requirements/risks с обязательным scenario checklist.
- GitHub is executable authority; зеркала либо используют те же scripts, либо
  явно retired. Final green получен после clean checkout/container build.
- STATE/COVERAGE/test docs синхронизированы с новым baseline без ложного claim
  о сохранении legacy coverage/parity.

После Phase 141: evidence-based re-audit B1–B6, затем productization
calendar → intranet → accounting → CRM → HR → bases, boundaries/schema
ownership и production acceptance.

---

*Канон обновлён по прямому founder decision: retain units, delete legacy
non-units, greenfield Python QA; parity и 82% migration gate отменены.*
