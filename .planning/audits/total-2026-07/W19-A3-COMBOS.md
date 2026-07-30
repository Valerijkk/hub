# W19-A3 — Combinatorial defects (W18 FAILs × Stage B residual)

> 2026-07-28 · Агент **W19-A3** (security-auditor) · ветка `cursor-grok-4.5` · **no push master**  
> Задача: [TASK-008](../../backlog/archive/legacy/TASK-008-total-audit-program.md) § Stage C / W19  
> Фокус: пары, где **finding/fix X + Y** на **текущем** коде после Stage B дают дефект тяжелее каждого по отдельности.  
> Якоря: W18 FAILs **014, 019, 041, 042, 044, 045, 046, 052, 054** + catch-up [TASK-079](../../backlog/archive/legacy/TASK-079-gdpr-mongo-eraser-fail-closed.md).  
> Метод: readonly code review + W18 mutation evidence; product code не правим.  
> **Re-verify:** 2026-07-28 19:58 UTC+3 — все `path:line` ниже сверены с HEAD `cursor-grok-4.5` (W19-A3 pass; Stage B residual confirmed).

## Сводка по severity

| Severity | Count |
|----------|------:|
| CRITICAL | 2 |
| HIGH     | 6 |
| MEDIUM   | 5 |
| LOW      | 3 |
| **Total** | **16** |

Actionable (не closed-reference): **15** (C-016 = FE adjacent reference). CRITICAL/HIGH = блокер «чистого» Stage C.

### Матрица W18 FAIL → combo

| W18 FAIL | Combo IDs | Суть усилителя |
|----------|-----------|----------------|
| **014** | C-003, C-004, C-010 | AddTeamToProject без team ACL × RBAC/sprint/complete |
| **019** | C-005 | To4 mutation-blind × Meili/push SSRF surfaces |
| **041** | C-002, C-010 | Dual StartSprint race unproven × no DB unique |
| **042** | C-011 | via_invite static assert × consumers trust flag |
| **044** | C-005, C-006, C-012 | Meili/stream hubtest gap × AI fail-open / soft-delete index |
| **045** | C-007, C-013 | Confirm-without-object unproven × empty/legacy entity |
| **046** | C-002 | Schema grep ≠ runtime unique × sprint active race |
| **052** | C-014 | expectTypeOf no-op × version DTO omitempty drift |
| **054** | C-015 | Docs metrics drift × false gate confidence |
| **079** (catch-up) | **C-001** | Mongo eraser `(0,nil)` × HardDelete cascade |

---

## Комбинации

### C-001 · CRITICAL — GDPR Mongo no-op + HardDelete cascade (TASK-079)

| | |
|--|--|
| **Пара** | W18-C5-001 (`gdpr_mongo.go` silent success) **+** P1-BE-34 cascade abort-on-err (TASK-023 PASS) |
| **Почему хуже суммы** | Cascade abort proven when eraser returns **error**. `DeleteMessagesBySender` при `coll==nil` возвращает `(0, nil)` → cascade считает Mongo OK → **HardDelete user** при живых messages. Right-to-erasure провален без ошибки. |
| **Файлы** | `backend/internal/audit/gdpr_mongo.go:27-30` · `backend/internal/audit/service.go:523-544` |
| **Stage A / B** | Catch-up W18-C5 · TASK-034/023 PASS · residual **TASK-079** |
| **OWASP** | A01 · A04 (incomplete erasure) · ASVS V8.3 |
| **Mutation idea** | `NewMongoMessageEraser(nil).DeleteMessagesBySender` → `err != nil`; cascade: nil eraser wired → HardDelete **не** вызывается. Filter `"sender_id"`→`"user_id"` → hubtest count RED. |

---

### C-002 · CRITICAL — Dual StartSprint race (041) + no unique active (046)

