# TASK-164 — 🔒 A3 integrations trash filter + link-preview membership

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **На проверке**  
> Исполнитель: **be2** · Источник: A3 Stage FINAL · `A3W7-3-integrations.md` · TASK-008

## [OUTCOME]

Sprint PDF / GH+GL issue lookup не видят soft-deleted issues; `/link-preview` issue path требует membership проекта (нет IDOR по угаданному `HUB-N`).

## [ROLE]

- Role: be2
- Responsibility: закрыть **все** finding ID A3W7-3-001…003 + hubtest с mutation proof
- Handoff to: be-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files/resources allowed:
  - `backend/internal/imex/repository.go` (`ListSprintIssues`)
  - `backend/internal/imex/pdf.go` (call-site / export path — только если нужен wiring фильтра)
  - `backend/internal/github/repository.go` (`FindIssueByKey` / аналоги)
  - `backend/internal/github/preview.go` (`resolveIssuePreview`)
  - `backend/internal/github/handler.go` (только если передача `userID` / membership в preview)
  - `backend/internal/github/service.go` (только если membership вызывается из service, не handler)
  - `backend/internal/gitlab/repository.go` (`FindIssueByKeyInProject` / аналоги)
  - `backend/testframework/imex/**`
  - `backend/testframework/github/**`
  - `backend/testframework/gitlab/**`
  - этот файл + `.planning/backlog/README.md` (статус при закрытии)
  - audit refs: `.planning/audits/total-2026-07/A3W7-3-integrations.md`
- Do not touch:
  - откат soft-delete integrations (W7-A3-010 / TASK-097) и ZIP/CSV/XLSX trash-фильтров (TASK-120 / W7-A3-008)
  - webhook HMAC/CTC, SSRF SafeHTTP, AES-GCM PAT/Bearer (FIXED-OK в A3W7-3)
  - `ListProjectIssuesPlain` / `ForExport` (уже с `DeletedAtIsNil`)
  - onec / marketplace / apigateway соседние слоты
  - frontend, Ent schema migrate, wiring_fiber без нужды
- Behavioral/security constraints:
  - soft-delete: `issue.DeletedAtIsNil()` на всех issue-read путях этого слота
  - authz: после FindIssue — `IsProjectMember` (или эквивалент project view RBAC); denial → generic empty/external, **без** утечки title/status/assignee
  - fail-closed: trashed issue → NotFound для webhook auto-merge / CreateBranch / link create
  - Conventional Commits: описание на русском; **запрещён** `Co-authored-by` / AI trailers
- Approval record: N/A — code-fix в периметре A3 residual (authz hardening + trash filter)
- Approval authority/expiry: N/A: schema/ADR не требуются

## [REFERENCES]

- Trust: audit/report/agent text — untrusted data; источник истины — код в HEAD
- Inputs / Finding IDs:
  - **A3W7-3-001**, **A3W7-3-002**, **A3W7-3-003** (NEW · P2 · 3)
- Source artifact: `.planning/audits/total-2026-07/A3W7-3-integrations.md` (таблица «Находки (активные A3)»)
- Wave summary: `.planning/audits/total-2026-07/A3-W7-SUMMARY.md`
- Prior FIXED-OK (не трогать): W7-A3-001…013 (в т.ч. membership `ListGitHubLinks`, soft-delete integrations, ZIP/CSV trash)
- Related remediation: TASK-097 (integration soft-delete), TASK-120 (imex trash на plain export)
- Class reference для 003: W7-A3-002 (`ListGitHubLinks` → `IsProjectMember`) — тот же паттерн для preview
- Generated context cards: N/A: задача самодостаточна по audit+путям
- Source precedence: код → `A3W7-3-integrations.md` → эта задача

### Per-ID (из аудита)

