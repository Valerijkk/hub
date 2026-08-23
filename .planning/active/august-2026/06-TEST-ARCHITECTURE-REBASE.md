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
  src/hub_testkit/**            future reusable installable framework package
  tests/{contract,api,integration,e2e}/**
                                future executable SUT pytest suites
  selftests/**                  framework/package contracts; never SUT behavior
  cases/**                      repository-owned TC-* intent and risk linkage
  schemas/**                    closed machine-readable case/run/ownership contracts
  policies/**                   marker, environment, redaction and profile policy data
  load/**                       separate bounded Locust lifecycle
  docker/**                     immutable execution assets/container contracts only
  runner/**                     sole launcher boundary; private static candidate exists
  reports/<task-run>/**         generated owner-marked evidence for one run
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
   расширяет уже зафиксированный structure-only root scaffold и переключает
   entrypoints.
4. **Новые проверки не являются переводом старых.** Test design выводится из
   актуальных requirements, API/specs, задач, threat model, accessibility и
   observable runtime behavior.
5. **Один безопасный outer runner.** Все target jobs идут через versioned
   launcher с outer wall-clock timeout, exit-code preservation, task-owned
   resource names и unconditional exact cleanup в `finally`. Private source
   candidate существует и принят только статически; до binding и separate
   parser/private-contract/Docker acceptance launcher = `NOT_READY`,
   `PS5.1_PARSE_NOT_RUN`, `PRIVATE_CONTRACT_NOT_RUN`, runtime = `NOT_RUN`; прямые
   snippets не fallback.
6. **Task-run reports isolated.** Каждый запуск пишет только в
   `testframework/reports/<task-run>/` после проверки owner marker с task и
   exact revision. Shared reports root и чужие artifacts read/write запрещены.

### 2.1 Founder maximum-coverage source-authoring override (2026-08-15)

Founder authorizes the ordered source-authoring waves in
[`07-TESTFRAMEWORK-MAX-COVERAGE-IMPLEMENTATION.md`](07-TESTFRAMEWORK-MAX-COVERAGE-IMPLEMENTATION.md):
F0 package/bootstrap/traceability/plugin/health/framework selftests; F1
HTTP/identity/authz/platform/OpenAPI/security API and contract surfaces; F2
implemented PostgreSQL/Mongo/Redis/NATS/storage/WebSocket/work/business
spine; F3 browser/a11y/privacy/load source. This narrowly supersedes the
former `G129-C`, `G130-H`, and per-feature `G-ACTIVATE-*` predecessor stops
**only while authoring and independently reviewing source**.

It does not accept Phase 129 recovery, a private launcher, a profile run,
dependency resolution, an image, Docker execution, CI, or runtime behavior.
Package, launcher, and every profile remain `NOT_READY`; runtime remains
`NOT_RUN` until one exact revision has separately accepted A1 versioned-runner
admission, R Docker resolver admission, immutable image evidence, generated
`uv.lock`, and Docker-only evidence. A source file, source review, or source
commit is never runtime evidence.

The scope is implemented observable behavior only. Magic-link TTL remains out
until a product decision; single-tenant deployments must not gain invented
true-tenant assertions. Current authenticated Calendar aggregation/export and
implemented call API authz/signature behavior remain in scope. Calendar-v2
writes/resources/availability/subscriptions, Wiki, WebRTC/media capacity, and
other unimplemented call proposals remain planned-only until delivery,
accepted API/UI behavior, fixture ownership, and observable oracles exist.
This override creates no authority for Phase 129 deletion/cutover or deployment
work.

## 3. Обязательный контракт каждого нового Python test case

Каждый новый test case имеет стабильный ID, matching
`^TC-[A-Z][A-Z0-9]*(?:-[A-Z][A-Z0-9]*)*-[0-9]{4}$`, и metadata/marker.
`TC-HEALTH-0001` is canonical. Historical `TC-AUTH-03` examples and
`HUB-TC-*` inventory identifiers do not match the new greenfield grammar and
are not accepted for new `testframework/cases/**` records:

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

> **Current recovery override (2026-08-14):** founder selected forward-only
> Option A after the current worktree accumulated bulk moves. The prescribed
> per-batch execution history was not produced. `STATIC_PATH_MAPPING=PASS`,
> but `CUTOVER_ACCEPTANCE=FAIL/BLOCKED` and Phase 129 is `RECOVERY_BLOCKED`.
> The immutable v1 manifest stays `PLANNED_UNCHANGED`; no receipt, verifier
> verdict, phase completion, or runtime result may be inferred from the bulk
> commits. The 241 staged legacy deletions must be handled in explicit ordered
> successor work, never accepted as one bulk deletion. See
> `../../audits/test-rebase/RECOVERY-OPTION-A-2026-08-14.md`.

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

For source authoring only, the ordered founder waves are an overlay on this
graph, not evidence that any predecessor runtime gate passed:

```text
F0 package + traceability + plugin + health selftests
  -> F1 HTTP + identity + authz/platform + OpenAPI/security
      -> F2 implemented integration/API spine
          -> F3 browser + a11y + privacy + load source
              -> A1 + R + image + lock + Docker evidence (separate runtime admission)
```

The literal 47-profile registry continues to be owned by the live Phase 130
plan; its registration does not authorize a run or create a 48th profile.
Source authoring uses only profile identities already in that registry. The
detailed maximum-risk suite map, exact source slices, commit/review order, and
exclusions are canonical in the 2026-08-15 plan linked above.

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