| | |
|--|--|
| **Пара** | TASK-041 FAIL (Ent concurrent GREEN on SQLite; conflict guard unproven) **+** TASK-046 FAIL (schema text grep, not Ent-runtime unique) **+** schema fact |
| **Почему хуже суммы** | Единственная app-защита — `ErrActiveSprintConflict` в `StartSprintAtomic` (`repository.go:216-227`) после `UPDATE project` lock (`:198-205`). На PG lock, вероятно, сериализует happy-path; **W18 не доказал** RED при снятии guard (SQLite serializes `Tx`). В `ent/schema/sprint.go:57-64` индекс `project+status` **не Unique** / нет partial unique на `status=active`. Вместе: регресс/обход guard (прямой Update, другой код-путь, будущий drop check) оставляет **два active навсегда** без DB safety net — planning/board/CompleteSprint семантика ломается. |
| **Файлы** | `backend/internal/sprint/repository.go:189-238` · `backend/ent/schema/sprint.go:57-64` |
| **Stage A / B** | W1-A4-009 · TASK-041 **Открыта** · TASK-046 **Открыта** |
| **OWASP** | A04 Insecure Design |
| **Mutation idea** | (a) Postgres hubtest: dual goroutine StartSprint → ровно один active; drop conflict check → RED. (b) Partial unique index / ent unique where active → duplicate Create/Update RED at DB. SQLite-only concurrent = insufficient. |

---

### C-003 · HIGH — AddTeamToProject no team ACL (014) + PermTeamManage only

| | |
|--|--|
| **Пара** | TASK-014 FAIL: service без `requireTeamMember` **+** route ACL только `PermTeamManage` |
| **Почему хуже суммы** | `AddTeamMember`/`RemoveTeamMember` защищены membership. `AddTeamToProject` — нет: любой с `team.manage` на проекте (обычно owner) тянет **чужую** org-team и bulk-inject участников как `developer` (`AddMember` errors swallowed, `repository.go:616-622`). RBAC «закрывает» endpoint, но не consent/team-scope → forced membership / lateral privilege. Fiber v3 order OK (`handler.go:58` — mw после handler). |
| **Файлы** | `backend/internal/project/service.go:571-579` · `handler.go:58,688-701` · `repository.go:600-624` |
| **Stage A / B** | P1-BE-08 PARTIAL · TASK-014 **Открыта** |
| **OWASP** | A01 Broken Access Control |
| **Mutation idea** | Stranger (not team member) + PermTeamManage → 403; drop `requireTeamMember` (после фикса) → RED. Repo: assert no silent AddMember for foreign team. |

---

### C-004 · MEDIUM — AddTeamToProject bulk (014) + CompleteSprint mid-fail unproven (014)

| | |
|--|--|
| **Пара** | Bulk developer inject **+** CompleteSprintAtomic Ent mid-fail не краснеет (W18: mock embeds rollback) |
| **Почему хуже суммы** | Prod **имеет** Ent-tx (`sprint/repository.go:301-331`) — partial complete не «открытый баг tx», а **дыра доказательности**. Вместе с forced team attach: на проекте с несогласованными developers сложнее отличить inject от легитимной гонки/ретрая Complete; CI не ловит регресс снятия tx. |
| **Файлы** | `backend/internal/sprint/repository.go:301-331` · `project/repository.go:616-622` |
| **Stage A / B** | TASK-014 P1-BE-05 **FAIL** |
| **OWASP** | A01 · A04 |
| **Mutation idea** | Ent hubtest: fail after move issues, before status=completed → assert full rollback; remove tx / split commits → RED. Mock-only mid-fail ≠ sufficient. |

---

### C-005 · HIGH — To4 residual (019) + Meili SSRF untested (044)

