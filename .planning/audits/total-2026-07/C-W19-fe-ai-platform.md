# C-W19 — Red-team FE / AI / platform (Stage C reopen)

> TASK-008 · Stage C · волна **W19** · агент **независимый red-team** (`cursor-grok-4.5`)  
> Дата: **2026-07-29** · ветка `cursor-grok-4.5` · HEAD: `617af223`  
> Режим: **READ-ONLY** по продукту · артефакт: только этот файл · **не коммитить** · TASK-008 **оставить открытой**  
> Периметр: FE (routes/stores/XSS) + AI + platform/apigateway/search · волны A3 **W4 / W5 / W9 / W10 / W11**

## Вердикт

| Проверка | Результат |
|----------|-----------|
| Orphans A3W4*/5*/9*/10*/11* вне TASK-124…176 | **0** (dedup `unmapped=0` подтверждён; все 91 ID этих волн в кластерах 124–176) |
| Комбо на текущем HEAD | **4 actionable** (1 HIGH · 2 MEDIUM · 1 LOW) + 1 adjacent P3 |
| Docs drift residual после TASK-132/133 | **OPEN** (siblings + AGENTS + recount drift) |
| Чистый Stage C по периметру | **НЕТ** — FE↔BE webhook trim, dead-code XSS, docs inventory |

**TASK-008 не закрывать.** Remediation A3 (124…176) частично на «На проверке»; Stage C видит **остатки и комбо**, не покрытые scope отдельных TASK.

---

## 1. Orphans A3W4*/A3W5*/A3W9*/A3W10*/A3W11* вне TASK-124…176

| Источник | Σ ID (волны) | В TASK-124…176 | Вне map |
|----------|-------------:|---------------:|--------:|
| A3-DEDUP + CSV | 91 (W4=12 · W5=20 · W9=23 · W10=25 · W11=11) | 91 | **0** |

Сирот P0/P1/P2/P3 из исходных A3-отчётов этих волн **нет**. Ниже — только **NEW residual / combo** после фиксов, не заведённые как отдельные finding ID в 124…176.

### Audit miss (A3W10-1)

`A3W10-1-fe-routes-modules.md` пометил `modules/messenger/link-preview.tsx` как «чисто», при том что `href={url}` без whitelist. TASK-174 scope = только `components/messenger/*` → gap **вне** закрытия A3W9-3-003.

---

## 2. Комбо (обязательные сценарии)

### C-W19-001 · HIGH — FE `sprint.completed` × BE allowlist trim (A3W4-5-003 partial)

| | |
|--|--|
| **Пара** | TASK-152 trim `allowedWebhookEvents` (без `sprint.completed`) **+** FE `AVAILABLE_EVENTS` всё ещё с `sprint.completed` |
| **Почему хуже суммы** | BE честно убрал мёртвый event; UI предлагает подписку → Create/Update webhook с `sprint.completed` → **400** «неизвестное событие». Пользователь видит «фича есть», API отвергает. Класс false-close A3W4-5-003: trim без FE/docs sync. |
| **Файлы** | `backend/internal/apigateway/webhook.go:448-460` · `frontend/src/modules/settings/webhooks.tsx:18-24` |
| **Связь** | A3W4-5-003 / TASK-152 (scope явно «не FE») → residual **вне** TASK-124…176 map |
| **Фикс** | Убрать `sprint.completed` из FE `AVAILABLE_EVENTS` (+ wiki/API если обещает) **или** вернуть producer + allowlist вместе |
| **Mutation** | Create webhook events=`["sprint.completed"]` → 400; после FE-trim опция исчезает из UI |

### C-W19-002 · MEDIUM — LinkPreview XSS path × membership (A3W9-3 + A3W7-3)

