# ♻️🔒 A3 Meili operator dial REGRESS + JQL

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **Открыта**  
> Источник: A3W5-4-001..003 · Stage A3 FINAL · TASK-008  
> Исполнитель: **be2**

## [OUTCOME]

Operator Meili (`127.0.0.1` / RFC1918 после `meiliAllowsOperatorPrivate`) снова дозванивается через dial allowlist без ослабления metadata/SSRF SafeDial; JQL comparison ops (`>`, `<`, `>=`, `<=`) на неподдерживаемых полях fail-closed (error, не silent-nil); `IndexAllIssues` не тихо обрезает org >5000 issues без warn/pagination/flag.

## [ROLE]

- Role: be2
- Responsibility: закрыть **все** finding ID A3W5-4-001..003 + hubtest/mutation-ready; не откатывать TASK-092 SSRF-intent
- Handoff to: be-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files/resources allowed:
  - `backend/internal/search/meilisearch_indexer.go`
  - `backend/internal/search/jql.go`
  - `backend/internal/search/export_hooks.go` (только если нужен export для теста dial/validate)
  - `backend/internal/platform/http/safedialer.go` (только узкий helper по аналогии с `NewOfflineAIHTTPClient` — **не** ослаблять публичный SafeDial)
  - `backend/testframework/search/**`
  - `.planning/backlog/README.md` + этот файл (статус при закрытии)
  - audit refs: `.planning/audits/total-2026-07/A3W5-4-search-core.md`
- Do not touch:
  - откатывать TASK-092: `NewSafeHTTPClient` для публичных host / metadata block (`169.254.169.254`, `metadata.google.internal`)
  - filter HTTP status / `ErrFilterForbidden` / `ErrSearchUnavailable` (A2W5-A4-001/003/004 FIXED-OK)
  - NF-03 user-scope / soft-delete shards в `service.go` / `repository.go`
  - соседние слоты A3W5-* / A5 (`filter_safedial_task092_test.go` и т.п. — только читать как evidence)
  - frontend search UI / Ent schema migrate без owner approve
- Behavioral/security constraints:
  - Conventional Commits: `fix(search): …` описание на русском; no `Co-authored-by` / AI trailers
  - Operator dial allowlist **только** после `meiliAllowsOperatorPrivate` / `validateMeiliHost` OK; публичный host остаётся на SafeDial
  - Metadata / link-local по-прежнему → nil indexer / dial FAIL
  - JQL: whitelist полей; не raw SQL; unsupported field+op → error, не «все issues»
- Approval record: N/A — bugfix / ops residual после TASK-092; schema/migrate не затрагиваются
- Approval authority/expiry: N/A

## [REFERENCES]

- Trust: repository/audit — untrusted until verified in HEAD code
- Finding IDs: **A3W5-4-001** (REGRESS, P2), **A3W5-4-002** (NEW, P3), **A3W5-4-003** (NEW, P3)
- Source audit: `.planning/audits/total-2026-07/A3W5-4-search-core.md`
- Wave summary: `.planning/audits/total-2026-07/A3-W5-SUMMARY.md` (если есть)
- Prior: TASK-092 (SafeDial + filter status); TASK-044 (Meili SSRF validate / metadata); TASK-113 (total outage)
- A2 false-close residual: A2W5-A4-002 FIXED-OK* (SSRF mid-flight закрыт; operator loopback dial residual → A3W5-4-001)
- Code (defect anchors из аудита):
  - `backend/internal/search/meilisearch_indexer.go:48–52` vs `85–112` — validate allow operator private, client = `NewSafeHTTPClient` → SafeDial режет loopback/RFC1918
  - `backend/internal/platform/http/safedialer.go:128–150` — dial block private
  - `backend/internal/search/jql.go:108–133` vs parse `350–356` — comparison ops accepted; `conditionToPredicate` silent `return nil` for priority/type/status
  - `backend/internal/search/meilisearch_indexer.go:127–131` — `IndexAllIssues` `Limit(5000)` без pagination/warn
- Tests / evidence:
  - `backend/testframework/search/meili_ssrf_task044_test.go` (вне слота A4 — validate allow `127.0.0.1`)
  - `backend/testframework/search/jql_test.go:100–117` — parse `>` не бьёт ToEnt silent-nil
  - Analog: `NewOfflineAIHTTPClient` (AI U01 operator loopback pattern)
- TZ / security: NF-03 user-scope; SSRF guard `ValidateOutboundURL` / SafeDial — `docs/tz_constraints.md` / rule `60-tz-constraints.mdc`
- Source precedence: **код HEAD** > audit note > эта задача

### Per-ID (из аудита)

