# Reduced GitHub Module — blueprint (расширенный)

> Продуктово-архитектурный blueprint **урезанного аналога GitHub-flow** внутри корпоративного хаба.  
> **Код не пишем** в горизонте AI-expansion (код-freeze продукта). Это карта для будущей фазы реализации.  
> Канон микросервисов: `.planning/decisions/ADR-ARCH-MICROSERVICES-2026.md`.  
> Связанные: [`04-PORTABLE-HUB-ARCH.md`](04-PORTABLE-HUB-ARCH.md), [`05-PORTABLE-KIT.md`](05-PORTABLE-KIT.md),  
> существующий интеграционный модуль `backend/internal/github/` (forge adapter, не замена).

- **Статус:** Blueprint / docs-only (2026-07)
- **Владелец:** product + arch · **Верифицирует:** arch-ver (+ sa при детализации API)
- **Граница:** process-of-record PR/checks/review **в Hub**; git storage остаётся на внешнем forge (GitHub/GHE/GitLab)

---

## 1. Цель и JTBD

Закрыть JTBD команды 10–50: **задача → ветка → PR → checks → review → merge** внутри Hub, без обязательного внешнего GitHub UI как process-of-record.

| JTBD | Результат в Hub |
|------|-----------------|
| Связать работу с изменениями | Issue ↔ branch/PR metadata |
| Видеть готовность к merge | Checks + required reviewers на одном экране |
| Соблюдать политику ветки | Merge gate (green + N approvals) |
| Узнать о review | In-app / Telegram / email через notification |
| (Later) AI-ревью | Defect-first комментарии через hub-ai |

**Не цель:** заменить git-hosting (Gitea/GitLab/Actions marketplace). Forge = storage + CI runners; Hub = orchestration + policy + UX.

---

## 2. Два слоя (важно не смешивать)

| Слой | Что это | Где живёт сейчас / будет |
|------|---------|---------------------------|
| **A. Forge adapter** | OAuth/token, webhook от GitHub/GitLab, link PR→issue | `internal/github`, `internal/gitlab` (уже в hub-core) |
| **B. DevFlow (reduced GitHub)** | Собственные сущности PR/Check/Review/MergePolicy как source of truth process | **Новый** пакет `devflow` в **hub-core** (Later код) |

Сегодня (MVP docs/AI): слой B **эмулируется агентами** (atomic commits, code-reviewer, CI-aware ship).  
Later: слой B как runtime-модуль; слой A остаётся адаптером к forge.

---

## 3. MVP сейчас (docs / process — без кода продукта)

| Capability | Как закрываем сегодня в Hub | В portable kit |
|------------|----------------------------|----------------|
| Issues | Tracker issues + US/AC | Agent `product-owner` + tracker FR map |
| Branches | git-workflow conventional | playbook git-workflow |
| PR description | шаблон team-lead / PR rules | prompts/13 corporate + templates |
| Review | `code-reviewer` agent | PE `security-reviewer` + review prompts |
| CI checks | GHA ci/pr-guard/deploy | devops agent + path-filter pattern |
| Merge policy | user merges master, rebase-merge | rules в kit (50-git) |
| Atomic commits | conventional commits | agents builders + GSD-секции |

---

## 4. Functional Requirements (FR) — target product

### 4.1 Модули и FR

| Module ID | FR ID | Requirement | Priority | Depends |
|-----------|-------|-------------|----------|---------|
| `devflow.issue_link` | FR-DF-01 | Связь issue ↔ branch name / head SHA / forge URL | Must | tracker |
| `devflow.issue_link` | FR-DF-02 | Автолинк по ключу issue в branch/PR title (`ABC-123`) | Should | tracker |
| `devflow.pr` | FR-DF-10 | PR entity: title, body, status, head/base, author | Must | auth |
| `devflow.pr` | FR-DF-11 | Список PR по project / issue / author | Must | rbac |
| `devflow.pr` | FR-DF-12 | Статусы: `draft` \| `open` \| `approved` \| `merged` \| `closed` | Must | — |
| `devflow.checks` | FR-DF-20 | Ingest CI status (webhook или polling forge) | Must | apigateway / safedialer |
| `devflow.checks` | FR-DF-21 | Агрегат `checks_state`: pending/success/failure | Must | — |
| `devflow.checks` | FR-DF-22 | Идемпотентность по `(provider, external_run_id)` | Must | — |
| `devflow.review` | FR-DF-30 | Approve / request-changes / comment | Must | rbac |
| `devflow.review` | FR-DF-31 | Required reviewers N (project policy) | Must | project |
| `devflow.review` | FR-DF-32 | Self-approve запрещён (NFR policy) | Must | — |
| `devflow.merge_gate` | FR-DF-40 | Block merge если checks ≠ success | Must | checks |
| `devflow.merge_gate` | FR-DF-41 | Block merge если approvals < N | Must | review |
| `devflow.merge_gate` | FR-DF-42 | Protected branches config per project | Should | project |
| `devflow.notify` | FR-DF-50 | События review/check → notification | Must | hub-notification |
| `devflow.ai` | FR-DF-60 | Опциональный AI review comment (hub-ai) | Could | hub-ai |

