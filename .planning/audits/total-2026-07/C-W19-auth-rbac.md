# C-W19 — Red-team: auth / rbac / audit / ws / search

> TASK-008 · Stage C · Wave 19 · READ-ONLY · 2026-07-29  
> Агент: независимый security/red-team (`cursor-grok-4.5`) · **не автор** фиксов A3 / TASK-146,156–161  
> HEAD проверен: `617af223` · Audit baseline A3: `ee353287`  
> Scope: `auth` · `rbac` · `audit` · `ws` · `search`  
> Источники A3: `A3W6-*`, `A3W3-4-ws.md`, `A3W5-4-search-core.md` (+ A3W5-5 для NF-03 тестов)  
> Выход: **только этот файл** · коммит **не** делать · **TASK-008 не закрывать**

---

## Вердикт

| Вопрос | Ответ |
|--------|--------|
| P0/P1 из A3 слотов → не в TASK-124…176? | **0 orphans** — в A3W6 / A3W3-4 / A3W5-4/5 **не было** open P0/P1 |
| Все open A3 ID (P2/P3) в бэклоге? | **Да** — 0 unmapped (DEDUP + TASK-146,156–161) |
| Комбо 1–4 на HEAD | Исторические **MITIGATED** в prod; **1 NEW** (deactivate ≠ WS kill) + **1 residual** GDPR orphan-after-commit |
| Attack chains ordinary member | Классический IDOR (WS room / search scope / JQL cross-project) **закрыт**; остался mid-session WS после admin deactivate |

**Итог:** постановка бэклога по периметру **полная**. Код HEAD уже несёт remediation A3 (часто с маркерами `A3W6-*` / `A3W5-4-*` / `A3W3-4-*`). README ещё может держать TASK как «Открыта» при файлах «На проверке» — это процессный drift, не orphan finding. **TASK-008 не закрывать** без независимой W18-мутации.

---

## 1. Gaps: A3 P0/P1 / open ID → TASK-124…176

### 1.1 P0/P1

| Слот | P0 | P1 | Комментарий |
|------|---:|---:|-------------|
| A3W6-1 auth-core | 0 | 0 | open findings = 0 |
| A3W6-2 auth-tests | 0 | 0 | open = 4×(P2+P3) |
| A3W6-3 rbac | 0 | 0 | open = 3 |
| A3W6-4 audit-core | 0 | 0 | open = 3 |
| A3W6-5 audit-tests | 0 | 0 | open = 7 |
| A3W3-4 ws | 0 | 0 | open = 0 (A3W3-4-001 → FIXED TASK-146) |
| A3W5-4 search-core | 0 | 0 | open = 3 |
| A3W5-5 search-tests | 0 | 0 | open = 2 (P3) |

**Ответ на вопрос «что не попало»:** P0/P1 из этих A3-отчётов **не существовали** → в TASK-124…176 нечего «терять». Orphan P0/P1 = **нет**.

### 1.2 Маппинг open A3 → backlog (полнота)

| Источник | Open ID (на момент A3) | Sev | Backlog | Coverage |
|----------|------------------------|-----|---------|----------|
| A3W6-1 | — | — | — | N/A (0 open) |
| A3W6-2 | A3W6-2-001…004 | P2×2 / P3×2 | [TASK-158](../../backlog/TASK-158-2026-07-29-a3-6-2-auth-tests.md) | ✅ 4/4 |
| A3W6-3 | A3W6-3-001…003 | P2 / P3×2 | [TASK-159](../../backlog/TASK-159-2026-07-29-a3-6-3-rbac.md) | ✅ 3/3 |
| A3W6-4 | A3W6-4-001…003 | P2 / P3×2 | [TASK-160](../../backlog/TASK-160-2026-07-29-a3-6-4-audit-core.md) | ✅ 3/3 |
| A3W6-5 | A3W6-5-001…007 | P2×4 / P3×3 | [TASK-161](../../backlog/TASK-161-2026-07-29-a3-6-5-audit-tests.md) | ✅ 7/7 |
| A3W3-4 | A3W3-4-001 | P3 | [TASK-146](../../backlog/TASK-146-2026-07-29-a3-3-4-ws.md) | ✅ |
| A3W5-4 | A3W5-4-001…003 | P2 / P3×2 | [TASK-156](../../backlog/TASK-156-2026-07-29-a3-5-4-search-core.md) | ✅ 3/3 |
| A3W5-5 | A3W5-5-001…002 | P3×2 | [TASK-157](../../backlog/TASK-157-2026-07-29-a3-5-5-search-tests.md) | ✅ 2/2 |
| **Σ** | **23** | | **8 TASK** | **missed = 0** |

