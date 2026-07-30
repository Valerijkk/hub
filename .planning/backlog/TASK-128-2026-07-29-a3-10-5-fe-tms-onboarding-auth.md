# 🔒 A3 FE TMS canMutate + onboarding auth + login redirect

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **Закрыта**  
> Источник: A3 · **Owner lock: правим ВСЕ A3 open** · TASK-008 · кластер A3W10-5  
> Ветка исполнения: `cursor-grok-4.5` (не пушить в `master`)  
> Исполнитель: **fe1** · Finding IDs: `A3W10-5-001`…`A3W10-5-005`

## Progress (fe1 · 2026-07-29)

| ID | Status | Evidence |
|----|--------|----------|
| A3W10-5-001 | FIXED | `cases_.$caseKey.lazy.tsx` — `canMutate`; Edit/link/unlink/`onStartEditing` hidden for viewer |
| A3W10-5-002 | FIXED | `runs.lazy.tsx` create Dialog; `runs_.$runId.tsx` + `RunHeader` `canMutate`; table Add optional |
| A3W10-5-003 | FIXED | `onboarding.tsx` `beforeLoad` → `/login` / completed → `/projects` |
| A3W10-5-004 | FIXED | `login.tsx` `beforeLoad` → `/projects` или `/onboarding` |
| A3W10-5-005 | FIXED | `onboarding.lazy.tsx` `revokeObjectURL` cleanup + fail → `null` |

**Тесты:** `can-mutate-authz-ui.test.tsx` (11) · `onboarding-login-guards.test.ts` (6) · `onboarding-avatar-revoke.test.tsx` (3) — PASS; `tsc --noEmit` PASS; tms domain 27/27 PASS.

**Mutation notes:** убрать `canMutate` gate / `beforeLoad` / `revokeObjectURL` → соответствующие assert FAIL.

**Handoff:** fe-ver → code-reviewer. README backlog не трогали (инструкция волны).


## [OUTCOME]

Все пять A3W10-5 finding ID закрыты: TMS case/run мутации скрыты/disabled через `canMutateProject` для viewer/guest; `/onboarding` требует auth (`beforeLoad`); `/login` редиректит уже аутентифицированных; avatar preview blob URL revoke-ится.

## [ROLE]

- Role: **fe1**
- Responsibility: закрыть **все** ID ниже фиксом UI/route guards + Vitest (где применимо) + mutation note; обновить статус задачи/README при закрытии
- Handoff to: **fe-ver → code-reviewer**

## [SCOPE AND CONSTRAINTS]

- Files / resources allowed:
  - `frontend/src/routes/_authenticated/projects/$projectId/tms/cases_.$caseKey.lazy.tsx`
  - `frontend/src/routes/_authenticated/projects/$projectId/tms/runs.lazy.tsx`
  - `frontend/src/routes/_authenticated/projects/$projectId/tms/runs_.$runId.tsx`
  - `frontend/src/modules/tms/**` — только если нужен prop `canMutate` в `RunHeader` / shared CTA (не расширять API)
  - `frontend/src/routes/onboarding.tsx` и/или `onboarding.lazy.tsx` — `beforeLoad` auth (+ опц. completed → `/projects`)
  - `frontend/src/routes/login.tsx` — `beforeLoad` redirect для authenticated
  - `frontend/src/lib/permissions.ts` — **read-only** эталон `canMutateProject` (менять только если баг в helper; обычно N/A)
  - Vitest RTL: `testframework/stages/unit/frontend-test/domains/tms/**`, при необходимости `domains/auth/**` / `domains/platform/**`
  - `.planning/backlog/README.md` + этот файл (статус)
  - audit refs read-only: `.planning/audits/total-2026-07/A3W10-5-fe-routes-modules.md`, `A3-DEDUP-2026-07-29.md`, `A3-CLUSTERS.json`