| | |
|--|--|
| **Пара** | TASK-019: drop `ip.To4()` в `isBlockedIP` не RED на Go 1.25 Contains **+** TASK-044: нет hubtest `validateMeiliHost` / metadata |
| **Почему хуже суммы** | Push/OpenGraph опираются на `SafeDial`/`isBlockedIP`. Meili — **свой** `validateMeiliHost` с `To4()` для 169.254 (`meilisearch_indexer.go:70-72`) и клиент `http.Client` **без** SafeDial (`:47`). Mutation-blind To4 + отсутствие Meili hubtest → регресс IPv4-mapped metadata (`::ffff:169.254.169.254`) / DNS-rebinding на operator `MEILI_HOST` не поймается ни одним контуром. |
| **Файлы** | `backend/internal/platform/http/safedialer.go:69-78` · `search/meilisearch_indexer.go:35-74` · `notification/push.go` (ValidateOutboundURL) |
| **Stage A / B** | TASK-019 **Открыта** · TASK-044 **Открыта** |
| **OWASP** | A10 SSRF |
| **Mutation idea** | Hubtest: `validateMeiliHost("http://[::ffff:169.254.169.254]")` reject; drop To4 in Meili path → RED. Document Go 1.25 Contains vs Meili-local To4; ideally SafeDial client. |

---

### C-006 · HIGH — AI access nil fail-open + stream cancel gap (044)

| | |
|--|--|
| **Пара** | `IssueContext` skips authz when `access==nil`; `authorizeReview` returns nil when `access==nil` **+** TASK-044 no hubtest stream `ctx.Err` |
| **Почему хуже суммы** | Wiring ставит checker, но nil = полный IDOR title/desc / PR context через @hub-ai. Stream cancel guard в `adapter_stream.go:60-62` без mutation-proof → cancelled client может продолжать читать/бить provider (cost/DoS) пока privacy path уже fail-open в тестах. |
| **Файлы** | `backend/internal/ai/issue_lookup.go:56-65` · `code_review.go:66-69` · `adapter_stream.go:60-62` |
| **Stage A / B** | TASK-021 PASS · TASK-044 **Открыта** residual |
| **OWASP** | A01 · A04 |
| **Mutation idea** | Unit без SetAccessChecker → deny; cancel ctx mid-SSE → Read stops (RED if remove `ctx.Err` check). |

---

### C-007 · HIGH — Confirm empty entity + Confirm-without-object (045)

| | |
|--|--|
| **Пара** | `validateEntityScope("","")` → nil **+** TASK-045 FAIL: Confirm→404 / `isS3NotFound` mapping unproven |
| **Почему хуже суммы** | Empty entity = orphan confirmed metadata. `assertObjectWithinLimit` мапит 404→`ErrObjectNotFound` только при `s3Client != nil` (`service.go:154-158`); nil client / unproven mapping = ghost Confirm. Вместе: Confirm без binding и без доказанного object-exists → NF-09-класс + невидимый file row в project scope. |
| **Файлы** | `backend/internal/storage/service.go:144-171` · `:209-218` · `:247-248` |
| **Stage A / B** | W7-A4 · TASK-045 **Открыта** |
| **OWASP** | A01 · A04 |
| **Mutation idea** | Confirm `entityType=""` → 400; pending + Head 404 → `ErrObjectNotFound` (production path, not helper tautology). Drop mapping → RED. |

---

### C-008 · HIGH — Soft-delete project + webhooks remain ListActive

| | |
|--|--|
| **Пара** | Project soft-delete (TASK-009 PASS) **+** webhook ListActive без `project.deleted_at` |
| **Почему хуже суммы** | `DeleteProject` ставит `status=deleted` + `deleted_at` (`project/service.go:207-228`), **не** деактивирует webhook endpoints. `ListActiveWebhookEndpoints` фильтрует только `IsActive` + endpoint `DeletedAtIsNil` (`apigateway/repository.go:168-176`). `webhookEndpointProjectID` возвращает ID без проверки `Project.DeletedAt` (`webhook.go:423-428`). UI/корзина скрывают проект; outbox-события с тем же `project_id` продолжают уходить на внешний URL до 30д / restore. |
| **Файлы** | `project/service.go:207-228` · `apigateway/repository.go:168-176` · `apigateway/webhook.go:210-234,423-428` |
| **Stage A / B** | Residual combo (не W18 FAIL, усилитель soft-delete Stage B) |
| **OWASP** | A01 · A04 |
| **Mutation idea** | Soft-delete project → 0 dispatch jobs; JOIN/`DeletedAtIsNil` on project edge или deactivate endpoints on DeleteProject. |