Канон: [`A3-DEDUP-2026-07-29.md`](A3-DEDUP-2026-07-29.md) (unmapped 0).

### 1.3 Процессный gap (не finding-orphan)

| Симптом | Evidence |
|---------|----------|
| Файлы задач: evidence FIXED / «На проверке» | TASK-146,156–160 frontmatter + evidence блоки |
| Prod уже с A3-маркерами | см. §2–3 |
| README может отставать от статуса файла | `.planning/backlog/README.md` |

→ Red-team: **дыры «не завели задачу» нет**.

---

## 2. Findings (Stage C / HEAD `617af223`)

| ID | Severity | file:line | Evidence | Уже в backlog? |
|----|----------|-----------|----------|----------------|
| **C-W19-001** | **P2** | `backend/internal/audit/service.go:384-408` (+ `repository.go:889-892`; контраст `auth/service_session.go:203-218`, `283-321`; bridge `ws/bridge.go:249-252`) | `deactivateUser` / Ban: (1) DB `DeleteUserSessions` **без** Redis Del pipeline `SessionService.invalidateUserSessions`; (2) **нет** `publishSessionEnded(..., "revoked"|"logout")`. WS Bridge рвёт conn только на `revoked`/`logout`. Итог: после admin Deactivate/Ban **живой WS** продолжает получать события комнат до reconnect; REST/JWT ловит ME-13 (`auth/middleware.go:94-99`), upgrade WS — тоже (`ws/handler.go:177-188`), но **уже установленный** socket не рвётся. Каталог: `ECC-RBAC-025` (P0 corner, TODO) | **NEW** (не TASK-109: там только logout/revoke path; не TASK-158) |
| **C-W19-002** | **P2** | `backend/internal/audit/service.go:489-492`, `604-610` | A3W6-4-001 **закрыл** худший вектор (Mongo erase внутри Ent tx → user жив / chat стёрт). Выбранный фикс: Mongo **после** Commit. Trade-off **остаётся**: Mongo erase fail → user уже HardDelete, messages с `sender_id` **orphan** (PII GDPR); error возвращается admin, compensating outbox/retry **нет** | **Residual TASK-160** (задокументирован в коде; outbox retry **не** в ACCEPTANCE как must) → отметить как **NEW** если нужна отдельная задача на erase-retry |
| C-W19-003 | P3 | `audit/service.go:399-402` | `DeleteUserSessions` на deactivate — best-effort (`slog.Error`, return nil). При fail DB sessions живы; ME-13 + refresh `userIsDeactivated` всё ещё режут API. Усиливает C-W19-001 (WS + возможно DB session row), не открывает REST IDOR | **NEW** (узкий hygiene; можно смержить с C-W19-001) |
| A3W6-2-001…004 | — | login/registration/session/resend tests | На HEAD: prod guard `Status != Active` (`service_login.go:61-63`); `firstUserMu`+advisory (`service_registration.go:72-81`); тесты `DeactivatedUserDenied`, `ConcurrentOnlyOneFounder`, Redis Exists на Invalidate | **TASK-158** (FIXED evidence) |
| A3W6-3-001…003 | — | rbac Delete/RequireSystemRole/cache | Delete: best-effort Invalidate (`custom_roles.go:155-161`); RequireSystemRole admin≡IsSystemAdmin (`middleware.go:67-72`) | **TASK-159** |
| A3W6-4-001…003 | — | GDPR/Ban/layering | Mongo post-Commit; Ban `errors.Is`; handler без `ent.IsNotFound` | **TASK-160** (+ residual C-W19-002) |
| A3W3-4-001 | — | `ws/handler.go:157-160` + `handler_auth_test.go` QueryToken+Ticket | Guard + mutation-тест на месте | **TASK-146** |
| A3W5-4-001…003 | — | Meili dial / JQL / IndexAllIssues | Operator dial client; `ErrUnsupportedJQLOperator`; pagination batch | **TASK-156** |

---

## 3. Combinations / attack chains

### 3.1 Запрошенные комбо (X+Y хуже суммы)

#### C-W19-C1 — deactivated JWT × founder race