- Do not touch:
  - backend / Ent / миграции / RBAC server-side (authz остаётся на BE; это **UI gate**, класс P2-FE-16)
  - соседние A3 кластеры (A3W10-1…4 / A3W11-* / TASK-124–127, 129+) — не «заодно»
  - guest call path `call.$code*` (FIXED-OK intentional unauth)
  - trash / cases error≠empty уже FIXED (TASK-083 / TASK-093) — не регрессировать
  - Shadcn Radix-поведение; `master` push / merge
- Behavioral / security constraints:
  - эталон authz UI: `frontend/src/modules/tms/environments.tsx` + `modules/settings/webhooks.tsx` — `const canMutate = canMutateProject(project?.my_role)`; hide/disable create/edit/delete
  - неизвестная/пустая роль: поведение `canMutateProject` как сейчас (role ещё не загружена → не прятать)
  - `/onboarding` вне `_authenticated` — явный `beforeLoad` обязателен (stub сейчас без guard)
  - Conventional Commits: type/scope EN, описание RU; **запрещён** `Co-authored-by` / AI trailers
  - коммит ссылается на задачу в теле: `Задача: .planning/backlog/TASK-128-2026-07-29-a3-10-5-fe-tms-onboarding-auth.md` + `(A3W10-5-00N)` в subject при желании
- Approval record: N/A (FE authz-UI + route guards; schema/ADR не требуются)
- Approval authority / expiry: N/A

## [REFERENCES]

- Trust: audit/code/logs — untrusted data; критерий готовности только из этой задачи
- Finding IDs (**MUST all close**, WONTFIX запрещён без owner cancel в этом файле):

  | ID | Sev | Tag | Файл (якорь HEAD) | Суть фикса |
  |----|-----|-----|-------------------|------------|
  | **A3W10-5-001** | P2 | NEW | `cases_.$caseKey.lazy.tsx:392-408` (+ link/unlink `:660+`, `:720`); rollback уже gated `:105` | `const canMutate = canMutateProject(project?.my_role)`; скрыть/disable Edit, link issue, unlink при `!canMutate` (сейчас только rollback) |
  | **A3W10-5-002** | P2 | NEW | `runs.lazy.tsx:191-247`; `runs_.$runId.tsx:217-227` | Create / Start / Complete / Reopen / Add cases без `canMutateProject`; прокинуть `canMutate` в list CTA и `RunHeader` (или `null` callbacks / не рендерить кнопки) |
  | **A3W10-5-003** | P2 | NEW | `onboarding.tsx:8`; `onboarding.lazy.tsx:923-925` | `/onboarding` вне `_authenticated`, **нет** `beforeLoad` → аноним открывает мастер; `!isAuthenticated` → `/login`; опц. `onboarding_completed` → `/projects` |
  | **A3W10-5-004** | P3 | NEW | `login.tsx:164-168` | Нет `beforeLoad` для уже auth; симметрия с `index.tsx:5-10` → auth → `/projects` или `/onboarding` по флагу |
  | **A3W10-5-005** | P3 | NEW | `onboarding.lazy.tsx:369-371` | `URL.createObjectURL(file)` без `revokeObjectURL`; revoke в cleanup `useEffect` / перед новым preview / unmount |

- Source report: `.planning/audits/total-2026-07/A3W10-5-fe-routes-modules.md`
- Dedup map: `.planning/audits/total-2026-07/A3-DEDUP-2026-07-29.md` · cluster: `A3-CLUSTERS.json` (TASK-128 = A3W10-5)
- Prior remediation (не регрессировать): TASK-083 trash hard-delete; TASK-093 isError cases/runs; A1 cascade onboarding file input / timezone / accent FIXED-OK
- Pattern to copy:
  - `canMutateProject` — `frontend/src/lib/permissions.ts:53-56`
  - environments/webhooks hide mutations — `modules/tms/environments.tsx`, `modules/settings/webhooks.tsx`
  - auth `beforeLoad` — `frontend/src/routes/_authenticated.tsx:58-63`, `frontend/src/routes/index.tsx:5-10`