### 4.2 Non-Functional Requirements (NFR)

| NFR ID | Requirement | Связь с TZ / ADR |
|--------|-------------|------------------|
| NFR-DF-01 | p95 API list/get PR < 500ms | NFR-01 Hub |
| NFR-DF-02 | Webhook ingest: verify signature **всегда**, skip запрещён | SEC-02 |
| NFR-DF-03 | Исходящие URL forge/CI → `ValidateOutboundURL` | SSRF / safedialer |
| NFR-DF-04 | UUID v4 PK, UTC, `version` optimistic lock, soft-delete 30д | TZ Data |
| NFR-DF-05 | Ошибки `{"error","request_id"}` | TZ API |
| NFR-DF-06 | Rate: webhook endpoint отдельный лимит (не auth 3/10) | TZ Rate |
| NFR-DF-07 | Events только через **outbox → NATS JetStream** | ADR §4 |
| NFR-DF-08 | Нет cross-schema JOIN к notification/ai | ADR §1 schema-per-service |
| NFR-DF-09 | Audit: merge attempt / policy change | audit module |
| NFR-DF-10 | Secrets forge token: at-rest encryption, не в логах | security guidelines |

---

## 5. Non-goals

- Полный Git LFS / Actions marketplace clone
- Hosted git server (замена Gitea/GitLab) в MVP и M0–M3
- Social (stars, followers, public profiles)
- Вынос `devflow`/`github` из **hub-core** (в ADR они в ядре; tracker исключён — metadata рядом)
- Big-bang замена текущего `internal/github` без dual-read/адаптера
- Shared-таблицы с `notification` / `ai` (запрет ADR)

---

## 6. Mapping на микросервисы (ADR-ARCH-MICROSERVICES)

| Concern | Юнит | Почему |
|---------|------|--------|
| PR/Check/Review/MergePolicy, issue links | **hub-core** | Рядом с tracker/project/rbac; github/gitlab уже в ядре |
| Check webhooks + SSRF | hub-core + `platform/http` safedialer | SEC |
| Review/check notifications | **hub-notification** (пилот strangler) | Consumer `hub.devflow.*` |
| AI review agents | **hub-ai** (вынос №1 после notification) | BYOK/изоляция |
| WS push статуса PR на карточку issue | hub-ws (после realtime-трио) или core WS bridge | Later |
| Email digest review | hub-mail (уже вынесен) | через notification orchestration |

**Инварианты:** schema-per-service; CQRS-проекции через NATS, не JOIN; sync-хопов ≤1; Caddy route-flip не применяется к devflow до отдельного ADR (остаётся в core).

---

## 7. ERD sketch (логические сущности, Later)

> Не код Ent — эскиз. Все PK = UUID v4; везде `created_at`/`updated_at` UTC; `version` int; soft-delete где нужно.

```
Project 1──* DevflowRepoBinding
Issue 1──* DevflowIssueLink
DevflowPullRequest *──1 Issue? (optional FK logical; issue_id UUID)
DevflowPullRequest 1──* DevflowCheckRun
DevflowPullRequest 1──* DevflowReview
Project 1──1 DevflowMergePolicy (or 1──* per branch pattern)
```

| Entity | Key fields | Notes |
|--------|------------|-------|
| `DevflowRepoBinding` | id, project_id, provider(`github`\|`gitlab`\|`hub`), forge_repo, encrypted_token_ref, webhook_secret_hash, version | Связь project↔forge; token не plaintext |
| `DevflowIssueLink` | id, issue_id, link_type(`branch`\|`pr`\|`commit`), external_id, url, status, branch, version | Аналог GitHubLink/GitLabLink; унификация Later |
| `DevflowPullRequest` | id, project_id, issue_id?, number, title, body, status, head_ref, base_ref, head_sha, author_user_id, forge_url?, version | Process SoT |
| `DevflowCheckRun` | id, pr_id, name, status, conclusion, external_run_id, provider, started_at, completed_at, html_url, version | Idempotent unique(provider, external_run_id) |
| `DevflowReview` | id, pr_id, reviewer_user_id, state(`approved`\|`changes_requested`\|`commented`), body, version | Self-approve reject in service |
| `DevflowMergePolicy` | id, project_id, branch_pattern, required_approvals, require_green_checks, allow_admin_bypass, version | Policy |