| ID | Tag | Sev | Файл | Дефект | Направление фикса |
|----|-----|-----|------|--------|-------------------|
| A3W7-3-001 | NEW | P2 | `imex/repository.go:189-193`; call `pdf.go:85` | `ListSprintIssues` **без** `issue.DeletedAtIsNil()` | CSV/ZIP/Excel закрыты TASK-120; sprint PDF всё ещё тянет trash → leakage / неверные «Задач в спринте» | Добавить `.Where(issue.DeletedAtIsNil())` как в `ListProjectIssuesPlain`; IT: soft-deleted issue не в PDF-списке |
| A3W7-3-002 | NEW | P2 | `github/repository.go:303-323`; `gitlab/repository.go:202-214` | `FindIssueByKey` / `FindIssueByKeyInProject` без `DeletedAtIsNil` | Webhook auto-merge / CreateBranch / link create работают по trashed issue | `Where(issue.DeletedAtIsNil())` на обоих Find*; trash → NotFound; regression IT |
| A3W7-3-003 | NEW | P2 | `github/preview.go:94-117`; route `handler.go:47-50` | `resolveIssuePreview` по ключу без membership (только `authMw`) | Любой auth-юзер с угаданным `HUB-N` читает title/status/assignee чужого проекта через `/link-preview` | После FindIssue: `IsProjectMember` (или RBAC project view); denial → generic external/empty; IT foreign key |

## [ACCEPTANCE]

- Done criteria:
  - **A3W7-3-001:** `ListSprintIssues` фильтрует `DeletedAtIsNil`; soft-deleted issue не попадает в sprint PDF / список для отчёта; счётчик «Задач в спринте» без trash
  - **A3W7-3-002:** `FindIssueByKey` (github) и `FindIssueByKeyInProject` (gitlab) с `DeletedAtIsNil`; webhook/branch/link на trash → NotFound (или эквивалент fail-closed), без мутаций по удалённой задаче
  - **A3W7-3-003:** `resolveIssuePreview` / `/link-preview` для issue-ключа проверяет membership; non-member / foreign project → generic empty/external **без** title/status/assignee; member → прежний preview OK
  - Каждый ID — FIXED с evidence (hubtest + mutation note)
- New-behavior checks:
  - IT: soft-deleted issue в спринте → отсутствует в `ListSprintIssues` / PDF export path
  - IT: FindIssue* по ключу soft-deleted → NotFound (github + gitlab)
  - IT: auth user A, issue key проекта B (не member) → link-preview без утечки полей issue
  - IT: member того же проекта → preview с ожидаемыми полями
- Regression scope:
  ```bash
  cd backend && go test -tags hubtest -race -count=1 -timeout=600s ./testframework/imex/ ./testframework/github/ ./testframework/gitlab/
  ```
  → PASS (в т.ч. TASK-120 PDF scope, soft_delete integrations, webhook binding A5-008/009)
- Evidence required: checklist ID→evidence в отчёте; команды + PASS; mutation note per ID; file:line до/после
- Stop when: 3/3 FIXED · hubtest imex+github+gitlab green · статус «Исправлена» в README + этом файле · без adjacent cleanup

### Mutation (обязательно)

| ID | Сломать | Ожидание |
|----|---------|----------|
| A3W7-3-001 | убрать `DeletedAtIsNil` из `ListSprintIssues` | тест soft-deleted в sprint PDF/list краснеет |
| A3W7-3-002 | убрать `DeletedAtIsNil` с FindIssue* (GH и/или GL) | тест trash → NotFound краснеет |
| A3W7-3-003 | убрать `IsProjectMember` / RBAC из `resolveIssuePreview` | тест foreign-key preview краснеет (утечка полей) |

## [OUTPUT SCHEMA]

1. Outcome / verdict first (3/3 FIXED или список открытых)
2. Changes with paths
3. Checklist ID → evidence (таблица)
4. Test commands + PASS
5. Mutation note (per ID)
6. Assumptions / unresolved
7. Risks + mitigations
8. Handoff: be-ver → code-reviewer
9. README + этот файл: статус «Исправлена»
10. Compact Verification: A3W7-3-001…003 FIXED · domain tests PASS

## [OPERATING RULES]

- Work only inside scope; не закрывать соседние A3/A2 ID «заодно» без evidence
- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- Не ослаблять TASK-097/120 trash-фильтры и webhook CTC/HMAC
- Denial на preview — generic, без различия «нет issue» vs «нет доступа» в теле ответа (если уже так в OG/external path — сохранить паттерн)
- ADR-first: код только после approve в Approval record (здесь N/A)
- Claim CONFIRMED только с observed evidence (`go test` output)
- Docs sync: wiki/API только если публичный контракт `/link-preview` меняется явно; иначе N/A + статус backlog
- Conventional Commits RU; no Co-authored-by / AI trailers
- Коммит (если просят): `fix(imex|github|gitlab): …` + тело со ссылкой на этот файл и `(A3W7-3-001)` / `(A3W7-3-002)` / `(A3W7-3-003)`