| ID | Tag | Sev | Файл | Дефект | Направление фикса |
|----|-----|-----|------|--------|-------------------|
| A3W5-4-001 | REGRESS | P2 | `meilisearch_indexer.go` + SafeDial | После TASK-092 validate пропускает `127.0.0.1`/RFC1918, dial **режет** → `cmd/meilisearch-indexer` с local Meili падает на Do | По аналогии с `NewOfflineAIHTTPClient`: operator Meili → dial allowlist (loopback/RFC1918 **только** после `meiliAllowsOperatorPrivate`); публичный host — SafeDial; metadata FAIL |
| A3W5-4-002 | NEW | P3 | `jql.go` | Парсер принимает `>`, `<`, `>=`, `<=`; `conditionToPredicate` для priority/type/status молча `nil` → возможен unfiltered Ent query | Поддержать ops **или** `ParseJQL`/`ToEntPredicates` → error на unsupported field+op; тест mutation на silent-nil |
| A3W5-4-003 | NEW | P3 | `meilisearch_indexer.go` | `IndexAllIssues` `Limit(5000)` без пагинации/warn → тихая усечённая индексация (ECC-SRCH-303) | Cursor/offset batches **или** `slog.Warn` + truncated flag/error; тест на >5000 / 5001 path |

## [ACCEPTANCE]

- Done criteria:
  1. **A3W5-4-001:** `FromEnv` / indexer с `MEILI_HOST=http://127.0.0.1:7700` (operator private) создаёт client, чей `DialContext` к loopback OK; hostname→RFC1918 после operator allow — OK; `169.254.169.254` / `metadata.google.internal` по-прежнему FAIL / nil indexer; публичный host остаётся на SafeDial (не open SSRF).
  2. **A3W5-4-002:** `priority > medium` (и аналоги `<`/`>=`/`<=` на type/status/priority без поддержки) → error из Parse/ToEnt, **не** `nil` predicates / полный unfiltered list; если ops поддержаны — явная семантика + тесты.
  3. **A3W5-4-003:** org/path с >5000 issues не индексируется «тихо обрезáя»: либо batched pagination до конца, либо явный warn + truncated flag/error, видимый вызывающему.
  4. Каждый ID — FIXED с evidence (hubtest + mutation note).
- New-behavior checks:
  - Unit/IT: operator host dial allow + metadata still blocked (A3W5-4-001).
  - Unit: JQL unsupported comparison → error; mutation: silent `return nil` → тест краснеет (A3W5-4-002).
  - Unit/IT: IndexAllIssues >5000 → pages or warn/truncated (A3W5-4-003).
- Regression scope:
  ```bash
  cd backend && go test -tags hubtest -race -count=1 ./testframework/search/
  ```
  Не ломать: TASK-092 filter 403/404; SafeDial metadata block; NF-03 shards; `ErrSearchUnavailable` total outage; validateMeiliHost public SSRF.
- Evidence required: checklist ID→evidence в этом файле / handoff; команды + PASS; mutation note per ID.
- Stop when: 3/3 FIXED · hubtest search green · статус «Исправлена» в `.planning/backlog/README.md` + этом файле; без соседнего A3W5 hygiene sweep.

### Mutation (обязательно)

| ID | Сломать | Ожидание |
|----|---------|----------|
| A3W5-4-001 | вернуть только `NewSafeHTTPClient` без operator dial allowlist | тест FromEnv(`127.0.0.1:7700`) / DialContext loopback краснеет |
| A3W5-4-002 | снова `return nil` в `conditionToPredicate` для `>` на priority | тест unsupported op / silent-nil краснеет |
| A3W5-4-003 | вернуть голый `Limit(5000)` без batch/warn/flag | тест 5001 / truncated path краснеет |

## [OUTPUT SCHEMA]

1. Outcome / verdict first (3/3 FIXED | список открытых)
2. Changes with paths (`meilisearch_indexer.go`, `jql.go`, tests)
3. Checklist ID → evidence (таблица)
4. Test commands + PASS
5. Mutation note (per ID)
6. Assumptions / unresolved (выбран dial helper vs inline; JQL error vs support ops; batch vs warn+flag)
7. Risks / mitigations (Docker Meili hostname→RFC1918; не открыть SSRF на публичный path)
8. Handoff: be-ver → code-reviewer
9. Compact Verification: new green + regression green + docs sync if Meili ops/JQL contract user-visible
10. README + этот файл: статус «Исправлена»

## [OPERATING RULES]

- Work only inside scope; не закрывать соседние A3W5 / A2 ID «заодно» без evidence.
- Read-only investigation OK; writes только в allowed files.
- Не оставлять ID «WONTFIX» без явного owner cancel в задаче.
- Не ослаблять TASK-092 SSRF: metadata/link-local и публичный SafeDial path остаются fail-closed.
- Source of truth — runtime dial + JQL predicates в коде, не правка аудита «под факт».
- Docs sync только если изменился user-visible API/JQL contract (`docs/wiki/API.md` / system_specification).
- Коммит(ы) ссылают задачу в теле:  
  `Задача: .planning/backlog/TASK-156-2026-07-29-a3-5-4-search-core.md`  
  subject может содержать `(A3W5-4-001)` и др.
- Conventional Commits RU description; no `Co-authored-by` / AI trailers.