| | |
|--|--|
| **Пара** | Prod `components/messenger` whitelist (TASK-174) **+** dead `modules/messenger/link-preview.tsx` без sanitize **+** BE `test_case` preview без membership |
| **Почему хуже суммы** | (a) XSS в prod-пути закрыт (`isSafeUrl` / `sanitizeUrl`). (b) Параллельный модуль: `href={url}` + fetch без whitelist — latent XSS при любом wire `MessageBubble`. A3W10 назвал файл «чисто» → false confidence. (c) Issue-preview: `IsProjectMember` OK; **test_case** отдаёт `Type=test_case` + title=ключ любому auth user без membership → enum TC-ключей + отличимый ответ от generic external. XSS-карточка + membership-дыра = разные поверхности одной MSG-11 фичи. |
| **Файлы** | `frontend/src/components/messenger/link-preview.tsx` (FIXED) · `frontend/src/modules/messenger/link-preview.tsx:225-228` · `frontend/src/modules/messenger/message-bubble.tsx` (imports unsanitized preview; **не** referenced из routes) · `backend/internal/github/preview.go:61-69` vs `:99-110` |
| **Связь** | A3W9-3-003 FIXED partial · membership issue = A3W7-3-003 · test_case gap **вне** TASK-174/124…176 |
| **Фикс** | Удалить/совместить modules LinkPreview с `@/lib/safe-url`; BE: test_case → membership project или fail-closed nil→generic |
| **Mutation** | (FE) `javascript:` url → null / no href; (BE) non-member TC URL → не `type=test_case` |

**Prod-path XSS (components):** PASS на HEAD.  
**OG `img src`:** FE не гоняет `imageSrc` через `isSafeUrl`; BE `ValidateOutboundURL` на og:image — defense-in-depth P3, не поднимаем до HIGH.

### C-W19-003 · MEDIUM — AI consent Default(false) × stream sanitize (A3W5-3)

| | |
|--|--|
| **Пара** | Schema/create deny external (005) **+** WS stream error sanitize (007) |
| **Статус на HEAD** | **HOLD — комбо закрыт в коде** |
| **Evidence** | `ent/schema/aiuserpreference.go:30-32` `Default(false)` · `repository_ent.go:391-396` create → false · `service.go:34,541-542` `msgAIStreamFailed` · FE `BotMarkdown` — text children, без `dangerouslySetInnerHTML` · preflight `resolveProviders` до stream 202 |
| **Residual (LOW)** | Исторические строки БД с `allow_external_ai=true` (созданные при старом Default(true)) **не** мигрированы data-fix — только schema default для новых. Нет evidence SQL `UPDATE … SET allow_external_ai=false`. Opt-in нарушен для legacy rows. |
| **Mutation (уже есть)** | TASK-155 hubtest: model-only save → false; strip sanitize → client видит raw err |

### C-W19-004 · MEDIUM — Webhook soft-delete filter × VersionEQ (A3W4-5-001/002)

| | |
|--|--|
| **Пара** | `ListActive` + `HasProjectWith(DeletedAtIsNil)` **+** `VersionEQ`/`AddVersion` на Update |
| **Статус на HEAD** | **HOLD на happy-path** |
| **Evidence** | `repository.go:170-178` soft-delete project filter · `:230-269` VersionEQ + affected==0 → `ErrVersionConflict` · `webhook.go` version ≥1 · sprint.completed trimmed |
| **Residual combo** | `SoftDeleteWithCrossLinkCleanup` **не** деактивирует endpoints (`project/repository.go:680-685` только project DeletedAt). Защита = один Where на ListActive. Регресс фильтра → outbound POST + decrypt секретов для корзины (W19 C-008 класс). VersionEQ не помогает: stale concurrent PUT на **живом** endpoint корзины всё ещё возможен через `ListWebhookEndpointsByProject` (нет `project.DeletedAtIsNil` на list-by-project). |
| **Фикс (defense-in-depth)** | В soft-delete tx: `IsActive=false` / soft-delete endpoints; list-by-project тоже фильтровать живой project |
| **Mutation** | Soft-delete project → ListActive len=0; drop `HasProjectWith` → RED; stale version → 409 |

### C-W19-005 · LOW — error≠empty miss residual (вне закрытых A3W9/10 ID)

| | |
|--|--|
| **Пара** | QueryStateView/канон error-first **+** точечные miss вне TASK-124/127/172 |
| **HOLD** | accounting secondary (`trialBalance`/`invoices`/`incomeStatement`) — alert на HEAD · bases detail/table — QueryErrorAlert · admin/routes из TASK-172/127 — по session notes FIXED |
| **NEW miss** | `leave-requests.tsx`: `myError` обработан (`:229+`), но admin `listPendingLeave` **без** `isError` — секция только при `pendingItems.length > 0` (`:177`). Fail очереди = «нет заявок на согласование» (тот же класс, что A3W10-1-004 marketplace pending). A3W10-1 FIXED-OK «leave» — **overclaim**. |
| **Adjacent** | `pipeline-status.tsx:265` `href={run.html_url}` без `sanitizeUrl` (GitHub Actions URL; trust BE; P3 XSS defense-in-depth) |
| **Фикс** | `isError` + alert на pending leave; sanitize CI urls как в TASK-125 |