**Границы данных (ADR):** все сущности в схеме **hub-core** (public или `core`). Notification хранит только delivery/read-model по event_id; AI — только job/result refs, не копию PR body без политики retention.

---

## 8. API sketch (REST, Later)

Префикс: `/api/devflow/...` · Auth: JWT · Ошибки: `{"error","request_id"}` · Пагинация: offset/limit для списков.

| Method | Path | FR | RBAC perm (sketch) |
|--------|------|----|---------------------|
| GET | `/projects/:pid/prs` | FR-DF-11 | `devflow.pr.read` |
| POST | `/projects/:pid/prs` | FR-DF-10 | `devflow.pr.create` |
| GET | `/prs/:id` | FR-DF-10 | `devflow.pr.read` |
| PATCH | `/prs/:id` | FR-DF-12 | `devflow.pr.update` (author/admin) |
| POST | `/prs/:id/reviews` | FR-DF-30 | `devflow.review.submit` |
| GET | `/prs/:id/checks` | FR-DF-21 | `devflow.pr.read` |
| POST | `/webhooks/devflow/:provider` | FR-DF-20 | signature only (no user JWT) |
| GET | `/projects/:pid/merge-policy` | FR-DF-42 | `devflow.policy.read` |
| PUT | `/projects/:pid/merge-policy` | FR-DF-42 | `devflow.policy.admin` |
| POST | `/prs/:id/merge` | FR-DF-40/41 | `devflow.merge` + gate |
| GET | `/issues/:id/devflow` | FR-DF-01 | issue read ∩ pr read |

**Webhook contract (inbound):**
- Headers: `X-Hub-Signature-256` (или provider-native + verify) — **never skip**
- Body size limit + `LimitReader`
- Response 2xx только после outbox write; retry-safe

**Outbound forge API:** только `safehttp.NewSafeHTTPClient` + `ValidateOutboundURL`.

---

## 9. RBAC

| Permission | Who (typical) | Allows |
|------------|---------------|--------|
| `devflow.pr.read` | member project | view PR/checks |
| `devflow.pr.create` | developer | open PR / link |
| `devflow.pr.update` | author, lead | edit title/body/draft |
| `devflow.review.submit` | developer ≠ author | approve/request-changes |
| `devflow.merge` | lead / release | merge if gate green |
| `devflow.policy.admin` | project admin / founder | required N, protected branches |
| `devflow.webhook.manage` | admin | rotate webhook secret |

**Инварианты:**
- Review rights **≠** org admin (FR AC).
- `founder`/`system_role` не обходит merge gate по умолчанию (`allow_admin_bypass` — явный флаг policy + audit).
- Search/list — **project membership scope** (аналог NF-03).

---

## 10. Domain events (NATS JetStream + outbox)

Subjects (канон `hub.<domain>.<event>`):

| Subject | Payload (минимум) | Consumers |
|---------|-------------------|-----------|
| `hub.devflow.pr.opened` | pr_id, project_id, issue_id?, author_id | notification, search-proj, audit |
| `hub.devflow.pr.updated` | pr_id, status, version | ws-bridge, search |
| `hub.devflow.check.updated` | pr_id, check_id, conclusion | notification (on failure), ws |
| `hub.devflow.review.submitted` | pr_id, reviewer_id, state | notification → author |
| `hub.devflow.merge.blocked` | pr_id, reasons[] | notification, audit |
| `hub.devflow.merge.completed` | pr_id, merger_id, sha | tracker projection?, audit |
| `hub.devflow.pr.deleted` | pr_id | orphan reconcile (ADR UUID orphans) |

**Правила ADR:** publish только из outbox в той же TX, что и write; consumers идемпотентны по `event_id`; notification **не** читает таблицы core напрямую.

---

## 11. Acceptance Criteria (будущая реализация)