| | A3W6-2-001 (deactivated login) | A3W6-2-002 (founder race) |
|--|-------------------------------|---------------------------|
| Исторический gap | Тест только `StatusPending` → мутация `== Pending` зелёная, deactivated получает JWT | Race-suite concurrent только `SetRegistrationMode`; снятие `firstUserMu`/tx не краснеет → dual founder (SEC-06) |
| Комбо (хуже суммы) | Dual-founder org **плюс** deactivated/former-admin с выданным JWT до ME-13 → окно привилегий + сломанный trust model ORG-04/06 | Не «просто» missing assert |
| HEAD prod | `service_login.go:61-63` deny non-Active; middleware ME-13; refresh `validateSessionForRefresh` + deactivated | `firstUserMu` + `WithPostgresAdvisoryTx` / EntTx |
| HEAD tests | `TestLoginWithMagicLink_DeactivatedUserDenied` | `TestCreateFirstOrMemberUser_ConcurrentOnlyOneFounder` |

**Вердикт комбо:** **MITIGATED** (prod + tests).  
**Ordinary user:** founder race **недоступен** (нужен empty-org registration). Deactivated не логинится и не refresh'ит.

**Связка с C-W19-001:** даже при корректном deny login, **admin Deactivate** не рвёт уже открытый WS — это **отдельная** цепочка (ниже), не mutual founder race.

#### C-W19-C2 — WS ticket × query `?token=`

| | A2/TASK-103 | A3W3-4-001 |
|--|-------------|------------|
| Было | Query JWT forbidden; ticket subprotocol канон | Тест бил только `?token=` **без** ticket → мутация «убрать query-guard, оставить ticket-required» false-green |
| Комбо | Valid ticket **+** `?token=` → JWT снова в access/proxy URI при ослабленном guard | Severity > P3 hygiene: regression handshake + log leak |
| HEAD | `handler.go:157-160` reject **до** parse ticket; `TestWSHandler_QueryToken_WithValidTicket_Denied` | |

**Вердикт комбо:** **MITIGATED**.  
**Ordinary user:** не может «протащить» JWT в query при валидном ticket.

#### C-W19-C3 — search NF-03 scope × JQL silent-nil

| | NF-03 Ctrl+K (`search/service.go`) | A3W5-4-002 JQL |
|--|-------------------------------------|----------------|
| Было | `projectIDs` на issues/TC/files/…; empty → nothing | `priority > x` → `conditionToPredicate` silent `nil` → unfiltered preds |
| Комбо (исторический) | JQL в tracker (`SearchIssuesJQL`) **всегда** с `issueProjectIDEQ(projectID)` (`repo_issues.go:492`) — silent-nil **не** давал cross-project IDOR, но **снимал фильтр внутри проекта** (все issues проекта вместо ожидаемого subset). Вместе с ложным ощущением «JQL безопасен» → data over-disclosure member'у своего проекта | Не сумма P3+«чисто» |
| HEAD | ParseJQL → `ErrUnsupportedJQLOperator` (`jql.go:105-113`); never-match `IDEQ(Nil)` defense-in-depth; Ctrl+K scope без изменений | RBAC `PermProjectView` на `/jql` |

**Вердикт комбо:** **MITIGATED** (cross-tenant IDOR не открывался; within-project filter-bypass закрыт).  
**Ordinary user:** чужие проекты через Ctrl+K / JQL — нет (membership + `projectIDEQ`).

#### C-W19-C4 — GDPR Mongo orphan

| | До TASK-160 | После (HEAD) |
|--|-------------|--------------|
| Вектор A | Mongo erase **внутри** Ent tx → Commit fail → user жив, chat стёрт | **Закрыт** (eraser после Commit; тесты `gdpr_cascade_p1be34_test.go`, `task114_residual_test.go`) |
| Вектор B (trade-off) | — | Mongo fail **после** Commit → user удалён, messages orphan (**C-W19-002**) |
| Комбо | A хуже для continuity/availability chat; B хуже для GDPR erasure completeness | Выбор A→B осознанный; без retry B остаётся |

**Вердикт:** исходный A3 residual **MITIGATED**; **новый residual P2** = C-W19-002.  
**Ordinary user:** GDPR erase — **admin-only**; member не инициирует.

### 3.2 Attack chains — ordinary authenticated user (member, не admin/founder)