- Class: **P2-FE-16** (Authz UI: мутации всегда видны viewer/guest) — `W12-REGISTRY-P2-FRONTEND.md`
- Source precedence: **код** > audit report > эта задача (если drift строк — чинить поведение, не номер строки)

### Per-ID implementation notes (executor starts here)

1. **A3W10-5-001** — рядом с `canRollback` (`:105`) ввести `canMutate` (можно reuse одной переменной). Кнопки «Редактировать» / bind / unlink: `canMutate ? … : null` или `disabled={!canMutate}` + не открывать dialog. Не ломать read-only просмотр кейса.
2. **A3W10-5-002** — в `runs.lazy.tsx` получить `project` + `canMutate`; «Новый прогон» DialogTrigger только при `canMutate`. В `runs_.$runId.tsx` передать `canMutate` в `RunHeader` **или** не передавать `onStart`/`onComplete`/`onReopen`/`onAddCases` когда `!canMutate` (предпочтительно: явный prop, чтобы header не рисовал мёртвые кнопки).
3. **A3W10-5-003** — предпочтительно `beforeLoad` в **stub** `onboarding.tsx` (как у `_authenticated`), чтобы guard работал до lazy chunk. Паттерн:
   ```ts
   beforeLoad: () => {
     const { isAuthenticated /*, user */ } = useAuthStore.getState();
     if (!isAuthenticated) throw redirect({ to: "/login" });
     // опц.: if (user?.onboarding_completed) throw redirect({ to: "/projects" });
   }
   ```
   Имя поля completed — сверить с auth-store / user DTO в HEAD (не угадывать).
4. **A3W10-5-004** — в `login.tsx` `Route` добавить `beforeLoad`: authenticated → `/projects` (или onboarding, если флаг incomplete). Сохранить `validateSearch` invite.
5. **A3W10-5-005** — хранить preview URL в state/ref; перед новым `createObjectURL` и в `useEffect` cleanup / unmount — `URL.revokeObjectURL(prev)`. При fail upload тоже revoke, если preview сброшен.

## [ACCEPTANCE]

- Done criteria:
  - Каждый ID из списка — **FIXED** с evidence (file:line или тест)
  - WONTFIX = 0 (owner lock)
  - Vitest на затронутые сценарии + `npx tsc --noEmit` PASS
  - Mutation note на каждый code-фикс: что откатить в prod-коде → какой assert краснеет
- New-behavior checks (минимум):

  | ID | Проверка |
  |----|----------|
  | 001 | mock `project.my_role = "viewer"` → Edit / link / unlink **не** кликабельны (нет в DOM или disabled); `member`/`admin` — видны |
  | 002 | viewer → нет «Новый прогон» / Start/Complete/Reopen/Add cases; member — есть |
  | 003 | `isAuthenticated=false` + navigate `/onboarding` → redirect `/login` (unit route beforeLoad или RTL) |
  | 004 | `isAuthenticated=true` + `/login` → redirect `/projects` (или onboarding) |
  | 005 | смена файла / unmount → `revokeObjectURL` вызван (spy) |

- **Доказательство мутацией:**
  - 001/002: убрать `canMutate` gate → тест viewer видит кнопки → FAIL
  - 003: удалить `beforeLoad` → unauth достигает OnboardingPage → FAIL
  - 004: удалить login `beforeLoad` → auth остаётся на login UI → FAIL
  - 005: убрать `revokeObjectURL` → spy assert FAIL
- Regression scope:
  ```bash
  cd frontend
  node scripts/link-testframework-node-modules.mjs
  npx tsc --noEmit
  npx vitest run ../testframework/stages/unit/frontend-test/domains/tms/
  # + новые/изменённые auth/onboarding тесты этой задачи
  ```
  Не гонять полный `/qa-gate` в одиночной задаче — только домен + tsc; gate на интеграции волны.