1. Пользователь видит PR status (`open` / checks aggregate / approvals) из карточки issue.
2. Merge в protected branch только при `checks_state=success` **и** `approvals ≥ N`.
3. Все webhook URL / outbound forge URL проходят `ValidateOutboundURL`; signature check без skip.
4. RBAC: `devflow.review.submit` ≠ org admin privilege; self-approve → 403.
5. Падение notification-сервиса не блокирует create PR (async); eventual delivery.
6. Optimistic lock: concurrent review/merge → 409 version conflict.
7. Soft-delete PR: 30 дней в корзине; hard delete — Admin/GDPR.
8. Contract-тест: producer `hub.devflow.review.submitted` ↔ notification consumer.
9. Path-filtered CI для `backend/internal/devflow/**` не раздувает бюджет GHA.
10. Docs sync: `docs/wiki/API.md` + STATE в той же сессии, что код (когда код появится).

---

## 12. Фазы реализации M0–M3 (код Later; сейчас — план)

Согласованы с Track P ADR (M0 infra → M1 hardening → M2 notification pilot → …), но **scoped на DevFlow**:

### M0 — Blueprint & contracts (текущий горизонт AI-expansion)
- Docs: этот файл + portable arch + PE templates PR/review.
- Зафиксировать FR/NFR/events/RBAC без Ent.
- Agent-эмуляция GitHub-discipline (atomic, review, CI-aware).
- **Go/No-Go docs:** kit переносится ≤10 шагов; AC понятны sa/be.

### M1 — Core domain skeleton (после снятия code-freeze / July return)
- Ent schemas + repository interfaces в hub-core (`devflow`).
- API CRUD PR + issue link; без merge gate.
- Dual-read с существующими `GitHubLink`/`GitLabLink` (миграционный мост).
- TD-02: никаких `*ent.Client` в handler/service.
- Unit/hubtest домена; docs API.

### M2 — Checks + webhooks + merge gate
- Webhook ingest + signature; check aggregation.
- MergePolicy + `POST /merge` gate.
- Outbox events; contract-тесты.
- **Параллельно с пилотом hub-notification:** consumer на `hub.devflow.*` (не блокирует M2 core).
- Security-auditor pass на webhook/SSRF.

### M3 — Review UX + AI optional + projections
- Full review workflow UI (fe) + required reviewers.
- WS/read-model обновление карточки issue.
- Опционально: hub-ai review job (только после JWKS gate ADR до hub-ai).
- Метрики: merge lead-time, % blocked merges, webhook failure rate.
- Retrospective: нужен ли отдельный сервис (по умолчанию **нет** — остаёмся в hub-core).

**После M3:** не выносить tracker/devflow без нового ADR; AI — только через события в hub-ai.

---

## 13. Связь с существующим `internal/github`

| Уже есть | Роль в DevFlow |
|----------|----------------|
| Integration + webhook + PR/commit links | Forge adapter (слой A) |
| WorkflowRun DTOs | Источник для `DevflowCheckRun` mapping |
| SafeHTTPClient к GitHub API | Обязательный паттерн для M2 |
| NATS publish из github service | Унифицировать subjects под `hub.devflow.*` / `hub.github.*` с версионированием |

**Стратегия:** не удалять github/gitlab в M1; ввести anti-corruption layer `devflow` ← adapters.

---

## 14. Portable kit / AI-эмуляция (сейчас)

Пока нет runtime-кода — **агенты эмулируют GitHub-discipline**:
- atomic conventional commits;
- defect-first `code-reviewer`;
- CI-aware `gsd-ship` / pr-guard;
- шаблоны PR (`reduced-github-pr-flow`, corporate-wave);
- merge policy: только user → master, rebase-merge.

См. `prompt-engineering/agents/templates/reduced-github-pr-flow.md`.

---

## 15. Риски и митигации

| Risk | Mitigation |
|------|------------|
| Дубль SoT (forge PR vs Hub PR) | Явно: Hub = process SoT; forge = git; sync one-way+webhook |
| Shared DB temptation с notification | Только events; ADR §1 |
| SSRF via webhook SSRF-callback | ValidateOutboundURL + deny private ranges |
| CI minute blow-up | path-filtered jobs (ADR M0) |
| Scope creep → mini-GitHub | Non-goals + product MoSCoW |

---

## 16. Verification (для arch-ver)

- [ ] Все FR/NFR трассируются на AC
- [ ] Границы юнитов = ADR catalog (core / notification / ai)
- [ ] Нет предложения shared tables / Redis pub/sub / gorilla WS
- [ ] LiveKit/mediasoup не затронуты
- [ ] Фазы M0–M3 не требуют кода в AI-expansion Must
- [ ] ≥ альтернатива «оставить только forge links» рассмотрена и отклонена для target (нужен process SoT в Hub)

**Handoff →** `arch-ver` → при детализации контрактов `sa1` → реализация после freeze: `be1`/`fe1` с границами файлов `devflow/**` vs `github/**`.