| # | Цепочка | Результат на HEAD | Sev |
|---|---------|-------------------|-----|
| A1 | `GET /ws?token=<jwt>` + valid `hub.ws.ticket.*` | **401** query forbidden | — |
| A2 | Redeem чужой/replay ticket | one-shot Redis; invalid → 401 | — |
| A3 | `subscribe` `channel:`/`project:`/`call:` чужого UUID | `CanSubscribe` fail-closed → `subscribe_denied` | — |
| A4 | Ctrl+K query без membership | `getUserProjectIDs` → empty shards / no foreign hits | — |
| A5 | `POST …/issues/jql` `priority > medium` | **400** unsupported op (не full dump) | — |
| A6 | JQL на project без `PermProjectView` | RBAC **403** | — |
| A7 | Magic-link login deactivated (свой аккаунт после ban) | `ErrAccountNotActive`; sessions==0 | — |
| A8 | Refresh после deactivate | DB session gone **или** deactivated check → fail | — |
| A9 | **Admin деактивирует жертву при открытом WS** | REST **403**; **WS продолжает** push до reconnect (**C-W19-001**) | **P2** |
| A10 | Concurrent self-registration на empty org | N/A для ordinary member уже в org; founder path защищён mu+advisory | — |
| A11 | API key deactivated user | `ErrAPIKeyUserDeactivated` (`apikey.go:176-178`) | — |

**Единственная устойчивая member-facing дыра в периметре:** A9 / C-W19-001 (требует **admin** как триггер, эксплуатируется сессией жертвы).

---

## 4. Clean checklist

| Проверка | Статус | Evidence |
|----------|--------|----------|
| Magic Link only / нет Telegram login | ✅ | A3W6-1 FIXED-OK; files absent; denial TASK-102 |
| JWT 15m / refresh Redis / CTC | ✅ | session service + magiclink CTC |
| Auth rate 3/10m | ✅ | wiring + magiclink constants (A3W6-1) |
| ME-13 deactivated на REST | ✅ | `RequireAuth` + status cache invalidate on deactivate |
| ME-13 на WS **upgrade** | ✅ | `handler.go:177-188` |
| ME-13 на WS **mid-session после deactivate** | ❌ | **C-W19-001** |
| Deactivate → Redis refresh Del + NATS revoke | ❌ | audit path ≠ SessionService |
| Founder SEC-06 race | ✅ | mu + advisory/Ent tx + concurrent test |
| Fiber v3 RBAC order | ✅ | A3W6-3 FIXED-OK |
| DeleteCustomRole cache-fail ≠ 500 | ✅ | TASK-159 |
| RequireSystemRole("admin") + founder | ✅ | IsSystemAdmin branch |
| GDPR Ent cascade tx | ✅ | TASK-114/160 |
| GDPR Mongo pre-commit erase | ✅ закрыт | post-Commit |
| GDPR Mongo post-commit orphan retry | ❌ residual | **C-W19-002** |
| Ban sentinels / handler без ent | ✅ | TASK-160 |
| WS `?token=` deny (+ ticket combo test) | ✅ | TASK-146 |
| NF-02 CanSubscribe fail-closed | ✅ | `authorizer.go` |
| NF-03 search project scope | ✅ | `Search` + empty projectIDs |
| JQL whitelist / unsupported comparison fail-closed | ✅ | TASK-156 |
| Meili metadata SSRF block | ✅ | validate + SafeDial public |
| Meili operator loopback dial | ✅ | `NewOperatorMeiliHTTPClient` |
| IndexAllIssues pagination | ✅ | batch loop |
| Ordinary user cross-project search/WS/JQL IDOR | ✅ чисто | §3.2 A1–A8 |
| A3 open ID unmapped → backlog | ✅ 0 | §1.2 |

---

## 5. Рекомендации (не исполнять в этой роли)

1. **NEW backlog:** deactivate/ban → для каждой session: Redis Del + `publishSessionEnded(..., "revoked")` **или** `UnregisterByUser` в hub; IT: open WS → Deactivate → conn dead / no further events (`ECC-RBAC-025`).
2. Опционально: outbox/retry `DeleteMessagesBySender` после GDPR Commit (закрыть C-W19-002).
3. Не закрывать TASK-008 / не мержить «A3 done» без W18 mutation на C-W19-001.

---

*Red-team Stage C W19 · auth/rbac/audit/ws/search · READ-ONLY · TASK-008 open*