---

### C-009 · HIGH — Soft-delete SQL OK + Meili upsert orphans (044 adjacent)

| | |
|--|--|
| **Пара** | TASK-022 PASS SQL `deleted_at` **+** Meili `IndexAllIssues` upsert-only (TASK-044 surface) |
| **Почему хуже суммы** | `IndexAllIssues` читает только `DeletedAtIsNil` и **upsert** (`meilisearch_indexer.go:84-114`) — нет DELETE stale docs. Ctrl+K (SQL/Mongo) чист; Meili держит trash docs → SEARCH-01/CLI leak; маскирует регресс SQL. |
| **Файлы** | `backend/internal/search/service.go` · `meilisearch_indexer.go:84-114` |
| **Stage A / B** | TASK-022 PASS · Meili delete path residual · TASK-044 |
| **OWASP** | A01 (search scope leak via secondary index) |
| **Mutation idea** | Soft-delete issue → reindex → Meili doc 404; или DELETE documents on soft-delete hook. |

---

### C-010 · MEDIUM — Forced team attach (014) + dual-active sprint residual (041)

| | |
|--|--|
| **Пара** | C-003 inject developers **+** C-002 dual active possible |
| **Почему хуже суммы** | Несогласованная толпа members видит/голосует в poker/board при двух active — blast radius membership × sprint invariant. |
| **Файлы** | project AddTeamToProject · sprint StartSprintAtomic |
| **Stage A / B** | TASK-014 + TASK-041 |
| **OWASP** | A01 · A04 |
| **Mutation idea** | Integration: attach foreign team + dual start attempt → assert invariants (members + ≤1 active). |

---

### C-011 · MEDIUM — via_invite static assert (042) + audit/WS consumers

| | |
|--|--|
| **Пара** | TASK-042: `via_invite` test на handmade JSON, не NATS из `AcceptInvite` **+** prod payload `via_invite: true` |
| **Почему хуже суммы** | Production сейчас ставит `true` (`service_channel_invites.go:194`). Регресс `true→false` **не краснеет**. Consumers (audit trail / automations / guest UX) теряют признак invite-join → wrong accountability без failing CI. |
| **Файлы** | `backend/internal/messenger/service_channel_invites.go:190-196` · testframework accept-invite payload test |
| **Stage A / B** | TASK-042 **Открыта** |
| **OWASP** | A09 (accountability / integrity of audit signal) |
| **Mutation idea** | AcceptInvite → capture PublishJSON → assert `via_invite==true`; flip prod → RED. |

---

### C-012 · MEDIUM — Search total-outage empty + soft-delete UX (044 adjacent)

| | |
|--|--|
| **Пара** | shardOK==0 → empty + partial, не hard error **+** soft-delete «0 hits = чисто» |
| **Почему хуже суммы** | Outage неотличим от «корзина пуста» без FE `partial` banner (`search/service.go:178-184`). |
| **Файлы** | `backend/internal/search/service.go:164-184` · FE Ctrl+K |
| **Stage A / B** | TASK-044 residual / hard-error follow-up |
| **OWASP** | A04 (fail-open UX) |
| **Mutation idea** | All shards down → 503; FE banner on `partial`. |

---

### C-013 · MEDIUM — Legacy entity Confirm + soft-deleted message (045 residual)

| | |
|--|--|
| **Пара** | `message|avatar|channel|voice` = only non-empty ID **+** soft-deleted target still UUID-addressable |
| **Почему хуже суммы** | Issue path checks live project match. Legacy types bind file to any ID → attach to soft-deleted/foreign message. Stacks with C-007 object gap. |
| **Файлы** | `backend/internal/storage/service.go:240-245` |
| **Stage A / B** | TASK-045 residual |
| **OWASP** | A01 |
| **Mutation idea** | Confirm `entityType=message` + soft-deleted/non-member → 404/403. |

---

### C-014 · LOW — expectTypeOf version (052) + DTO `omitempty` zero