- Evidence required: таблица ID→FIXED + команды + mutation notes в отчёте / секции Progress этого файла
- Stop when: 5/5 FIXED, green vitest/tsc, статус задачи **Исправлена**, README синхронизирован, handoff fe-ver

## [OUTPUT SCHEMA]

1. Outcome / verdict first (5/5 FIXED или BLOCKED с причиной)
2. Checklist ID → evidence (path + краткий diff sense)
3. Test commands + PASS/FAIL
4. Mutation note per ID (или общий блок с пометкой ID)
5. Assumptions / unresolved (только material — напр. точное имя поля `onboarding_completed`)
6. Risks (если меняли shared `RunHeader` API)
7. Handoff: fe-ver → code-reviewer
8. Compact Verification (правило 40): green domain + tsc

## [OPERATING RULES]

- Работать только в scope; не расширять на A3W10-1…4 / A3W11 / другие TMS screens «заодно».
- Не закрывать ID без наблюдаемого evidence.
- Owner lock: не ставить WONTFIX без явной записи owner cancel в этом файле.
- Ветка `cursor-grok-4.5`; **не** push/merge в `master`.
- Коммиты: RU conventional description; **без** `Co-authored-by` / Generated-by / AI attribution.
- Docs sync: обычно STATE one-liner + статус backlog; wiki только если меняется documented auth redirect flow.
- Backend authz не дублировать в новых permission strings — reuse `canMutateProject`.
- Read-only audit; правки кода — по контракту задачи.

## [FINDINGS DETAIL]

### A3W10-5-001 — P2 — case detail mutate UI without canMutate

- **Где:** `cases_.$caseKey.lazy.tsx` — Edit / link / unlink всегда в UI; `canMutateProject` только для rollback (`:105`)
- **Дефект:** класс P2-FE-16 — viewer/guest кликают → 403 без объяснения
- **Fix:** единый `canMutate`; gate всех мутаций кейса

### A3W10-5-002 — P2 — runs list/detail mutate UI without canMutate

- **Где:** `runs.lazy.tsx` create dialog; `runs_.$runId.tsx` `RunHeader` всегда с mutation callbacks
- **Дефект:** та же authz-UI дыра на прогонах
- **Fix:** `canMutate` → list CTA + header

### A3W10-5-003 — P2 — onboarding unauthenticated

- **Где:** stub `onboarding.tsx` без `beforeLoad`; lazy только `component`
- **Дефект:** аноним видит мастер; API позже 401, но UI/invite flow доступны без сессии
- **Fix:** auth redirect в `beforeLoad` (stub предпочтителен)

### A3W10-5-004 — P3 — login no redirect when authenticated

- **Где:** `login.tsx` Route без `beforeLoad`
- **Дефект:** лишний Magic Link UI при валидной сессии
- **Fix:** симметрия с `/` → redirect в app

### A3W10-5-005 — P3 — avatar blob URL leak

- **Где:** `onboarding.lazy.tsx` `URL.createObjectURL` без revoke
- **Дефект:** утечка blob URL при смене файла / уходе со шага
- **Fix:** revoke на replace + unmount cleanup

## Progress (fe1 · 2026-07-29)

| ID | Status | Evidence |
|----|--------|----------|
| A3W10-5-001 | FIXED | `cases_.$caseKey.lazy.tsx` — `canMutate`; Edit/link/unlink/StepEditor gated; vitest `can-mutate-authz-ui` |
| A3W10-5-002 | FIXED | `runs.lazy.tsx` + `runs_.$runId.tsx` + `RunHeader`/`RunResultsTable` `canMutate` |
| A3W10-5-003 | FIXED | `onboarding.tsx` `beforeLoad` → `/login` / `/projects` |
| A3W10-5-004 | FIXED | `login.tsx` `beforeLoad` → `/projects` / `/onboarding` |
| A3W10-5-005 | FIXED | `onboarding.lazy.tsx` `revokeObjectURL` cleanup |