---

## 3. Docs drift residual после TASK-132 / TASK-133

Замер 2026-07-29 (этот агент):

| Метрика | Docs (после 132/133) | Факт HEAD | Drift |
|---------|----------------------|-----------|:-----:|
| BE `backend/testframework/**/*_test.go` | **805** (README / test_plan / TESTING / strategy / TROUBLESHOOTING) | **807** | ±2 |
| Vitest files `stages/unit/frontend-test` | **167** (test_plan / TESTING / strategy / user_needs / vision) | **170** | ±3 |
| `AGENTS.md` inventory | **515** files · Vitest **99/99** | 807 / 170 | **stale** |
| `docs/system_specification.md` | Vitest **94** files (таблица :690); NFR-30 «**99+** tests» | 170 files | **stale** |
| `docs/user_needs_analysis_2026.md:64` | `master` · HEAD `bbdaf16` | ветка `cursor-grok-4.5` · `617af223` | **stale SHA** (132 чинил competitive, не user_needs) |
| `docs/vision_and_scope.md` US-M06 | webcam **Phase 16** | webcam UI shipped (competitive/BUSINESS) | **product drift** |
| `docs/competitive_audit.md` | gap Calls OK; backlog «Vitest **99/99**» | 170 files | cosmetic |

**Вывод:** TASK-132/133 закрыли канон inventory в основном кластере docs, но **siblings + AGENTS + system_spec metrics + vision webcam** остались. Recount уже снова разошёлся (805→807, 167→170) — inventory без процесса «пересчёт при каждом fix-wave» снова поедет.

---

## 4. Сводка находок Stage C (этот отчёт)

| ID | Sev | Tag | Суть | Вне 124…176? |
|----|-----|-----|------|:------------:|
| C-W19-001 | HIGH | NEW combo | FE `sprint.completed` после BE trim | да |
| C-W19-002 | MEDIUM | NEW combo | modules LinkPreview XSS + TC preview без membership | да |
| C-W19-003 | — | HOLD (+LOW legacy rows) | AI Default(false)+stream sanitize | hold |
| C-W19-004 | MEDIUM | RESIDUAL combo | soft-delete filter без deactivate endpoints | да (defense) |
| C-W19-005 | LOW | NEW / overclaim | HR pending leave error≠empty | да |
| C-W19-DOC | P2 | RESIDUAL | docs/AGENTS/spec/vision drift после 132/133 | да |

P0 в периметре: **0**.

---

## 5. Что перепроверено и чисто (не изобретать)

| Зона | Evidence |
|------|----------|
| Apigateway SEC-02 / SSRF | outbound sign + ValidateOutboundURL — без регресса |
| ListActive soft-delete project | `HasProjectWith(DeletedAtIsNil)` присутствует |
| Webhook VersionEQ | Where+AddVersion+409 path |
| AI schema consent default | `Default(false)` + create Set false |
| AI stream WS error | фиксированная `AI stream failed` |
| LinkPreview prod (components) | `isSafeUrl` / `sanitizeUrl` |
| Issue link-preview membership | `IsProjectMember` + per-user cache key |
| Meili operator dial | `meiliAllowsOperatorPrivate` + SafeDial split (TASK-156 territory) |
| QueryStateView | error before empty — канон OK |
| client.ts wire→ApiError | `mapWireToApiError` (TASK-129) |
| BotMarkdown XSS | React text nodes, no raw HTML |

---

## 6. Приоритет для оркестратора

1. **C-W19-001** — FE trim `sprint.completed` (1 файл) или явный docs «не поддерживается»  
2. **C-W19-002** — modules LinkPreview delete/alias + BE test_case membership  
3. **C-W19-DOC** — recount 807/170 + AGENTS + system_spec + user_needs SHA + vision US-M06  
4. **C-W19-004** — deactivate webhooks в project soft-delete tx  
5. **C-W19-005** — pending leave `isError`

Не заводить дубликаты A3W*-* ID, уже в TASK-124…176. Новые ID — только Stage C (`C-W19-*`).

---

## 7. Handoff

- **team-lead:** завести тонкие TASK / дописать Progress к 152/174/132 только если owner lock «все residual»; иначе отдельный C-wave backlog.  
- **Не закрывать TASK-008.**  
- Mutation/верификация фиксов 124…176 — отдельные агенты Stage C (не автор фикса).

---

*Артефакт C-W19-fe-ai-platform · TASK-008 · readonly · RU · 2026-07-29*