| | |
|--|--|
| **Пара** | TASK-052: `expectTypeOf` не RED без tsc **+** `UpdateBaseRequest.Version` `json:"version,omitempty"` |
| **Почему хуже суммы** | Backend fail-closed (`Version < 1` → conflict, `bases/service.go:92-94`). FE `updateBase` требует `version: number` (`frontend/src/api/bases.ts:72-76`). Combo = **confidence gap**: type drift (`version?:`) не ловится Vitest → silent client bugs / wrong 409 UX, не authz bypass. |
| **Файлы** | `frontend/src/api/bases.ts:72-76` · `backend/internal/bases/dto.go:84-87` · `bases/service.go:92-94` |
| **Stage A / B** | TASK-052 **Открыта** |
| **OWASP** | A04 (integrity of optimistic lock contract) |
| **Mutation idea** | tsc-gated test or `satisfies`; mutate API to optional version → CI RED. |

---

### C-015 · LOW — Docs metrics drift (054) + gate false confidence

| | |
|--|--|
| **Пара** | TASK-054: нет автотеста docs drift **+** agents/CI читают stale counts (515 vs wrong) |
| **Почему хуже суммы** | Не exploit. Вместе с mutation gaps (041/045/019) создаёт иллюзию «покрыто N тестами» при дырявых ClassID. |
| **Файлы** | `docs/TROUBLESHOOTING.md` · `docs/test_plan.md` · TASK-080 |
| **Stage A / B** | TASK-054 **Открыта** · TASK-080 |
| **OWASP** | A05 (process / assurance misconfiguration) |
| **Mutation idea** | Script/CI: grep measured counts vs `find` actual; drift → job fail. |

---

### C-016 · LOW — CallChat empty+mutate (near FE-P3; reference)

| | |
|--|--|
| **Пара** | list fail → `messages=[]` без isError **+** send enabled |
| **Почему** | Sensitive send into «empty» history; не W18 FAIL (TASK-073). Фиксируем как FE empty+mutate class рядом с W19 focus. |
| **Файлы** | `frontend/src/components/calls/CallChat.tsx:60` (default `[]`), `:99-127` (send enabled) |
| **Mutation idea** | Reject list → error UI; send disabled. |

---

## Рекомендуемый порядок фикса (handoff)

1. **be** — **C-001** (TASK-079), **C-002** (041+046: PG race test + unique active)  
2. **be** — **C-003** (TASK-014 team ACL); **C-004** mid-fail Ent proof  
3. **be** — **C-007/C-013** (TASK-045 Confirm object+entity), **C-005/C-006** (019/044 SSRF+AI)  
4. **be** — C-008/C-009/C-011/C-012  
5. **fe/docs** — C-014 (052), C-015 (054/080), C-016 (073)

---

## Вердикт

## Security Audit: W19-A3 combos (W18 FAILs + TASK-079)

- [CRITICAL] `audit/gdpr_mongo.go:27-30` + cascade — `(0,nil)` → HardDelete без erase messages → TASK-079 (OWASP A01/A04)
- [CRITICAL] `sprint/repository.go:189-238` + `ent/schema/sprint.go:57-64` — unproven race + non-unique active (OWASP A04)
- [HIGH×6] C-003, C-005, C-006, C-007, C-008, C-009 — team ACL / SSRF / AI nil / Confirm / webhook / Meili
- [MEDIUM×5] C-004, C-010, C-011, C-012, C-013
- [LOW×3] C-014, C-015, C-016
- Вердикт: `verified: false` — 2026-07-28 19:58 — открыты CRITICAL/HIGH без mitigation

- Комбинаторных дефектов: **16** (CRITICAL **2** · HIGH **6** · MEDIUM **5** · LOW **3**).  
- Все 9 W18 FAIL + TASK-079 покрыты матрицей.  
- `verified: false` для чистого Stage C — открыты CRITICAL/HIGH без mitigation.  
- Re-verify path:line: **2026-07-28 19:58** (UTC+3).

*Артефакт W19-A3 · TASK-008 · readonly product review · docs-only commit.*