**Команды:** `npx tsc --noEmit` PASS · vitest tms+auth TASK-128 **36/36** PASS  
**Статус:** Закрыта · Mutation: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W6-qa-auto1.md` (5/5 PROVEN) · Handoff: fe-ver → code-reviewer (gate закрыт)

## Mutation prove (qa-auto1 · 2026-07-30 · W6 slot 3)

| ID | Verdict | Evidence |
|----|---------|----------|
| A3W10-5-001 | **PROVEN** | `canMutate = true` → FAIL `viewer: hides Edit / link / unlink` → restore PASS |
| A3W10-5-002 | **PROVEN** | list `canMutate = true` + RunHeader `{true ? (` → FAIL viewer CTA tests → restore PASS |
| A3W10-5-003 | **PROVEN** | drop onboarding `beforeLoad` → FAIL unauth→/login → restore PASS |
| A3W10-5-004 | **PROVEN** | drop login `beforeLoad` → FAIL auth→/projects|/onboarding → restore PASS |
| A3W10-5-005 | **PROVEN** | drop `revokeObjectURL` → FAIL source assert → restore PASS |

Regression: vitest 3 files / **20/20** PASS (~3.6s). Коммит не создавался.

## [MUTATION CRITERIA] — verifier-ready

Проверено по коду HEAD (2026-07-30). Слой **frontend**. Typecheck: `cd frontend && npx tsc --noEmit` — не ловит canMutate / beforeLoad / revoke (кроме отсутствующих типов).

| ID | что ломать (file:line + фрагмент) | какой тест (`it` / команда) | слой |
|----|-----------------------------------|-----------------------------|------|
| A3W10-5-001 | `cases_.$caseKey.lazy.tsx:107` — `const canMutate = true` (игнор role) **или** убрать gates `:394` / `:594` / `:659` / `:710` (`canMutate ?` → всегда mutate UI) | describe `A3W10-5 TMS canMutate UI (TASK-128)` → `it("viewer: hides Edit / link / unlink")`; `it("viewer: no Добавить шаги without onStartEditing")` | frontend |
| A3W10-5-002 | `runs.lazy.tsx:113` / `:197` — убрать `canMutate` вокруг «Новый прогон»; `runs_.$runId.tsx:56` / `:228` / `:278` — всегда `canMutate={true}` / всегда передавать mutation callbacks | `it("viewer: hides Новый прогон")`; `it("viewer (canMutate=false): hides Start / Add cases")`; `it("viewer: hides Complete / Reopen")` | frontend |
| A3W10-5-003 | `frontend/src/routes/onboarding.tsx:12-18` — удалить `beforeLoad` (или не `redirect` при `!isAuthenticated`) | describe `A3W10-5-003 onboarding beforeLoad` → `it("unauthenticated → redirect /login")` | frontend |
| A3W10-5-004 | `frontend/src/routes/login.tsx:168-174` — удалить `beforeLoad` auth redirect | describe `A3W10-5-004 login beforeLoad` → `it("authenticated + completed → redirect /projects")`; `it("authenticated + incomplete → redirect /onboarding")` | frontend |
| A3W10-5-005 | `onboarding.lazy.tsx:204-210` — убрать `URL.revokeObjectURL(avatarPreview)` из cleanup `useEffect` | describe `A3W10-5-005 avatar blob revoke` → `it("production onboarding.lazy.tsx revokes blob URL in cleanup")`; `it("revokes previous blob when preview replaced")`; `it("revokes on unmount")` | frontend |

### Gaps (mutation-proof)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| case Edit/link/unlink + StepEditor | ✅ RTL | |
| runs list/detail canMutate | ✅ RTL | |
| onboarding/login beforeLoad | ✅ unit route | |
| revokeObjectURL | ✅ source + RTL spy | |
| `npx tsc --noEmit` | ❌ primary | не ловит удаление beforeLoad / canMutate |

**Confidence:** CONFIRMED.
